package replier

import (
	model "GO/Judge/Model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
)

func recoverFromPanic() {
	if r := recover(); r != nil {
		fmt.Println("Recovered:", r)
	}
}

func PythonJudge(msg amqp.Delivery, cli *client.Client, submission model.SubmissionMessage) {
	defer recoverFromPanic()
	outputs, cli, resp, err := Run(cli, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to marshal output\n", err)
		err := msg.Ack(true)
		if err != nil {
			return
		}
		return
	}
	outputs = CheckTestCases(cli, resp.ID, outputs, submission)
	err = SendResult(outputs, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to send result\n", err)
	}
	RemoveDir("Submissions/" + submission.ProblemID + "/")
	err = msg.Ack(true)
	if err != nil {
		return
	}
}

func SendResult(res map[string]string, submission model.SubmissionMessage) error {
	defer recoverFromPanic()
	result := make(map[string]interface{})
	result["submission_id"] = submission.SubmissionID
	result["problem_id"] = submission.ProblemID
	result["user_id"] = submission.UserID
	result["results"] = res

	fmt.Println(result)

	env := NewEnv()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	ctx := context.Background()
	resultJson, err := json.Marshal(result)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(ctx, "", "result", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        resultJson,
	})

	if err != nil {
		return err
	}
	return nil
}

func SendToJudge(msg amqp.Delivery, minioClient *minio.Client, cli *client.Client) error {
	var submission model.SubmissionMessage
	fmt.Println(string(msg.Body))
	err := json.Unmarshal(msg.Body, &submission)
	if err != nil {
		err := msg.Ack(true)
		if err != nil {
			return err
		}
		log.Printf("%s: %s", "Failed to unmarshal message\n", err)
		return err
	}
	err = Download(context.Background(), minioClient, "problems", "problem"+submission.ProblemID, "Problems")
	if err != nil {
		err := msg.Ack(true)
		if err != nil {
			return err
		}
		log.Printf("%s: %s", "Failed to download problem\n", err)
		return err
	}
	err = Download(context.Background(), minioClient, "submissions", submission.ProblemID+"/"+submission.UserID+"/"+strconv.FormatInt(submission.TimeStamp, 10), "Submissions")
	if err != nil {
		err := msg.Ack(true)
		if err != nil {
			return err
		}
		log.Printf("%s: %s", "Failed to download submission\n", err)
		return err
	}
	switch submission.Type {
	case "python":
		go PythonJudge(msg, cli, submission)
	case "csv":
	}
	return nil
}
