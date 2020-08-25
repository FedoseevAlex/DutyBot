package calendar

import (
	//"net/http"
	"fmt"
	"strings"
	"time"
)

const (
	calendarUrl string = "http://isdayoff.ru/"
	DateFormat string = "02-01-2006"
)

func CheckDay(date time.Time) int {
	var url_builder strings.Builder;
	url_builder.WriteString(calendarUrl)

	fmt.Println(date.Year())
	url_builder.WriteString(string(date.Year()))
	fmt.Println(url_builder.String())

	fmt.Println(date.Month())
	url_builder.WriteString(string(date.Month()))
	fmt.Println(url_builder.String())

	fmt.Println(date.Day())
	url_builder.WriteString(string(date.Day()))
	fmt.Println(url_builder.String())


	fmt.Println(url_builder.String())
	//http.Get()
	return 100
}
