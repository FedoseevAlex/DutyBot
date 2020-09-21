package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Operator struct {
	id        int64
	UserName  string
	FirstName string
	LastName  string
}

type DBModel interface {
	Insert()
	Delete()
}

func (op *Operator) Insert() {
	op.id = createOperator(op.UserName, op.FirstName, op.LastName)
}

func (op *Operator) Delete() {
	deleteOperator(op.id)
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
    lastname TEXT
);
`
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

func createOperator(username string, firstname string, lastname string) int64 {
	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var query string = "INSERT INTO operators(username, firstname, lastname) VALUES (?, ?, ?)"
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	res, err := tx.Exec(query, username, firstname, lastname)
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	inserted_id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return inserted_id
}

func deleteOperator(id int64) {
	db, err := sql.Open("sqlite3", "duty.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var query string = "delete from operators where id=?"
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(query, id)
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
