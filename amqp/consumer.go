package amqp

import (
	"log"

	"github.com/Sterks/rXmlReader/config"
	"github.com/streadway/amqp"
)

// ConsumerMQ Подписчик
type ConsumerMQ struct {
	conn    *amqp.Connection
	config  *config.Config
	channel *amqp.Channel
}

// NewConsumer Инициализация
func NewConsumer(config *config.Config) *ConsumerMQ {
	return &ConsumerMQ{
		conn:    nil,
		channel: nil,
		config:  config,
	}
}

// ConnectRabbitMQ ...
func (c *ConsumerMQ) ConnectRabbitMQ(config *config.Config) *amqp.Connection {
	conn, err := amqp.Dial(config.Rabbit.ConnectRabbit)
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}
	c.conn = conn
	// defer conn.Close()
	return conn
}

// ChannelMQ ...
func (c *ConsumerMQ) ChannelMQ() (*amqp.Channel, amqp.Queue) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("connection.channel - %s", err)
	}
	// defer ch.Close()

	q, err2 := ch.QueueDeclare(
		"Files",
		false,
		false,
		false,
		false,
		nil,
	)
	if err2 != nil {
		log.Fatalf("connection.QueueDeclare - %s", err2)
	}
	return ch, q
}

// ConsumerMQNow ...
func (c *ConsumerMQ) ConsumerMQNow(ch *amqp.Channel, q amqp.Queue) <-chan amqp.Delivery {
	msqs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("connection.consumer - %s", err)
	}
	return msqs
}
