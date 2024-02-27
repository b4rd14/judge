package replier

import (
	model "GO/Judge/Model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

func PythonJudge(msg amqp.Delivery, cli *client.Client, submission model.SubmissionMessage) {
	outputs, cli, resp, err := Run(cli, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to marshal output\n", err)
		msg.Ack(true)
		return
	}
	outputs = CheckTestCases(cli, resp.ID, outputs, submission)
	err = SendResult(outputs, submission)
	if err != nil {
		log.Printf("%s: %s", "Failed to send result\n", err)
	}
	RemoveDir("Submissions/" + submission.ProblemID + "/")
	msg.Ack(true)
}

func SendResult(res map[string]string, submission model.SubmissionMessage) error {
	result := make(map[string]interface{})
	result["submission_id"] = submission.SubmissionID
	result["problem_id"] = submission.ProblemID
	result["user_id"] = submission.UserID
	result["results"] = res

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

func RemoveDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("%s: %s", "Failed to remove directory", err)
	}
}
