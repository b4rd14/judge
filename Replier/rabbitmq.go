package replier

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	*amqp.Connection
}
type RabbitChannel struct {
	*amqp.Channel
}

func (ch *RabbitChannel) ReadQueue(queueName string) (<-chan amqp.Delivery, error) {
	defer RecoverFromPanic()
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("%s: %s", "Failed to register a consumer", err)
		return nil, err
	}
	return msgs, nil
}

func (ch *RabbitChannel) AddQueue(queueName string) {
	defer RecoverFromPanic()
	_, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return
	}
}

func NewRabbitMQConnection() (*RabbitMQConnection, error) {
	env := NewEnv()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", env.RabbitmqUsername, env.RabbitmqPassword, env.RabbitmqUrl))
	return &RabbitMQConnection{
		conn,
	}, err
}

func (conn *RabbitMQConnection) NewChannel() (*RabbitChannel, error) {
	ch, err := conn.Channel()
	return &RabbitChannel{
		ch,
	}, err
}
