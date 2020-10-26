package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	hoursInDay int = 24
	daysInWeek int = 7
)

var DB *sql.DB

func InitDB(connStr string) error {
	var err error

	DB, err = sql.Open("sqlite3", connStr)
	if err != nil {
		return err
	}

	err = createSchema()
	if err != nil {
		return err
	}

	return nil
}

func GetAllTodaysOperators() (as []*Assignment, err error) {
	year, month, day := time.Now().Date()
	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	today := time.Date(year, month, day, 0, 0, 0, 0, utcLocation)

	rows, err := DB.Query(
		`SELECT id, dutydate, chat_id, operator
         FROM assignments
         WHERE dutydate=?`,
		today.Unix(),
	)
	if err != nil {
		return
	}
	if rows.Err() != nil {
		err = rows.Err()
		return
	}
	defer rows.Close()

	for rows.Next() {
		op := &Operator{}
		a := &Assignment{Operator: op}
		err = rows.Scan(&a.ID, &a.DutyDate, &a.ChatID, &op.ID)
		if err != nil {
			return nil, err
		}

		err = op.GetByID()
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return
}

func GetTodaysAssignment(chatID int64) (*Assignment, error) {
	return GetAssignmentByDate(chatID, time.Now())
}

func GetAssignmentByDate(chatID int64, date time.Time) (as *Assignment, err error) {
	year, month, day := date.Date()
	dutydate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	row := DB.QueryRow(
		`SELECT id, dutydate, chat_id, operator
         FROM assignments
         WHERE dutydate=? AND chat_id=?`,
		dutydate.Unix(),
		chatID,
	)
	log.Println("today", dutydate.Unix(), "chatid", chatID)
	if err != nil {
		return
	}

	op := &Operator{}
	as = &Assignment{Operator: op}
	err = row.Scan(&as.ID, &as.DutyDate, &as.ChatID, &op.ID)
	if err != nil {
		err = fmt.Errorf("Assignment scan: %s", err)
		return nil, err
	}

	err = op.GetByID()
	if err != nil {
		err = fmt.Errorf("Operator get: %s", err)
		return nil, err
	}
	return
}

func GetAssignmentSchedule(weeks int, chatID int64) (as []*Assignment, err error) {
	hoursInWeek := time.Duration(hoursInDay*daysInWeek) * time.Hour
	year, month, day := time.Now().Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	// Get future date "weeks" from now
	future := today.Add(hoursInWeek * time.Duration(weeks))

	rows, err := DB.Query(
		`SELECT id, dutydate, chat_id, operator
         FROM assignments
         WHERE chat_id=? AND
               dutydate BETWEEN ? AND ?
         ORDER BY dutydate`,
		chatID,
		today.Unix(),
		future.Unix(),
	)
	if err != nil {
		log.Print(err)
		return
	}

	if rows.Err() != nil {
		log.Print(rows.Err())
		return
	}
	defer rows.Close()

	for rows.Next() {
		op := &Operator{}
		a := &Assignment{Operator: op}
		err = rows.Scan(&a.ID, &a.DutyDate, &a.ChatID, &op.ID)
		if err != nil {
			log.Print(err)
			return
		}
		err = op.GetByID()
		if err != nil {
			log.Print(err)
			return
		}
		as = append(as, a)
	}
	return
}
