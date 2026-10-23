# ektevalgomat harvester

Offline lead-finder against `data.stortinget.no`, built for
`docs/roadmap.md`'s "Next candidate work" item 1: query Stortinget's own
JSON export API directly — not the JS-rendered web search UI, which the
2026-09-23 research pass hit real limits with — across sessions, and
surface candidate cases per open statement topic.

**This tool never writes into `internal/data/content/`.** It writes a
Markdown report plus per-lead evidence-YAML *skeletons* under
`reports/<date>/`, with only the fields the API can actually confirm
(case id, `henvisning`, title, URL) filled in for real. Every field that
requires reading the case — `party_id`, `position`, `verbatim_excerpt`,
`context_note` — stays an explicit `TODO`, and `position` is always
`null`. That's deliberate: a skeleton must never be mistaken for real
evidence, and it must be structurally impossible for a lead to become a
coded position without a human (or an agent doing the same
primary-source read utsagn-01/03/06 went through) actually opening the
case and reading it — same discipline as every other evidence file in
this repo, see `docs/roadmap.md`'s kjernekraft and partiloven entries
for the bar a real evidence file has to clear.

## Setup

```sh
cd tools/harvester
uv sync
```

## Usage

```sh
uv run harvest.py --dry-run                    # print the session list, no network
uv run harvest.py                               # scan every session since 2009 for every open topic
uv run harvest.py --start-year 2020             # narrower session range
uv run harvest.py --topic utsagn-04              # one topic only
uv run harvest.py --no-cache                    # ignore cached session JSON, re-fetch
```

Output goes to `reports/<today>/REPORT.md` plus one skeleton `.yaml` per
lead. Session JSON is cached under `.cache/` so re-running after editing
`topics.yaml` doesn't re-hit the API for sessions already fetched.

## Adding a topic

Edit `topics.yaml`. One entry per statement that already has a real
`statement.yaml` under `internal/data/content/statements/` but is
missing some or all party positions — **do not** add a topic for a
statement number that doesn't exist yet (see `topics.yaml`'s own header
comment for why: inventing a sub-topic ahead of drafting the actual
statement text is the exact mistake `docs/roadmap.md`'s "Correction"
section documents fixing once already).

## What this tool cannot do

Keyword matching runs against each case's `tittel`, `korttittel`, and
`emne_liste[].navn` only — the API's case list doesn't include the
committee merknad text itself (that's in the innstilling PDF). A real
case can have a title that gives no hint of its content: utsagn-03's
best evidence (Innst. 528 L, the actual anti-anonymous-funding vote) has
a title about "bakenforliggende givere," not "partifinansiering" or
"åpenhet," and was found by web search, not blind title scanning here.
Treat a 0-lead run as "no title-level match," never as "confirmed
nothing exists" — the same caveat `docs/roadmap.md` already carries for
the earlier JS-search null results.

## Review workflow

1. Run a harvest for the topic(s) you're working on.
2. Read `REPORT.md`. For each lead worth pursuing, open the real case
   page (or its PDF, via `pdftotext -layout`) and read it — the same way
   every existing evidence file in this repo was built.
3. Fill in the skeleton's `TODO` fields from what you actually read,
   with a verbatim excerpt, not a paraphrase.
4. Verify the cited URL is live (`scripts/research/verify_evidence.py`).
5. Move the finished file into `internal/data/content/evidence/` with
   the next free number, run `go test ./...`, update the count assertion
   in `internal/data/loader_test.go`.
6. Delete the skeleton from `reports/` (or let `.gitignore` handle it —
   `reports/` is never committed).
