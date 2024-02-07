package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SubmissionResult struct {
	SubmissionId string
	ProblemId    string
	Output       string
}
type SubmissionMessage struct {
	SubmissionId   string
	ProblemId      string
	TestCaseNumber int
	TimeLimit      time.Duration
}

const TestCaseNumber = 3

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("%s: %s", "Failed to close connection to RabbitMQ", err)
		}
	}(conn)

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("%s: %s", "Failed to close channel", err)
		}
	}(ch)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	msgs, err := ch.Consume("submit", "", false, false, false, false, nil)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to register a consumer", err)
	}

	for msg := range msgs {
		var submission SubmissionMessage
		err := json.Unmarshal(msg.Body, &submission)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to unmarshal message", err)
		}
		msg := msg
		go func() {
			outputs, cli, resp, err := Run(submission)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to marshal output", err)
			}
			outputs = checkTestCases(cli, resp.ID, outputs, submission)
			fmt.Println(outputs)
			msg.Ack(true)
		}()

	}

	select {}

}

func Run(submission SubmissionMessage) (map[string]string, *client.Client, container.CreateResponse, error) {

	Outputs := make(map[string]string)

	cli, err := client.NewClientWithOpts(client.WithVersion("1.41"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	config := &container.Config{
		Image: "python",
		Cmd:   []string{"sh", "-c", "while true; do sleep 1; done"},
	}
	resp, err := cli.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {

		log.Fatal(err)
	}
	dest := "/home"
	ProblemSRC := fmt.Sprintf("Problems/Problem%s/in", submission.ProblemId)
	SubmissionSRC := fmt.Sprintf("Submissions/%s/%s.py", submission.SubmissionId, submission.SubmissionId)
	err = copyDirToContainer(ctx, ProblemSRC, dest, cli, resp.ID)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to copy problem to container", err)

	}
	err = copyDirToContainer(ctx, SubmissionSRC, dest, cli, resp.ID)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to copy submission to container", err)
	}

	for i := 0; i < TestCaseNumber; i++ {
		newCTX := context.WithValue(ctx, "TestCase", i+1)
		output, err := createExec(newCTX, cli, resp.ID, fmt.Sprintf("python3 %s.py < in%d.txt > out%d.txt 2>out%d.txt ; echo done", submission.SubmissionId, i+1, i+1, i+1), submission)
		if err != nil {
			return nil, cli, resp, err
		}
		Outputs[fmt.Sprintf("TestCase%d", i+1)] = output
	}

	err = cli.ContainerKill(ctx, resp.ID, "SIGKILL")
	if err != nil {
		log.Fatalf("%s: %s", "Failed to kill container", err)
	}

	return Outputs, cli, resp, nil
}

func createExec(ctx context.Context, cli *client.Client, containerID, command string, submission SubmissionMessage) (string, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"sh", "-c", command},
		WorkingDir:   "/home",
	}

	execResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to create exec", err)
	}
	cancelCtx, cancel := context.WithTimeout(ctx, submission.TimeLimit)
	defer cancel()

	outCH := make(chan []byte)
	go func() {
		execStartResp, err := cli.ContainerExecAttach(cancelCtx, execResp.ID, types.ExecStartCheck{})
		if err != nil {
			log.Fatalf("%s: %s", "Failed to attach exec", err)
		}
		output := make([]byte, 4096)
		_, err = execStartResp.Reader.Read(output)
		if err != nil && err != io.EOF {
			log.Fatalf("%s: %s", "Failed to read from exec", err)
		}
		outCH <- output
		execStartResp.Close()
	}()

	select {
	case <-cancelCtx.Done():
		return "Time Limit Exceeded", nil
	case output1 := <-outCH:
		return string(output1), nil
	}

}

func checkTestCases(cli *client.Client, containerID string, output map[string]string, submission SubmissionMessage) map[string]string {
	outputs := make(map[string]string)
	ctx := context.Background()

	for i := 0; i < TestCaseNumber; i++ {
		if output[fmt.Sprintf("TestCase%d", i+1)] == "Time Limit Exceeded" {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Time Limit Exceeded"
			continue
		}
		src := fmt.Sprintf("/home/out%d.txt", i+1)
		fromContainer, _, _ := cli.CopyFromContainer(ctx, containerID, src)

		TarToTxt(fromContainer, submission.SubmissionId)

		if checkRunTime(fmt.Sprintf("Submissions/%s/out%d.txt", submission.SubmissionId, i+1)) {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Runtime Error"
			continue
		}

		outputs[fmt.Sprintf("TestCase%d", i+1)] = CompareOutputs(fmt.Sprintf("Problems/Problem%s/out/%d.txt", submission.ProblemId, i+1), fmt.Sprintf("Submissions/%s/out%d.txt", submission.SubmissionId, i+1))

	}
	return outputs
}

func copyDirToContainer(ctx context.Context, srcDir, destDir string, cli *client.Client, id string) error {

	archivePath := filepath.Join(os.TempDir(), "archive.tar")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(archivePath)
	defer func(archiveFile *os.File) {
		err := archiveFile.Close()
		if err != nil {

		}
	}(archiveFile)

	tw := tar.NewWriter(archiveFile)

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {

				}
			}(file)

			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	err = tw.Close()
	if err != nil {
		return err
	}
	archiveFile, err = os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func(archiveFile *os.File) {
		err := archiveFile.Close()
		if err != nil {

		}
	}(archiveFile)

	err = cli.CopyToContainer(ctx, id, destDir, archiveFile, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	return nil
}

func TarToTxt(reader io.ReadCloser, ID string) {
	read := tar.NewReader(reader)
	for {
		header, err := read.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		if header.Typeflag == tar.TypeReg {
			file, err := os.Create("/home/b4rd14/go/judge/Submissions/" + ID + "/" + header.Name)
			if err != nil {
				fmt.Println(err)
			}
			_, err = io.Copy(file, read)
			if err != nil {
				fmt.Println(err)
			}
			err = file.Close()
			if err != nil {
				fmt.Println(err)
			}

		}
	}
}

func checkRunTime(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	out := make([]byte, 4096)
	_, err = file.Read(out)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to read from file", err)
	}
	if strings.Contains(string(out), "Traceback (most recent call last):") {
		return true
	}
	return false
}

func CompareOutputs(output1 string, output2 string) string {
	out1, err := os.Open(output1)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file", err)
	}
	defer func(out1 *os.File) {
		err := out1.Close()
		if err != nil {

		}
	}(out1)

	out2, err := os.Open(output2)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open file", err)
	}
	defer func(out2 *os.File) {
		err := out2.Close()
		if err != nil {

		}
	}(out2)

	out1Bytes, err := io.ReadAll(out1)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to read from file", err)
	}

	out2Bytes, err := io.ReadAll(out2)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to read from file", err)
	}
	if strings.TrimSpace(string(out1Bytes)) == strings.TrimSpace(string(out2Bytes)) {
		return "Accepted"
	} else {
		return "Wrong Answer"
	}
}
