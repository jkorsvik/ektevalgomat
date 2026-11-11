# The two-track loop

`LOOP.md` is one iteration of the evidence loop. This file runs two tracks in
parallel, because they now have different bottlenecks.

- **Track H — historikk.** The cells are deep enough on one date and empty on
  the next. The raw material exists: 25 773 votes across 16 sessions. The work
  is to give every coded cell a second and third dated observation.
- **Track L — lobby og personer.** A new domain. Who the elected people are,
  what roles they hold over time, which companies and organisations they are
  tied to, and which organisations get heard. This is person data, so it has a
  different model, a different gate and a different boundary.

The two tracks share one driver, one stop rule set, and one rule: nothing is
content until a mechanical gate re-fetches the source itself.

## Roles

| role | agent | may write | job | model |
|---|---|---|---|---|
| **driver** | this session | queue, content, commits | gate, promote, verify, commit, report | any |
| **scout** | `ekv-scout` | nothing | leads for lane H | deepseek |
| **historian** | `ekv-historian` (new) | `staging/<item-id>/` | turn history.py candidates into per-party vote evidence | deepseek |
| **verifier** | `ekv-verifier` | `staging/<item-id>/VERIFIER.md` | falsify a lane-H batch | deepseek |
| **lobby-scout** | `ekv-lobby-scout` (new) | `lobby/reports/` | find and validate sources; leads only, no person records | **opus** |
| **person-tracker** | `ekv-person-tracker` (new) | `lobby/staging/<person-id>/` | record one person's dated role timeline | **opus** |
| **person-verifier** | `ekv-person-verifier` (new) | `lobby/staging/<person-id>/VERIFIER.md` | falsify a person record | **opus** |
| **bias-analyst** | `bias-analyst` | `staging/<id>/` report | coverage asymmetry | any |

**The model split is not a preference.** Lane L handles named individuals.
`USER.md` forbids DeepSeek and GLM from seeing PII. Every lane-L agent is
pinned to `anthropic/claude-opus-5-5`. The driving session must not read a
person record either: it sees counts, ids and gate results, never names.
Lane H is parties and votes only, so deepseek is fine there.

## One parallel iteration

Run the two lanes at the same time, then join.

1. **Read state.** `git status` clean or explained, `git log -1`, both queues
   (`queue-history.yaml`, `lobby/queue.yaml`) and the last report.
2. **Pick.** One item per lane. Skip `blocked` and `failed` unless the blocker
   is cleared in the file itself.
3. **Fan out, both lanes at once.**
   - H: spawn `ekv-historian` for the next statement needing a second date.
   - L: spawn `ekv-person-tracker` for the next person in scope, and
     `ekv-lobby-scout` on a rotating source-discovery item.
   - Never two writers in one staging directory.
4. **Gate each lane.** `gate.py` for H; `gate_people.py` for L. PASS only.
   A FAIL is a result; quote it back or mark the item `failed`.
5. **Verify each lane.** `ekv-verifier` for H, `ekv-person-verifier` for L.
   `WRONG` means the file does not ship.
6. **Promote serially.** One item at a time, per lane. `promote.py` for H.
   `promote_people.py` for L, and only after the Go person schema exists.
7. **Join.** Re-run `tools/verify/verify_excerpts.py` and
   `tools/verify/verify_votes.py`, then `go test ./... -count=1`.
8. **Measure.** `uv run tools/analytics/trends.py` and read the two numbers:
   cells with two or more dates, and persons with two or more dated roles.
9. **Commit.** One commit per risk. Name the paths. Do not push.
10. **Report.** What shipped per lane, what is blocked, the exact ask, next item.

## Stop rules, both lanes

- A zero result is a result. Record what was searched and what was not found.
- Never invent a vote id, a case id, a role, a company or a date.
- A split vote is not a position. An empty cell stays empty. `0` means neutral.
- Never lower a gate or the acquiescence ratchet to make a change pass.
- Every lane-L fact needs `source_url`, `date` and `retrieved_at`. No source,
  no record.

## Lane L boundary (DPIA)

The site already says it never names a private citizen. Lane L keeps that:

- **In scope:** sitting and former Storting representatives, sitting and former
  ministers and state secretaries, in their public roles.
- **Out of scope:** private citizens, family members, private addresses,
  health, private finances beyond declared interests, and any unverified
  account.
- **Each record:** a public role with a start date and, where the source gives
  it, an end date. Declared interests only from an official register.
- **Deletion:** a person may ask for removal. The record is deleted, and the
  request is logged in `lobby/corrections/`.

## Where the work comes from

Lane H, ranked by value per hour:
1. The 29 cells that already have two dates: verify the series before widening.
2. The 488 cells with one date: one historical vote each, earliest session first.
3. Cells whose only evidence is a party merknad: a vote is the stronger date.

Lane L, in order:
1. The source catalog: which of the candidate registers actually answer, and
   what each one yields (`lobby/sources.md`).
2. The roster: 169 sitting representatives, party and constituency.
3. Committee and role history per person.
4. Hearings and the organisations heard: access, not opinion.
5. Declared interests and company roles, from the official registers.
6. Job changes over time: a new role with a start date is a fact.

## Failure modes this loop has already met

- `gate.py` with a bare item name reads as empty and exits 2. Pass the full path.
- `history.py` is leads only. A candidate case is not evidence until a human
  opens the vote and reads the forslag.
- A historical vote is a second date only when its forslag asks the same
  question. Two votes in one case can be two questions.
- Stortinget's `person` export carries no `verv`. Role history is a separate
  page or register; the lobby-scout must prove which before lane L scales.
- Brreg gives roles per company (`/enheter/{orgnr}/roller`), not companies per
  person. A person-to-company reverse lookup is a discovery item, not an
  assumption.

## Lane H, measured on 2026-11-09

- 517 coded cells: 488 hold one distinct year, 21 hold two, 8 hold three or more.
- 29 cells (6%) can speak about consistency today.
- 57 statements need a second date. 55 of them walk through `history.py` from an
  `Innst.` reference. 2 (utsagn-02, utsagn-45) cite only `votering_id`, and there
  is no `votering?voteringid=` endpoint to reverse a vote to its case. Those two
  are `votering-only` in the queue and stay blocked until that route is built.
- `tools/analytics/trends.py` prints the same numbers, so nobody re-measures by
  memory.

## Lane L, measured on 2026-11-09

- Roster: 169 sitting representatives, fetched with party and constituency.
- Hearings: 351 in session 2026-2027, with committee and case; participants are
  not in the list payload.
- Company roles: `data.brreg.no/enhetsregisteret/api/enheter/{orgnr}/roller`
  answers 200. Person-to-company reverse is not this endpoint.
- Minister appointments: 245 dated rows already in the repo.
- The person gate passes a valid record and rejects a planted record on all six
  checks (scope, forbidden field, missing retrieved_at, tier C, to-before-from,
  missing orgnr).

## Owner decisions, 2026-11-09

- **Lane L starts with the verv/economic-interest register.** That is the first
  source the lobby-scout must prove; hearings and Brreg company roles follow.
- **Cadence is manual first.** No schedule runs the loop. A human fires one
  iteration, reads the report, and decides whether to continue.

## Measured state, 2026-11-09 (lane L first full pass)

The verv register is one dated PDF for the whole Storting, so it is parsed once
into `internal/data/verv/index.json` by `tools/analytics/build_verv_index.py`:
**267 people, 201 with roles, 371 role entries**, snapshot 2026-10-30. Sections
§5–§11 are dropped at index time and counted in `counts.dropped_sections`; §9
keeps company names only, with share counts and kroner stripped.

All **169 sitting representatives** now have a record in
`internal/data/content/people/`, written from the roster (mandate), the verv
index (§2 posts) and `ministers.yaml` (statsråd). **525 dated roles**, of which
115 people have at least one verv and 26 have held government office.

Three defect classes were found by adversarial sampling (3 of 12 wrong) and then
scanned across all 169:

1. `party_id` held the roster API code (`A`) instead of the dataset id (`ap`):
   40 records.
2. A `statsrad` role copied department and dates but not the post: 8 roles.
3. A `verv` organisation was clipped, or dated with the snapshot although the
   register states a year range: 2 joins, 15 dates. Where the register states
   "2022 - 2025" or "Fra 29. oktober 2026", that date is now used; the snapshot
   date is kept only where the register states nothing, and the page labels it
   `registrert`.

The page is `/personer`: a full table, one row per role, every row citing its
source. A mandate with no end date reads `sitter`; a register role with no end
date reads `ikke avsluttet i registeret`, because the register going quiet is
not the same as the post still running.

Gates: `gate_people.py --no-fetch` 169/169 PASS; `TestNoForbiddenFieldInPeopleFiles`
scans the raw YAML; `TestRegisterSnapshotMatchesIndex` fails if the pinned
snapshot drifts from the index; `TestPersonerPageShowsEveryRecord` fails if a
record cannot be reached from the page.

Not done yet: Brreg company roles (needs the bulk `roller/totalbestand`
inversion on name + birth date), per-person hearing participation, and the
`statssekretar` scope. Lane H's open leads are unchanged.

## Register history, 2026-11-09

A verv role is snapshot-dated because the register is a snapshot. Two dated
editions turn that into a real answer without inventing a date. The National
Library's web archive lists earlier editions; `pr-19-april-2024.pdf` is the
oldest usable copy we could fetch (the archive blocks bulk requests as bot
traffic, so this is one edition per run, not a batch).

`tools/analytics/build_verv_history.py` compares two editions and writes
`internal/data/content/verv/history.json`. Between 2024-04-19 and 2026-10-30:
123 people appear in both, 50 verv carried, **45 verv gone**, 44 new.
The page shows the 45 as "står i 2024-04-19, ikke i 2026-10-30" and never as an
end date, because the register publishes none.

To extend: add older editions as `--snapshot DATE PDF` pairs. Each run compares
the oldest and newest only, so adding one edition refines the interval rather
than multiplying tables.

## Measured dead ends, 2026-11-11

Recorded so the next session does not spend a day rediscovering them. Each was
probed, not assumed.

**Committee membership is not in the API.** `komiteer?sesjonid=` returns 16
committees with only `id`, `navn` and `versjon` — no member list. Asking for one
committee by id fails with the server's own error:

```
GET /eksport/komite?id=ARBSOS&format=json
{"feilkode":0,"feilmelding":"Forespoersel med ukjent parameter:
 'Metode=HentKomiteMedlemmer' feilet.","metode":""}
```

The message names the internal method the request reached, so the data exists
behind a route that cannot be called. `komitemedlemmer` is 404. The verv register
does not carry committee posts either, and only 8 of 169 person records mention
the word "komite" at all, all incidentally. So committee membership is absent,
and it is absent because the source withholds it, not because nobody built the
loader.

**Subject tags cannot find the earlier document.** Measured across 15 sessions:
7,124 innstillinger have a resolvable PDF URL, but only 165 distinct subject tags
exist, and the best overlap between a statement's own wording and an earlier
innstilling's tags plus short title is 4. Nothing reaches 5, and 6 of 13
statements score zero. `tools/analytics/gen_merknad_candidates.py` enumerates a
session; it does not locate a document that asks the same question, and its
docstring says so.

**Sak relations do not cross sessions.** `sak_relasjon_liste` links siblings
inside one session — four saker rows share innstilling 16082 — and never a
previous session's version of the same case.

The consequence for lane H is the important one: breadth is solved and cheap,
selection needs a reader. That is the whole reason the source feeder exists.

## Open rule question: utsagn-25 / rodt, reverted 2026-11-11

A row dated 2023-11-27 was promoted as +1 at 13:22:44 and withdrawn by revert at
13:23:34, fifty seconds later (66b6abd, then 038cf94; times read from the commit
records, not estimated). The reason it was withdrawn is worth more than the row.

My argument for +1 had two legs. One was affirmative: Rødt 2023 proposes
phasing by covering part of the bill for those with the largest costs, 40 pct.
refund above a 2,500 kr deductible, more than 3 bn for the first step, which is
the statement's own instrument. The other leg was not: I said the text does not
commit to a universal right by 2030. The source feeder named that as the defect.
Absence of a commitment is not evidence for the statement's negative clause, and
agreement with the non-discriminating half of a statement is not agreement with
the statement. A row cannot rest on what a document fails to say.

The disputed sentence is «Målet er å likestille tannhelse med andre
helsetjenester i løpet av fire år», next to «offentlig finansiert gratis
tannhelse» and «gratis tannhelse for alle». Three readings are defensible and
they point in different directions:

1. It is a universal commitment with a deadline earlier than 2030, so -1.
   This is what both blind reviewers read.
2. It is fiscal equalisation, how dental care is paid for, not a right, so the
   sentence says nothing about the negative clause and the row is uncoded.
3. The mechanism is the coded clause, so +1, which is what the sibling row
   ev-2025-12-08-rodt-tannhelse-innst11s already does.

The sibling makes this urgent rather than academic: it codes +1, and its own
note leans on Rødt not binding itself to 2030, which is the same absence defect.
So the pair cannot be read with one rule until that question is answered, and
until it is, the proposed 2023 observation stays withheld. The cell is not
uncoded: the 2025 row at +1 is published and stands. What is withheld is the
second date, and with it both the reversal and the consistency claim that date
would produce. Publishing either now would be a coin flip dressed as evidence.

The rule to settle first: does the corpus code the discriminating clause of a
statement, or the mechanism it names? Pick one, apply it to both rows, and
re-examine the sibling before admitting anything here.

## Scope of the clause rule, measured and then corrected 2026-11-11

A first attempt at scoping this question reported that only one statement
carried a two-clause structure. That was wrong, and the way it was wrong matters
more than the number.

WHAT WAS ACTUALLY MEASURED. Every `text` field was scanned for four literal verbs
(«ikke forplikte», «ikke binde», «ikke pålegge», «ikke love») plus a stated
deadline («innen 2030», «innen N år», «senest»). Under those two literal
patterns, exactly one statement matched both: utsagn-25. One more, utsagn-26, has
a deadline and no negative half, and its instrument and commitment point the same
way.

WHY THAT SUPPORTS NOTHING. The scan's scope was four verbs. The report drawn from
it was semantic: «the only two-clause statement», «the rule touches nothing
else». Those are different claims, and the second is false. The source feeder
read all sixty texts and named seven counterexamples, each verified here:

- 04  «... er tilstrekkelig, uten et eget oppdrag ...»
- 13  «... vedta lovendringer ..., ikke vente på en stortingsmelding.»
- 18  «... videreføre dagens hjemmelslengde ..., ikke gjeninnføre et strengere
  28-meter meterskille.»
- 20  «... flere leverandører, ikke en rask, tidfestet exit-strategi ...»
- 21  «... selv avgjøre sin organisasjonsform, uten en bindende føring ...»
- 31  «... i de årlige statsbudsjettene, ikke lovfestes med en automatisk årlig
  prisjustering.»
- 37  «... utredes og skje gradvis, ikke vedtas nå som en økning fra 2 til
  4 prosent.»

So the contrastive form «do X, not Y» is the normal shape of these statements,
not an exception. The clause question is corpus-wide.

WHAT THIS COSTS. The earlier reading priced the rule at one statement and made a
per-statement field look like machinery for a single case. With at least eight
statements carrying a real contrast, that reasoning is void: which clause
discriminates has to be decided for many statements, and a per-statement field is
no longer disproportionate. The decision is the user's and the options are
unchanged; only their price was wrong.

MEASURED PROPERLY, as a count rather than a conclusion. Counting the literal
constructions `, ikke ` and `, uten ` across all 60 `text` fields returns 9
statements: 04, 07, 13, 18, 20, 21, 25, 31, 37. That is a lower bound on the
contrastive form, and it is the number to use when pricing the rule. Nine of
sixty is a review pass, not a schema change and not a one-off.

One further trap, from the source feeder's own scan: a stem match on `lov\w*`
catches `lovfestes` and would inflate the count. Word boundaries are what make
the count honest, whichever pattern is used.

A word scan cannot answer a question about meaning. If scoping this again, read
the texts, or count the «, ikke » and «, uten » constructions as patterns to
review rather than as a complete set.
