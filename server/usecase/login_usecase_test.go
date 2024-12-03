package usecases_test

import (
	"encoding/json"
	"testing"

	"net/http/httptest"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/usecases"
	"github.com/stretchr/testify/assert"
)

func TestCreateSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// ctrl := gomock.NewController(t)
	// defer ctrl.Finish()

	// モックセッションストアの作成
	store := cookie.NewStore([]byte("secret"))
	sessionName := "session"
	r := gin.Default()
	r.Use(sessions.Sessions(sessionName, store))

	// テスト用のリクエストとコンテキストの作成
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// セッションミドルウェアを適用
	r.Use(sessions.Sessions(sessionName, store))
	r.HandleContext(ctx)

	// テスト用のユーザー
	user := domain.User{
		ID:    1,
		Name:  "test user",
		Email: "test@example.com",
	}

	// テスト対象のUsecase
	loginUsecase := usecases.NewLoginUsecase(nil, 0)

	// テスト実行
	err := loginUsecase.CreateSession(ctx, user)

	// アサーション
	assert.NoError(t, err)

	// セッションに正しいユーザー情報が保存されているか確認
	session := sessions.Default(ctx)
	bUser, _ := json.Marshal(user)
	assert.Equal(t, string(bUser), session.Get("userInfo"))

}
