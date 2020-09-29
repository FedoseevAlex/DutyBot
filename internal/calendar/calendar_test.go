package calendar

import (
	"fmt"
	"testing"
	"time"
)

type TestData struct {
	Date   time.Time
	Answer string
}

func TestBuildURLForDate(t *testing.T) {
	datesAnswers := []TestData{
		{
			time.Date(2020, 9, 1, 0, 0, 0, 0, time.Now().Location()),
			fmt.Sprintf("%s%s", calendarURL, "20200901")},
		{
			time.Date(2020, 11, 1, 0, 0, 0, 0, time.Now().Location()),
			fmt.Sprintf("%s%s", calendarURL, "20201101")},
		{
			time.Date(2020, 9, 11, 0, 0, 0, 0, time.Now().Location()),
			fmt.Sprintf("%s%s", calendarURL, "20200911")},
		{
			time.Date(2020, 11, 11, 0, 0, 0, 0, time.Now().Location()),
			fmt.Sprintf("%s%s", calendarURL, "20201111")},
	}
	for _, test := range datesAnswers {
		result := buildURLForDate(test.Date)
		if result != test.Answer {
			t.Errorf("got %s expected %s", result, test.Answer)
		}
	}
}
