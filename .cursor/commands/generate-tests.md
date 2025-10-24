# generate-tests
# Usage:
#   cursor-agent run --prompt "$(cat .cursor/prompts/generate_tests.txt)" --print --var filepath=path/to/file.go
#
# This command generates or updates tests (and mocks if needed) for the specified Go file.
# Replace "path/to/file.go" with the relative path of the file to test.
