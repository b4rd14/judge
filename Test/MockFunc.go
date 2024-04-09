package Test

import (
	replier "GO/Judge/Replier"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
)

func MockSendToJudge(msg amqp.Delivery, minioClient *minio.Client, cli *client.Client) (map[string]string, error) {
	var submission replier.SubmissionMessage
	fmt.Println(string(msg.Body))
	err := json.Unmarshal(msg.Body, &submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to unmarshal message\n", err)
		return nil, err
	}
	err = replier.Download(context.Background(), minioClient, "problems", "problem"+submission.ProblemID, "Problems")
	if err != nil {
		log.Printf("%s: %s", "Failed to download problem\n", err)
		return nil, err
	}
	err = replier.Download(context.Background(), minioClient, "submissions", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, "Submissions")
	if err != nil {
		log.Printf("%s: %s", "Failed to download submission\n", err)
		return nil, err
	}
	switch submission.Type {
	case "python":
		outChan := make(chan map[string]string)
		go func() {
			outChan <- MockPythonJudge(cli, submission)
		}()
		select {
		case outputs := <-outChan:
			return outputs, nil
		}
	case "csv":
	}
	return nil, nil
}

func MockPythonJudge(cli *client.Client, submission replier.SubmissionMessage) map[string]string {
	defer replier.RecoverFromPanic()
	outputs, resp, err := submission.Run(cli)
	if err != nil {
		log.Printf("%s: %s", "Failed to marshal output\n", err)
	}
	result := outputs.CheckTestCases(cli, resp.ID, submission)
	replier.RemoveDir("Submissions/" + submission.ProblemID + "/")
	return result
}
