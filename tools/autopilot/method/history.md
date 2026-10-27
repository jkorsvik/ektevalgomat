# Method: giving a coded cell a second date

A cell can only say whether a party stayed consistent if it has two dated
observations. Most do not. Measured from the content on 2026-10-27: 58 statements
carry coded cells, and 49 of them carry exactly ONE distinct date. The konsistens
score therefore reports "utilstrekkelig grunnlag" almost everywhere. That is not a
bug in the score, it is the shape of the data, and this method is the fix.

The second date comes from an earlier session's vote on the same question.

## The chain, every step a real API field

`tools/autopilot/history.py` walks it; do not re-implement it by hand:

```
evidence file --source_name-->  "Innst. 535 L (2024-2025)"
              --henvisning-->   sak id in that session's list
              --emne_liste-->   the case's subject ids
              --same emne-->    saker in EARLIER sessions
              --voteringer-->   the votes taken, with forslag text and tally
```

Run it for one statement:

```sh
uv run tools/autopilot/history.py --statement utsagn-54 --top 8
```

Output is a report under `tools/autopilot/reports/history/<today>/`, which is
gitignored. It is leads only. Nothing in that report is evidence.

A zero-candidate run is a RESULT: record what was searched and that nothing on the
same emne was found. Never widen the match by taste to manufacture a candidate,
and never invent a past position to fill the second date.

## Getting the votes for a candidate

```sh
uv run tools/autopilot/votes.py <sakid> [<sakid> ...]
```

Writes one JSON line per vote to `tools/autopilot/staging/votes/<sakid>.jsonl`:
`votering_id`, `forslag` (the proposal as the API states it), `vedtatt`, the
official totals, and the party tally. Two honest states it reports rather than
hides: `votering_resultat_type` 5 ("Enstemmig vedtatt") means no vote was taken,
so the tally is empty and the totals are `-1`; and a case with 58 votes is not
worth the fetch budget, which is why `history.py` caps at 12 per case.

Party codes must be translated before a human reads them: the API says `A`, `KrF`,
`R`; the dataset says `ap`, `krf`, `rodt`. `history.py` and `verify_votes.py` both
do this. Anything you write by hand must too.

## Writing the evidence file, one per party per vote

One file per party per vote, `evidence_type: voteringsspor`. The gate requires
every field below to be present and non-empty (`party_id` may be empty for an
institutional row):

```
id                 ev-<date>-<party>-<slug>
statement_id       the cell this gives a second date
party_id           the party, or "" for an institutional row
actor              the party's name as the record states it
date               the vote's date (YYYY-MM-DD), NOT the read date
evidence_type      voteringsspor
evidence_class     fakta | partiets_begrunnelse | analyse | uklart
scan_summary       one line: what happened
claim_action       what the party actually did, naming the forslag it bore on
verbatim_excerpt   the forslag fragment that carries the claim
status             vedtatt | forkastet | stemte_for | stemte_mot | ...
position           -2..2, or absent when no defensible coding exists
source_tier        A (the vote record is the primary source)
source_name        the innstilling or voteringsdetaljer reference
source_url         the page the excerpt is read from
claim_confidence   dokumentert | sannsynlig | sannsynliggjort | usikkert | ukjent
context_note       why this reading, and what the other votes in the case were
retrieved_at       the date you read it, never earlier than the source's own date
```

## Rules the gate will enforce, and the ones it cannot

- **`status: stemte_for`/`stemte_mot` must cite a votering id** in `source_name`,
  `source_url` or `context_note`, and the id must match the official record. The
  gate re-fetches and compares.
- **A `verbatim_excerpt` must be a substring of what `source_url` serves today.**
  The gate fetches the page, not the API. So cite the votering-detaljer page the
  excerpt actually came from, and keep the excerpt to the fragment that carries
  the claim. Fragments far apart in one document merge into one window and fail
  `TestWindowsStayBounded` at 6000 chars; the fix is the quote, not the bound.
- **Read the tilråding inverted.** A vote on "Komiteens tilråding: ... – vedtas
  ikke" means voting FOR the tilråding is voting AGAINST the proposal.
  `verify_votes.rejects_proposal` applies this; apply it yourself too.
- **A vote on a different paragraph is a different question.** Two votes in one
  case are one question only when their forslag text addresses the same thing.
  Read both forslag before deciding they belong to the same cell.
- **A split is not a position.** 3 for and 3 against states that the party did
  not take one side; record it as such and code nothing from it.
- **Changing an existing cell's `position` is the owner's call**, never a
  worker's. Report the recommendation with its vote and stop.
- **Evidence dated after `SnapshotDate`** (`internal/server/viewmodel.go`, today
  `2026-10-12`) fails `TestNoEvidenceAfterSnapshotDate`. An EARLIER session's
  vote satisfies this trivially, which is the point of this method.
- **The acquiescence ratchet is not lowered.** If coding the new rows pushes the
  always-agree spread over 12, the item is `blocked` with the measured spread; the
  fix is a statement phrased the opposite way, which is the owner's decision.

## Before promotion

`uv run tools/autopilot/gate.py tools/autopilot/staging/<item-id>` (the full path;
a bare name reads as an empty item and exits 2). PASS only. Then a verifier reads
the batch with the opposite intent, then `promote.py` runs serially. A FAIL is a
result: quote it back and fix the files, or mark the item `failed`.
