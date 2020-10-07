package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

// Configuration structure
type Config struct {
	// How often to announce a new duty
	DutyCycle time.Duration `yaml:"duty_cycle"`
	// When duty shift starts
	DutyStartAt time.Time `yaml:"duty_start_at"`
	// Telegram bot token
	BotToken string `yaml:"bot_token"`
}

var Cfg Config

// Read config from file and fill Cfg var
func ReadConfig(path string) {

	configdata, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(configdata, &Cfg)
	if err != nil {
		log.Fatal(err)
	}
}
