package calendar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Address     string
	Endpoint    []string
	QueryParams map[string]string
	Answers     []string
}

func TestBuildURLForDate(t *testing.T) {
	datesAnswers := []TestData{
		{
			Address:     "https://someaddress.ru",
			Endpoint:    nil,
			QueryParams: nil,
			Answers:     []string{"https://someaddress.ru"},
		},
		{
			Address: "https://someaddress.ru",
			Endpoint: []string{
				"api",
				"getdata",
			},
			QueryParams: nil,
			Answers:     []string{"https://someaddress.ru/api/getdata"},
		},
		{
			Address: "https://someaddress.ru",
			Endpoint: []string{
				"api",
				"getdata",
			},
			QueryParams: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
			Answers: []string{
				"https://someaddress.ru/api/getdata?param1=value1&param2=value2",
				"https://someaddress.ru/api/getdata?param2=value2&param1=value1",
			},
		},
		{
			Address:  "https://someaddress.ru",
			Endpoint: nil,
			QueryParams: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
			Answers: []string{
				"https://someaddress.ru?param1=value1&param2=value2",
				"https://someaddress.ru?param2=value2&param1=value1",
			},
		},
	}
	for _, test := range datesAnswers {
		assert.Contains(t, test.Answers, buildQueryString(test.Address, test.Endpoint, test.QueryParams))
	}
}
