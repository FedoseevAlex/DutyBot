package database

import (
	"dutybot/internal/utils"
	"fmt"
	"log"
	"time"
)

type Assignment struct {
	ID int64
	// Assignment day
	DutyDate time.Time
	// From which chat assignment came from
	ChatID int64
	// Assignee for duty
	Operator *Operator
}

// DBModel interface implementation

func (a *Assignment) Insert() (err error) {
	res, err := db.Exec(
		`INSERT INTO assignments(dutydate,
                                 operator,
                                 chat_id)
         VALUES (?, ?, ?)`,
		a.DutyDate.Format(utils.DateFormat),
		a.Operator.ID,
		a.ChatID,
	)
	if err != nil {
		log.Print(err)
		return
	}
	a.ID, err = res.LastInsertId()
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func (a *Assignment) Delete() (err error) {
	_, err = db.Exec(
		"DELETE FROM assignments WHERE id=?",
		a.ID,
	)
	log.Print(err)
	return
}

func GetAssignmentSchedule(weeks int, chatID int64) (as []*Assignment, err error) {
	hoursInWeek := time.Duration(hoursInDay*daysInWeek) * time.Hour
	today := utils.GetToday()
	// Get future date "weeks" from now
	future := today.Add(hoursInWeek * time.Duration(weeks))

	rows, err := db.Query(
		`SELECT id, dutydate, chat_id, operator
         FROM assignments
         WHERE chat_id=? AND
               dutydate BETWEEN ? AND ?
         ORDER BY dutydate`,
		chatID,
		today.Format(utils.DateFormat),
		future.Format(utils.DateFormat),
	)
	if err != nil {
		log.Print(err)
		return
	}

	if rows.Err() != nil {
		log.Print(rows.Err())
		return
	}
	defer utils.Close(rows)

	for rows.Next() {
		op := &Operator{}
		a := &Assignment{Operator: op}
		var dutyDate string
		err = rows.Scan(&a.ID, &dutyDate, &a.ChatID, &op.ID)
		if err != nil {
			log.Print(err)
			return
		}

		a.DutyDate, err = time.Parse(utils.DateFormat, dutyDate)
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

func GetTodaysAssignment(chatID int64) (*Assignment, error) {
	return GetAssignmentByDate(chatID, utils.GetToday())
}

func GetAssignmentByDate(chatID int64, date *time.Time) (as *Assignment, err error) {
	row := db.QueryRow(
		`SELECT id, dutydate, chat_id, operator
         FROM assignments
         WHERE dutydate=? AND chat_id=?`,
		utils.GetDate(date).Format(utils.DateFormat),
		chatID,
	)
	if err != nil {
		return
	}

	op := &Operator{}
	as = &Assignment{Operator: op}
	var dutyDate string
	err = row.Scan(&as.ID, &dutyDate, &as.ChatID, &op.ID)
	if err != nil {
		err = fmt.Errorf("assignment scan: %s", err)
		return nil, err
	}

	as.DutyDate, err = time.Parse(utils.DateFormat, dutyDate)
	if err != nil {
		err = fmt.Errorf("duty date parse error: %s", err)
		return nil, err
	}

	err = op.GetByID()
	if err != nil {
		err = fmt.Errorf("operator get: %s", err)
		return nil, err
	}
	return
}
