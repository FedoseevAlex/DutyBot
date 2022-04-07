package config

import (
	"github.com/spf13/viper"
)

// Public function to read config from standard location
func ReadConfig() {
	viper.BindEnv("DBConnectString", "DB_CONNECT_STRING")
	viper.BindEnv("BotToken", "BOT_TOKEN")

	viper.BindEnv("DBDriver", "DB_DRIVER")
	viper.SetDefault("DBDriver", "postgres")

	viper.BindEnv("LogPath", "LOG_PATH")
	viper.SetDefault("LogPath", "/var/log/dutybot.log")

	viper.BindEnv("ListenAddress", "LISTEN_ADDRESS")
	viper.SetDefault("ListenAddress", "0.0.0.0:8443")

	viper.BindEnv("CertPath", "CERT_PATH")
	viper.SetDefault("CertPath", "/etc/dutybot/pub.pem")

	viper.BindEnv("KeyPath", "KEY_PATH")
	viper.SetDefault("KeyPath", "/etc/dutybot/priv.key")

	viper.BindEnv("DutyAnnounceSchedule", "ANNOUNCE_SCHEDULE")
	viper.SetDefault("DutyAnnounceSchedule", "0 10 * * *")

	viper.BindEnv("FreeSlotsWarnSchedule", "FREE_SLOTS_SCHEDULE")
	viper.SetDefault("FreeSlotsWarnSchedule", "0 10 * * FRI")

	viper.AutomaticEnv()
}
