#!/usr/bin/env bash
set -euo pipefail

FILE="$1"  # e.g. internal/health/handler.go
PROMPT_FILE=".cursor/prompts/generate_tests.txt"

# Ensure cursor-agent is installed
if ! command -v cursor-agent &>/dev/null; then
  echo "ERROR: cursor-agent not found in PATH" >&2
  exit 2
fi

# Ensure the prompt file exists
if [[ ! -f "$PROMPT_FILE" ]]; then
  echo "ERROR: Prompt file not found at $PROMPT_FILE" >&2
  exit 3
fi

# Prepare a temporary prompt file with the filepath substituted
TMP_PROMPT="$(mktemp)"
sed "s|{{filepath}}|${FILE}|g" "$PROMPT_FILE" > "$TMP_PROMPT"

echo "Generating tests for: $FILE"

# Run cursor-agent using the prompt
cursor-agent run "$(cat "$TMP_PROMPT")" --force

# Cleanup
rm -f "$TMP_PROMPT"
echo "Done generating for $FILE"
