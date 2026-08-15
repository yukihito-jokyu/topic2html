package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestApplyAdminAuthSchema(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pool      fakePool
		wantError bool
	}{
		{"first apply", fakePool{
			transaction: &fakeTx{
				rows: []error{pgx.ErrNoRows},
			},
		}, false},
		{"already applied", fakePool{
			transaction: &fakeTx{},
		}, false},
		{"begin failure", fakePool{
			err: errors.New("boom"),
		}, true},
		{"metadata failure", fakePool{
			transaction: &fakeTx{
				exec: []error{errors.New("boom")},
			},
		}, true},
		{"lookup failure", fakePool{
			transaction: &fakeTx{
				rows: []error{errors.New("boom")},
			},
		}, true},
		{"DDL failure", fakePool{
			transaction: &fakeTx{
				rows: []error{pgx.ErrNoRows},
				exec: []error{nil, errors.New("boom")},
			},
		}, true},
		{"record failure", fakePool{
			transaction: &fakeTx{
				rows: []error{pgx.ErrNoRows},
				exec: []error{nil, nil, errors.New("boom")},
			},
		}, true},
		{"commit failure", fakePool{
			transaction: &fakeTx{
				rows:   []error{pgx.ErrNoRows},
				commit: errors.New("boom"),
			},
		}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := applyAdminAuthSchema(context.Background(), tt.pool)
			if (err != nil) != tt.wantError {
				t.Fatalf("error=%v", err)
			}
			if tt.name == "already applied" {
				for _, call := range tt.pool.transaction.calls {
					if strings.Contains(call, "admin_oauth_transactions") || strings.Contains(call, "INSERT INTO schema_migrations") {
						t.Fatalf("already applied migration executed DDL or record insert: %q", call)
					}
				}
			}
		})
	}
}

func TestApplyMigrations(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pool      fakePool
		wantError bool
	}{
		{
			name: "all migrations apply",
			pool: fakePool{transaction: &fakeTx{rows: []error{pgx.ErrNoRows, pgx.ErrNoRows}}},
		},
		{
			name: "second migration fails",
			pool: fakePool{transaction: &fakeTx{
				rows: []error{pgx.ErrNoRows, pgx.ErrNoRows},
				exec: []error{nil, nil, nil, nil, errors.New("boom")},
			}},
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyMigrations(context.Background(), tt.pool); (err != nil) != tt.wantError {
				t.Fatalf("error=%v", err)
			}
		})
	}
	if err := applyAdminAuthSchemaWithDDL(context.Background(), fakePool{transaction: &fakeTx{rows: []error{pgx.ErrNoRows}}}, "CREATE TABLE example (id INT);"); err != nil {
		t.Fatal(err)
	}
}
