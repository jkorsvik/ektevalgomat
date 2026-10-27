# The standing loop

One iteration per fire. This file is the whole instruction: a session that
starts with "run one iteration of tools/autopilot/LOOP.md" must not need the
chat that created it. Read this, then `queue.yaml`, then work.

The goal it serves: a Norwegian political god's-eye view. For every party,
dimension and year, a sourced answer to "what did they do and how do we know".
Derived views over the dataset are generated, never hand-written.

## Roles

Different roles run in parallel because they fail differently. A scout cannot
ship anything, so it cannot fabricate content. A harvester drafts but decides
nothing. A verifier reads what a harvester wrote with the opposite intent. The
driver is the only role that can move a file into `internal/data/content/`.

| role | agent | may write | job | parallelism |
|---|---|---|---|---|
| **driver** | this session | content, queue, commits | gate, promote, verify, commit, report | one, never parallel with itself |
| **scout** | `ekv-scout` (inline) | nothing | find candidate sources and cases; leads only, never evidence YAML | up to 3 at once, different streams |
| **harvester** | `ekv-harvester` (thread) | `staging/<item-id>/` only | draft evidence against one queue item, per its method file | one thread per item, never two in one dir |
| **verifier** | `ekv-verifier` (thread) | `staging/<item-id>/VERIFIER.md` | falsify a batch: real quote attached to the wrong claim, magnitude not supported | one per batch, before promotion |
| **bias-analyst** | `bias-analyst` (thread) | `staging/<id>/` report only | measure coverage asymmetry per party, loaded wording, evidence strength | one, every few iterations |

Rules that hold for every role:

- Nothing is content until `gate.py` re-fetched the source itself. A worker's
  own "verified" is a receipt, never evidence.
- A zero result is a result: record what was searched and what was not found.
  Never invent a quote, a vote id, a number or a position.
- A cell with no defensible coding stays empty. `0` means "neutral", never
  "unknown".
- Never lower a gate or the acquiescence ratchet to make a change pass.

## One iteration

1. **Read state.** `git status` (must be clean or explain why),
   `tools/autopilot/queue.yaml`, and the last commit. Do not trust a summary.
2. **Pick.** The next `pending` item in queue order. Skip `blocked` and
   `failed` unless the blocker is cleared. One content item per iteration,
   plus its verifier.
3. **Fan out.** Spawn the roles that fit the item: harvester for the draft,
   scout for a stream with no item yet, bias-analyst on a rotating schedule.
   Give every spawn a self-contained task: absolute repo path, the method file,
   the item id, the statement id, the output contract, and the snapshot cutoff
   (`SnapshotDate` in `internal/server/viewmodel.go`; evidence dated after it
   fails `TestNoEvidenceAfterSnapshotDate`).
4. **Gate.** `uv run tools/autopilot/gate.py tools/autopilot/staging/<item-id>`
   (the full path: a bare name reads as "nothing staged" and exits 2). PASS
   only. A FAIL is a result, not a disaster: quote the error back to the worker
   or mark the item `failed`.
5. **Verify.** For any batch that carries a coded `position`, run
   `ekv-verifier` on it and read its verdict. `WRONG` means the file does not
   ship; move it to `staging/<item-id>/not-promoted/` with a `REASON.md`.
6. **Promote.** `uv run tools/autopilot/promote.py tools/autopilot/staging/<item-id>`,
   serially, one item at a time. It assigns numbers, rebuilds the archive when a
   quote is added, bumps the documented counts, runs `go test ./...`, and rolls
   back on failure. A pattern IDENTICAL line earlier in a file can absorb an
   edit: read the rollback, do not assume.
7. **Verify the whole set.** `uv run tools/verify/verify_excerpts.py` and
   `uv run tools/verify/verify_votes.py`. Both must end 0 failed / 0 mismatch.
8. **Commit.** One commit per risk. Follow the repo's existing style (imperative,
   sentence case, a body that says why). Never `git add -A`; name the paths.
   Do not push. Pushing, merging and deleting need the owner.
9. **Report.** Five lines in the session: what shipped, what is blocked, the
   exact ask that unblocks it, and the next item.

## The three gates promotion has, all learned the hard way

`promote.py` now handles all three. Do not re-derive them, but do not be
surprised by them either:

1. `TestDocsStateTheRealCounts` pins README and `docs/roadmap.md` to the real
   totals. `promote.py` rewrites them and advances their "Current state" date.
   `test_promote.py` reads the Go test's own claim table and fails if
   `promote.py` cannot rewrite a claim it asserts on.
2. `TestArchiveCoversEveryQuotedClaim` needs one stored window per quoted claim.
   `promote.py` rebuilds `internal/data/archive` when a promoted file carries an
   excerpt.
3. `TestWindowsStayBounded` caps a window at 6000 chars. A quote whose fragments
   sit far apart in the document produces one merged window and fails. Fix the
   quote, not the bound: quote the fragment that carries the claim and state the
   other figure in `claim_action`.

Two more that bite during collection rather than promotion:

- `TestAcquiescenceSpreadDoesNotGrow` is a ratchet. Coding a statement in only
  one direction can push the always-agree spread past 12 and block the whole
  batch. The fix is a statement phrased the opposite way, which is the owner's
  decision. Record the item as `blocked` with the measured spread.
- `TestNoEvidenceAfterSnapshotDate` compares event dates as strings. New events
  need `SnapshotDate` advanced, which is sanctioned: history is added, never
  rewritten.

## Where the work comes from

Ranked by value per hour. `queue.yaml` carries the concrete items.

1. **Voteringsspor, complete.** The spine. Every vote per representative, per
   session, joined to the statement it bears on. This is the only stream that
   gives a cell a second date, and 510 of 518 coded cells hold exactly one item.
   `tools/autopilot/votes.py` fetches a case's votes; `tools/autopilot/history.py`
   finds the earlier cases worth fetching.
2. **Government composition over time.** Who governed when, which minister, which
   party, from which date. One small authoritative table. Prevents a whole class
   of attribution errors: a 2026 proposition is an Ap-led government's, never
   "Ap/Sp", and Sp left on 2025-02-04.
3. **Alternative budgets and faginnstilling merknader.** What a party did when
   money had to be divided. Tier A, and today only five rows carry it.
4. **Party finance.** SSB's party-finance statistics, the Partiloven register and
   the parties' own accounts. Closes the anonym-financing dimension with numbers
   instead of documents.
5. **Representantforslag backfill.** Positions on questions that were never
   enacted, so no vote exists to cite.
6. **Election results per constituency, 2009-2025.** Seats, margins, overhang:
   the power map a reader can place themselves in.
7. **Committee membership over time.** Merknader are signed by named members; a
   merknad cannot be attributed without it.
8. **Ownership (Brønnøysund).** The actor layer: who owns the companies behind
   the money. The largest new build, and explicitly unbuilt today.
9. **Independent audits.** Riksrevisjonen (Dokument 3), ESA/EØS cases, court
   rulings. Third-party verdicts on whether a policy worked, the strongest tier
   there is.

## Refused on purpose

Record these as refused, with the reason, rather than leaving them as gaps:

- **Polling.** Not an action, not verifiable, and it rewards horserace reading.
- **Social-media sentiment and anonymous accounts.** The DPIA rule: work stops
  the day a case needs to name a private citizen or an unverified account.
- **Model summaries as evidence.** A summary is a receipt.
- **Press as tier A.** Context only; the tier ladder already says so.

## Failure modes this loop has already met

- A bare item name to `gate.py` reads as an empty item and exits 2.
- `voteringer?sakid=` carries neither the forslag text nor the tally: they live
  behind `voteringsforslag` and `voteringsresultat`. `votes.py` reads all three.
- The API speaks party codes (`A`, `KrF`, `SV`); the dataset speaks ids (`ap`,
  `krf`, `sv`). Translate in anything a human reads.
- The export's `sesjonid` filter is not disjoint: cases that span two sessions
  appear in both, so dedupe on votering id, never on sak id.
- `data.stortinget.no` rate-limits hard. The fetch cache under
  `~/.cache/ektevalgomat-sources/` is what makes a rerun cheap; a run that must
  re-fetch everything will look hung.
