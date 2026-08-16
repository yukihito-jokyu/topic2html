#!/usr/bin/env bash
set -euo pipefail

database_url="${TEST_DATABASE_URL:-}"
container=""

cleanup() {
	if [[ -n "$container" ]]; then
		docker rm -f "$container" >/dev/null 2>&1 || true
	fi
}
trap cleanup EXIT

if [[ -z "$database_url" ]]; then
	container="topic2html-pg-integration-$$"
	docker run --rm --detach --name "$container" \
		-e POSTGRES_USER=topic2html \
		-e POSTGRES_PASSWORD=topic2html \
		-e POSTGRES_DB=topic2html \
		-p 127.0.0.1::5432 \
		postgres:16-alpine >/dev/null
	port="$(docker port "$container" 5432/tcp | sed -E 's/.*:([0-9]+)$/\1/')"
	for _ in {1..30}; do
		if docker exec "$container" pg_isready -U topic2html -d topic2html >/dev/null 2>&1 && (: >"/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1; then
			break
		fi
		sleep 1
	done
	if ! docker exec "$container" pg_isready -U topic2html -d topic2html >/dev/null 2>&1 || ! (: >"/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1; then
		echo 'PostgreSQL did not become reachable' >&2
		exit 1
	fi
	database_url="postgres://topic2html:topic2html@127.0.0.1:${port}/topic2html?sslmode=disable"
fi

go test -count=1 ./repository/postgres -run '^(TestAdminAuthSchemaIntegration|TestAdminSessionCSRFCiphertextMigrationRollbackIntegration|TestGenerationRequestSchemaIntegration|TestGenerationRequestSchemaRollbackAndConstraintsIntegration|TestGenerationT1SessionRollbackIntegration|TestGenerationTransactionRollbackIntegration)$' -args "-integration-database-url=$database_url"
go test -count=1 ./integration -run '^TestAdminSessionHTTPPostgresIntegration$' -args "-integration-database-url=$database_url"
