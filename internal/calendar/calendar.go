package calendar

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	calendarUrl string = "http://isdayoff.ru/"
	DateFormat  string = "02-01-2006"
)

func buildURLForDate(date time.Time) string {
	var (
		url_builder strings.Builder
		month, day  string
	)

	url_builder.WriteString(calendarUrl)
	url_builder.WriteString(strconv.Itoa(date.Year()))

	month = strconv.Itoa(int(date.Month()))
	if len(month) == 1 {
		url_builder.WriteString("0")
	}
	url_builder.WriteString(month)

	day = strconv.Itoa(int(date.Day()))
	if len(day) == 1 {
		url_builder.WriteString("0")
	}
	url_builder.WriteString(day)

	return url_builder.String()
}

/// This function requests isdayoff.ru service to
/// determine if specified date is working day.
/// Isdayoff returns 0 if requested day is working day and
/// 1 if holiday.
/// Detailed information about API is here:
/// https://isdayoff.ru/desc/
func IsWorkingDay(date time.Time) (bool, error) {
	var url string = buildURLForDate(date)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}

	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	result, err := strconv.Atoi(string(answer))
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
