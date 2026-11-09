# Method: the historian, one statement at a time

`method/history.md` says how to write ONE second-date evidence file. This
method says how the historian works a whole statement: how to turn the leads in
`tools/autopilot/reports/history/<today>/candidates.json` into a set of dated
observations for every party that has a coded cell on that statement.

Read `method/history.md` first. Everything there still applies.

## The one outcome

For statement S, add one evidence row per party per historical vote, where the
vote asks the same question as the statement. After promotion, a cell on S has
two or more distinct `date` years, and the konsistens label can say
"Samme posisjon i YYYY og YYYY" or "Endret posisjon mellom YYYY og YYYY".

A statement where no earlier session voted on the same question is a result:
record it with `outcome: no-earlier-vote` and stop. Do not widen the match.

## Input

The queue item gives: `statement`, and the path to today's
`candidates.json` if the scout already ran. If it did not, run:

```sh
uv run tools/autopilot/history.py --statement <id> --top 8
```

The report is leads. A candidate is admitted only when all of these are true:

1. It sits in a session strictly earlier than the earliest session the cell's
   own evidence already cites.
2. The forslag text of the vote overlaps the statement's subject. Title overlap
   alone is not enough; `history.py` already drops the cases whose vote text
   never mentions the subject, and you must not re-admit them by taste.
3. The vote actually divides the parties. A unanimous vote (`resultat_type` 5)
   has no tally and says nothing about a party's position.

## Per admitted vote

Fetch the votes for the case:

```sh
uv run tools/autopilot/votes.py <sakid>
```

Then, for each party with a coded cell on S:

- Read the tally for that party: `for` / `mot` / `ikke_tilstede`.
- Read the forslag. Decide which side of the statement a FOR vote is on. A vote
  on "Komiteens tilråding: ... – vedtas ikke" is inverted: FOR is AGAINST the
  proposal. `verify_votes.rejects_proposal` applies the same rule.
- Code the position only when the forslag asks the statement's question. If it
  asks an adjacent question, do not code it; write the row without `position`
  and say why in `context_note`, or skip it.
- A split (some for, some mot within one party) is not a position. Record it as
  split and code nothing.

Write one YAML per party per vote into `staging/<item-id>/`, with the field set
`method/history.md` lists. Use `evidence_type: voteringsspor`. The `date` is the
vote date, never the read date. `source_tier: A`.

## Two routes, and one that is not built

The queue item carries `route`:

- `route: innstilling` — the cell's evidence cites an `Innst. N X (YYYY-YYYY)`.
  `history.py` resolves it to a case and walks back by subject. This is the
  normal route, and 55 of 57 statements have it.
- `route: votering-only` — the cell cites a `votering_id` and no innstilling
  (utsagn-02, utsagn-45). `history.py` resolves nothing for these: 0 of 0
  innstilling references. There is no `votering?voteringid=` endpoint to reverse
  a vote to its case. Do not run `history.py` on this route and call the empty
  result a "no earlier vote". Mark the item `blocked` with
  `blocked_on: votering-to-sak reverse route not built`, and stop.

## What you must not do


- Do not change an existing cell's `position`. A historical vote ADDS a dated
  observation; it never overwrites the newest one. If the vote contradicts the
  current position, that is exactly the "Endret posisjon" result, and the
  arithmetic in `internal/scoring` will show it from the dates.
- Do not code a party that has no cell on S. A second date is for an existing
  cell.
- Do not touch `internal/data/content/`. Write only under `staging/<item-id>/`.
- Do not lower the acquiescence ratchet. If the batch pushes the always-agree
  spread over 12, stop and mark the item `blocked` with the measured spread.

## Output contract

`staging/<item-id>/` contains:

- one `<party>-<votering_id>.yaml` per admitted party-vote, and
- `meta.yaml` with `id`, `kind: voteringsspor`, `statement`, `votes_read`,
  `votes_admitted`, `outcome` (`added`, `no-earlier-vote`, `blocked`), and a
  one-line `note`.

Then the driver runs `gate.py`, `ekv-verifier`, and `promote.py`. Nothing you
write is content until those three pass.
