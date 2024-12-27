package repository_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/repository"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGrantPermission(t *testing.T) {
	type args struct {
		ctx            context.Context
		taskPermission *domain.TaskPermission
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
				taskPermission: &domain.TaskPermission{
					TaskID:  1,
					UserID:  1,
					CanEdit: true,
					CanRead: true,
				},
			},
			`INSERT INTO "task_permissions" ("task_id","user_id","can_edit","can_read") VALUES ($1,$2,$3,$4)`,
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return tx, true
				}
			},
			nil,
		},
		{
			"failed to get transaction",
			args{
				ctx: context.TODO(),
				taskPermission: &domain.TaskPermission{
					TaskID:  1,
					UserID:  1,
					CanEdit: true,
					CanRead: true,
				},
			},
			"",
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return nil, false
				}
			},
			myerror.ErrTransactionNotFound,
		},
		{
			"failed to grant permission",
			args{
				ctx: context.TODO(),
				taskPermission: &domain.TaskPermission{
					TaskID:  1,
					UserID:  1,
					CanEdit: true,
					CanRead: true,
				},
			},
			`INSERT INTO "task_permissions" ("task_id","user_id","can_edit","can_read") VALUES ($1,$2,$3,$4)`,
			func(tx *gorm.DB) {
				repository.GetTxFunc = func(ctx context.Context) (*gorm.DB, bool) {
					return tx, true
				}
			},
			myerror.ErrGrantPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			switch tt.wantError {
			case myerror.ErrTransactionNotFound:
			case myerror.ErrGrantPermission:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskPermission.TaskID, tt.args.taskPermission.UserID, tt.args.taskPermission.CanEdit, tt.args.taskPermission.CanRead).
					WillReturnError(tt.wantError)
				mock.ExpectRollback()
			default:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskPermission.TaskID, tt.args.taskPermission.UserID, tt.args.taskPermission.CanEdit, tt.args.taskPermission.CanRead).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			}

			// run
			tt.setGetTxFunc(db)
			r := repository.NewTaskPermissionRepository(db)
			err := r.GrantPermission(tt.args.ctx, tt.args.taskPermission)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestFetchTaskIDByUserID(t *testing.T) {
	type args struct {
		ctx     context.Context
		userID  int
		canEdit bool
		canRead bool
	}

	tests := []struct {
		title       string
		args        args
		query       string
		wantTaskIDs []int
		wantError   error
	}{
		{
			"success",
			args{
				ctx:     context.TODO(),
				userID:  1,
				canEdit: true,
				canRead: true,
			},
			`SELECT "task_id" FROM "task_permissions" WHERE user_id = $1 AND (can_edit = $2 AND can_read = $3)`,
			[]int{1, 2},
			nil,
		},
		{
			"permission not found",
			args{
				ctx:     context.TODO(),
				userID:  1,
				canEdit: true,
				canRead: true,
			},
			`SELECT "task_id" FROM "task_permissions" WHERE user_id = $1 AND (can_edit = $2 AND can_read = $3)`,
			nil,
			myerror.ErrPermissionNotFound,
		},
		{
			"failed to fetch task",
			args{
				ctx:     context.TODO(),
				userID:  1,
				canEdit: true,
				canRead: true,
			},
			`SELECT "task_id" FROM "task_permissions" WHERE user_id = $1 AND (can_edit = $2 AND can_read = $3)`,
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			switch tt.wantError {
			case myerror.ErrPermissionNotFound:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.userID, tt.args.canEdit, tt.args.canRead).
					WillReturnError(gorm.ErrRecordNotFound)

			case myerror.ErrQueryFailed:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.userID, tt.args.canEdit, tt.args.canRead).
					WillReturnError(fmt.Errorf("fetch task failed"))
			default:
				mock.MatchExpectationsInOrder(false)
				rows := sqlmock.NewRows([]string{"task_id"})
				for _, id := range tt.wantTaskIDs {
					rows.AddRow(id)
				}
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.userID, tt.args.canEdit, tt.args.canRead).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewTaskPermissionRepository(db)
			taskIDs, err := r.FetchTaskIDByUserID(tt.args.ctx, tt.args.userID, tt.args.canEdit, tt.args.canRead)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTaskIDs, taskIDs)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestFetchPermissionByTaskID(t *testing.T) {
	type args struct {
		ctx    context.Context
		taskID int
		userID int
	}

	tests := []struct {
		title              string
		args               args
		query              string
		wantTaskPermission *domain.TaskPermission
		wantError          error
	}{
		{
			"success",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			`SELECT * FROM "task_permissions" WHERE task_id = $1 AND user_id = $2 LIMIT $3`,
			&domain.TaskPermission{
				TaskID:  1,
				UserID:  1,
				CanEdit: true,
				CanRead: true,
			},
			nil,
		},
		{
			"permission not found",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			`SELECT * FROM "task_permissions" WHERE task_id = $1 AND user_id = $2 LIMIT $3`,
			nil,
			myerror.ErrPermissionNotFound,
		},
		{
			"failed to fetch task",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			`SELECT * FROM "task_permissions" WHERE task_id = $1 AND user_id = $2 LIMIT $3`,
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			db, mock, tearDown := helper.GetDBMock(t)
			defer tearDown()

			switch tt.wantError {
			case myerror.ErrPermissionNotFound:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID, tt.args.userID).
					WillReturnError(gorm.ErrRecordNotFound)

			case myerror.ErrQueryFailed:
				mock.MatchExpectationsInOrder(false)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID, tt.args.userID).
					WillReturnError(fmt.Errorf("fetch task failed"))

			default:
				mock.MatchExpectationsInOrder(false)
				rows := sqlmock.NewRows([]string{"task_id", "user_id", "can_edit", "can_read"}).
					AddRow(tt.wantTaskPermission.TaskID, tt.wantTaskPermission.UserID, tt.wantTaskPermission.CanEdit, tt.wantTaskPermission.CanRead)
				mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
					WithArgs(tt.args.taskID, tt.args.userID, 1).
					WillReturnRows(rows)
			}

			// run
			r := repository.NewTaskPermissionRepository(db)
			taskPermission, err := r.FetchPermissionByTaskID(tt.args.ctx, tt.args.taskID, tt.args.userID)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTaskPermission, taskPermission)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
