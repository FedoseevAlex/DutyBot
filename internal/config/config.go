package config

import (
	"github.com/spf13/viper"
)

// Public function to read config from standard location
func ReadConfig() error {
	if err := viper.BindEnv("DBConnectString", "DB_CONNECT_STRING"); err != nil {
		return err
	}
	if err := viper.BindEnv("BotToken", "BOT_TOKEN"); err != nil {
		return err
	}

	viper.SetDefault("HookMode", false)
	if err := viper.BindEnv("HookMode", "HOOK_MODE"); err != nil {
		return err
	}
	viper.SetDefault("CertPath", "cert")
	if err := viper.BindEnv("CertPath", "CERT_PATH"); err != nil {
		return err
	}

	viper.SetDefault("KeyPath", "key")
	if err := viper.BindEnv("KeyPath", "KEY_PATH"); err != nil {
		return err
	}

	viper.SetDefault("DBDriver", "postgres")
	if err := viper.BindEnv("DBDriver", "DB_DRIVER"); err != nil {
		return err
	}

	viper.SetDefault("ListenAddress", "0.0.0.0:8080")
	if err := viper.BindEnv("ListenAddress", "LISTEN_ADDRESS"); err != nil {
		return err
	}

	viper.SetDefault("ExternalAddress", "https://0.0.0.0:8443")
	if err := viper.BindEnv("ExternalAddress", "EXTERNAL_ADDRESS"); err != nil {
		return err
	}

	viper.SetDefault("DutyAnnounceSchedule", "0 7 * * *")
	if err := viper.BindEnv("DutyAnnounceSchedule", "ANNOUNCE_SCHEDULE"); err != nil {
		return err
	}

	viper.SetDefault("FreeSlotsWarnSchedule", "0 10 * * FRI")
	if err := viper.BindEnv("FreeSlotsWarnSchedule", "FREE_SLOTS_SCHEDULE"); err != nil {
		return err
	}

	viper.AutomaticEnv()
	return nil
}
