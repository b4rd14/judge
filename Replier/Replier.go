package replier

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"time"
)

type SubmissionResult struct {
	SubmissionID string
	ProblemID    string
	Output       string
}
type SubmissionMessage struct {
	SubmissionID   string
	ProblemID      string
	TestCaseNumber int
	TimeLimit      time.Duration
	MemoryLimit    int64
}

func NewClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.41"))
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func RunTestCases(ctx context.Context, cli *client.Client, respID string, Outputs map[string]string, submission SubmissionMessage) map[string]string {
	for i := 0; i < submission.TestCaseNumber; i++ {
		newCTX := context.WithValue(ctx, "TestCase", i+1)
		output, err := RunExec(newCTX, cli, respID, fmt.Sprintf("python3 %s.py < input%d.txt > out%d.txt 2>out%d.txt ; echo done", submission.SubmissionID, i+1, i+1, i+1), submission)
		if err != nil {
			return nil
		}
		Outputs[fmt.Sprintf("TestCase%d", i+1)] = output
	}
	return Outputs
}

func Reply() {

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

	cli, err := NewClient()

	for msg := range msgs {
		var submission SubmissionMessage
		err := json.Unmarshal(msg.Body, &submission)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to unmarshal message", err)
		}
		msg := msg
		go func() {
			outputs, cli, resp, err := Run(cli, submission)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to marshal output", err)
			}
			outputs = CheckTestCases(cli, resp.ID, outputs, submission)
			fmt.Println(outputs)
			msg.Ack(true)
		}()

	}

	select {}

}
func Run(cli *client.Client, submission SubmissionMessage) (map[string]string, *client.Client, container.CreateResponse, error) {
	ProblemSRC := fmt.Sprintf("Problems/Problem%s/in", submission.ProblemID)
	SubmissionSRC := fmt.Sprintf("Submissions/%s/%s.py", submission.SubmissionID, submission.SubmissionID)
	dest := "/home"
	Outputs := make(map[string]string)

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

	err = CopyDirToContainer(ctx, ProblemSRC, dest, cli, resp.ID)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to copy problem to container", err)

	}
	err = CopyDirToContainer(ctx, SubmissionSRC, dest, cli, resp.ID)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to copy submission to container", err)
	}

	Outputs = RunTestCases(ctx, cli, resp.ID, Outputs, submission)

	err = cli.ContainerKill(ctx, resp.ID, "SIGKILL")
	if err != nil {
		log.Fatalf("%s: %s", "Failed to kill container", err)
	}

	return Outputs, cli, resp, nil
}

func RunExec(ctx context.Context, cli *client.Client, containerID, command string, submission SubmissionMessage) (string, error) {
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

	for {
		stats, err := cli.ContainerStats(ctx, containerID, false)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to get container stats", err)
		}
		var memStats types.MemoryStats
		err = json.NewDecoder(stats.Body).Decode(&memStats)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to decode memory stats", err)
		}
		if memStats.Usage > uint64(submission.MemoryLimit) {
			return "Memory Limit Exceeded", nil
		}
		select {
		case <-cancelCtx.Done():
			return "Time Limit Exceeded", nil
		case output1 := <-outCH:
			return string(output1), nil
		}
	}

}

func CheckTestCases(cli *client.Client, containerID string, output map[string]string, submission SubmissionMessage) map[string]string {
	outputs := make(map[string]string)
	ctx := context.Background()

	for i := 0; i < submission.TestCaseNumber; i++ {
		if output[fmt.Sprintf("TestCase%d", i+1)] == "Time Limit Exceeded" {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Time Limit Exceeded"
			continue
		}
		src := fmt.Sprintf("/home/out%d.txt", i+1)
		fromContainer, _, _ := cli.CopyFromContainer(ctx, containerID, src)

		TarToTxt(fromContainer, submission.SubmissionID)

		if CheckRunTime(fmt.Sprintf("Submissions/%s/out%d.txt", submission.SubmissionID, i+1)) {
			outputs[fmt.Sprintf("TestCase%d", i+1)] = "Runtime Error"
			continue
		}
		outputs[fmt.Sprintf("TestCase%d", i+1)] = CompareOutputs(fmt.Sprintf("Problems/Problem%s/out/output%d.txt", submission.ProblemID, i+1), fmt.Sprintf("Submissions/%s/out%d.txt", submission.SubmissionID, i+1))
	}
	return outputs
}
