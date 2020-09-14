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

// PublishSend ...
func (pr *ProducerMQ) PublishSend(config *config2.Config, info os.FileInfo, nameQueue string, in []byte, id int, region string, fullpath string) {

	conn, err := amqp.Dial(config.Rabbit.ConnectRabbit)
	if err != nil {
		log.Fatalf("Не могу подключиться к RabbitMQ - %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Не могу открыть канал - %v", err)
	}
	defer ch.Close()

	q, err2 := ch.QueueDeclare(
		nameQueue, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err2 != nil {
		log.Fatalf("Не могу открыть очередь - %v", err)
	}

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

	if err4 := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bodyJSON,
		}); err4 != nil {
		log.Fatalf("Не могу опубликовать - %v", err4)
	}
	log.Printf(" [x] Sent %s - название очереди %s", body.NameFile, nameQueue)
}
