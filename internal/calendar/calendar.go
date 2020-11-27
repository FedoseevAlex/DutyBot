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
	calendarURL string = "http://isdayoff.ru"
	dateFormat  string = "20060102"
)

type TimeSet map[time.Time]struct{}

func buildQueryString(address string, endpoint []string, queryParams map[string]string) string {
	parts := make([]string, 0, len(endpoint)+1)
	parts = append(parts, address)
	parts = append(parts, endpoint...)

	url := strings.Join(parts, "/")
	if queryParams == nil {
		return url
	}

	queryParts := make([]string, 0, len(queryParams))
	for key, value := range queryParams {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}
	query := strings.Join(queryParts, "&")

	return fmt.Sprintf("%s?%s", url, query)
}

// This function requests isdayoff.ru service to
// determine if specified date is working day.
// Isdayoff returns 0 if requested day is working day and
// 1 if holiday.
// Detailed information about API is here:
// https://isdayoff.ru/desc/
func IsHoliday(date time.Time) (isHoliday bool) {
	isHoliday = date.Weekday() > time.Friday
	if isHoliday {
		return
	}

	client := http.DefaultClient
	URL := buildQueryString(calendarURL, []string{date.Format(dateFormat)}, nil)
	req, err := http.NewRequestWithContext(
		context.TODO(),
		http.MethodGet,
		URL,
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

// Get working days as map time.Time: bool.
// True value means that day in key is holiday.
func GetWorkingDays(start time.Time, stop time.Time) (TimeSet, error) {
	client := http.DefaultClient

	URL := buildQueryString(
		calendarURL,
		[]string{
			"api",
			"getdata",
		},
		map[string]string{
			"date1": start.Format(dateFormat),
			"date2": stop.Format(dateFormat),
		})

	req, _ := http.NewRequestWithContext(
		context.TODO(),
		http.MethodGet,
		URL,
		http.NoBody,
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer utils.Close(resp.Body)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	calendar := TimeSet{}
	for date, i := start, 0; !date.After(stop); date, i = date.Add(utils.DayDuration), i+1 {
		if respData[i] == '1' {
			continue
		}
		calendar[date] = struct{}{}
	}

	return calendar, nil
}
