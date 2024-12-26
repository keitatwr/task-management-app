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
	type args struct {
		user  *domain.User
		query string
	}
	tests := []struct {
		title     string
		args      args
		wantError error
	}{
		{
			"success",
			args{
				user: &domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
				query: `INSERT INTO "users" ("name","email","password","created_at") VALUES ($1,$2,$3,$4)`,
			},
			nil,
		},
		{
			"create user failed",
			args{
				user: &domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
				query: `INSERT INTO "users" ("name","email","password","created_at") VALUES ($1,$2,$3,$4)`,
			},
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			mock.MatchExpectationsInOrder(false)

			switch tt.wantError {
			case myerror.ErrQueryFailed:
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.user.Name, tt.args.user.Email, tt.args.user.Password, helper.AnyTime{}).
					WillReturnError(tt.wantError)
				mock.ExpectRollback()
			default:
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.user.Name, tt.args.user.Email, tt.args.user.Password, helper.AnyTime{}).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewUserReposiotry(db)
			err := r.Create(context.TODO(), tt.args.user)

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

func TestFetchUserByEmail(t *testing.T) {
	type args struct {
		email   string
		query   string
		mockRow []driver.Value
	}
	tests := []struct {
		title     string
		args      args
		wantUser  *domain.User
		wantError error
	}{
		{
			"success",
			args{
				email:   "test@example.com",
				query:   `SELECT * FROM "users" WHERE email = $1 LIMIT $2`,
				mockRow: []driver.Value{1, "test", "test@example.com", "hashedPassword", time.Time{}},
			},
			&domain.User{
				ID:        1,
				Name:      "test",
				Email:     "test@example.com",
				Password:  "hashedPassword",
				CreatedAt: time.Time{},
			},
			nil,
		},
		{
			"user not found",
			args{
				email: "test@example.com",
				query: `SELECT * FROM "users" WHERE email = $1 LIMIT $2`,
			},
			nil,
			myerror.ErrUserNotFound,
		},
		{
			"fetch user failed",
			args{
				email: "test@example.com",
				query: `SELECT * FROM "users" WHERE email = $1 LIMIT $2`,
			},
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			mock.MatchExpectationsInOrder(false)

			switch tt.wantError {
			case myerror.ErrUserNotFound:
				mock.ExpectQuery(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.email, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			case myerror.ErrQueryFailed:
				mock.ExpectQuery(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.email, 1).
					WillReturnError(fmt.Errorf("fetch user failed"))
			default:
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "created_at"}).AddRow(tt.args.mockRow...)
				mock.ExpectQuery(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.email, 1).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewUserReposiotry(db)
			user, err := r.FetchUserByEmail(context.TODO(), tt.args.email)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, user)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	type args struct {
		id    int
		query string
	}
	tests := []struct {
		title     string
		args      args
		wantError error
	}{
		{
			"success",
			args{
				id:    1,
				query: `DELETE FROM "users" WHERE id = $1`,
			},
			nil,
		},
		{
			"delete user failed",
			args{
				id:    1,
				query: `DELETE FROM "users" WHERE id = $1`,
			},
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := GetDbMock(t)
			defer tearDown()
			mock.MatchExpectationsInOrder(false)

			switch tt.wantError {
			case myerror.ErrQueryFailed:
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.id).
					WillReturnError(fmt.Errorf("delete user failed"))
				mock.ExpectRollback()
			default:
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.args.query)).
					WithArgs(tt.args.id).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewUserReposiotry(db)
			err := r.Delete(context.TODO(), tt.args.id)

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
