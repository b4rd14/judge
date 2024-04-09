package replier

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"log"
)

func RecoverFromPanic() {
	if r := recover(); r != nil {
		fmt.Println("Recovered:", r)
	}
}

func PythonJudge(cli *client.Client, submission SubmissionMessage) map[string]string {
	defer RecoverFromPanic()
	outputs, resp, err := submission.Run(cli)
	if err != nil {
		log.Printf("%s: %s", "Failed to marshal output\n", err)
	}
	result := outputs.CheckTestCases(cli, resp.ID, submission)
	_, err = SendResult(result, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to send result\n", err)
	}
	RemoveDir("Submissions/" + submission.ProblemID + "/")
	return result
}

func SendResult(res map[string]string, submission SubmissionMessage) (map[string]interface{}, error) {
	defer RecoverFromPanic()
	result := make(map[string]interface{})
	result["submission_id"] = submission.SubmissionID
	result["problem_id"] = submission.ProblemID
	result["user_id"] = submission.UserID
	result["results"] = &res

	env := NewEnv()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resultJson, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	err = ch.PublishWithContext(ctx, "", "results", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        resultJson,
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func SendToJudge(msg amqp.Delivery, minioClient *minio.Client, cli *client.Client, rds *redis.Client) (map[string]string, error) {
	var submission SubmissionMessage
	fmt.Println(string(msg.Body))
	ctx := context.Background()
	err := json.Unmarshal(msg.Body, &submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to unmarshal message\n", err)
		return nil, err
	}
	if _, err := getProblem(ctx, rds, submission.ProblemID); err != nil {
		err = Download(context.Background(), minioClient, "problems", "problem"+submission.ProblemID, "Problems")
		if err != nil {
			log.Printf("%s: %s", "Failed to download problem\n", err)
			return nil, err
		}
		err := setProblem(ctx, rds, submission.ProblemID)
		if err != nil {
			return nil, err
		}
	}
	err = Download(context.Background(), minioClient, "submissions", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, "Submissions")
	if err != nil {
		log.Printf("%s: %s", "Failed to download submission\n", err)
		return nil, err
	}
	switch submission.Type {
	case "python":
		outChan := make(chan map[string]string)
		go func() {
			outChan <- PythonJudge(cli, submission)
		}()
		select {
		case outputs := <-outChan:
			return outputs, nil
		}
	case "csv":
	}
	return nil, nil
}
