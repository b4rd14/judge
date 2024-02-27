package requester

import (
	replier "GO/Judge/Replier"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func Request() error {
	env := replier.NewEnv()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))

	if err != nil {
		return err
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
		return err
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("%s: %s", "Failed to close channel", err)
		}
	}(ch)

	_, err = ch.QueueDeclare(
		"result",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	return nil
}
