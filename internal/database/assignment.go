package database

import (
	"fmt"
	"sort"
	"time"

	"github.com/FedoseevAlex/DutyBot/internal/calendar"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
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
		`INSERT INTO assignments(
			dutydate,
			operator,
			chat_id)
		VALUES (?, ?, ?)`,
		a.DutyDate.Format(utils.DateFormat),
		a.Operator.ID,
		a.ChatID,
	)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return
	}
	a.ID, err = res.LastInsertId()
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return
	}
	return
}

func (a *Assignment) Delete() (err error) {
	_, err = db.Exec(
		"DELETE FROM assignments WHERE id=?",
		a.ID,
	)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
	}
	return
}

// Return assignment for specified chat and number of weeks ahead
func GetAssignmentSchedule(weeks int, chatID int64) (as []*Assignment, err error) {
	hoursInWeek := time.Duration(utils.HoursInDay*utils.DaysInWeek) * time.Hour
	today := utils.GetToday()
	// Get future date "weeks" from now
	future := today.Add(hoursInWeek * time.Duration(weeks))

	rows, err := db.Query(
		`SELECT id, dutydate, chat_id, operator
		 FROM assignments
		 WHERE chat_id=? AND dutydate BETWEEN ? AND ?
		 ORDER BY dutydate`,
		chatID,
		today.Format(utils.DateFormat),
		future.Format(utils.DateFormat),
	)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return
	}

	if rows.Err() != nil {
		logger.Log.Error().Stack().Err(rows.Err()).Send()
		return
	}
	defer utils.Close(rows)

	for rows.Next() {
		op := &Operator{}
		a := &Assignment{Operator: op}
		var dutyDate string
		err = rows.Scan(&a.ID, &dutyDate, &a.ChatID, &op.ID)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return
		}

		a.DutyDate, err = time.Parse(utils.DateFormat, dutyDate)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return
		}

		err = op.GetByID()
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
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

	var dutyDate string
	op := &Operator{}
	as = &Assignment{Operator: op}

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

// Return all chat ids
func GetAllChats() ([]int64, error) {
	res, err := db.Query(`SELECT DISTINCT(chat_id) FROM assignments`)
	if err != nil {
		return nil, err
	}
	defer utils.Close(res)

	chats := make([]int64, 0)

	for res.Next() {
		var chatID int64

		err = res.Scan(&chatID)
		if err != nil {
			return nil, err
		}

		chats = append(chats, chatID)
	}
	return chats, nil
}

// Return free duty slots for
// specified number of weeks
func GetFreeSlots(weeks int, chatID int64) (freedates []time.Time, err error) {
	start := utils.GetToday()
	stop := start.Add(time.Duration(utils.HoursInDay*utils.DaysInWeek*weeks) * time.Hour)
	dates, err := calendar.GetWorkingDays(*start, stop)
	if err != nil {
		return
	}

	res, err := db.Query(`SELECT dutydate FROM assignments
                                WHERE chat_id=?
                                AND
                                dutydate BETWEEN ? AND ?`,
		chatID,
		start.Format(utils.DateFormat),
		stop.Format(utils.DateFormat),
	)
	if err != nil {
		return
	}
	defer utils.Close(res)

	var buf string

	for res.Next() {
		err = res.Scan(&buf)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return
		}

		date, err := time.Parse(utils.DateFormat, buf)
		if err != nil {
			return nil, err
		}

		dates.Remove(date)
	}

	for freedate := range *dates {
		freedates = append(freedates, freedate)
	}

	sort.Slice(freedates, func(i int, j int) bool {
		return freedates[i].Before(freedates[j])
	})
	return
}
