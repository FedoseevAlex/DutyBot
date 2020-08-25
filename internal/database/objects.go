package database

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
