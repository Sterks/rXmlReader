package main

import (
	"github.com/BurntSushi/toml"
	"github.com/Sterks/rXmlReader/config"
	"github.com/Sterks/rXmlReader/rabbit"
	"github.com/Sterks/rXmlReader/services"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "user_ro"
	password = "4r2w3e1q"
	dbname   = "freader"
)

func init() {
	customFormat := new(logrus.TextFormatter)
	customFormat.TimestampFormat = "2006-01-02 15:04:05"
	customFormat.FullTimestamp = true
	customFormat.ForceColors = true
	log.SetFormatter(customFormat)
}

func main() {
	//TODO Перенести конфиг в корень
	configPath := "config/config.prod.toml"
	config := config.NewConf()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Println(err)
	}

	c := make(chan bool)
	consumer := rabbit.NewConsumer(config)
	msgs1 := consumer.ConsumerMQNow(config, "Notifications44")
	msgs2 := consumer.ConsumerMQNow(config, "Notifications223")
	msgs3 := consumer.ConsumerMQNow(config, "Protocols44")
	msgs4 := consumer.ConsumerMQNow(config, "Protocols223")
	forever := make(chan bool)
	xmlReader := services.NewXMLReader(config)
	xml := xmlReader.Start(config)
	go xml.UnzipFiles(msgs1, forever, config)
	go xml.UnzipFiles(msgs2, forever, config)
	go xml.UnzipFiles(msgs3, forever, config)
	go xml.UnzipFiles(msgs4, forever, config)
	<-c
	// 	c := make(chan bool)
	// 	consumer := rabbit.NewConsumer(config)
	// 	msgs := consumer.ConsumerMQNow(config, "Files")
	// 	forever := make(chan bool)
	// 	xmlReader := services.NewXMLReader(config)
	// 	xml := xmlReader.Start(config)
	// 	go xml.UnzipFiles(msgs, forever, config)

	// 	consumer2 := rabbit.NewConsumer(config)
	// 	msgs2 := consumer2.ConsumerMQNow(config, "Files223")
	// 	forever2 := make(chan bool)
	// 	xmlReader2 := services.NewXMLReader(config)
	// 	xml2 := xmlReader2.Start(config)
	// 	go xml2.UnzipFiles(msgs2, forever2, config)
	// 	<-c
}
