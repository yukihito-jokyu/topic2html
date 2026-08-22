#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
container="topic2html-pg-e2e-$$"
database_port=""
google_endpoint_file="$(mktemp)"
backend_binary="$(mktemp)"
google_pid=""
backend_pid=""
frontend_pid=""

cleanup() {
	for pid in "$frontend_pid" "$backend_pid" "$google_pid"; do
		if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
			kill "$pid" >/dev/null 2>&1 || true
		fi
	done
	wait "$frontend_pid" "$backend_pid" "$google_pid" 2>/dev/null || true
	rm -f "$google_endpoint_file"
	rm -f "$backend_binary"
	docker rm --force "$container" >/dev/null 2>&1 || true
}

wait_for() {
	local endpoint="$1"
	for _ in {1..60}; do
		if curl --fail --silent --show-error --insecure "$endpoint" >/dev/null 2>&1; then
			return
		fi
		sleep 1
	done
	echo "Timed out waiting for $endpoint" >&2
	exit 1
}

wait_for_postgres() {
	for _ in {1..90}; do
		if ! docker inspect --format '{{.State.Running}}' "$container" 2>/dev/null | grep -qx true; then
			break
		fi
		if docker exec "$container" pg_isready -U topic2html -d topic2html >/dev/null 2>&1; then
			return
		fi
		sleep 1
	done

	echo "PostgreSQL did not become reachable; container logs follow:" >&2
	docker logs "$container" >&2 || true
	exit 1
}

wait_for_postgres_port() {
	for _ in {1..30}; do
		if nc -z 127.0.0.1 "$database_port" >/dev/null 2>&1; then
			return
		fi
		sleep 1
	done
	echo "PostgreSQL host port did not become reachable: $database_port" >&2
	exit 1
}

trap cleanup EXIT INT TERM

docker run --rm --detach --name "$container" \
	-e POSTGRES_USER=topic2html \
	-e POSTGRES_PASSWORD=topic2html \
	-e POSTGRES_DB=topic2html \
	-p 127.0.0.1::5432 \
	postgres:16-alpine >/dev/null
database_port="$(docker port "$container" 5432/tcp | sed -E 's/.*:([0-9]+)$/\1/')"
wait_for_postgres
wait_for_postgres_port

export TOPIC2HTML_TRUSTED_APP_ORIGIN="https://localhost:5173"
export TOPIC2HTML_GOOGLE_CLIENT_ID="e2e-client-id"
export TOPIC2HTML_GOOGLE_CLIENT_SECRET="e2e-client-secret"
export TOPIC2HTML_ALLOWED_EMAIL="admin@example.test"
export TOPIC2HTML_DATABASE_URL="postgres://topic2html:topic2html@127.0.0.1:${database_port}/topic2html?sslmode=disable"
export TOPIC2HTML_PROTECTION_KEY="e2e-only-protection-key"
export TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT="unix:///tmp/topic2html-e2e-broker-$$.sock"

export TOPIC2HTML_E2E_GOOGLE_ENDPOINT_FILE="$google_endpoint_file"
node "$root_dir/frontend/scripts/e2e-google-double.mjs" &
google_pid=$!
for _ in {1..30}; do
	if [[ -s "$google_endpoint_file" ]]; then
		break
	fi
	sleep 1
done
if [[ ! -s "$google_endpoint_file" ]]; then
	echo "Google OAuth test double did not start" >&2
	exit 1
fi
export TOPIC2HTML_GOOGLE_DISCOVERY_ENDPOINT="$(<"$google_endpoint_file")"
wait_for "$TOPIC2HTML_GOOGLE_DISCOVERY_ENDPOINT"

(
	cd "$root_dir/backend"
	go run ./cmd/migrate
	go build -tags e2e -o "$backend_binary" ./cmd/server
)
"$backend_binary" &
backend_pid=$!
wait_for "http://127.0.0.1:8080/health"

bash "$root_dir/scripts/create-dev-certificate.sh"
(
	cd "$root_dir/frontend"
	TOPIC2HTML_DEV_TLS=1 npm run dev
) &
frontend_pid=$!
wait_for "https://localhost:5173/admin"

cd "$root_dir/frontend"
npx playwright test --config playwright.config.ts
