#!/usr/bin/env bash
# Poll-and-deploy for this repo, run every few minutes from cron:
#
#   */3 * * * * $HOME/ektevalgomat/poll-deploy.sh >> $HOME/logs/ektevalgomat-deploy.log 2>&1
#
# Two properties it enforces, both learnt the hard way:
#
# 1. Converge on every run; never gate the deploy on "did HEAD move".
#    A script that compares HEAD before and after its own pull skips
#    itself whenever something else advanced HEAD first, and a failed
#    build never retries. `docker compose up -d` is a no-op when nothing
#    changed and the build cache makes a same-source rebuild take
#    seconds, so there is nothing to gain by gating. Only the log line is
#    gated, to keep cron output quiet.
#
# 2. `flock` around the git section, and a second lock around the whole
#    run. Two overlapping cron runs collide on .git/index.lock, and
#    `set -e` turns that into a failed deploy.
#
# This repo has its own clone, so REPO_DIR is this script's own directory.
set -euo pipefail

REPO_DIR="${REPO_DIR:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)}"
CONTAINER="${CONTAINER:-ektevalgomat}"
BRANCH="${BRANCH:-main}"
LOG_PREFIX="[poll-deploy $CONTAINER]"

if [ ! -f "$REPO_DIR/docker-compose.yml" ]; then
  echo "$LOG_PREFIX no docker-compose.yml at $REPO_DIR" >&2
  exit 64
fi

cd "$REPO_DIR"

# Keep one run active through Git, build, up and health. FD 9 stays open
# until exit, including in child commands; do not unlink the lock file
# after use.
exec 9>"$REPO_DIR/.git/$CONTAINER-deploy.lock"
if flock --nonblock --conflict-exit-code 75 9; then
  :
else
  status=$?
  if [ "$status" -eq 75 ]; then
    echo "$LOG_PREFIX another run is active; skipped (next cron run retries)"
    exit 0
  fi
  echo "$LOG_PREFIX cannot acquire deploy lock (exit $status)" >&2
  exit "$status"
fi

before="$(git rev-parse HEAD)"
# The git lock is separate from the deploy lock. Wait at most 120s for
# another consumer of this clone; a timeout remains a failure.
flock --timeout 120 "$REPO_DIR/.git" bash -c '
  set -euo pipefail
  git fetch origin "'"$BRANCH"'" --quiet
  git merge --ff-only "origin/'"$BRANCH"'" --quiet
'
after="$(git rev-parse HEAD)"

[ "$before" != "$after" ] && echo "$LOG_PREFIX $before -> $after"

docker compose build --quiet
docker compose up -d --quiet-pull

# Always check health, even when another run already advanced HEAD or this
# run only recovered a stopped container.
sleep 5
if ! docker ps --filter "name=$CONTAINER" --filter status=running \
     --format '{{.Names}}' | grep -qx "$CONTAINER"; then
  echo "$LOG_PREFIX $CONTAINER is not running after deploy — recent logs:" >&2
  docker logs "$CONTAINER" --tail 20 >&2 || true
  exit 1
fi

# A passing build says nothing about what the container answers right now.
if ! docker exec "$CONTAINER" /server healthcheck; then
  echo "$LOG_PREFIX $CONTAINER is running but not serving — recent logs:" >&2
  docker logs "$CONTAINER" --tail 20 >&2 || true
  exit 1
fi

if [ "$before" != "$after" ]; then
  echo "$LOG_PREFIX deployed $after, $CONTAINER running and serving"
fi
