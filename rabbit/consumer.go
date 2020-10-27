package rabbit

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

func (c *ConsumerMQ) ConsumerMQNow(config *config.Config, nameQueue string) <-chan amqp.Delivery {
	conn, err := amqp.Dial(config.Rabbit.ConnectRabbit)
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("connection.channel - %s", err)
	}
	//defer ch.Close()

	q, err2 := ch.QueueDeclare(
		nameQueue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err2 != nil {
		log.Fatalf("connection.QueueDeclare - %s", err2)
	}

	err = ch.Qos(1, 0, false)
	failOnError(err, "Qos - не работает")

	msqs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("connection.consumer - %s", err)
	}

	stopChan := make(chan bool)

	go func() {

	}()

	<-stopChan
	return msqs
}
