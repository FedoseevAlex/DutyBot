package assignment

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/FedoseevAlex/DutyBot/internal/calendar"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var _ AssignmentRepoer = &AssignmentRepoData{}

// errors
var (
	ErrNotInserted = errors.New("pgx CommandTag is not INSERT")
	ErrNotDeleted  = errors.New("pgx CommandTag is not DELETE")
)

func (asr *AssignmentRepoData) AddAssignment(ctx context.Context, as Assignment) error {
	sql, params, err := goqu.Insert(assignmentsTableName).
		Rows(as).ToSQL()
	if err != nil {
		return err
	}
	logger.Log.Debug().Str("sql", sql).Send()

	result, err := asr.conn.Exec(ctx, sql, params)
	if err != nil {
		return err
	}
	if !result.Insert() {
		return ErrNotInserted
	}
	return nil
}

func (asr *AssignmentRepoData) DeleteAssignment(ctx context.Context, uid uuid.UUID) error {
	sql, params, err := goqu.Delete(assignmentsTableName).
		Where(goqu.Ex{
			"uuid": uid.String(),
		}).ToSQL()
	if err != nil {
		return err
	}
	logger.Log.Debug().Str("sql", sql).Send()
	result, err := asr.conn.Exec(ctx, sql, params...)
	if err != nil {
		return err
	}
	if !result.Delete() {
		return ErrNotDeleted
	}
	return nil
}

// Return assignments due specified date and for specified chat
func (asr *AssignmentRepoData) GetAssignmentSchedule(
	ctx context.Context,
	due time.Time,
	chatID int64) ([]Assignment, error) {
	today := utils.GetToday()

	sql, params, err := goqu.From(assignmentsTableName).
		Select("id", "at", "chat_id", "operator").
		Where(goqu.Ex{
			"chat_id": chatID,
			"at": goqu.I("at").
				Between(
					exp.NewRangeVal(
						today.Format(utils.DateFormat),
						due.Format(utils.DateFormat),
					)),
		}).
		Order(goqu.I("at").Desc()).
		ToSQL()
	logger.Log.Debug().Str("sql", sql).Send()

	if err != nil {
		return nil, err
	}

	rows, err := asr.conn.Query(ctx, sql, params...)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return []Assignment{}, err
	}

	if rows.Err() != nil {
		logger.Log.Error().Stack().Err(rows.Err()).Send()
		return []Assignment{}, err
	}
	defer rows.Close()

	as := make([]Assignment, 0, 10)
	for rows.Next() {
		a := Assignment{}
		err = rows.Scan(a)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return []Assignment{}, err
		}
		as = append(as, a)
	}
	return as, nil
}

// Get assignments for all chats due specified date
func (asr *AssignmentRepoData) GetAssignmentScheduleAllChats(
	ctx context.Context,
	due time.Time) ([]Assignment, error) {
	today := utils.GetToday()

	sql, params, err := goqu.From(assignmentsTableName).
		Select("id", "at", "chat_id", "operator").
		Where(goqu.Ex{
			"at": goqu.I("at").
				Between(
					exp.NewRangeVal(
						today.Format(utils.DateFormat),
						due.Format(utils.DateFormat),
					)),
		}).
		Order(goqu.I("at").Desc()).
		ToSQL()
	logger.Log.Debug().Str("sql", sql).Send()

	if err != nil {
		return nil, err
	}

	rows, err := asr.conn.Query(ctx, sql, params...)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return []Assignment{}, err
	}

	if rows.Err() != nil {
		logger.Log.Error().Stack().Err(rows.Err()).Send()
		return []Assignment{}, err
	}
	defer rows.Close()

	as := make([]Assignment, 0, 10)
	for rows.Next() {
		a := Assignment{}
		err = rows.Scan(a)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return []Assignment{}, err
		}
		as = append(as, a)
	}
	return as, nil
}

func (asr *AssignmentRepoData) GetAssignmentByDate(
	ctx context.Context,
	date time.Time,
	chatID int64) (Assignment, error) {

	a := Assignment{}
	sql, params, err := goqu.From(assignmentsTableName).
		Select(Assignment{}).
		Where(goqu.Ex{
			"at":      utils.GetDate(date).Format(utils.DateFormat),
			"chat_id": chatID,
		}).
		ToSQL()
	if err != nil {
		return Assignment{}, err
	}

	row := asr.conn.QueryRow(ctx, sql, params...)

	err = row.Scan(a)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return Assignment{}, nil
	case err != nil:
		return Assignment{}, err
	default:
	}
	return a, nil
}

// Return free duty slots for
// specified number of weeks
func (asr *AssignmentRepoData) GetFreeSlots(
	ctx context.Context,
	due time.Time,
	chatID int64) ([]time.Time, error) {

	today := utils.GetToday()
	dates, err := calendar.GetWorkingDays(today, due)
	if err != nil {
		return []time.Time{}, err
	}

	sql, params, err := goqu.From(assignmentsTableName).
		Select("at").
		Where(goqu.Ex{
			"chat_id": chatID,
			"at": goqu.I("at").
				Between(
					exp.NewRangeVal(
						today.Format(utils.DateFormat),
						due.Format(utils.DateFormat),
					)),
		}).
		ToSQL()

	rows, _ := asr.conn.Query(ctx, sql, params...)
	defer rows.Close()
	switch {
	case errors.Is(rows.Err(), pgx.ErrNoRows):
		return []time.Time{}, nil
	case rows.Err() != nil:
		return []time.Time{}, err
	default:
	}

	for rows.Next() {
		date := time.Time{}
		err = rows.Scan(&date)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return []time.Time{}, err
		}

		dates.Remove(date)
	}

	freedates := make([]time.Time, 0, 10)
	for freedate := range *dates {
		freedates = append(freedates, freedate)
	}

	sort.Slice(freedates, func(i int, j int) bool {
		return freedates[i].Before(freedates[j])
	})
	return freedates, nil
}

func (asr AssignmentRepoData) GetAllChats(ctx context.Context) ([]int64, error) {
	sql, params, err := goqu.From(assignmentsTableName).
		Select("chat_id").
		Distinct().
		ToSQL()

	if err != nil {
		return []int64{}, err
	}

	rows, _ := asr.conn.Query(ctx, sql, params...)
	defer rows.Close()
	switch {
	case errors.Is(rows.Err(), pgx.ErrNoRows):
		return []int64{}, nil
	case rows.Err() != nil:
		logger.Log.Error().Stack().Err(err).Send()
		return []int64{}, err
	default:
	}
	chats := make([]int64, 0, 10)
	for rows.Next() {
		var chatID int64
		err := rows.Scan(&chatID)
		if err != nil {
			logger.Log.Error().Stack().Err(err).Send()
			return []int64{}, err
		}
		chats = append(chats, chatID)
	}
	return chats, nil
}

func (asr AssignmentRepoData) Close(ctx context.Context) error {
	asr.conn.Close()
	return nil
}
