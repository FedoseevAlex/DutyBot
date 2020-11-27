package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	DateFormat      string        = "2006-01-02"
	HumanDateFormat string        = "Mon Jan 02 2006"
	HoursInDay      int           = 24
	DaysInWeek      int           = 7
	DayDuration     time.Duration = time.Duration(HoursInDay) * time.Hour
)

type PrettyTable struct {
	rows [][]string
}

func (pt *PrettyTable) String() (string, error) {
	buf := make([]byte, 0)
	b := bytes.NewBuffer(buf)
	t := tabwriter.NewWriter(
		b,
		0,
		4,
		2,
		' ',
		tabwriter.TabIndent,
	)

	for _, rowParts := range pt.rows {
		row := strings.Join(rowParts, "\t")
		_, err := fmt.Fprintf(t, "%s\n", row)
		if err != nil {
			return "", err
		}
	}

	err := t.Flush()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (pt *PrettyTable) AddRow(row []string) {
	pt.rows = append(pt.rows, row)
}

func NewPrettyTable() *PrettyTable {
	return &PrettyTable{}
}

// This function returns time.Time object
// representing current date.
func GetToday() time.Time {
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return today
}

// This function returns time.Time object
// representing tomorrow date.
func GetTomorrow() time.Time {
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	dayDuration, _ := time.ParseDuration("1d")
	tomorrow := today.Add(dayDuration)
	return tomorrow
}

// Function that strips hours minutes and seconds
// from given time.Time. Returns pointer to time.Time object
// representing only date.
func GetDate(date *time.Time) *time.Time {
	y, m, d := date.Date()
	onlyDate := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &onlyDate
}

// Simple wrapper for objects that need to be closed.
// Could be used with defer statement to avoid unhandled
// error from Close function.
func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}
