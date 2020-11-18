package calendar

import (
	"context"
	"dutybot/internal/utils"
	"fmt"
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
	var URLBuilder strings.Builder

	URLBuilder.WriteString(calendarURL)
	URLBuilder.WriteString(strconv.Itoa(date.Year()))
	URLBuilder.WriteString(fmt.Sprintf("%02d", date.Month()))
	URLBuilder.WriteString(fmt.Sprintf("%02d", date.Day()))
	return URLBuilder.String()
}

/// This function requests isdayoff.ru service to
/// determine if specified date is working day.
/// Isdayoff returns 0 if requested day is working day and
/// 1 if holiday.
/// Detailed information about API is here:
/// https://isdayoff.ru/desc/
func IsHoliday(date time.Time) (isHoliday bool) {
	isHoliday = date.Weekday() > time.Friday
	if isHoliday {
		return
	}

	client := http.DefaultClient
	url := buildURLForDate(date)
	req, err := http.NewRequestWithContext(
		context.TODO(),
		http.MethodGet,
		url,
		http.NoBody,
	)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer utils.Close(resp.Body)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	answer, err := strconv.Atoi(string(respData))
	if err != nil {
		return
	}
	return answer == 1
}
