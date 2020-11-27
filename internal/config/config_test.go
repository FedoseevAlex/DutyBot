package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var configString = `
bot_token: 1234567890:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
free_slots_warn_schedule: "0 7 * * FRI"
duty_announce_schedule: "0 7 * * *"
db_connect_string: user:passwd@tcp(10.100.0.100:3306)/database
db_driver: mysql
`

func TestReadConfigFromBytes(t *testing.T) {
	expected := Config{
		DBConnectString:       "user:passwd@tcp(10.100.0.100:3306)/database",
		DBDriver:              "mysql",
		BotToken:              "1234567890:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		FreeSlotsWarnSchedule: "0 7 * * FRI",
		DutyAnnounceSchedule:  "0 7 * * *",
	}
	testConfig := []byte(configString)
	result := readConfigFromBytes(&testConfig)

	assert.Equal(t, expected, *result)
}
