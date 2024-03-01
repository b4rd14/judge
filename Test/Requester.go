package Test

import (
	replier "GO/Judge/Replier"
	"context"
	"encoding/json"
	"fmt"
	echo "github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func Request(submissionMsg map[string]interface{}) error {
	env := replier.NewEnv()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))

	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("%s: %s", "Failed to close connection to RabbitMQ", err)
		}
	}(conn)

	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("%s: %s", "Failed to close channel", err)
		}
	}(ch)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	submissionBytes, err := json.Marshal(submissionMsg)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx, "", "submit", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        submissionBytes,
	})

	if err != nil {
		return err
	}

	return nil
}

func Submit(c echo.Context) error {
	submissionMsg := make(map[string]interface{})
	err := c.Bind(&submissionMsg)
	if err != nil {
		return err
	}
	err = Request(submissionMsg)
	fmt.Println(submissionMsg)
	if err != nil {
		return err
	}
	return c.JSON(200, submissionMsg)
}

func StartServer() {
	e := echo.New()
	e.POST("/submit", Submit)
	err := e.Start(":8080")
	if err != nil {
		return
	}
}
