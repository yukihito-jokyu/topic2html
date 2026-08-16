package postgres

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const AdminAuthSchemaMigration = "001_admin_auth_schema"
const AdminSessionCSRFCiphertextMigration = "002_admin_session_csrf_ciphertext"
const GenerationRequestSchemaMigration = "003_generation_request_schema"

//go:embed migrations/001_admin_auth_schema.sql
var adminAuthSchemaDDL string

//go:embed migrations/002_admin_session_csrf_ciphertext.sql
var adminSessionCSRFCiphertextDDL string

//go:embed migrations/003_generation_request_schema.sql
var generationRequestSchemaDDL string

// ApplyAdminAuthSchemaはmigrationを一つの明示transactionで冪等に適用します。
func ApplyAdminAuthSchema(ctx context.Context, database *pgxpool.Pool) error {
	return ApplyMigrations(ctx, database)
}

// ApplyMigrationsは承認済みmigrationを順序どおり適用します。
func ApplyMigrations(ctx context.Context, database *pgxpool.Pool) error {
	return applyMigrations(ctx, newPGXPool(database))
}

func applyMigrations(ctx context.Context, database pool) error {
	for _, migration := range []struct {
		version string
		ddl     string
	}{
		{
			version: AdminAuthSchemaMigration,
			ddl:     adminAuthSchemaDDL,
		},
		{
			version: AdminSessionCSRFCiphertextMigration,
			ddl:     adminSessionCSRFCiphertextDDL,
		},
		{
			version: GenerationRequestSchemaMigration,
			ddl:     generationRequestSchemaDDL,
		},
	} {
		if err := applyMigrationWithDDL(ctx, database, migration.version, migration.ddl); err != nil {
			return err
		}
	}

	return nil
}

func applyAdminAuthSchema(ctx context.Context, database pool) error {
	return applyMigrationWithDDL(ctx, database, AdminAuthSchemaMigration, adminAuthSchemaDDL)
}

// applyAdminAuthSchemaWithDDLはテスト用DDLを適用します。
func applyAdminAuthSchemaWithDDL(ctx context.Context, database pool, ddl string) error {
	return applyMigrationWithDDL(ctx, database, AdminAuthSchemaMigration, ddl)
}

func applyMigrationWithDDL(ctx context.Context, database pool, versionName, ddl string) error {
	transaction, err := database.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if _, err := transaction.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT NOT NULL PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL)`); err != nil {
		return err
	}
	var version string
	err = transaction.QueryRow(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, versionName).Scan(&version)
	if err == nil {
		return transaction.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err := transaction.Exec(ctx, ddl); err != nil {
		return err
	}
	if _, err := transaction.Exec(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, CURRENT_TIMESTAMP)`, versionName); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}
