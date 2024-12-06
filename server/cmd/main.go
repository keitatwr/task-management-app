package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/api/route"
	"github.com/keitatwr/todo-app/bootstrap"
)

func main() {
	app := bootstrap.App()

	env := app.Env

	db := app.Postgres

	timeout := time.Duration(env.ContextTimeout) * time.Second

	gin.SetMode(gin.ReleaseMode)
	gin := gin.Default()

	route.Setup(timeout, db, gin)

	gin.Run(":8080")
}
