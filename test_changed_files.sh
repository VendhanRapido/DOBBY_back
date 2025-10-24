#!/usr/bin/env bash
set -euo pipefail

# Simulate changed files (multiline string)
CHANGED_FILES=$'internal/helloworld/handler.go
internal/helloworld/service.go'

echo "📂 Files to process:"
echo "$CHANGED_FILES"


# Loop through each file
while IFS= read -r file; do
  if [[ -n "$file" ]]; then
    echo "Generating tests for: $file"
    # scripts/generate_tests_for_file.sh "$file"
  fi
done <<< "$CHANGED_FILES"
