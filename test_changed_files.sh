#!/usr/bin/env bash
set -euo pipefail

CHANGED_FILES=$'internal/helloworld/handler.go
internal/helloworld/service.go'

echo "📂 Files to process:"
echo "$CHANGED_FILES"

files=()
while IFS= read -r line; do
  [[ -n "$line" ]] && files+=("$line")
done <<< "$CHANGED_FILES"

for file in "${files[@]}"; do
  echo "Generating tests for: $file"
  # ./scripts/generate_tests_for_file.sh "$file"
done
