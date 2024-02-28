package replier

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func DeployRabbitMq() (<-chan amqp.Delivery, error) {
	defer recoverFromPanic()
	conn, err := NewRabbitMQConnection()
	if err != nil {
		log.Printf("%s: %s", "Failed to connect to RabbitMQ", err)
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("%s: %s", "Failed to open a channel", err)
		return nil, err
	}
	if err != nil {
		log.Printf("%s: %s", "Failed to declare a queue", err)
		return nil, err
	}
	msgs, err := ch.Consume("submit", "", false, false, false, false, nil)
	if err != nil {
		log.Printf("%s: %s", "Failed to register a consumer", err)
		return nil, err
	}
	return msgs, nil
}

func NewRabbitMQConnection() (*amqp.Connection, error) {
	env := NewEnv()
	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))
}
