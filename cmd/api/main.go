package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PedroDrago/greenlight/internal/data"
	"github.com/PedroDrago/greenlight/internal/data/jsonlog"
	_ "github.com/lib/pq"
)

const (
	version = "1.0.0"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	config config
	models data.Models
	logger *jsonlog.Logger
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB"), "DSN for connection with PostgreSQL")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL Max Open Connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL Max Idle Connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL Max Connection Idle Time")
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (dev | prod | stage)")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()
}

// NOTE:  last page: 135 - Chapter 7 - CRUD Operations
//  FIX:  last page: 135 - Chapter 7 - CRUD Operations
// TEST:  last page: 135 - Chapter 7 - CRUD Operations
// WARN:  last page: 135 - Chapter 7 - CRUD Operations
// TODO:  last page: 135 - Chapter 7 - CRUD Operations

func newApplication() (*application, *sql.DB) {
	var cfg config
	parseFlags(&cfg)
	app := application{
		config: cfg,
		logger: jsonlog.New(os.Stdout, jsonlog.LevelInfo),
	}
	db, err := openDB(cfg)
	if err != nil {
		app.logger.Fatal(err, nil)
	}
	app.models = data.NewModels(db)
	return &app, db
}

func newServer(app *application) *http.Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(app.logger, "", 0),
	}
	return srv
}

func main() {
	app, db := newApplication()
	defer db.Close()
	srv := newServer(app)
	app.logger.Info("Connection with DB established", nil)
	app.logger.Info("Starting %s server on %s", map[string]string{"addr": srv.Addr, "env": app.config.env})
	err := srv.ListenAndServe()
	app.logger.Fatal(err, nil)
}
