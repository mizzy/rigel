#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "Running all Expect tests in $(pwd)"
status=0

for f in *.exp; do
  echo "\n--- Running $f ---"
  if expect -f "$f"; then
    echo "PASS: $f"
  else
    echo "FAIL: $f"
    status=1
  fi
done

exit $status

