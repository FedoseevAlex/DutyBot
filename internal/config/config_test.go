package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var configString = `
duty_shift: 24h
bot_token: 1234567890:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
duty_start_at: 19:17:30
db_connect_string: user:passwd@tcp(10.100.0.100:3306)/database
`

func TestReadConfigFromBytes(t *testing.T) {
	expectedDuration, _ := time.ParseDuration("24h")

	expectedStartTime, _ := time.Parse("15:04:05", "19:17:30")
	var expectedStartAt Clock
	expectedStartAt.Time = expectedStartTime

	expected := Config{
		DBConnectString: "user:passwd@tcp(10.100.0.100:3306)/database",
		DutyShift:       expectedDuration,
		DutyStartAt:     expectedStartAt,
		BotToken:        "1234567890:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	testConfig := []byte(configString)
	result := readConfigFromBytes(&testConfig)

	assert.Equal(t, &expected, result)
}
