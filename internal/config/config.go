package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

type Clock struct {
	time.Time
}

// Configuration structure
type Config struct {
	DBConnectString string `yaml:"db_connect_string"`
	// How often to announce a new duty
	DutyShift time.Duration `yaml:"duty_shift"`
	// Time when duty shift starts as time.Time
	DutyStartAt Clock `yaml:"duty_start_at"`
	// Telegram bot token
	BotToken string `yaml:"bot_token"`
}

var Cfg Config

func (c *Clock) UnmarshalYAML(unmarshal func (interface {}) error) error {
	var tmp string
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	c.Time, err = time.Parse("15:04:05", tmp)
	if err != nil {
		return err
	}
	return nil
}

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
