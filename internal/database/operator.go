package database

// Basic info of Telegram user that can be assigned for duty
type Operator struct {
	ID        int64
	UserName  string
	FirstName string
	LastName  string
}

// DBModel interface implementation

func (op *Operator) Insert() (err error) {
	insertedID, err := execQuery(
		"INSERT INTO operators(username, firstname, lastname) VALUES (?, ?, ?)",
		op.UserName,
		op.FirstName,
		op.LastName,
	)
	if err != nil {
		return
	}
	op.ID = insertedID
	return
}

func (op *Operator) Delete() (err error) {
	_, err = execQuery(
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
	opData := DB.QueryRow(
		"select id, firstname, lastname from operators where username=?",
		op.UserName,
	)
	err = opData.Scan(&op.ID, &op.FirstName, &op.LastName)
	if err != nil {
		return
	}
	return
}

func (op *Operator) GetByID() (err error) {
	opData := DB.QueryRow(
		"select username, firstname, lastname from operators where id=?",
		op.ID,
	)
	err = opData.Scan(&op.UserName, &op.FirstName, &op.LastName)
	if err != nil {
		return
	}
	return
}
