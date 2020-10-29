package main

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Sterks/rXmlReader/config"
	"github.com/Sterks/rXmlReader/rabbit"
	"github.com/Sterks/rXmlReader/services"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func init() {
	customFormat := new(logrus.TextFormatter)
	customFormat.TimestampFormat = "2006-01-02 15:04:05"
	customFormat.FullTimestamp = true
	customFormat.ForceColors = true
	log.SetFormatter(customFormat)
}

func main() {

	time.Sleep(30 * time.Second)
	// TODO Перенести конфиг в корень
	configPath := "config/config.prod.toml"
	config := config.NewConf()
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Println(err)
	}

	c := make(chan bool)
	consumer := rabbit.NewConsumer(config)
	msgs1 := consumer.ConsumerMQNow(config, "notifications44")
	msgs2 := consumer.ConsumerMQNow(config, "notifications223")
	msgs3 := consumer.ConsumerMQNow(config, "protocols44")
	msgs4 := consumer.ConsumerMQNow(config, "protocols223")
	forever := make(chan bool)
	xmlReader := services.NewXMLReader(config)
	xml := xmlReader.Start(config)
	go xml.UnzipFiles(msgs1, forever, config, "notifications44")
	go xml.UnzipFiles(msgs2, forever, config, "notifications223")
	go xml.UnzipFiles(msgs3, forever, config, "protocols44")
	go xml.UnzipFiles(msgs4, forever, config, "protocols223")
	<-c
}
