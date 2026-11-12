#!/bin/sh
# Build and test the COMMITTED tree, not the working tree.
#
# Why: `git add <explicit paths>` can leave a file behind that the working tree
# still has. The working tree then compiles and its tests pass, while the commit
# panics at startup. On 2026-11-12 exactly that shipped: nine templates called
# {{explore}} and the line registering the function was never committed, so the
# binary died at template parse, the healthcheck failed, and the deploy host kept
# the previous container serving for 19 minutes.
#
# A test can only fail on what is on disk. This extracts HEAD into a temp dir and
# runs the same build and test there, which puts the committed tree under the
# check the working tree already passed.
#
# Usage: tools/verify/check_head.sh [ref]   (default: HEAD)
set -eu

REF="${1:-HEAD}"
ROOT=$(git rev-parse --show-toplevel)
TMP=$(mktemp -d "${TMPDIR:-/tmp}/ekv-head-XXXXXX")
trap 'rm -rf "$TMP"' EXIT INT TERM

# git sets GIT_DIR/GIT_INDEX_FILE when it runs a hook; scrub them so the archive
# and the build in $TMP do not resolve against the caller's environment.
unset GIT_DIR GIT_INDEX_FILE GIT_WORK_TREE

git -C "$ROOT" archive "$REF" | tar -x -C "$TMP"

cd "$TMP"

if ! command -v go >/dev/null 2>&1; then
	echo "check_head: go not on PATH; cannot verify $REF" >&2
	exit 127
fi

echo "check_head: building $REF from an archive ($(git -C "$ROOT" rev-parse --short "$REF"))"
go build ./... || { echo "check_head: FAIL build" >&2; exit 1; }

go test ./... -count=1 2>&1 | tee "$TMP/.test.log" >/dev/null || true
PKGS=$(grep -cE '^ok[[:space:]]' "$TMP/.test.log" || true)
FAILED=$(grep -cE '^(FAIL|---[[:space:]]*FAIL)' "$TMP/.test.log" || true)

if [ "$FAILED" -gt 0 ]; then
	grep -E '^(FAIL|---[[:space:]]*FAIL)' "$TMP/.test.log" >&2
	echo "check_head: FAIL tests ($PKGS packages ok, $FAILED failing)" >&2
	exit 1
fi
if [ "$PKGS" -eq 0 ]; then
	# An empty result is not a pass: it means the test run produced nothing the
	# grep recognises, which is a broken invocation, not a green tree.
	echo "check_head: FAIL no package reported ok; the test run did not run" >&2
	exit 1
fi

echo "check_head: OK $REF builds and $PKGS packages pass"
