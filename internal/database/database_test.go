package database

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	_ = os.Remove("duty.db")
	CreateSchema()
	retCode := m.Run()
	os.Exit(retCode)
	// os.Remove("duty.db")
}

func TestExecQueryInsert(t *testing.T) {
	id, err := execQuery(
		"INSERT INTO operators(username, firstname, lastname) VALUES (?, ?, ?)",
		"username1",
		"firstname1",
		"lastname1",
	)
	if err != nil {
		t.Errorf("Got error %s", err)
	}
	if id != 1 {
		t.Errorf("Wrong id got: %d, expected: %d", id, 1)
	}
}

func TestExecQueryDelete(t *testing.T) {
	id, err := execQuery(
		"DELETE FROM operators WHERE username=? AND firstname=? AND lastname=?",
		"username1",
		"firstname1",
		"lastname1",
	)
	if err != nil {
		t.Errorf("Got error %s", err)
	}
	if id != -1 {
		t.Errorf("Wrong id got: %d, expected: %d", id, -1)
	}
}

func TestOperatorInsert(t *testing.T) {
	var op = Operator{UserName: "operator1", FirstName: "op1_name", LastName: "op1_surname"}
	err := op.Insert()
	if err != nil {
		t.Fail()
	}
}

func TestAssignmentInsert(t *testing.T) {
	var op = Operator{UserName: "operator2", FirstName: "op1_name", LastName: "op1_surname"}
	err := op.Insert()
	if err != nil {
		t.Fail()
	}

	date := time.Date(2020, 9, 26, 0, 0, 0, 0, time.Now().Location())
	var a = Assignment{DutyDate: date.Unix(), Operator: &op}
	err = a.Insert()
	if err != nil {
		t.Fail()
	}
}
