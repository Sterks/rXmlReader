package rabbit

import (
	"encoding/json"
	"log"
	"os"
	"time"

	config2 "github.com/Sterks/rXmlReader/config"
	"github.com/streadway/amqp"
)

type ProducerMQ struct {
	config *config2.Config
	am     *amqp.Connection
	amq    *amqp.Channel
}

type InformationFile struct {
	FileID   int
	NameFile string
	SizeFile int64
	DateMode time.Time
	Fullpath string
	Region   string
	FileZip  []byte
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (pr *ProducerMQ) PublishSend(config *config2.Config, info os.FileInfo, nameQueue string, in []byte, id int, region string, fullpath string) {
	conn, err := amqp.Dial(config.Rabbit.ConnectRabbit)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		nameQueue, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := &InformationFile{
		FileID:   id,
		DateMode: info.ModTime(),
		NameFile: info.Name(),
		FileZip:  in,
		SizeFile: info.Size(),
		Fullpath: fullpath,
		Region:   region,
	}

	bodyJSON, err3 := json.Marshal(body)
	if err3 != nil {
		log.Println(err3)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bodyJSON,
		})
	log.Printf(" [x] Sent %s - название очереди %s", body.NameFile, nameQueue)
	failOnError(err, "Failed to publish a message")
}
