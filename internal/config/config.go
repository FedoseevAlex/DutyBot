package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

const (
	configPath string = "/etc/dutybot/dutybot.yaml"
)

// Configuration structure
type Config struct {
	DBConnectString string `yaml:"db_connect_string"`
	DBDriver        string `yaml:"db_driver"`
	// Telegram bot token
	BotToken string `yaml:"bot_token"`
	// Logfile path
	LogPath string `yaml:"log_path"`
	// Web hook listen address
	ListenAddr string `yaml:"listen_addr"`
	// Certificate file path
	CertPath string `yaml:"cert_path"`
	// Key file path
	KeyPath string `yaml:"key_path"`

	// TODO: Candidates to per chat settings
	// Cron pattern to notify current Duty
	DutyAnnounceSchedule string `yaml:"duty_announce_schedule"`
	// Cron pattern to warn about free duty slots
	FreeSlotsWarnSchedule string `yaml:"free_slots_warn_schedule"`
}

var Cfg *Config

// Public function to read config from standard location
func ReadConfig() {
	configdata, err := ioutil.ReadFile(configPath)
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
