package database

import (
	"log"
	"time"
)

type Assignment struct {
	ID int64
	// Assignment day
	// DutyDate time.Time
	DutyDate int64
	// From which chat assignment came from
	ChatID int64
	// Assignee for duty
	Operator *Operator
}

// DBModel interface implementation

func (a *Assignment) Insert() (err error) {
	insertedID, err := execQuery(
		`INSERT INTO assignments(dutydate,
                                 operator,
                                 chat_id,
                                 created_at)
         VALUES (?, ?, ?, ?)`,
		a.DutyDate,
		a.Operator.ID,
		a.ChatID,
		time.Now().Format(time.RFC1123),
	)
	if err != nil {
		log.Print(err)
		return
	}
	a.ID = insertedID
	return
}

func (a *Assignment) Delete() (err error) {
	_, err = execQuery(
		"DELETE FROM assignments WHERE id=?",
		a.ID,
	)
	log.Print(err)
	return
}
