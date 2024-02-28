package replier

import (
	model "GO/Judge/Model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"log"
	"strconv"
	"time"
)

func Reply() {
	defer recoverFromPanic()
	msgs, err := DeployRabbitMq()
	if err != nil {
		return
	}
	cli, err := NewDockerClint()
	if err != nil {
		log.Printf("%s: %s", "Failed to create docker client", err)
		return
	}
	minioClient, err := NewMinIoClient()
	if err != nil {
		log.Printf("%s: %s", "Failed to create minio client", err)
		return
	}

	for msg := range msgs {
		err := SendToJudge(msg, minioClient, cli)
		if err != nil {
			continue
		}
	}
	select {}

}
func Run(cli *client.Client, submission model.SubmissionMessage) (map[string]string, *client.Client, container.CreateResponse, error) {
	defer recoverFromPanic()
	ProblemSRC := fmt.Sprintf("Problems/problem%s/in", submission.ProblemID)
	SubmissionSRC := fmt.Sprintf("Submissions/%s/%v.py", submission.ProblemID+"/"+submission.UserID+"/"+strconv.FormatInt(submission.TimeStamp, 10), submission.TimeStamp)
	dest := "/home"
	Outputs := make(map[string]string)

	ctx := context.Background()
	config := &container.Config{
		Image: "python",
		Cmd:   []string{"sh", "-c", "while true; do sleep 1; done"},
	}
	resp, err := cli.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		return nil, nil, container.CreateResponse{}, err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		KillContainer(cli, ctx, resp.ID)
		return nil, nil, container.CreateResponse{}, err
	}

	err = CopyDirToContainer(ctx, ProblemSRC, dest, cli, resp.ID)
	if err != nil {
		KillContainer(cli, ctx, resp.ID)
		log.Printf("%s: %s", "Failed to copy problem to container", err)
		return nil, nil, container.CreateResponse{}, err
	}
	err = CopyDirToContainer(ctx, SubmissionSRC, dest, cli, resp.ID)

	if err != nil {
		KillContainer(cli, ctx, resp.ID)
		log.Printf("%s: %s", "Failed to copy submission to container", err)
		return nil, nil, container.CreateResponse{}, err
	}

	Outputs = RunTestCases(ctx, cli, resp.ID, Outputs, submission)

	KillContainer(cli, ctx, resp.ID)

	return Outputs, cli, resp, nil
}

func RunExec(ctx context.Context, cli *client.Client, containerID, command string, submission model.SubmissionMessage) (string, error) {
	defer recoverFromPanic()
	memCh := make(chan struct{})
	errCh := make(chan struct{})

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"sh", "-c", command},
		WorkingDir:   "/home",
	}

	execResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		KillContainer(cli, ctx, containerID)
		log.Printf("%s: %s", "Failed to create exec", err)
		return "", err
	}

	go func() {
		for {
			stats, err := cli.ContainerStats(ctx, containerID, false)
			if err != nil {
				fmt.Printf("%s: %s", "Failed to get container stats", err)
				errCh <- struct{}{}
				return
			}
			var memStats types.MemoryStats
			err = json.NewDecoder(stats.Body).Decode(&memStats)
			if err != nil {
				log.Printf("%s: %s", "Failed to decode memory stats", err)
				errCh <- struct{}{}
				return
			}
			if memStats.Usage > uint64(submission.MemoryLimit) {
				memCh <- struct{}{}
				return
			}
		}
	}()
	outCH := make(chan []byte)
	go func() {
		execStartResp, err := cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
		if err != nil {
			KillContainer(cli, ctx, containerID)
			log.Printf("%s: %s", "Failed to attach exec", err)
			errCh <- struct{}{}
			return
		}
		output := make([]byte, 4096)
		_, err = execStartResp.Reader.Read(output)
		if err != nil && err != io.EOF {
			KillContainer(cli, ctx, containerID)
			log.Printf("%s: %s", "Failed to read from exec", err)
			errCh <- struct{}{}
			return
		}
		outCH <- output
		execStartResp.Close()
	}()
	for {
		select {
		case <-errCh:
			return "", err
		case <-memCh:
			return "Memory Limit Exceeded", nil
		case <-time.After(submission.TimeLimit + 200*time.Millisecond):
			return "Time Limit Exceeded", nil
		case output1 := <-outCH:
			return string(output1), nil
		}
	}

}

func RunTestCases(ctx context.Context, cli *client.Client, respID string, Outputs map[string]string, submission model.SubmissionMessage) map[string]string {
	defer recoverFromPanic()
	for i := 0; i < submission.TestCaseNumber; i++ {
		newCTX := context.WithValue(ctx, "TestCase", i+1)
		output, err := RunExec(newCTX, cli, respID, fmt.Sprintf("python3 %s.py < input%d.txt > out%d.txt 2>out%d.txt ; echo done", strconv.FormatInt(submission.TimeStamp, 10), i+1, i+1, i+1), submission)
		if err != nil {
			return nil
		}
		Outputs[fmt.Sprintf("TestCase%d", i+1)] = output
	}
	return Outputs
}

func CheckTestCases(cli *client.Client, containerID string, output map[string]string, submission model.SubmissionMessage) map[string]string {
	defer recoverFromPanic()
	outputs := make(map[string]string)
	ctx := context.Background()

	for i := 0; i < submission.TestCaseNumber; i++ {
		if output[fmt.Sprintf("TestCase%d", i+1)] == "Time Limit Exceeded" {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Time Limit Exceeded"
			continue
		}
		src := fmt.Sprintf("/home/out%d.txt", i+1)
		fromContainer, _, _ := cli.CopyFromContainer(ctx, containerID, src)

		TarToTxt(fromContainer, submission)

		if CheckRunTime(fmt.Sprintf("Submissions/%s/out%d.txt", submission.ProblemID+"/"+submission.UserID+"/"+strconv.FormatInt(submission.TimeStamp, 10), i+1)) {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Runtime Error"
			continue
		}
		outputs[fmt.Sprintf("TestCase%d", i+1)] = CompareOutputs(fmt.Sprintf("Problems/problem%s/out/output%d.txt", submission.ProblemID, i+1), fmt.Sprintf("Submissions/%s/out%d.txt", submission.ProblemID+"/"+submission.UserID+"/"+strconv.FormatInt(submission.TimeStamp, 10), i+1))
	}
	return outputs
}
