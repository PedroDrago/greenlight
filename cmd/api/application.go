package main

import (
	"database/sql"
	"flag"
	"os"

	"github.com/PedroDrago/greenlight/internal/data"
	"github.com/PedroDrago/greenlight/internal/data/jsonlog"
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
