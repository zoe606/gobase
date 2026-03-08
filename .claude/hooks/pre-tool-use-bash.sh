#!/bin/bash
#
# Claude Code pre-tool-use hook for Bash commands
# Blocks dangerous git operations and enforces quality checks before commits
#
# Receives tool input JSON on stdin: {"command": "..."}

set -euo pipefail

# Read the command from stdin JSON
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"command"[[:space:]]*:[[:space:]]*"//;s/"$//' || true)

if [ -z "$COMMAND" ]; then
    exit 0
fi

# --- Block --no-verify in git context ---
if echo "$COMMAND" | grep -qE 'git\s+(commit|push).*--no-verify'; then
    echo "BLOCKED: --no-verify is prohibited. Quality checks must not be bypassed." >&2
    echo "Run 'make check-all' first, then commit normally." >&2
    exit 2
fi

# --- Block force push ---
if echo "$COMMAND" | grep -qE 'git\s+push\s+.*(-f|--force)\b'; then
    echo "BLOCKED: Force push is prohibited. It can destroy remote history." >&2
    exit 2
fi

# --- Block destructive git operations ---
if echo "$COMMAND" | grep -qE 'git\s+reset\s+--hard'; then
    echo "BLOCKED: git reset --hard is prohibited. Use safer alternatives." >&2
    exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+checkout\s+--\s+\.'; then
    echo "BLOCKED: git checkout -- . is prohibited. It discards all changes." >&2
    exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+clean\s+-f'; then
    echo "BLOCKED: git clean -f is prohibited. It removes untracked files." >&2
    exit 2
fi

# --- Block git commit without quality checks ---
if echo "$COMMAND" | grep -qE 'git\s+commit\b'; then
    SENTINEL=".claude/.checks-passed"
    if [ ! -f "$SENTINEL" ]; then
        echo "BLOCKED: Cannot commit without running quality checks first." >&2
        echo "Run 'make check-all' before committing." >&2
        exit 2
    fi

    # Check sentinel age (must be less than 10 minutes old)
    if [ "$(uname)" = "Darwin" ]; then
        FILE_AGE=$(( $(date +%s) - $(stat -f %m "$SENTINEL") ))
    else
        FILE_AGE=$(( $(date +%s) - $(stat -c %Y "$SENTINEL") ))
    fi

    if [ "$FILE_AGE" -gt 600 ]; then
        echo "BLOCKED: Quality checks are stale (>10 minutes old)." >&2
        echo "Run 'make check-all' again before committing." >&2
        rm -f "$SENTINEL"
        exit 2
    fi
fi

exit 0
