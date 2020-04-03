package main

import (
	"github.com/BurntSushi/toml"
	"github.com/Sterks/rXmlReader/rabbit"
	"github.com/Sterks/rXmlReader/config"
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
	configPath := "config/config.prod.toml"
	config := config.NewConf()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Println(err)
	}

	consumer := rabbit.NewConsumer(config)
	msgs := consumer.ConsumerMQNow(config, "Files")

	forever := make(chan bool)

	xmlReader := services.NewXMLReader(config)
	//fmt.Println(xmlReader)
	xml := xmlReader.Start(config)

	xml.UnzipFiles(msgs, forever, config)
}
