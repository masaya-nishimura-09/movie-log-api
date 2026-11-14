#!/bin/bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

if [ -n "$(git status --porcelain)" ]; then
  echo "error: uncommitted changes" >&2
  exit 1
fi

TAG=$(git describe --tags --exact-match)

docker buildx build --platform linux/arm64 -t "movie-log-api:${TAG}" --load .
echo "built movie-log-api:${TAG}"
