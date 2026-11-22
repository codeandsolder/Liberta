#!/usr/bin/env bash
set -euo pipefail
go test ./lib/storage ./app/vmstorage
go build -o "$RUNNER_TEMP/victoria-metrics" ./app/victoria-metrics
