package replier

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func ReadQueue(queueName string, conn *amqp.Connection) (<-chan amqp.Delivery, error, *amqp.Connection, *amqp.Channel) {
	defer RecoverFromPanic()
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("%s: %s", "Failed to open a channel", err)
		return nil, err, nil, nil
	}
	if err != nil {
		log.Printf("%s: %s", "Failed to declare a queue", err)
		return nil, err, nil, nil
	}
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("%s: %s", "Failed to register a consumer", err)
		return nil, err, nil, nil
	}
	return msgs, nil, conn, ch
}

func AddQueue(queueName string, conn *amqp.Connection) {
	defer RecoverFromPanic()
	ch, err := conn.Channel()
	_, err = ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return
	}
	err = ch.Close()
	if err != nil {
		return
	}
}

func NewRabbitMQConnection() (*amqp.Connection, error) {
	env := NewEnv()
	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))
}
