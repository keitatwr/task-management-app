package repository_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getDbMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	// SQLMockの初期化
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to create SQL mock")

	// GORMの初期化
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err, "failed to open gorm DB connection")

	tearDown := func() {
		db.Close()
	}

	return gdb, mock, tearDown
}

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

var AnyDate domain.DateOnly

func TestCreateTask(t *testing.T) {
	tests := []struct {
		title        string
		setGetTxFunc func(*gorm.DB)
		task         *domain.Task
		query        string
		wantError    error
	}{
		{
			"success",
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return tx, true
				}
			},
			&domain.Task{
				Title:       "test",
				Description: "test",
				Completed:   false,
				CreatedBy:   1,
				DueDate:     AnyDate,
			},
			`INSERT INTO "tasks" ("title","description","completed","created_by","due_date","created_at") VALUES ($1,$2,$3,$4,$5,$6)`,
			nil,
		},
		{
			"create task failed",
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return tx, true
				}
			},
			&domain.Task{
				Title:       "test",
				Description: "test",
				Completed:   false,
				CreatedBy:   1,
				DueDate:     AnyDate,
			},
			`INSERT INTO "tasks" ("title","description","completed","created_by","due_date","created_at") VALUES ($1,$2,$3,$4,$5,$6)`,
			myerror.ErrQueryFailed,
		},
		{
			"transaction not found",
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return nil, false
				}
			},
			&domain.Task{
				Title:       "test",
				Description: "test",
				Completed:   false,
				CreatedBy:   1,
				DueDate:     AnyDate,
			},
			`INSERT INTO "tasks" ("title","description","completed","created_by","due_date","created_at") VALUES ($1,$2,$3,$4,$5,$6)`,
			myerror.ErrTransactionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := getDbMock(t)
			defer tearDown()

			if tt.wantError != myerror.ErrTransactionNotFound {
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				if tt.wantError != nil {
					mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
						WithArgs(tt.task.Title, tt.task.Description, tt.task.Completed,
							tt.task.CreatedBy, tt.task.DueDate, AnyTime{}).
						WillReturnError(tt.wantError)
					mock.ExpectRollback()

				} else {
					mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
						WithArgs(tt.task.Title, tt.task.Description, tt.task.Completed,
							tt.task.CreatedBy, tt.task.DueDate, AnyTime{}).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
					mock.ExpectCommit()
				}
			}

			// run
			ctx := context.TODO()
			tt.setGetTxFunc(db)
			r := repository.NewTaskRepository(db)
			id, err := r.Create(ctx, tt.task)

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

// func TestGetTodoByID(t *testing.T) {
// 	now := time.Now()
// 	coloumns := []string{"id", "title", "description", "completed",
// 		"user_id", "created_at", "updated_at"}
// 	tests := []struct {
// 		title         string
// 		id            int
// 		query         string
// 		mockRow       []driver.Value
// 		expected      *domain.Task
// 		expectedError bool
// 	}{
// 		{
// 			"get todo by id successfully",
// 			1,
// 			`SELECT * FROM "todos" WHERE id = $1 LIMIT $2`,
// 			[]driver.Value{1, "test", "test", false, 1, now, now},
// 			&domain.Task{
// 				ID:          1,
// 				Title:       "test",
// 				Description: "test",
// 				Completed:   false,
// 				UserID:      1,
// 				CreatedAt:   now,
// 				UpdatedAt:   now,
// 			},
// 			false,
// 		},
// 		{
// 			"get todo by id with error",
// 			1,
// 			`SELECT * FROM "todos" WHERE id = $1 LIMIT $2`,
// 			nil,
// 			nil,
// 			true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.title, func(t *testing.T) {
// 			// mock
// 			db, mock, tearDown := GetDbMock(t)
// 			defer tearDown()
// 			mock.MatchExpectationsInOrder(false)
// 			if tt.expectedError {
// 				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
// 					WithArgs(tt.id, 1).
// 					WillReturnError(fmt.Errorf("get todo error"))
// 			} else {
// 				rows := sqlmock.NewRows(coloumns).AddRow(tt.mockRow...)
// 				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
// 					WithArgs(tt.id, 1).
// 					WillReturnRows(rows)
// 			}

// 			// run
// 			r := repository.NewTodoRepository(db)
// 			todo, err := r.GetTodoByID(context.TODO(), tt.id)

// 			// assert
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 				assert.Equal(t, tt.expected, todo)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.expected, todo)
// 			}

// 			if err := mock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	}
// }

// func TestGetAllTodoByUserID(t *testing.T) {
// 	now := time.Now()
// 	columns := []string{"id", "title", "description", "completed",
// 		"user_id", "created_at", "updated_at"}
// 	tests := []struct {
// 		title         string
// 		id            int
// 		query         string
// 		mockRow       [][]driver.Value
// 		expected      []domain.Task
// 		expectedError bool
// 	}{
// 		{
// 			"get all todo by user id successfully",
// 			1,
// 			`SELECT * FROM "todos" WHERE user_id = $1`,
// 			[][]driver.Value{
// 				[]driver.Value{1, "test", "test", false, 1, now, now},
// 				[]driver.Value{2, "test", "test", false, 1, now, now},
// 			},
// 			[]domain.Task{
// 				{
// 					ID:          1,
// 					Title:       "test",
// 					Description: "test",
// 					Completed:   false,
// 					UserID:      1,
// 					CreatedAt:   now,
// 					UpdatedAt:   now,
// 				},
// 				{
// 					ID:          2,
// 					Title:       "test",
// 					Description: "test",
// 					Completed:   false,
// 					UserID:      1,
// 					CreatedAt:   now,
// 					UpdatedAt:   now,
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"get all todo by user id with error",
// 			1,
// 			`SELECT * FROM "todos" WHERE user_id = $1`,
// 			nil,
// 			nil,
// 			true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.title, func(t *testing.T) {
// 			// mock
// 			db, mock, tearDown := GetDbMock(t)
// 			defer tearDown()
// 			mock.MatchExpectationsInOrder(false)
// 			if tt.expectedError {
// 				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
// 					WithArgs(tt.id).
// 					WillReturnError(fmt.Errorf("get all todo error"))
// 			} else {
// 				rows := sqlmock.NewRows(columns)
// 				for _, row := range tt.mockRow {
// 					rows.AddRow(row...)
// 				}
// 				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
// 					WithArgs(tt.id).
// 					WillReturnRows(rows)
// 			}

// 			// run
// 			r := repository.NewTodoRepository(db)
// 			todos, err := r.GetAllTodoByUserID(context.TODO(), tt.id)

// 			// assert
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 				assert.Equal(t, tt.expected, todos)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.expected, todos)
// 			}

// 			if err := mock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	}
// }

// func TestUpdateTodo(t *testing.T) {
// 	tests := []struct {
// 		title         string
// 		id            int
// 		expectedError bool
// 	}{
// 		{
// 			"update todo successfully",
// 			1,
// 			false,
// 		},
// 		{
// 			"update todo with error",
// 			1,
// 			true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.title, func(t *testing.T) {
// 			// mock
// 			db, mock, tearDown := GetDbMock(t)
// 			defer tearDown()
// 			mock.MatchExpectationsInOrder(false)
// 			mock.ExpectBegin()
// 			if tt.expectedError {
// 				// regexp.QuoteMetaを使うと._+などのメタ文字がエスケープされるため、
// 				// テストが期待通りに動作しない
// 				mock.ExpectExec("UPDATE \"todos\" SET .+").
// 					WillReturnError(fmt.Errorf("update todo error"))
// 				mock.ExpectRollback()
// 			} else {
// 				mock.ExpectExec("UPDATE \"todos\" SET .+").
// 					WillReturnResult(sqlmock.NewResult(1, 1))
// 				mock.ExpectCommit()
// 			}

// 			// run
// 			r := repository.NewTodoRepository(db)
// 			err := r.Update(context.TODO(), tt.id)

// 			// assert
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}

// 			if err := mock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	}
// }

// func TestDeleteTodo(t *testing.T) {
// 	tests := []struct {
// 		title         string
// 		id            int
// 		query         string
// 		expectedError bool
// 	}{
// 		{
// 			"delete todo successfully",
// 			1,
// 			`DELETE FROM "todos" WHERE id = $1`,
// 			false,
// 		},
// 		{
// 			"delete todo with error",
// 			1,
// 			`DELETE FROM "todos" WHERE id = $1`,
// 			true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.title, func(t *testing.T) {
// 			// mock
// 			db, mock, tearDown := GetDbMock(t)
// 			defer tearDown()
// 			mock.MatchExpectationsInOrder(false)
// 			mock.ExpectBegin()
// 			if tt.expectedError {
// 				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
// 					WithArgs(tt.id).
// 					WillReturnError(fmt.Errorf("delete error"))
// 				mock.ExpectRollback()
// 			} else {
// 				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
// 					WithArgs(tt.id).
// 					WillReturnResult(sqlmock.NewResult(1, 1))
// 				mock.ExpectCommit()
// 			}

// 			// run
// 			r := repository.NewTodoRepository(db)
// 			err := r.Delete(context.TODO(), tt.id)

// 			// assert
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}

// 			if err := mock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	}
// }
