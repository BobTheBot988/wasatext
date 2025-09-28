package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gitlab.com/mycompany8201046/myProject/service/api"
	"gitlab.com/mycompany8201046/myProject/service/api/globaltime"
	"gitlab.com/mycompany8201046/myProject/service/database"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
	}
}

func run() error {

	rand.Seed(globaltime.Now().UnixNano())
	// Load Configuration and defaults
	cfg, err := loadConfiguration()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return err
	}

	// Init logging
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Infof("application initializing")

	// Start Database
	logger.Println("initializing database support")
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename+"?_journal_mode=WAL&_busy_timeout=10000&_synchronous=NORMAL")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = dbconn.Close()
	}()
	db, err := database.New(dbconn)
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return fmt.Errorf("creating AppDatabase: %w", err)
	}
	// Define which tables/columns to validate
	validations := map[string][]string{
		"User":    {"userName"},
		"Message": {"content"},
		"Comment": {"content"},
		"GroupTB": {"Name", "Description"},
	}

	// Add the hook
	db.AddPreCommitHook(db.CreateSanitizeHook(validations))

	// Start (main) API server
	logger.Info("initializing API server")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Create the API router
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: db,
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("creating the API server instance: %w", err)
	}
	router := apirouter.Handler()

	router, err = registerWebUI(router)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return fmt.Errorf("registering web UI handler: %w", err)
	}

	// Apply CORS policy
	router = applyCORSHandler(router)

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests in a separate goroutine
	go func() {
		logger.Infof("API listening on %s", apiserver.Addr)
		serverErrors <- apiserver.ListenAndServe()
		logger.Infof("stopping API server")
	}()

	// Waiting for shutdown signal or POSIX signals
	select {
	case err = <-serverErrors:
		// Non-recoverable server error
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infof("signal %v received, start shutdown", sig)

		// Asking API server to shut down and load shed.
		err = apirouter.Close()
		if err != nil {
			logger.WithError(err).Warning("graceful shutdown of apirouter error")
		}

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and load shed.
		err = apiserver.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Warning("error during graceful shutdown of HTTP server")
			err = apiserver.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
