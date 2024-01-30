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
	"time"
)

type SubmissionResult struct {
	SubmissionId string
	ProblemId    string
	Output       string
}
type SubmissionMessage struct {
	SubmissionId string
	ProblemId    string
	TimeLimit    time.Time
}

const TestCaseNumber = 3

func main() {

	//cleanCh := make(chan *os.File, 30)

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
			outputs, err := Run(submission)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to marshal output", err)
			}
			fmt.Println(outputs)
			//outputs = checkTestCases(outputs, submission, cleanCh)
			//fmt.Println(outputs)
			//cleanUp(cleanCh)
			msg.Ack(true)

		}()

	}

	select {}

}

func Run(submission SubmissionMessage) (map[string]string, error) {

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
	SubmissionSRC := fmt.Sprintf("Submissions/%s.py", submission.SubmissionId)
	err = copyDirToContainer(ctx, ProblemSRC, dest, cli, resp.ID)
	if err != nil {
		log.Fatal(err)

	}
	err = copyDirToContainer(ctx, SubmissionSRC, dest, cli, resp.ID)

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < TestCaseNumber; i++ {
		output, err := createExec(ctx, cli, resp.ID, fmt.Sprintf("python3 %s.py < in%d.txt", submission.SubmissionId, i+1))
		if err != nil {
			return nil, err
		}
		Outputs[fmt.Sprintf("TestCase%d", i+1)] = output
	}

	err = cli.ContainerKill(ctx, resp.ID, "SIGKILL")
	if err != nil {
		log.Fatal(err)
	}

	return Outputs, nil
}

func createExec(ctx context.Context, cli *client.Client, containerID, command string) (string, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"sh", "-c", command},
		WorkingDir:   "/home",
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		log.Fatal(err)
	}

	execStartResp, err := cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		log.Fatal(err)
	}
	defer execStartResp.Close()

	output := make([]byte, 4096)
	_, err = execStartResp.Reader.Read(output)
	if err != nil {
		log.Fatal(err)
	}
	return string(output), nil
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

	// Close the tar writer
	err = tw.Close()
	if err != nil {
		return err
	}

	// Open the created archive file for reading
	archiveFile, err = os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func(archiveFile *os.File) {
		err := archiveFile.Close()
		if err != nil {

		}
	}(archiveFile)

	// Copy the archive to the container
	err = cli.CopyToContainer(ctx, id, destDir, archiveFile, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	return nil
}

//func checkTestCases(output map[string]string, submission SubmissionMessage, cleanCh chan *os.File) map[string]string {
//	src := fmt.Sprintf("Problems/Problem%s/out", submission.ProblemId)
//	outputs := make(map[string]string)
//
//	fmt.Println(output)
//
//	for i := 0; i < TestCaseNumber; i++ {
//		file, err := os.Open(fmt.Sprintf("%s/%d.txt", src, i+1))
//		if err != nil {
//			log.Fatal(err)
//		}
//		cleanCh <- file
//		out := make([]byte, 4096)
//		_, err = file.Read(out)
//		if err != nil {
//			return nil
//		}
//		if strings.EqualFold(string(out), output[fmt.Sprintf("TestCase%d", i+1)]) {
//			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Wrong Answer"
//		} else {
//			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Accepted"
//		}
//	}
//	cleanCh <- nil
//	return outputs
//}
//
//func cleanUp(cleanCh <-chan *os.File) {
//	for file := range cleanCh {
//		if file == nil {
//			break
//		}
//		err := file.Close()
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
//}
