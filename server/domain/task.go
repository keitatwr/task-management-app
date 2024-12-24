package domain

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedBy   int       `json:"createdBy"`
	DueDate     DateOnly  `json:"dueDate"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DateOnly struct {
	time.Time
}

func NewDateOnly(t string) DateOnly {
	parseTime, _ := time.Parse("2006-01-02", t)
	return DateOnly{Time: parseTime}
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	parsedTime, err := time.Parse(`"2006-01-02"`, string(b))
	if err != nil {
		return &time.ParseError{Value: string(b), Layout: "2006-01-02", LayoutElem: "2006", ValueElem: string(b)}
	}
	d.Time = parsedTime
	return nil
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format("2006-01-02"))), nil
}

func (d *DateOnly) Scan(value interface{}) error {
	d.Time = value.(time.Time)
	return nil
}

func (d DateOnly) Value() (driver.Value, error) {
	return d.Time, nil
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) (int, error)
	FetchAllTaskByTaskID(ctx context.Context, taskIDs ...int) ([]Task, error)
	FetchTaskByTaskID(ctx context.Context, taskID int) (*Task, error)
	Update(ctx context.Context, task *Task) error
	// GetAllTaskByUserID(ctx context.Context, id int) ([]Task, error)
	// Update(ctx context.Context, id int) error
	// Delete(ctx context.Context, id int) error
}

type TaskUsecase interface {
	Create(ctx context.Context, title string, description string, userID int, due_date DateOnly) error
	FetchAllTaskByUserID(ctx context.Context, userID int) ([]Task, error)
	FetchTaskByTaskID(ctx context.Context, taskID, userID int) (*Task, error)
	Update(ctx context.Context, taskID, userID int, title, description string, due_date DateOnly) error
	// GetAllTaskByUserID(ctx context.Context, id int) ([]Task, error)
	// Update(ctx context.Context, id int) error
	// Delete(ctx context.Context, id int) error
}
