package domain

import (
	"context"
	"time"
)

type User struct {
	ID        int       `json:"id"`         // ユーザーID
	Name      string    `json:"name"`       // 名前
	Email     string    `json:"email"`      // メールアドレス
	Password  string    `json:"password"`   // パスワード（ハッシュ化されていることを想定）
	CreatedAt time.Time `json:"created_at"` // 作成日時
}

type UserRepository interface {
	Create(context.Context, *User) error
	GetUser(context.Context) (User, error)
	GetAllUser(context.Context) ([]User, error)
	Delete(context.Context, int) error
}

type UserUsecases interface {
	Create(context.Context, *User) error
	GetUser(context.Context) (User, error)
	GetAllUser(context.Context) ([]User, error)
	Delete(context.Context, int) error
}
