package postgres

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const AdminAuthSchemaMigration = "001_admin_auth_schema"

//go:embed migrations/001_admin_auth_schema.sql
var adminAuthSchemaDDL string

// ApplyAdminAuthSchemaはmigrationを一つの明示transactionで冪等に適用します。
func ApplyAdminAuthSchema(ctx context.Context, database *pgxpool.Pool) error {
	return applyAdminAuthSchemaWithDDL(ctx, newPGXPool(database), adminAuthSchemaDDL)
}

func applyAdminAuthSchema(ctx context.Context, database pool) error {
	return applyAdminAuthSchemaWithDDL(ctx, database, adminAuthSchemaDDL)
}

// applyAdminAuthSchemaWithDDLはテスト用DDLを適用します。
func applyAdminAuthSchemaWithDDL(ctx context.Context, database pool, ddl string) error {
	transaction, err := database.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if _, err := transaction.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT NOT NULL PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL)`); err != nil {
		return err
	}
	var version string
	err = transaction.QueryRow(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, AdminAuthSchemaMigration).Scan(&version)
	if err == nil {
		return transaction.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err := transaction.Exec(ctx, ddl); err != nil {
		return err
	}
	if _, err := transaction.Exec(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, CURRENT_TIMESTAMP)`, AdminAuthSchemaMigration); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}
