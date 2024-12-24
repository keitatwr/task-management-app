package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/route"
	"github.com/keitatwr/task-management-app/bootstrap"
	"github.com/keitatwr/task-management-app/internal/logger"
)

const logDir = "log"
const logFile = "log/app.log"
const hasStack = false

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

	logger.SetupLogger(logger.ModeDebug, hasStack, logfile)
	logger.AddKey("TraceID")
	logger.I(nil, "setup logger")

	logger.I(nil, "loading application...")
	app, err := bootstrap.App()
	if err != nil {
		logger.E(nil, "failed to loading application", err)
		logger.I(nil, "shutting down...")
		os.Exit(1)
	}
	env := app.Env
	db := app.Postgres
	timeout := time.Duration(env.ContextTimeout) * time.Second

	logger.I(nil, "set up server...")
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

		logger.I(nil, "Server is shutting down...")
		if err := server.Shutdown(ctx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.I(nil, "HTTP server Shutdown: timeout")
			} else {
				logger.W(nil, "HTTP server Shutdown", err)
			}
			close(idleConnsClosed)
			return
		}

		logger.I(nil, "Server is shut down")
		close(idleConnsClosed)
	}()

	logger.I(nil, fmt.Sprintf("Server is running on %s", env.ServerAddress))
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.E(nil, "HTTP server ListenAndServe", err)
		os.Exit(1)
	}

	// wait for idleConnsClosed to be closed
	<-idleConnsClosed
}
