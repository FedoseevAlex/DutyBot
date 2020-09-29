package calendar

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	calendarURL string = "http://isdayoff.ru/"
	// Date format is DD-MM-YYYY
	DateFormat string = "02-01-2006"
)

func buildURLForDate(date time.Time) string {
	var (
		URLBuilder strings.Builder
		month, day string
	)

	URLBuilder.WriteString(calendarURL)
	URLBuilder.WriteString(strconv.Itoa(date.Year()))

	month = strconv.Itoa(int(date.Month()))
	if len(month) == 1 {
		URLBuilder.WriteString("0")
	}
	URLBuilder.WriteString(month)

	day = strconv.Itoa(date.Day())
	if len(day) == 1 {
		URLBuilder.WriteString("0")
	}
	URLBuilder.WriteString(day)

	return URLBuilder.String()
}

/// This function requests isdayoff.ru service to
/// determine if specified date is working day.
/// Isdayoff returns 0 if requested day is working day and
/// 1 if holiday.
/// Detailed information about API is here:
/// https://isdayoff.ru/desc/
func IsWorkingDay(date time.Time) (isWorkingDay bool, err error) {
	var url string = buildURLForDate(date)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	answer, err := strconv.Atoi(string(respData))
	if err != nil {
		return
	}
	isWorkingDay = answer == 1
	return
}
