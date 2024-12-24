package repository

import (
	"context"

	"github.com/keitatwr/task-management-app/internal/logger"
	"github.com/keitatwr/task-management-app/transaction"
	"gorm.io/gorm"
)

var txKey = struct{}{}

type tx struct {
	db *gorm.DB
}

func NewTransaction(db *gorm.DB) transaction.Transaction {
	return &tx{db: db}
}

func (t *tx) DoInTx(ctx context.Context, f func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	tx := t.db.WithContext(ctx).Begin()

	ctx = context.WithValue(ctx, &txKey, tx)

	v, err := f(ctx)

	if err != nil {
		tx.Rollback()
		logger.I(ctx, "rollback")
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.I(ctx, "rollback")
		return nil, err
	}
	logger.I(ctx, "commit")

	return v, nil
}

func GetTx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(&txKey).(*gorm.DB)
	return tx, ok
}
