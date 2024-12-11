package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/api/route"
	"github.com/keitatwr/todo-app/bootstrap"
	"github.com/keitatwr/todo-app/internal/logger"
)

const logDir = "log"
const logFile = "log/app.log"

func main() {
	// setup logger
	root, _ := os.Getwd()
	os.Mkdir(fmt.Sprintf("%s/%s", root, logDir), os.ModePerm)

	logfile, err := os.OpenFile(fmt.Sprintf("%s/%s", root, logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	defer func() {
		if err := logfile.Close(); err != nil {
			log.Fatalf("failed to close log file: %v", err)
		}
	}()

	customLogger := logger.NewLogger(logger.ModeDebug, os.Stdout)
	slog.SetDefault(customLogger)
	logger.Info(nil, "setup logger")

	logger.Info(nil, "loading application...")
	app, err := bootstrap.App()
	if err != nil {
		logger.Errorf(nil, "failed to loading application: %v", err)
		logger.Info(nil, "shutting down...")
		os.Exit(1)
	}
	env := app.Env
	db := app.Postgres
	timeout := time.Duration(env.ContextTimeout) * time.Second

	logger.Info(nil, "set up server...")
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	route.Setup(timeout, db, router)
	server := &http.Server{
		Addr:    env.ServerAddress,
		Handler: router,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		// make chan to receive signal
		c := make(chan os.Signal, 1)
		// notify signal to chan and waiting for signal
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger.Info(nil, "Server is shutting down...")
		if err := server.Shutdown(ctx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.Info(nil, "HTTP server Shutdown: timeout")
			} else {
				logger.Infof(nil, "HTTP server Shutdown: %v", err)
			}
			close(idleConnsClosed)
			return
		}

		logger.Info(nil, "Server is shut down")
		close(idleConnsClosed)
	}()

	logger.Infof(nil, "Server is running on %s", env.ServerAddress)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf(nil, "HTTP server ListenAndServe: %v", err)
		os.Exit(1)
	}

	// wait for idleConnsClosed to be closed
	<-idleConnsClosed
}
