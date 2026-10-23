# Research briefs — delegating evidence work to other AI tools

This folder is a **handoff package**, not app content — nothing here is
read by the server. It exists so evidence research for the remaining 37 of
42 statements (and deepening the 5 already live) can be delegated to
whatever AI tools you have access to, run in parallel, outside this
project's own coding-agent session.

## Files

- **`CONTEXT.md`** — the shared "constitution": schema, source-tier ladder,
  attribution-confidence ladder, party list, verification discipline, a
  worked example, and the two real mistakes this project already made
  (a fabricated case ID, a fabricated dimension name) told as concrete
  cautionary examples. Paste this into every delegated session — it's the
  part that keeps independent research sessions from re-making either
  mistake.
- **`TASKS.md`** — the actual work, broken into ten independently-delegable
  units (deepen statements 1-5, finish three partial leads, then one task
  per remaining dimension).

## How to run this

1. Pick one task from `TASKS.md`.
2. Open a new session in whatever AI tool you're using (one with live web
   search/browsing — this doesn't work as a pure-memory chat, the whole
   point is fetching and re-verifying real sources).
3. Paste `CONTEXT.md` in full, then the one task block from `TASKS.md`, as
   your first message (or two messages back to back).
4. Let it work. Ask it to show every source URL it opened, not just the
   ones it ended up citing — that's what makes your own spot-check fast.
5. Bring the output back here (paste it into a RAIWork/coding-agent
   session, or drop it in a file) for review against its cited sources
   before anything is merged into `internal/data/content/`.

Running several tasks in parallel across different tools (or several
sessions of the same tool) is the point — each dimension is independent,
and the whole reason this is being delegated is that the source memo itself
estimates this as 70-90 person-weeks of work (see `../research-memo.md`).
Nothing here needs to happen in the numbered order in `TASKS.md` except
Task 0 and Task A being the most bounded, highest-confidence starting
points.

## What "done" looks like for one task

Per statement: drafted text (if it was undrafted) grounded in a real,
verified-current topic, one YAML evidence block per finding across all nine
parties, and an explicit list of which parties came back empty. Not every
statement will get nine parties' worth of evidence — a thorough, honest
search that finds real evidence for three parties and genuinely nothing for
six is a complete, useful result. A search that manufactures something for
all nine to look complete is the failure mode this whole package exists to
prevent.

## Before you merge anything

Spot-check a handful of the returned source URLs yourself — open them,
confirm the case number, date, and quoted text actually match. This project
caught its own fabricated case-ID and fabricated dimension-name mistakes
exactly this way: not by trusting a confident-sounding writeup, but by
re-opening the primary source a second time and checking it said what was
claimed.
