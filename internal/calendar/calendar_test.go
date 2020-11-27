package calendar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Address     string
	Endpoint    []string
	QueryParams map[string]string
	Answer      string
}

func TestBuildURLForDate(t *testing.T) {
	datesAnswers := []TestData{
		{
			Address:     "https://someaddress.ru",
			Endpoint:    nil,
			QueryParams: nil,
			Answer:      "https://someaddress.ru",
		},
		{
			Address: "https://someaddress.ru",
			Endpoint: []string{
				"api",
				"getdata",
			},
			QueryParams: nil,
			Answer:      "https://someaddress.ru/api/getdata",
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
			Answer: "https://someaddress.ru/api/getdata?param1=value1&param2=value2",
		},
		{
			Address:  "https://someaddress.ru",
			Endpoint: nil,
			QueryParams: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
			Answer: "https://someaddress.ru?param1=value1&param2=value2",
		},
	}
	for _, test := range datesAnswers {
		assert.Equal(t, buildQueryString(test.Address, test.Endpoint, test.QueryParams), test.Answer)
	}
}
