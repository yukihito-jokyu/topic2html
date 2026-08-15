#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
certificate_directory="$root_dir/.certs"
certificate="$certificate_directory/topic2html-dev-cert.pem"
private_key="$certificate_directory/topic2html-dev-key.pem"

if [[ -f "$certificate" && -f "$private_key" ]] && \
	openssl x509 -in "$certificate" -noout >/dev/null 2>&1 && \
	openssl pkey -in "$private_key" -noout >/dev/null 2>&1; then
	exit 0
fi

mkdir -p "$certificate_directory"
rm -f "$certificate" "$private_key"
openssl req -x509 -newkey rsa:2048 -sha256 -nodes \
	-keyout "$private_key" \
	-out "$certificate" \
	-days 30 \
	-subj '/CN=localhost' \
	-addext 'subjectAltName=DNS:localhost,IP:127.0.0.1' >/dev/null 2>&1
chmod 600 "$private_key"
