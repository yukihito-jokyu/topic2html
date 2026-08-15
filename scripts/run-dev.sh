#!/usr/bin/env bash
set -euo pipefail

backend_pid=''
frontend_pid=''

cleanup() {
	for pid in "$backend_pid" "$frontend_pid"; do
		if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
			kill "$pid" >/dev/null 2>&1 || true
		fi
	done
	wait "$backend_pid" "$frontend_pid" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

(
	cd backend
	go run ./cmd/server
) &
backend_pid=$!

(
	cd frontend
	TOPIC2HTML_DEV_TLS=1 npm run dev
) &
frontend_pid=$!

while kill -0 "$backend_pid" >/dev/null 2>&1 && kill -0 "$frontend_pid" >/dev/null 2>&1; do
	sleep 1
done

if ! kill -0 "$backend_pid" >/dev/null 2>&1; then
	wait "$backend_pid"
fi

wait "$frontend_pid"
