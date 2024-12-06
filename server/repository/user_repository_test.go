package repository_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDbMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

func TestCreateUser(t *testing.T) {
	now := time.Now()
	tests := []struct {
		title         string
		user          *domain.User
		query         string
		expectedError bool
	}{
		{
			"create a user successfully",
			&domain.User{
				Name:      "sample name",
				Email:     "test@test.co.jp",
				Password:  "secret",
				CreatedAt: now,
			},
			`INSERT INTO "users" ("name","email","password","created_at") VALUES ($1,$2,$3,$4)`,
			false,
		},
		{
			"create a user with error",
			&domain.User{
				Name:      "sample name",
				Email:     "test@test.co.jp",
				Password:  "secret",
				CreatedAt: now,
			},
			`INSERT INTO "users" ("name","email","password","created_at") VALUES ($1,$2,$3,$4)`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			mock.MatchExpectationsInOrder(false)
			mock.ExpectBegin()
			if tt.expectedError {
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.user.Name, tt.user.Email, tt.user.Password, tt.user.CreatedAt).
					WillReturnError(fmt.Errorf("create user error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.user.Name, tt.user.Email, tt.user.Password, tt.user.CreatedAt).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewUserReposiotry(db)
			err := r.Create(context.TODO(), tt.user)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	now := time.Now()
	columns := []string{"id", "name", "email", "password", "created_at"}
	tests := []struct {
		title         string
		id            int
		query         string
		mockRow       []driver.Value
		expected      *domain.User
		expectedError bool
	}{
		{
			"get a user successfully",
			1,
			`SELECT * FROM "users" WHERE id = $1 LIMIT $2`,
			[]driver.Value{1, "sample name", "test@test.co.jp", "secret", now},
			&domain.User{
				ID:        1,
				Name:      "sample name",
				Email:     "test@test.co.jp",
				Password:  "secret",
				CreatedAt: now,
			},
			false,
		},
		{
			"get a user with error",
			1,
			`SELECT * FROM "users" WHERE id = $1 LIMIT $2`,
			[]driver.Value{1, "sample name", "test@test.co.jp", "secret", now},
			&domain.User{
				ID:        1,
				Name:      "sample name",
				Email:     "test@test.co.jp",
				Password:  "secret",
				CreatedAt: now,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			mock.MatchExpectationsInOrder(false)
			if tt.expectedError {
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.id, 1).
					WillReturnError(fmt.Errorf("get user error"))
			} else {
				rows := sqlmock.NewRows(columns).AddRow(tt.mockRow...)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.id, 1).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewUserReposiotry(db)
			actual, err := r.GetUserByID(context.TODO(), tt.id)

			// assert

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetAllUser(t *testing.T) {
	now := time.Now()
	columns := []string{"id", "name", "email", "password", "created_at"} // 期待するカラム数

	tests := []struct {
		title         string
		query         string
		mockRows      [][]driver.Value
		expected      []domain.User
		expectedError bool
	}{
		{
			"get all users successfully",
			`SELECT * FROM "users"`,
			[][]driver.Value{
				[]driver.Value{1, "sample name 1", "test1@test.co.jp", "secret1", now},
				[]driver.Value{2, "sample name 2", "test2@test.co.jp", "secret2", now},
			},
			[]domain.User{
				{
					ID:        1,
					Name:      "sample name 1",
					Email:     "test1@test.co.jp",
					Password:  "secret1",
					CreatedAt: now,
				},
				{
					ID:        2,
					Name:      "sample name 2",
					Email:     "test2@test.co.jp",
					Password:  "secret2",
					CreatedAt: now,
				},
			},
			false,
		},
		{
			"get all users with error",
			`SELECT * FROM "users"`,
			[][]driver.Value{
				[]driver.Value{1, "sample name 1", "test1@test.co.jp", "secret1", now},
				[]driver.Value{2, "sample name 2", "test2@test.co.jp", "secret2", now},
			},
			[]domain.User{
				{
					ID:        1,
					Name:      "sample name 1",
					Email:     "test1@test.co.jp",
					Password:  "secret1",
					CreatedAt: now,
				},
				{
					ID:        2,
					Name:      "sample name 2",
					Email:     "test2@test.co.jp",
					Password:  "secret2",
					CreatedAt: now,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			if tt.expectedError {
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WillReturnError(fmt.Errorf("get all user error"))
			} else {
				rows := sqlmock.NewRows(columns)
				for _, row := range tt.mockRows {
					rows.AddRow(row...)
				}
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewUserReposiotry(db)
			actual, err := r.GetAllUser(context.TODO())

			// assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		title         string
		id            int
		query         string
		expectedError bool
	}{
		{
			"delete a user successfully",
			1,
			`DELETE FROM "users" WHERE id = $1`,
			false,
		},
		{
			"delete a user with error",
			1,
			`DELETE FROM "users" WHERE id = $1`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			mock.ExpectBegin()
			if tt.expectedError {
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.id).
					WillReturnError(fmt.Errorf("delete error"))
				// mock.ExpectCommit()
			} else {
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.id).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewUserReposiotry(db)
			err := r.Delete(context.TODO(), tt.id)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
