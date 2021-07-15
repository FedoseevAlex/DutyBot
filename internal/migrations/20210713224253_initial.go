package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upInitial, downInitial)
}

func upInitial(tx *sql.Tx) error {
	createOperators := `
	CREATE TABLE operators (
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(32) UNIQUE NOT NULL,
		firstname TEXT,
		lastname TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := tx.Exec(createOperators)
	if err != nil {
		return err
	}

	createAssignments := `
	CREATE TABLE assignments (
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		dutydate DATE NOT NULL,
		operator INTEGER,
		chat_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(operator) REFERENCES operators(id),
		UNIQUE(operator, chat_id, dutydate)
	)
	`
	_, err = tx.Exec(createAssignments)
	if err != nil {
		return err
	}

	return nil
}

func downInitial(tx *sql.Tx) error {
	dropOperators := "DROP TABLE operators"
	_, err := tx.Exec(dropOperators)
	if err != nil {
		return err
	}

	dropAssignments := "DROP TABLE assignments"
	_, err = tx.Exec(dropAssignments)
	if err != nil {
		return err
	}
	return nil
}
