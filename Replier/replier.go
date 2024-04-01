package replier

import (
	model "GO/Judge/Model"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"time"
)

func Reply() {
	defer RecoverFromPanic()
	rds := NewRedisClient()
	err := rds.Ping(context.Background()).Err()
	if err != nil {
		fmt.Println("you are donkey")
		return
	}
	conn, err := NewRabbitMQConnection()
	if err != nil {
		return
	}
	//AddQueue("results", conn)

	msgs, err, conn, ch := ReadQueue("submit", conn)
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			return
		}
	}(ch)
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
		msg := msg
		go func() {
			start := time.Now()
			outputs, err := SendToJudge(msg, minioClient, cli, rds)
			err = msg.Ack(true)
			if err != nil {
				return
			}
			since := time.Since(start)
			fmt.Println(outputs)
			fmt.Println(since)
		}()

	}
	select {}

}
func Run(cli *client.Client, submission model.SubmissionMessage) (map[string]string, *client.Client, container.CreateResponse, error) {
	defer RecoverFromPanic()
	ProblemSRC := fmt.Sprintf("Problems/problem%s/in", submission.ProblemID)
	SubmissionSRC := fmt.Sprintf("Submissions/%s/%v.py", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, submission.TimeStamp)
	memoryCommand := fmt.Sprintf("chmod +x memory.sh ; ./memory.sh %v.py", submission.TimeStamp)
	memorySrc := fmt.Sprintf("memory.sh")
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
	err = CopyDirToContainer(ctx, memorySrc, dest, cli, resp.ID)
	err = RunMemoryExec(ctx, cli, resp.ID, memoryCommand)
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
	defer RecoverFromPanic()
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

	outCH := make(chan []byte)
	go func() {
		execStartResp, err := cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
		if err != nil {
			log.Printf("%s: %s", "Failed to attach exec", err)
			errCh <- struct{}{}
			return
		}
		output := make([]byte, 4096)
		_, err = execStartResp.Reader.Read(output)
		if err != nil && err != io.EOF {
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
		case <-time.After(submission.TimeLimit + 200*time.Millisecond):
			return "Time Limit Exceeded", nil
		case output1 := <-outCH:
			return string(output1), nil
		}
	}

}

func RunTestCases(ctx context.Context, cli *client.Client, respID string, Outputs map[string]string, submission model.SubmissionMessage) map[string]string {
	defer RecoverFromPanic()
	for i := 0; i < submission.TestCaseNumber; i++ {
		newCTX := context.WithValue(ctx, "TestCase", i+1)
		output, err := RunExec(newCTX, cli, respID, fmt.Sprintf("python3 %s.py < input%d.txt > out%d.txt 2>out%d.txt ; echo done", submission.TimeStamp, i+1, i+1, i+1), submission)
		if err != nil {
			return nil
		}
		Outputs[fmt.Sprintf("TestCase%d", i+1)] = output
	}
	return Outputs
}

func CheckTestCases(cli *client.Client, containerID string, output map[string]string, submission model.SubmissionMessage) map[string]string {
	defer RecoverFromPanic()
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

		if CheckRunTimeError(fmt.Sprintf("Submissions/%s/out%d.txt", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, i+1)) {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Runtime Error"
			continue
		}
		if CheckMemoryLimitError(fmt.Sprintf("Submissions/%s/out%d.txt", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, i+1)) {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Memory Limit Exceeded"
			continue
		}
		outputs[fmt.Sprintf("TestCase%d", i+1)] = CompareOutputs(fmt.Sprintf("Problems/problem%s/out/output%d.txt", submission.ProblemID, i+1), fmt.Sprintf("Submissions/%s/out%d.txt", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, i+1))
	}
	return outputs
}

func RunMemoryExec(ctx context.Context, cli *client.Client, containerID, command string) error {
	defer RecoverFromPanic()
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"sh", "-c", command},
		WorkingDir:   "/home",
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		fmt.Println("Error creating exec instance:", err)
		return err
	}

	resp, err := cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		fmt.Println("Error attaching to exec instance:", err)
		return err
	}
	defer resp.Close()

	return err
}
