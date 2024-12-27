package repository_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/repository"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var AnyDate domain.DateOnly

func TestCreateTask(t *testing.T) {
	type args struct {
		ctx  context.Context
		task *domain.Task
	}

	tests := []struct {
		title        string
		args         args
		query        string
		setGetTxFunc func(*gorm.DB)
		wantError    error
	}{
		{
			"success",
			args{
				ctx: context.TODO(),
				task: &domain.Task{
					Title:       "test",
					Description: "test",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				},
			},
			`INSERT INTO "tasks" ("title","description","completed","created_by","due_date","created_at") VALUES ($1,$2,$3,$4,$5,$6)`,
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return tx, true
				}
			},
			nil,
		},
		{
			"create task failed",
			args{
				ctx: context.TODO(),
				task: &domain.Task{
					Title:       "test",
					Description: "test",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				},
			},
			`INSERT INTO "tasks" ("title","description","completed","created_by","due_date","created_at") VALUES ($1,$2,$3,$4,$5,$6)`,
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return tx, true
				}
			},
			myerror.ErrQueryFailed,
		},
		{
			"transaction not found",
			args{
				ctx: context.TODO(),
				task: &domain.Task{
					Title:       "test",
					Description: "test",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				},
			},
			`INSERT INTO "tasks" ("title","description","completed","created_by","due_date","created_at") VALUES ($1,$2,$3,$4,$5,$6)`,
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return nil, false
				}
			},
			myerror.ErrTransactionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			switch tt.wantError {
			case myerror.ErrTransactionNotFound:
			case myerror.ErrQueryFailed:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.task.Title, tt.args.task.Description, tt.args.task.Completed,
						tt.args.task.CreatedBy, tt.args.task.DueDate, helper.AnyTime{}).
					WillReturnError(tt.wantError)
				mock.ExpectRollback()
			default:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.task.Title, tt.args.task.Description, tt.args.task.Completed,
						tt.args.task.CreatedBy, tt.args.task.DueDate, helper.AnyTime{}).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			}

			// run
			tt.setGetTxFunc(db)
			r := repository.NewTaskRepository(db)
			id, err := r.Create(tt.args.ctx, tt.args.task)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, -1, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, id)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestFetchAllTaskByTaskID(t *testing.T) {
	// now := time.Now()
	type args struct {
		ctx     context.Context
		taskIDs []int
	}

	tests := []struct {
		title     string
		args      args
		query     string
		mockRow   [][]driver.Value
		wantTasks []domain.Task
		wantError error
	}{
		{
			"success",
			args{
				ctx:     context.TODO(),
				taskIDs: []int{1, 2},
			},
			`SELECT * FROM "tasks" WHERE id IN ($1,$2)`,
			[][]driver.Value{
				[]driver.Value{1, "test", "test", false, 1, AnyDate, time.Time{}},
				[]driver.Value{2, "test", "test", false, 1, AnyDate, time.Time{}},
			},
			[]domain.Task{
				{ID: 1, Title: "test", Description: "test", Completed: false, CreatedBy: 1, DueDate: AnyDate, CreatedAt: time.Time{}},
				{ID: 2, Title: "test", Description: "test", Completed: false, CreatedBy: 1, DueDate: AnyDate, CreatedAt: time.Time{}},
			},
			nil,
		},
		{
			"tasks not found",
			args{
				ctx:     context.TODO(),
				taskIDs: []int{1, 2},
			},
			`SELECT * FROM "tasks" WHERE id IN ($1,$2)`,
			nil,
			nil,
			myerror.ErrTaskNotFound,
		},
		{
			"query failed",
			args{
				ctx:     context.TODO(),
				taskIDs: []int{1, 2},
			},
			`SELECT * FROM "tasks" WHERE id IN ($1,$2)`,
			nil,
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			mock.MatchExpectationsInOrder(false)

			switch tt.wantError {
			case myerror.ErrTaskNotFound:
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(1, 2).
					WillReturnError(gorm.ErrRecordNotFound)
			case myerror.ErrQueryFailed:
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(1, 2).
					WillReturnError(fmt.Errorf("failed to fetch tasks"))
			default:
				rows := sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "due_date", "created_at"})
				for _, row := range tt.mockRow {
					rows.AddRow(row...)
				}
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(1, 2).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewTaskRepository(db)
			tasks, err := r.FetchAllTaskByTaskID(tt.args.ctx, tt.args.taskIDs...)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Nil(t, tasks)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTasks, tasks)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestFetchTaskByTaskID(t *testing.T) {
	type args struct {
		ctx    context.Context
		taskID int
	}

	tests := []struct {
		title     string
		args      args
		query     string
		mockRow   []driver.Value
		wantTask  *domain.Task
		wantError error
	}{
		{
			"success",
			args{
				ctx:    context.TODO(),
				taskID: 1,
			},
			`SELECT * FROM "tasks" WHERE id = $1 LIMIT $2`,
			[]driver.Value{1, "test", "test", false, 1, AnyDate, time.Time{}},
			&domain.Task{ID: 1, Title: "test", Description: "test", Completed: false, CreatedBy: 1, DueDate: AnyDate, CreatedAt: time.Time{}},
			nil,
		},
		{
			"task not found",
			args{
				ctx:    context.TODO(),
				taskID: 1,
			},
			`SELECT * FROM "tasks" WHERE id = $1 LIMIT $2`,
			nil,
			nil,
			myerror.ErrTaskNotFound,
		},
		{
			"query failed",
			args{
				ctx:    context.TODO(),
				taskID: 1,
			},
			`SELECT * FROM "tasks" WHERE id = $1 LIMIT $2`,
			nil,
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			mock.MatchExpectationsInOrder(false)

			switch tt.wantError {
			case myerror.ErrTaskNotFound:
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			case myerror.ErrQueryFailed:
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID, 1).
					WillReturnError(fmt.Errorf("failed to fetch task"))
			default:
				rows := sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "due_date", "created_at"}).
					AddRow(tt.mockRow...)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID, 1).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewTaskRepository(db)
			task, err := r.FetchTaskByTaskID(tt.args.ctx, tt.args.taskID)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Nil(t, task)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTask, task)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

	}
}

func TestUpdateTask(t *testing.T) {
	type args struct {
		ctx          context.Context
		taskID       int
		updateFields map[string]any
	}

	tests := []struct {
		title         string
		args          args
		query         string
		expectedError error
	}{
		{
			"update task successfully",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				updateFields: map[string]any{
					"title":       "test",
					"description": "test",
					"due_date":    AnyDate,
				},
			},
			`UPDATE "tasks" SET "description"=$1,"due_date"=$2,"title"=$3 WHERE id = $4`,
			nil,
		},
		{
			"update task failed",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				updateFields: map[string]any{
					"title":       "test",
					"description": "test",
					"due_date":    AnyDate,
				},
			},
			`UPDATE "tasks" SET "description"=$1,"due_date"=$2,"title"=$3 WHERE id = $4`,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			switch tt.expectedError {
			case myerror.ErrQueryFailed:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs("test", AnyDate, "test", 1).
					WillReturnError(fmt.Errorf("update task error"))
				mock.ExpectRollback()
			default:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs("test", AnyDate, "test", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewTaskRepository(db)
			err := r.Update(tt.args.ctx, tt.args.taskID, tt.args.updateFields)

			// assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

	}
}

func TestDeleteTask(t *testing.T) {
	type args struct {
		ctx    context.Context
		taskID int
	}

	tests := []struct {
		title     string
		args      args
		query     string
		wantError error
	}{
		{
			"success",
			args{
				ctx:    context.TODO(),
				taskID: 1,
			},
			`DELETE FROM "tasks" WHERE id = $1`,
			nil,
		},
		{
			"delete task failed",
			args{
				ctx:    context.TODO(),
				taskID: 1,
			},
			`DELETE FROM "tasks" WHERE id = $1`,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			switch tt.wantError {
			case myerror.ErrQueryFailed:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID).
					WillReturnError(fmt.Errorf("delete task error"))
				mock.ExpectRollback()
			default:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewTaskRepository(db)
			err := r.Delete(tt.args.ctx, tt.args.taskID)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
