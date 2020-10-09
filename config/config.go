package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

//Config ...
type Config struct {
	MainSettings MainSettings
	Directory    Directory
	Tasks        Tasks
	Rabbit       Rabbit
	Postgres     Postgres
}

// Postgres ...
type Postgres struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
}

//MainSettings ...
type MainSettings struct {
	ServerConnect string `toml:"server_connect"`
	LogLevel      string `toml:"log_level"`
}

// Directory ...
type Directory struct {
	RootPath   string `toml:"root_path"`
	MainFolder string `toml:"main_folder"`
}

// Tasks ...
type Tasks struct {
	Notifications int64 `toml:"notifications"`
	Protocols     int64 `toml:"protocols"`
}

// Rabbit соединение для RabbitMQ
type Rabbit struct {
	ConnectRabbit string `toml:"connection_rabbit"`
}

// NewConf инициализация конфигурации
func NewConf() *Config {
	return &Config{
		MainSettings: MainSettings{},
		Directory:    Directory{},
		Tasks:        Tasks{},
		Postgres:     Postgres{},
	}
}

// ConfigConfigure ...
func (conf *Config) ConfigConfigure() {
	configPath := "config/config.prod.toml"
	_, err := toml.DecodeFile(configPath, conf)
	if err != nil {
		log.Printf("Ошибка - %v", err)
	}
}
