package assignment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const assignmentsTableName = "assignments"

type AssignmentRepoData struct {
	conn *pgxpool.Pool
}

var AssignmentRepo AssignmentRepoer

type AssignmentRepoer interface {
	AddAssignment(ctx context.Context, as Assignment) error
	DeleteAssignment(ctx context.Context, id uuid.UUID) error
	GetAssignmentSchedule(ctx context.Context, due time.Time, chatID int64) ([]Assignment, error)
	GetAssignmentScheduleAllChats(ctx context.Context, due time.Time) ([]Assignment, error)
	GetAssignmentByDate(ctx context.Context, due time.Time, chatID int64) (Assignment, error)
	GetFreeSlots(ctx context.Context, due time.Time, chatID int64) ([]time.Time, error)
	GetAllChats(ctx context.Context) ([]int64, error)
	Close(ctx context.Context) error
}

type Assignment struct {
	ID uuid.UUID `db:"uuid"`
	// Assignment day
	At time.Time `db:"at"`
	// From which chat assignment came from
	ChatID int64 `db:"chat_id"`
	// Assignee for duty
	Operator string `db:"operator"`
	// When assignment was created
	CreatedAt time.Time `db:"created_at"`
}

func InitAssignmentRepo(ctx context.Context, dsn string) (AssignmentRepoer, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	result := &AssignmentRepoData{conn: conn}
	AssignmentRepo = result
	return result, nil
}
