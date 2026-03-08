#!/bin/bash
#
# Claude Code stop hook
# Warns if there are uncommitted changes without quality checks having been run

set -euo pipefail

# Check if we're in a git repo
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    exit 0
fi

# Check for uncommitted changes
if ! git diff --quiet HEAD 2>/dev/null || [ -n "$(git ls-files --others --exclude-standard 2>/dev/null)" ]; then
    SENTINEL=".claude/.checks-passed"
    if [ ! -f "$SENTINEL" ]; then
        echo "WARNING: There are uncommitted changes and quality checks have not been run." >&2
        echo "Run 'make check-all' before committing." >&2
    fi
fi

exit 0
