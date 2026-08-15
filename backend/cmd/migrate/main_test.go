package main

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRun(t *testing.T) {
	originalNewPool, originalApply, originalClose := newPool, applyMigration, closePool
	t.Cleanup(func() { newPool, applyMigration, closePool = originalNewPool, originalApply, originalClose })
	lookup := func(string) (string, bool) { return "postgres://user:password@localhost:5432/topic2html", true }
	if err := run(context.Background(), func(string) (string, bool) { return "", false }); err == nil {
		t.Fatal("missing config succeeded")
	}
	newPool = func(context.Context, string) (*pgxpool.Pool, error) { return nil, errors.New("unavailable") }
	if err := run(context.Background(), lookup); err == nil {
		t.Fatal("connection succeeded")
	}
	newPool = func(context.Context, string) (*pgxpool.Pool, error) { return &pgxpool.Pool{}, nil }
	closePool = func(*pgxpool.Pool) {}
	applyMigration = func(context.Context, *pgxpool.Pool) error { return errors.New("migration failure") }
	if err := run(context.Background(), lookup); err == nil {
		t.Fatal("migration succeeded")
	}
	applyMigration = func(context.Context, *pgxpool.Pool) error { return nil }
	if err := run(context.Background(), lookup); err != nil {
		t.Fatal(err)
	}
}

func TestStart(t *testing.T) {
	originalPrint, originalExit := printError, exitProcess
	defer func() { printError, exitProcess = originalPrint, originalExit }()
	called := false
	printError = func(...any) { called = true }
	exitProcess = func(code int) {
		if code != 1 {
			t.Fatalf("exit code = %d", code)
		}
	}
	main()
	if !called {
		t.Fatal("error was not reported")
	}
}

func TestClosePool(t *testing.T) {
	t.Helper()
	defer func() { _ = recover() }()
	closePool(nil)
}
