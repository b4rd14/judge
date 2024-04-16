package replier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func RecoverFromPanic() {
	if r := recover(); r != nil {
		log.Printf("Recovered from panic: %v", r)
	}
}

func retry(ctx context.Context, fn func() error, maxAttempts int, interval time.Duration) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fn(); err == nil {
				return nil
			} else {
				log.Printf("Attempt %d failed: %s\n", attempt, err)
				time.Sleep(interval)
			}
		}
	}
	return fmt.Errorf("exceeded maximum number of attempts")
}

func Judge(cli *client.Client, rds *redis.Client, submission SubmissionMessage) map[string]string {
	defer RecoverFromPanic()
	outputs, resp, err := submission.Run(cli)
	if err != nil {
		log.Printf("%s: %s", "Failed to marshal output\n", err)
	}
	result := outputs.CheckTestCases(cli, resp.ID, submission)
	_, err = SendResult(rds, result, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to send result\n", err)
	}
	RemoveDir("Submissions/" + submission.ProblemID + "/")
	return result
}

func SendResult(rds *redis.Client, res map[string]string, submission SubmissionMessage) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["submission_id"] = submission.SubmissionID
	result["problem_id"] = submission.ProblemID
	result["user_id"] = submission.UserID
	result["results"] = &res

	conn, err := NewRabbitMQConnection()
	ch, err := conn.NewChannel()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resultJson, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	err = publishMessage(ch, ctx, resultJson, "results")
	if err != nil {
		if errors.Is(err, errors.New("exceeded maximum number of attempts")) {
			err := setResult(ctx, rds, resultJson, submission)
			if err != nil {
				return nil, err
			}
			panic("Failed to send result to RabbitMQ, saved to Redis instead")
		}
	}
	return result, nil
}

func Result(ctx context.Context, msg amqp.Delivery, minioClient *minio.Client, cli *client.Client, rds *redis.Client) (map[string]string, error) {
	var submission SubmissionMessage
	fmt.Println(string(msg.Body))
	err := json.Unmarshal(msg.Body, &submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to unmarshal message\n", err)
		return nil, err
	}
	if _, err := getProblem(context.Background(), rds, submission.ProblemID); err != nil {
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
	err = retry(ctx, func() error {
		return Download(ctx, minioClient, "submissions", submission.ProblemID+"/"+submission.UserID+"/"+submission.TimeStamp, "Submissions")
	}, 3, time.Second*5)
	if err != nil {
		panic("Failed to download submission")
	}

	outChan := make(chan map[string]string)
	go func() {
		outChan <- Judge(cli, rds, submission)
	}()
	select {
	case <-ctx.Done():
		return map[string]string{"msg": "Result took too long"}, ctx.Err()
	case outputs := <-outChan:
		return outputs, nil
	}

}
