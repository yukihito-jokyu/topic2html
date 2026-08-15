package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/repository/postgres"
)

var (
	newPool        = pgxpool.New
	applyMigration = postgres.ApplyAdminAuthSchema
	closePool      = func(pool *pgxpool.Pool) { pool.Close() }
	printError     = log.Print
	exitProcess    = os.Exit
)

func main() {
	start()
}

func start() {
	if err := run(context.Background(), os.LookupEnv); err != nil {
		printError(err)
		exitProcess(1)
	}
}

func run(ctx context.Context, lookup func(string) (string, bool)) error {
	url, ok := lookup("TOPIC2HTML_DATABASE_URL")
	if !ok || url == "" {
		return errors.New("migration configuration is invalid")
	}
	pool, err := newPool(ctx, url)
	if err != nil {
		return errors.New("migration database connection is unavailable")
	}
	defer closePool(pool)
	if err := applyMigration(ctx, pool); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
