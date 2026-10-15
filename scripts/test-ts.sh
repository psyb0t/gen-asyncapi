#!/bin/bash
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

# The generated files are checked on their own, without Node.js types, so they
# cannot use anything a browser lacks. The tests are checked with Node.js types
# and then run.
log INFO "type-checking the generated TypeScript"
if ! tsc -p tsconfig.json; then
	log ERROR "generated TypeScript failed the type check"
	exit 1
fi

log INFO "type-checking the TypeScript tests"
if ! tsc -p tsconfig.test.json; then
	log ERROR "TypeScript tests failed the type check"
	exit 1
fi

log INFO "running the TypeScript tests"
if ! node --test 'examples/**/*.test.ts'; then
	log ERROR "TypeScript tests failed"
	exit 1
fi

log INFO "TypeScript checks complete"
