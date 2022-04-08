package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReadConfigFromEnvironmentVariables(t *testing.T) {
	err := os.Setenv("DB_CONNECT_STRING", "foobarbaz_db_connect_wow")
	if err != nil {
		t.Error(err)
	}

	err = os.Setenv("BOT_TOKEN", "bot_token_official")
	if err != nil {
		t.Error(err)
	}

	if err := ReadConfig(); err != nil {
		t.Error(err)
	}

	assert.Equal(t, "foobarbaz_db_connect_wow", viper.GetString("DB_CONNECT_STRING"))
	assert.Equal(t, "bot_token_official", viper.GetString("BOT_TOKEN"))
}
