package database

import (
	"database/sql"
	"log"
	"time"
)

type DBModel interface {
	Insert() error
	Delete() error
}

func GetAllTodaysOperators() (as []*Assignment, err error) {
	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		return
	}
	defer db.Close()

	year, month, day := time.Now().Date()
	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	today := time.Date(year, month, day, 0, 0, 0, 0, utcLocation)

	rows, err := db.Query(
		"select id, dutydate, chat_id, operator from assignments where dutydate=?",
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


func GetTodaysOperator(chatID int64) (as []*Assignment, err error) {
	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		return
	}
	defer db.Close()

	year, month, day := time.Now().Date()
	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	today := time.Date(year, month, day, 0, 0, 0, 0, utcLocation)

	row := db.QueryRow(
		"select id, dutydate, chat_id, operator from assignments where dutydate=? and chat_id=?",
		today.Unix(),
		chatID,
	)
	log.Println("today", today.Unix(), "chatid", chatID)
	if err != nil {
		return
	}

	op := &Operator{}
	a := &Assignment{Operator: op}
	err = row.Scan(&a.ID, &a.DutyDate, &a.ChatID, &op.ID)
	if err != nil {
		return nil, err
	}

	err = op.GetByID()
	if err != nil {
		return nil, err
	}
	as = append(as, a)
	return
}

func GetAssignmentSchedule(weeks int, chatID int64) (as []*Assignment, err error) {
	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		log.Print(err)
		return
	}
	defer db.Close()

	year, month, day := time.Now().Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location())
	// Get future date "weeks" from now
	future := today.Add(time.Hour * time.Duration(weeks*7*24))

	rows, err := db.Query(
		"select id, dutydate, chat_id, operator from assignments where chat_id=? and dutydate BETWEEN ? and ?",
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

// Execute query without any resulting row
func execQuery(query string, args ...interface{}) (insertedID int64, err error) {
	insertedID = -1
	err = nil

	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		log.Print(err)
		return
	}
	defer db.Close()

	res, err := db.Exec(query, args...)
	if err != nil {
		log.Print(err)
		return
	}

	insertedID, err = res.LastInsertId()
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func CreateSchema() {
	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var schema string = `
CREATE TABLE IF NOT EXISTS operators (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    firstname TEXT,
    lastname TEXT,
    created_at TEXT
);

CREATE TABLE IF NOT EXISTS assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dutydate TEXT NOT NULL,
    operator INTEGER,
    chat_id INTEGER,
    created_at TEXT,
    FOREIGN KEY(operator) REFERENCES operators(id),
    UNIQUE(operator, chat_id, dutydate)
);
`
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}
