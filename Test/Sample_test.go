package Test

import (
	"GO/Judge/Replier"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type SubmissionMessage struct {
	SubmissionID   string `json:"submission_id"`
	ProblemID      string `json:"q_id"`
	UserID         string `json:"user_id"`
	TestCaseNumber int    `json:"TestCases"`
	Prompt         string `json:"prompt"`
}

func TestEnv(t *testing.T) {
	env := replier.NewEnv()
	assert.NotNil(t, env)
}

func TestRequestSubmission(t *testing.T) {
	go replier.Reply()
	submission := SubmissionMessage{
		SubmissionID:   "123",
		ProblemID:      "14",
		UserID:         "123",
		TestCaseNumber: 10,
		Prompt:         "i need a Python function that includes the following logic: The base form of the function for 0 input is 5.If the input 'i' is even, the function should be func(i−1)−21. If 'i' is odd, the function should be func(i−1)**2. Additionally, the function should print the final output at the end. and also get the input from the user and don't write any example and get the input without any text and print the output without any text and 10000 limit recursion",
	}
	Request(submission)

}

func Test5Submission(t *testing.T) {
	env := replier.NewEnv()
	conn, err := amqp.DialConfig(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl), amqp.Config{Heartbeat: 10 * time.Second})
	go replier.Reply()
	if err != nil {
		return
	}

	submission := SubmissionMessage{
		SubmissionID:   "123",
		ProblemID:      "12",
		UserID:         "123",
		TestCaseNumber: 10,
		Prompt:         "i need a Python function that includes the following logic: The base form of the function for 0 input is 5.If the input 'i' is even, the function should be func(i−1)−21. If 'i' is odd, the function should be func(i−1)**2. Additionally, the function should print the final output at the end. and also get the input from the user and don't write any example and get the input without any text and print the output without any text and 10000 limit recursion",
	}

	submissions := make([]SubmissionMessage, 5)

	for i := 0; i < 5; i++ {
		submissions[i] = submission
		submissions[i].UserID = fmt.Sprintf("%d", i)
		go Request(submissions[i])
	}

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	msgs, err := ch.Consume("result", "", false, false, false, false, nil)
	start := time.Now()
	check := make(map[string]bool)
	fmt.Println("Waiting for results")
	for msg := range msgs {
		msg := msg
		go func() {
			fmt.Println(string(msg.Body))
			value := make(map[string]interface{})
			err := json.Unmarshal(msg.Body, &value)
			if err != nil {
				return
			}
			fmt.Println(value["results"])
			check[value["user_id"].(string)] = true
			msg.Ack(true)
		}()
		if len(check) == 5 {
			since := time.Since(start)
			fmt.Println(since)
			break
		}
	}
}

func Request(submission SubmissionMessage) {
	client := resty.New()
	_, err := client.R().
		SetBody(submission).
		Post("https://fastapi-kodalab.chbk.run/api/v1/prompt/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
