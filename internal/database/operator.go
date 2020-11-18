package database

import (
	"dutybot/internal/utils"
	"time"
)

// Basic info of Telegram user that can be assigned for duty
type Operator struct {
	ID        int64
	UserName  string
	FirstName string
	LastName  string
}

// DBModel interface implementation

func (op *Operator) Insert() (err error) {
	res, err := db.Exec(
		"INSERT INTO operators(username, firstname, lastname) VALUES (?, ?, ?)",
		op.UserName,
		op.FirstName,
		op.LastName,
	)
	if err != nil {
		return
	}

	op.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	return
}

func (op *Operator) Delete() (err error) {
	_, err = db.Exec(
		"DELETE FROM operators WHERE id=?",
		op.ID,
	)
	if err != nil {
		return
	}
	return
}

// Methods

func (op *Operator) GetByUserName() (err error) {
	opData := db.QueryRow(
		"SELECT id, firstname, lastname FROM operators WHERE username=?",
		op.UserName,
	)
	err = opData.Scan(&op.ID, &op.FirstName, &op.LastName)
	if err != nil {
		return
	}
	return
}

func (op *Operator) GetByID() (err error) {
	opData := db.QueryRow(
		"SELECT username, firstname, lastname FROM operators WHERE id=?",
		op.ID,
	)
	err = opData.Scan(&op.UserName, &op.FirstName, &op.LastName)
	if err != nil {
		return
	}
	return
}

func GetAllTodaysOperators() (as []*Assignment, err error) {
	rows, err := db.Query(
		`SELECT id, dutydate, chat_id, operator
         FROM assignments
         WHERE dutydate=?`,
		utils.GetToday().Format(utils.DateFormat),
	)
	if err != nil {
		return
	}
	if rows.Err() != nil {
		err = rows.Err()
		return
	}
	defer utils.Close(rows)

	for rows.Next() {
		op := &Operator{}
		a := &Assignment{Operator: op}

		var dutyDate string

		err = rows.Scan(&a.ID, &dutyDate, &a.ChatID, &op.ID)
		if err != nil {
			return nil, err
		}

		a.DutyDate, err = time.Parse(utils.DateFormat, dutyDate)
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
