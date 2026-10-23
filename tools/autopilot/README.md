# ektevalgomat autopilot

The gated loop that builds out the evidence base. It exists because the
bottleneck was never computing — it was *reading*, and the project's own
history (three documented fabrication incidents, see
`docs/research-briefs/CONTEXT.md`) shows that reading by a fast model
invents plausible citations unless something mechanical re-checks them.

So the autopilot is not a research agent making claims. It is a **staging
area with a lock on the door**:

```
worker agent  ->  tools/autopilot/staging/<item>/   (a draft; not content)
gate.py       ->  re-fetches every source, decides PASS/FAIL
promote.py    ->  only a PASS moves into internal/data/content/, then go test
```

Nothing a worker writes is content until `gate.py` has re-fetched the source
itself. A worker's own "all verified" summary is never read.

## Files

| File | Role |
|---|---|
| `queue.yaml` | the work items, with status |
| `method/budget.md` | how to read the statsbudsjett (the owner's method, written down) |
| `method/program.md` | how to read a party programme |
| `gate.py` | schema + excerpt + vote checks, live |
| `promote.py` | number, move, bump counts, `go test`, rollback |
| `staging/` | worker output; gitignored; never committed |

## The gate (`gate.py`)

Three checks, the last two imported from `tools/verify/` so there is one
implementation, not two:

1. **schema** — every field the loader requires, every enum, and the loader's
   own reference rules (unknown statement/party/class/tier fail). Plus two the
   loader does not: a status that says a party voted must cite a vote id, and a
   `verbatim_excerpt` must have a `source_url`.
2. **excerpt** — every quoted excerpt must be a substring of its `source_url`,
   fetched now.
3. **vote** — every `stemte_for` / `stemte_mot` must match the party's real
   split on data.stortinget.no.

Exit 1 on any failure, 2 if it checked nothing (a run that checks nothing must
never look like success).

## Promotion (`promote.py`)

Assigns the next free evidence number, moves the files, updates the count
assertions in `internal/data/loader_test.go` to the real totals, runs
`go test ./...`, and **rolls everything back** if the build or tests fail.
Writes the new status back into `queue.yaml`.

## How to run a batch

See the `ektevalgomat-autopilot` skill for the full loop. In short: read the
queue, spawn `ekv-harvester` (deepseek-v4.1-flash) per item, gate, promote,
then run both repo-wide verifiers and commit.

## Why the workers are cheap models

The design does not need an expensive model, because the model is not the
check. A cheap model with a strict output contract plus a mechanical gate is
safer than a strong model trusted to mark its own homework: the gate catches a
fabricated quote from a cheap model exactly as fast as from an expensive one,
and the cheap model is affordable at hundreds of items. The verifier agent
(`ekv-verifier`) exists for the one thing the gate cannot see — a real quote
attached to the wrong claim — and is run on the items that will carry a coded
position.

## The vote stream (`votes.py`)

The statements carry two kinds of tier-A evidence: what was decided (a
vedtak, an innstilling) and how each party voted (a voteringsspor). The
second kind cannot be backfilled on demand — a vote happens once, on a date,
and has to be captured then. `votes.py` is the capture tool:

    uv run tools/autopilot/votes.py 200414 200418

Given one or more Stortinget **sak** ids, it fetches each case's voteringer
and writes `staging/votes/<sak_id>.jsonl`, one line per vote, with the
forslag voted on, the vedtatt flag, the official totals, and the per-party
tally (the codes: 2 for, 3 mot, 1 ikke tilstede — see `verify_votes.py`).
`inventory.json` in the same directory lists every sak in the session and
how many votes it has.

**State, 2026-10-12.** The 2026–2027 session has 184 saker and 59 votes, in
two cases only (the trontaledebatt, 58; a committee change, 1). The
statsbudsjett-2027 cases (200434–200453) have **zero** votes: the
innstillinger and voteringer are in **December 2026**, after the collection
date. So the budget stream currently has its proposal half (Prop. 1 S, as
evidence 749–752 and on `/budsjett`) and will get its vote half in
December. Re-run `votes.py` on 200434–200453 then; the tooling is ready.

Verified: `verify_votes.py` over every vote-bearing evidence file — 410
files against Stortinget's official record, 0 MISMATCH (23 documented MIXED
where a party voted differently across several cited votes, 2 SPLIT).
