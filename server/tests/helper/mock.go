package helper

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetMockUserRepository(t *testing.T) (*mock.MockUserRepository, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mock.NewMockUserRepository(ctrl), teardown
}

func GetDBMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	// SQLMockの初期化
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "failed to create SQL mock")

	// GORMの初期化
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err, "failed to open gorm DB connection")

	tearDown := func() {
		db.Close()
	}

	return gdb, mock, tearDown
}
