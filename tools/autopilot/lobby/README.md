# Lane L: lobby and persons

This directory is the second track. It answers a different question from the
evidence track:

> Who are the elected people, what roles do they hold over time, which
> organisations are they tied to, and which organisations actually get heard?

It complements the evidence track. That track says what a *party* did on a
statement. This track says what the *people* in those parties do, and who has
access to them.

## Why it is separate

Person data is not party data. It has a different source, a different gate and
a legal boundary. The site already promises never to name a private citizen.
This track keeps that promise with a mechanical gate (`gate_people.py`), not
with a reader's good intentions.

## Scope

`scope.yaml` is the rule. In one sentence: **public office holders, in their
public roles only.**

- Sitting and former Storting representatives.
- Sitting and former ministers and state secretaries.
- Their committee seats, board seats, declared interests and company roles,
  each with a public source and a date.
- Which organisations were invited to a hearing, and on which case.

Out: private citizens, family, addresses, health, private finance, unverified
accounts. A record with no public source is not a record.

## The model split

`USER.md` forbids DeepSeek and GLM from seeing PII. Every agent in this lane is
pinned to `anthropic/claude-opus-5-5`. The driving session reads counts, ids and
gate results, never names.

## Files

| file | what it is |
|---|---|
| `scope.yaml` | who is in scope, and the forbidden fields |
| `sources.md` | verified sources, what each yields, what still needs discovery |
| `schema.md` | the person record: fields and invariants |
| `gate_people.py` | the mechanical privacy and provenance gate |
| `promote_people.py` | moves a passed record into content, namespaced |
| `queue.yaml` | one item per person to track, generated from the roster |
| `staging/` | worker output, gitignored, never content |
| `reports/` | scout output, gitignored |

## The chain

```
roster (Stortinget representanter, government ministers)
  -> queue item per person
  -> ekv-person-tracker reads official registers and writes roles.yaml
  -> gate_people.py re-checks scope, provenance and forbidden fields
  -> ekv-person-verifier re-reads each role against its source
  -> promote_people.py writes content/people/<id>.yaml
```

Nothing reaches content until the gate passes. A hunter's "verified" is a
receipt, not evidence.

## Current state

The source catalog is deliberately honest about what is proven and what is not:

- **Proven:** Stortinget roster (`representanter?sesjonid=`, 169 rows with party
  and constituency), hearings (`horinger?sesjonid=`, 351 per session, with
  committee and case), and company roles (`data.brreg.no` `/enheter/{orgnr}/roller`).
- **Seed already in the repo:** `internal/data/content/government/ministers.yaml`,
  245 dated minister appointments with source URLs.
- **Needs discovery, not assumption:** the per-representative "verv og
  økonomiske interesser" register, the person-to-company reverse lookup, and the
  legal status of a Norwegian lobby register. The lobby-scout proves each before
  the tracker depends on it.

Run the lane with `tools/autopilot/LOOP-TRACKS.md`.
