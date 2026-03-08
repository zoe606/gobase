#!/bin/bash
#
# Claude Code post-tool-use hook for Bash commands
# Creates sentinel file after successful quality checks
#
# Receives tool input JSON on stdin: {"command": "...", "stdout": "...", "stderr": "...", "exitCode": 0}

set -euo pipefail

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"command"[[:space:]]*:[[:space:]]*"//;s/"$//' || true)
EXIT_CODE=$(echo "$INPUT" | grep -o '"exitCode"[[:space:]]*:[[:space:]]*[0-9]*' | sed 's/"exitCode"[[:space:]]*:[[:space:]]*//' || true)

if [ -z "$COMMAND" ]; then
    exit 0
fi

# Touch sentinel file after successful make check-all or make ci
if echo "$COMMAND" | grep -qE 'make\s+(check-all|ci)\b'; then
    if [ "$EXIT_CODE" = "0" ]; then
        mkdir -p .claude
        touch .claude/.checks-passed
    fi
fi

exit 0
