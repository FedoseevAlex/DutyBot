package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	ConfigPath string = "/etc/dutybot/dutybot.yaml"
)

type Clock struct {
	time.Time
}

// Configuration structure
type Config struct {
	DBConnectString string `yaml:"db_connect_string"`
	DBDriver        string `yaml:"db_driver"`
	// How often to announce a new duty
	DutyShift time.Duration `yaml:"duty_shift"`
	// Time when duty shift starts as time.Time
	DutyStartAt Clock `yaml:"duty_start_at"`
	// Telegram bot token
	BotToken string `yaml:"bot_token"`
}

var Cfg *Config

func (c *Clock) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

// Public function to read config from standard location
func ReadConfig() {
	configdata, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	Cfg = readConfigFromBytes(&configdata)
}

// Read config from file and fill Cfg var
func readConfigFromBytes(contents *[]byte) (config *Config) {
	err := yaml.Unmarshal(*contents, &config)
	if err != nil {
		log.Fatal(err)
	}
	return
}
