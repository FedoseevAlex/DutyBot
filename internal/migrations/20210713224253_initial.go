package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upInitial, downInitial)
}

func upInitial(tx *sql.Tx) error {
	createAssignments := `
	CREATE TABLE assignments (
		uuid UUID DEFAULT uuid_generate_v4(),
		at DATE NOT NULL,
		operator TEXT,
		chat_id BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(operator, chat_id, at)
	)
	`
	_, err := tx.Exec(createAssignments)
	if err != nil {
		return err
	}

	return nil
}

func downInitial(tx *sql.Tx) error {
	dropAssignments := "DROP TABLE assignments"
	_, err := tx.Exec(dropAssignments)
	if err != nil {
		return err
	}
	return nil
}
