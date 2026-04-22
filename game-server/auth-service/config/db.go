package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/constants"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/member"
	commonoutbox "github.com/darkphotonKN/cosmic-void-server/common/outbox"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// Importing for side effects - Dont Remove
	// This IS being used!
	_ "github.com/lib/pq"
)

/**
* Sets up the Database connection and provides its access as a singleton to
* the entire application.
**/
func InitDB() *sqlx.DB {
	// construct the db connection string
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// pass the db connection string to connect to our database
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	slog.Info("Connected to the database successfully")

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %v", err)
	}

	slog.Info("Successfully ran all migrations")
	return nil
}

func SeedDefaults(db *sqlx.DB) {
	// --- default members ---
	memberRepo := member.NewRepository(db)
	outboxRepo := commonoutbox.NewRepo(db)
	outboxService := commonoutbox.NewService(outboxRepo)
	memberService := member.NewService(db, memberRepo, nil, nil, outboxService)

	err := memberService.CreateDefaultMembers(constants.DefaultMembers)

	if err != nil {
		log.Fatal("Error when attempting to create default members:", err)
	}

	slog.Info("Successfully created all default members")

}
