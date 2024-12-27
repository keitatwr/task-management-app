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

func TestCreateUser(t *testing.T) {
	type args struct {
		ctx  context.Context
		user *domain.User
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
				ctx: context.TODO(),
				user: &domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
			},
			`INSERT INTO "users" ("name","email","password","created_at") VALUES ($1,$2,$3,$4)`,
			nil,
		},
		{
			"create user failed",
			args{
				ctx: context.TODO(),
				user: &domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
			},
			`INSERT INTO "users" ("name","email","password","created_at") VALUES ($1,$2,$3,$4)`,
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
			case myerror.ErrQueryFailed:
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.user.Name, tt.args.user.Email, tt.args.user.Password, helper.AnyTime{}).
					WillReturnError(tt.wantError)
				mock.ExpectRollback()
			default:
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.user.Name, tt.args.user.Email, tt.args.user.Password, helper.AnyTime{}).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewUserReposiotry(db)
			err := r.Create(tt.args.ctx, tt.args.user)

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
		ctx   context.Context
		email string
	}

	tests := []struct {
		title     string
		args      args
		query     string
		mockRow   []driver.Value
		wantUser  *domain.User
		wantError error
	}{
		{
			"success",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			`SELECT * FROM "users" WHERE email = $1 LIMIT $2`,
			[]driver.Value{1, "test", "test@example.com", "hashedPassword", time.Time{}},
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
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			`SELECT * FROM "users" WHERE email = $1 LIMIT $2`,
			nil,
			nil,
			myerror.ErrUserNotFound,
		},
		{
			"fetch user failed",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			`SELECT * FROM "users" WHERE email = $1 LIMIT $2`,
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
			case myerror.ErrUserNotFound:
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.email, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			case myerror.ErrQueryFailed:
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.email, 1).
					WillReturnError(fmt.Errorf("fetch user failed"))
			default:
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "created_at"}).AddRow(tt.mockRow...)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.email, 1).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewUserReposiotry(db)
			user, err := r.FetchUserByEmail(tt.args.ctx, tt.args.email)

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
		ctx context.Context
		id  int
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
				ctx: context.TODO(),
				id:  1,
			},
			`DELETE FROM "users" WHERE id = $1`,
			nil,
		},
		{
			"delete user failed",
			args{
				ctx: context.TODO(),
				id:  1,
			},
			`DELETE FROM "users" WHERE id = $1`,
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
			case myerror.ErrQueryFailed:
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.id).
					WillReturnError(fmt.Errorf("delete user failed"))
				mock.ExpectRollback()
			default:
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.id).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			// run
			r := repository.NewUserReposiotry(db)
			err := r.Delete(tt.args.ctx, tt.args.id)

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
