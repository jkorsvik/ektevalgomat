# ektevalgomat

**A real valgomat** — a Norwegian election-guide quiz that answers each
value statement against a *documented* party position (a motion, a vote,
a law) instead of the party's own program wording, and always shows the
source. Built from the research memo in `docs/research-memo.md`
(2026-09-23) — see `docs/roadmap.md` for the full build history and
exactly what's deliberately not built yet (a `rettelseslogg` UI,
`konsistens`/`beneficiaries` fields the memo describes but no evidence
yet needs). Current state (2026-10-28): 60 statements, 643 evidence
files. That is the memo's 42 (utsagn 1-42), 2 "Tillegg" extras (44, 45),
two owner-approved dimensions added after user feedback,
"Tilgang, integrering og rettsstat" (46-51) and "Finansiering av
fellesskapet" (52-58), utsagn-59 in "Norsk handlefrihet", and the two
statsbudsjett 2027 statements (60, 61) added 2026-10-12. There is
no utsagn 43, see roadmap "Phase 2r".
55 of 60 statements have a coded position for all 9 parties. The six
that do not are utsagn-06 (7 of 9), utsagn-52 (8 of 9), utsagn-55 (7 of 9)
and utsagn-57 (8 of 9), each one searched and left honestly empty, plus
utsagn-60 and 61, which carry institutional evidence only. roadmap.md's
"Current state" section says why each gap is a gap. Every statement is one
claim with a "Hva betyr et svar i praksis?" box, and 13 are phrased in the
opposite direction to hold the acquiescence spread at 12 points (was 41).
Tests enforce both, see the roadmap. Every verbatim excerpt and every
"stemte for/mot" is checked against its source by
`tools/verify/verify_excerpts.py` and `tools/verify/verify_votes.py`; 346 of
the 643 evidence rows carry such an excerpt, and the rest are votes and
documents where the claim is the action itself, not a quotation.

Live at **https://ektevalgomat.no**. A "small public Go server": Go
stdlib `net/http`, content `go:embed`ded into the binary, a `FROM
scratch` image, no database, no writable volume, and a `healthcheck`
subcommand so the deploy can ask the container what it actually answers.
The content is structured evidence data (`internal/data/content/*.yaml`).

## Why no database

The source memo asks for a "rettelseslogg, versjonshistorikk og
metodeendringer" (correction log, version history, method changes) and
recommends PostgreSQL for the full 42×9 project. This project uses
neither: every statement, party, and evidence object is version-controlled
YAML under `internal/data/content/`, embedded into the binary at build time
(see `internal/data/loader.go`). Git commit history **is** the correction
log — diff-able, tamper-evident, and every change already goes through
review before merge. Editing evidence means: change the YAML, commit,
push through review, and the existing 5-minute poll-deploy cron rebuilds
and redeploys automatically.

## The three scores, never merged

`internal/scoring/scoring.go` computes, per statement×party cell:

- **Saksmatch** — how close the user's answer is to the party's coded
  `position` (-2..+2), only when a position has actually been coded from
  a source. No coded position → "ikke vurdert", never a fabricated 0%.
- **Konsistens** — does documented action match stated principle over
  time? Nearly every phase-1 cell honestly reports "utilstrekkelig
  grunnlag" because the dataset is mostly single-snapshot evidence so
  far — this is a real reflection of what's been collected, not a bug.
- **Evidensstyrke** — how much and how authoritative the source is
  (kildeklasse A-E), independent of what it actually says.

A party can be highly consistent in a policy the user disagrees with (high
konsistens, low saksmatch), or match the user's values in its program
while never acting on them (the reverse). Collapsing these into one
number is exactly what the memo says a real valgomat must not do.

## Local development

```
go run ./cmd/server        # listens on :8080
PORT=3000 go run ./cmd/server
go test ./...              # loads the real embedded dataset, checks known invariants
go vet ./...
```

The same gofmt, vet and test run on every PR that touches this stack
(`.github/workflows/ci.yml`). The source verifiers in
`tools/verify/` fetch live pages, so run them locally before a PR that
changes evidence.

`/healthz` returns 200 once the server is up. `/utsagn/{number}` renders
the evidence-chain page for each statement. `/om` explains why this
exists, `/metode` explains the methodology (source tiers, attribution),
`/personvern` is the privacy page. `/data` lists the whole dataset and
links it as files: `/data/utsagn.json`, `/data/evidens.json` and
`/data/evidens.csv`, whose columns are documented on the page, because a
reader who wants to check the project should not have to scrape HTML.
`/bidra` lists what is missing right
now — which parties have no coded position on which statements, and which
sources are weak — generated from the dataset at build time and gated by a
test, because a hand-kept "help us" list drifts away from the real gaps and
sends volunteers after work that is already done. `/resultat` is your own
running match against documented party positions — same as `/`, still nothing ever
reaches the server. `/` is the quiz itself — answers are stored in
`localStorage` only.

## Adding or correcting evidence

1. Add or edit a YAML file under `internal/data/content/evidence/`.
   Fields are documented in `internal/data/types.go`. `party_id: ""`
   (or field omitted) means institutional/non-party evidence — it shows
   as context on the statement page but is never scored per party.
2. `go test ./...` — a malformed file, an unknown `party_id`, or an
   unknown `statement_id` fails the build, not just a lint warning.
3. Commit, push, open a PR. The existing tests in
   `internal/data/loader_test.go` lock in specific known facts (e.g.
   Venstre's, Rødt's, and MDG's coded positions on utsagn-01, or the
   addendum statements 44-45 having zero party positions) — update them
   deliberately if a fact genuinely changes, never silently.

## Building the image

Deployed by `poll-deploy.sh` from a cron entry on the host, with no
orchestrator involved. Clone this repository once, then point cron at the
script inside the clone:

```
*/3 * * * * $HOME/ektevalgomat/poll-deploy.sh >> $HOME/logs/ektevalgomat-deploy.log 2>&1
```

It fast-forwards its own clone under two locks, then builds, converges
and runs `/server healthcheck` inside the container on every run. It
converges instead of gating on "did HEAD move", so a run that finds the
clone already advanced still deploys, and a failed build retries on the
next tick. The same commands a human would run:

```
cp .env.example .env        # sets PUBLIC_DOMAIN, used by the Traefik host rule
docker compose build
docker compose up -d
```

Without `.env`, `PUBLIC_DOMAIN` resolves to nothing and the Traefik rule
renders as ``Host(`ektevalgomat.`)``, a host with a trailing dot that never
matches.

To build and smoke-test locally on amd64 hardware:

```
docker build -t ektevalgomat .
docker run --rm -p 8080:8080 ektevalgomat
curl localhost:8080/healthz
curl localhost:8080/                # quiz landing page
curl localhost:8080/utsagn/1        # evidence chain for statement 1
curl localhost:8080/metode          # methodology
curl localhost:8080/om              # why this exists
curl localhost:8080/personvern      # privacy page
curl localhost:8080/data            # the dataset, and its file downloads
curl localhost:8080/resultat        # your own running match
```

## Going live: DNS + Cloudflare Tunnel (one-time, manual)

`ektevalgomat.no` is already registered. Two steps only the domain owner
can do, both in the Cloudflare dashboard:

1. **Delegate DNS to Cloudflare**: add `ektevalgomat.no` as a site in
   Cloudflare, then update the domain's nameservers at whichever
   registrar it was bought through to the ones Cloudflare assigns.
2. **Add the public hostname**: Cloudflare dashboard → Zero Trust →
   Networks → Tunnels → your tunnel → Public Hostname
   tab → Add a public hostname. Domain: `ektevalgomat.no`. Service →
   Type: `HTTP` (not HTTPS — the container has no TLS listener; picking
   HTTPS here makes cloudflared attempt a TLS handshake against a
   plaintext service and fail with `tls: first record does not look like
   a TLS handshake`). URL: `ektevalgomat:8080` (the Docker service name,
   resolved over the `cloudflare-tunnel` network). Save — Cloudflare
   creates the DNS record and issues the edge TLS cert automatically.

Nothing else changes: adding a public hostname does not touch the tunnel
container or its token, and the service is reached over the shared
`cloudflare-tunnel` network created once on the host.

## Deploy

`poll-deploy.sh`, run from cron on the host every few minutes:

```
*/3 * * * * /path/to/ektevalgomat/poll-deploy.sh >> ~/logs/ektevalgomat-deploy.log 2>&1
```

It converges on every run rather than gating on "did HEAD move" (a script
that gates that way skips itself whenever something else advances HEAD
first, and a failed build never retries), takes a `flock` so overlapping
runs cannot collide on `.git/index.lock`, then builds, brings the
container up, and requires `/server healthcheck` to answer. Rollback is
`git revert` plus the next cron tick.

## Corrections and the record of change

The published history starts at the **initial public release**
(`fbc64d8`, 2026-10-23). Every change since is a commit on `main`, and
anything before that release is not published, so the history a reader can
see begins where the project became public rather than where it began.

Those commits are also not the record. A commit message is a note a
developer writes for other developers, and half of what went wrong here was
found by reading sources, not by reading diffs.

The record of what changed lives in two files that ship with the code:

- `docs/roadmap.md` — the day-by-day build log, including research
  reversals that never reached the site.
- `internal/data/content/corrections/` — the same mistakes in the
  project's own words, rendered at https://ektevalgomat.no/om for a
  reader who does not want to read a build log.

Neither is tamper-evident. That is the trade: an append-only record a
reader can check is worth more than a cryptographic one they cannot.

## License

**GNU Affero General Public License v3.0 or later (AGPL-3.0-or-later).**
Full text in `LICENSE`.

This is deliberately the *strongest* copyleft license, not a permissive
one. Plain GPL only triggers on distributing the binary — a hosted web
service never distributes anything, so anyone could take this code,
run their own closed valgomat off it, and never owe the public a single
line back. AGPL closes exactly that gap (§13): if you run a modified
version of this program and let others interact with it over a
network, you must offer them the modified source too. The idea here —
sourced, adversarial, "show the evidence chain, not the campaign
promise" political accountability tooling — is meant to spread and
improve in the open, not get forked into somebody's proprietary
product. Copyright (C) 2026 Jon-Mikkel Korsvik.
