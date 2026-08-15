#!/usr/bin/env bash
set -euo pipefail

image="${DB_IMAGE:-topic2html-postgres:local}"
container="${DB_CONTAINER:-topic2html-postgres}"
volume="${DB_VOLUME:-topic2html-postgres-data}"

docker build --tag "$image" --file docker/postgres/Dockerfile docker/postgres
if docker container inspect "$container" >/dev/null 2>&1; then
	docker start "$container" >/dev/null
else
	docker run --detach --name "$container" \
		-e POSTGRES_USER=topic2html \
		-e POSTGRES_PASSWORD=topic2html \
		-e POSTGRES_DB=topic2html \
		-p 127.0.0.1:5432:5432 \
		-v "$volume":/var/lib/postgresql/data \
		"$image" >/dev/null
fi

for _ in {1..30}; do
	if docker exec "$container" pg_isready -U topic2html -d topic2html >/dev/null 2>&1 && (: > /dev/tcp/127.0.0.1/5432) >/dev/null 2>&1; then
		exit 0
	fi
	sleep 1
done

echo 'PostgreSQL did not become reachable' >&2
exit 1
