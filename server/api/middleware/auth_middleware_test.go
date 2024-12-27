package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/middleware"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/stretchr/testify/assert"
)

func setup() *gin.Engine {
	// session store
	store := cookie.NewStore([]byte("secret"))

	// router
	r := gin.Default()
	// session middleware
	r.Use(sessions.Sessions("sessionid", store))

	// public route
	public := r.Group("/public")
	public.GET("", func(c *gin.Context) {
		user := domain.User{
			ID:       1,
			Name:     "test",
			Email:    "email",
			Password: "password",
		}
		session := sessions.Default(c)
		bUser, _ := json.Marshal(user)
		session.Set("userInfo", string(bUser))
		session.Save()
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// protected route
	protected := r.Group("/protected")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	return r

}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		title    string
		wantCode int
		wantBody interface{}
	}{
		{
			"success",
			http.StatusOK,
			`{"message":"success"}`,
		},
		{
			"authentication failed",
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "unauthorized",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeNoLogin),
						Message:     myerror.ErrMessages[myerror.CodeNoLogin],
						Description: "user not logged in",
					},
				},
			},
		},
	}

	gin.SetMode(gin.TestMode)

	r := setup()

	for _, tt := range tests {

		// test suite
		t.Run(tt.title, func(t *testing.T) {
			var sessionCookie string

			if tt.wantCode == http.StatusOK {
				// request to the public endpoint
				wPub := httptest.NewRecorder()
				reqPub, _ := http.NewRequest("GET", "/public", nil)
				reqPub.Header.Set("Content-Type", "application/json")
				r.ServeHTTP(wPub, reqPub)

				// Extract session cookie from the response
				sessionCookie = wPub.Header().Get("Set-Cookie")
				assert.NotEmpty(t, sessionCookie, "Session cookie should not be empty")
			}

			// request to the protected endpoint
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", "/protected", nil)
			req2.Header.Set("Content-Type", "application/json")
			if tt.wantCode == http.StatusOK {
				req2.Header.Set("Cookie", sessionCookie)
			}
			r.ServeHTTP(w2, req2)

			// Assert that the protected endpoint responds with success
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, tt.wantCode, w2.Code)
				assert.JSONEq(t, tt.wantBody.(string), w2.Body.String())
			} else {
				assert.Equal(t, tt.wantCode, w2.Code)
				helper.AssertResponse(t, tt.wantCode, tt.wantBody, w2)
			}
		})
	}
}
