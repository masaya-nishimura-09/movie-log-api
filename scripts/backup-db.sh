#!/bin/bash
set -euo pipefail

BUCKET="cinelier-backup-648358186766-ap-southeast-2-an"
FILE="movie_log-$(date -u +%Y%m%dT%H%M%SZ).dump"

cd "$(dirname "$0")/.."

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

docker compose exec -T postgres sh -c 'pg_dump -Fc -U "$POSTGRES_USER" -d "$POSTGRES_DB"' > "$TMP"
aws s3 cp "$TMP" "s3://${BUCKET}/${FILE}" --region ap-southeast-2
echo "uploaded ${FILE}"
