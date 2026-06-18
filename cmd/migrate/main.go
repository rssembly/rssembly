package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dir, dbURL string
	flag.StringVar(&dir, "dir", "internal/database/migrations", "path to migration files")
	flag.StringVar(&dbURL, "db", "", "database URL (required)")
	flag.Parse()

	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "error: -db flag or DATABASE_URL env var is required")
		os.Exit(1)
	}

	m, err := migrate.New("file://"+dir, dbURL)
	if err != nil {
		log.Fatalf("migrate: new instance: %v", err)
	}
	defer m.Close()

	cmd := flag.Arg(0)
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migration up complete")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("migration down complete")
	case "drop":
		if err := m.Drop(); err != nil {
			log.Fatalf("migrate drop: %v", err)
		}
		fmt.Println("migration drop complete")
	default:
		fmt.Fprintf(os.Stderr, "usage: migrate [-dir <path>] -db <url> <up|down|drop>\n")
		os.Exit(1)
	}
}