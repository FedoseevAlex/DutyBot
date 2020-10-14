package database

import (
	"fmt"
	"log"
)

type DBModel interface {
	Insert() error
	Delete() error
}

// Execute query without any resulting row
func execQuery(query string, args ...interface{}) (insertedID int64, err error) {
	insertedID = -1
	err = nil

	fmt.Println(args...)

	res, err := DB.Exec(query, args...)
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

func createSchema() error {
	var schema string = `
PRAGMA foreign_keys = ON;

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
	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}
	return nil
}
