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
	for _, testCase := range []struct {
		name      string
		lookup    func(string) (string, bool)
		configure func()
		wantError bool
	}{
		{
			name:      "missing configuration",
			lookup:    func(string) (string, bool) { return "", false },
			wantError: true,
		},
		{
			name:   "pool connection fails",
			lookup: lookup,
			configure: func() {
				newPool = func(context.Context, string) (*pgxpool.Pool, error) { return nil, errors.New("unavailable") }
			},
			wantError: true,
		},
		{
			name:   "migration fails",
			lookup: lookup,
			configure: func() {
				newPool = func(context.Context, string) (*pgxpool.Pool, error) { return &pgxpool.Pool{}, nil }
				closePool = func(*pgxpool.Pool) {}
				applyMigration = func(context.Context, *pgxpool.Pool) error { return errors.New("migration failure") }
			},
			wantError: true,
		},
		{
			name:   "migration succeeds",
			lookup: lookup,
			configure: func() {
				newPool = func(context.Context, string) (*pgxpool.Pool, error) { return &pgxpool.Pool{}, nil }
				closePool = func(*pgxpool.Pool) {}
				applyMigration = func(context.Context, *pgxpool.Pool) error { return nil }
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			newPool, applyMigration, closePool = originalNewPool, originalApply, originalClose
			if testCase.configure != nil {
				testCase.configure()
			}
			if err := run(context.Background(), testCase.lookup); (err != nil) != testCase.wantError {
				t.Fatalf("error=%v", err)
			}
		})
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
