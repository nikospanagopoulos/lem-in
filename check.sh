#!/usr/bin/env bash
# check.sh: run before every pull request.
# Checks formatting, vet, build and tests, and reports every failing step.
# Exit code is 0 only if all steps pass.
#
# Usage (from anywhere): ./check.sh

set -u

# Always work from the repo root, wherever the script is called from.
cd "$(dirname "$0")" || { echo "cannot enter the repo root"; exit 1; }

if ! command -v go >/dev/null 2>&1; then
    echo "go is not installed or not in PATH"
    exit 1
fi

failed=()

# run_step NAME COMMAND [ARGS...]: run a command, remember if it failed.
run_step() {
    local name="$1"
    shift
    echo "==> $name"
    if "$@"; then
        echo "    ok"
    else
        echo "    FAILED"
        failed+=("$name")
    fi
}

# gofmt -l prints the unformatted files but exits 0, so we test its output.
check_format() {
    local unformatted
    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        echo "These files need formatting (fix with: gofmt -w <file>):"
        echo "$unformatted" | sed 's/^/      /'
        return 1
    fi
}

run_step "gofmt"    check_format
run_step "go vet"   go vet ./...
run_step "go build" go build ./...
run_step "go test"  go test ./...

echo
if [ ${#failed[@]} -eq 0 ]; then
    echo "All checks passed."
    exit 0
fi

echo "Failed steps: ${failed[*]}"
exit 1