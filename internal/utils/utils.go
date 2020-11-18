package utils

import (
	"io"
	"log"
	"time"
)

const (
	DateFormat string = "2006-01-02"
)

/// This function returns time.Time object
/// representing current date.
func GetToday() *time.Time {
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &today
}

/// This function returns time.Time object
/// representing tomorrow date.
func GetTomorrow() *time.Time {
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	dayDuration, _ := time.ParseDuration("1d")
	tomorrow := today.Add(dayDuration)
	return &tomorrow
}

/// Function that strips hours minutes and seconds
/// from given time.Time. Returns pointer to time.Time object
/// representing only date.
func GetDate(date *time.Time) *time.Time {
	y, m, d := date.Date()
	onlyDate := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &onlyDate
}

/// Simple wrapper for objects that need to be closed.
/// Could be used with defer statement to avoid unhandled
/// error from Close function.
func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}
