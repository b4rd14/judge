package requester

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type SubmissionMessage struct {
	SubmissionID   string
	ProblemID      string
	TestCaseNumber int
	TimeLimit      time.Duration
}

func Request() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	fmt.Println("Connected to RabbitMQ")

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

	q, err := ch.QueueDeclare(
		"submit",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	submission := SubmissionMessage{
		SubmissionID:   "1",
		ProblemID:      "1",
		TimeLimit:      2 * time.Second,
		TestCaseNumber: 10,
	}
	submissionBytes, err := json.Marshal(submission)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to marshal submission", err)
	}

	err = ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        submissionBytes,
	})

	if err != nil {
		log.Fatalf("%s: %s", "Failed to publish a message", err)
	}

	fmt.Println("Submission sent")

}
