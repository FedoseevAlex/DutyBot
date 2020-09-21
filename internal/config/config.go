package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Configuration structure
type Config struct {
	// How often to remind about duty
	DutyCycle int
	// Telegram bot token
	BotToken string

}

// Read config from file and return Config struct
func ReadConfig(path string) Config {
	fmt.Println(path)
	var config Config

	configdata, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(configdata, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}
