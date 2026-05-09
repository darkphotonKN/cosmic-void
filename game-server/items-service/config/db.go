package config

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

func InitDB() *sqlx.DB {
	var (
		host     = getEnvAsString("DB_HOST", "localhost")
		port     = getEnvAsInt("DB_PORT", 5432)
		user     = getEnvAsString("DB_USER", "postgres")
		password = getEnvAsString("DB_PASSWORD", "password")
		dbname   = getEnvAsString("DB_NAME", "items_db")
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	runMigrations(db.DB, dbname)

	slog.Info("Connected to database", slog.String("dbname", dbname))
	return db
}

func runMigrations(db *sql.DB, dbname string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		dbname,
		driver,
	)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Could not run migrations: %v", err)
	}

	slog.Info("Migrations completed successfully")
}

func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
