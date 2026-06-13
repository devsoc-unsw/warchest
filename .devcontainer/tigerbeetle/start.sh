#!/usr/bin/env sh
set -e

DATA_FILE="${TIGERBEETLE_DATA_FILE:-/data/0_0.tigerbeetle}"

if [ ! -f "$DATA_FILE" ]; then
  /tigerbeetle format --cluster=0 --replica=0 --replica-count=1 --development "$DATA_FILE"
fi

exec /tigerbeetle start --development --addresses=0.0.0.0:3000 "$DATA_FILE"
