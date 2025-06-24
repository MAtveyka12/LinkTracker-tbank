package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	// Register PostgreSQL driver.
	_ "github.com/lib/pq"

	"github.com/pressly/goose/v3"
)

func RunMigrations(dsn, migrationsPath string) error {
	log.Printf("Starting migrations with DSN: %s", dsn)
	log.Printf("Migrations path: %s", migrationsPath)

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Printf("Failed to parse DSN: %v", err)
		return fmt.Errorf("parse dsn: %w", err)
	}

	log.Println("DSN parsed successfully")

	db := stdlib.OpenDB(*config)

	defer func() {
		db.Close()
		log.Println("Database connection closed")
	}()

	log.Println("Database connection established")

	goose.SetLogger(log.New(os.Stdout, "[goose] ", log.LstdFlags))
	log.Println("Goose logger configured")

	if err := goose.SetDialect("postgres"); err != nil {
		log.Printf("Failed to set dialect to postgres: %v", err)
		return fmt.Errorf("set dialect: %w", err)
	}

	log.Println("Database dialect set to postgres")
	log.Printf("Running migrations from: %s", migrationsPath)

	if err := goose.Up(db, migrationsPath); err != nil {
		log.Printf("Migration failed: %v", err)
		return fmt.Errorf("goose.Up: %w", err)
	}

	log.Println("Migrations completed successfully")

	return nil
}
