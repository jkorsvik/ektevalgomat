# Roadmap — what's built, what isn't, and why

`research-memo.md` is the original design source: a 12-week, 8-9 FTE,
70-90 person-week institutional research project. This repo builds a
narrow, real slice of it, run solo on a ZimaBoard (Celeron N3450). This
document exists so that gap never goes unstated — "ekte" (real) means
every claim is honestly sourced, and that discipline has to apply to
this project's own scope claims too, including its own mistakes (see
"Correction" below).

## Current state (2026-10-27)

Read this section first. The phase entries below are the build log, in
order, and their "FINAL STATE" lines were true on the day they were
written, not today.

- **60 statements, 619 evidence files.** Memo statements 1-42, two
  "Tillegg" extras (44, 45), two dimensions the owner approved after
  user feedback: "Tilgang, integrering og rettsstat" (46-51) and
  "Finansiering av fellesskapet" (52-58), and utsagn-59 in "Norsk
  handlefrihet". There is no utsagn 43 (see "Phase 2r").
- **54 of 60 statements have a coded position for all 9 parties.** The
  six that do not, each searched and left honestly empty (see "Phase 4b"):
  utsagn-06 7/9 (deliberate, see "Gap-fill pass"), utsagn-52 8/9 (Ap:
  "formuesskatt" appears nowhere in Innst. 2 S, Innst. 3 S or Ap's
  program), utsagn-55 7/9 (Sp and MDG have no total-level claim; R coded -2
  from its program, file 603), utsagn-57 8/9 (Ap: only a Sp minister's
  sentence about Ap, file 710, which does not set Ap's position), and
  utsagn-60 and 61, the two statsbudsjett statements, which carry
  institutional evidence only and no party positions by design.
  utsagn-45 was 7/9 when this block was first written and has since been
  coded for R and Sp, so it is not on this list any more; the phase entries
  below still describe it as it stood on their own date.
- **Every source claim is machine-checked.** `tools/verify/verify_excerpts.py`
  fetches each source (full document) and requires every
  `verbatim_excerpt` to be in it: 309 checked, 0 failed (2026-10-03).
  `tools/verify/verify_votes.py` compares every stemte_for/mot with the
  official record on data.stortinget.no: 410 checked, 0 mismatch, every
  one cites a vote id. Both need network, so they are not in `go test`.
  Run both before merging evidence changes.
- **Statement form.** Every statement is one claim, max 200 characters,
  no dash, no "selv om"/"fordi", and has a "Hva betyr et svar i
  praksis?" box (`practical_stakes`). Enforced by
  `internal/data/statement_form_test.go`.
- **Direction balance.** 13 statements are phrased in the opposite
  direction (`FLIPPED_2026_09_27` in `quiz.js`, which also migrates
  returning users' stored answers). A user who answers the same on
  everything now sees a 12-point gap between best and worst party (was
  41). `internal/scoring/acquiescence_test.go` is a ratchet on that
  number. `/metode` shows both numbers, and
  `internal/server/consistency_test.go` fails if its text drifts from
  the code.
- **Snapshot.** The badge reads "Øyeblikksbilde: hendelser til og med
  2026-10-12". Sources can be retrieved later (most were), but no
  evidence event may be dated after the cutoff. `TestNoEvidenceAfterSnapshotDate`
  enforces it.
- **Statement page headline** counts coded positions ("Kodet posisjon
  for N av 9 partier"), not parties with any source.
- **Corrections are published on the site.** `/om` carries a dated
  corrections log (`internal/data/content/corrections/`), seven entries
  taken from this build log: five `rettelse` (something a reader could
  see was wrong and changed) and two `nesten` (caught before it reached
  the site). This log plus this file are the readable public record, and
  the gate below holds the count. The repository's git history starts at
  the initial public release (`fbc64d8`, 2026-10-23): every change since is
  a commit, and nothing before the release is published. Commits are still
  not the record a reader is pointed to, because a reader cannot weigh one.
  `internal/data/corrections_test.go` and `TestOmPublishesEveryCorrection`
  are the gate.

## Bias baseline (2026-10-27) — what the content does unevenly

A measurement of the content itself, not of the site. Every figure below was
recomputed in the driving session from `internal/data/content/evidence/*.yaml`;
the worker's report and its script are in `tools/autopilot/staging/bias-baseline/`,
which is not versioned, so the numbers live here.

**Coverage is symmetric.** All nine parties cover 56-58 of 60 statements, and 54
of 60 statements are coded for all nine. No party is systematically thinner than
another. The asymmetry is elsewhere.

**Quote coverage is not symmetric, and this is the strongest finding.** Rows that
carry a verbatim excerpt, per party:

| SV | H | FrP | Ap | Rødt | MDG | Sp | V | KrF |
|---|---|---|---|---|---|---|---|---|
| 71.0 % | 68.9 % | 63.3 % | 58.3 % | 56.7 % | 56.7 % | 53.3 % | 33.9 % | 33.9 % |

SV's record rests on a quote twice as often as Venstre's and KrF's. This is a
collection bias, not a collection gap: the evidence exists, the sentence in the
source was not found for those parties. Closing it is excerpt-tail work with a
party filter (V and KrF), not new research.

**Direction is uneven by party.** Ap has 29 of 57 coded positions at −1 and only 5
at +2; SV, H and Rødt each have 22 at +2. Whether that is the parties or the
statements is not answerable from the data, but a reader comparing parties should
know the shape before reading the means.

**One open coding question.** Two `faktakontekst` rows carry a coded position
(0) with no party: `ev-2025-10-14-politisk-reklame-kontekst` (utsagn-02, Stortinget
i plenum) and `ev-2023-03-14-utenlandsk-paavirkning-kontekst` (utsagn-04,
Justiskomiteen). A body-of-the-Storting context row codes nobody, so either the
position field should be empty or every view must keep excluding it. The coded-cell
count is 518 or 516 depending on the answer, and the two figures appear in different
places today. This needs an owner decision, not a silent fix.

**The acquiescence spread stays at 12** (always-agree and always-disagree both 12
points), confirmed by the Go test. Leave-one-out points at utsagn-16, utsagn-39 and
utsagn-11 as the largest contributors to it.

## Budget lead closed with a negative result (2026-10-27) — Enova

The queue carried a lead: "Enova is cut while the government calls energy
efficiency a main argument for Norgespris." Half of it holds and half of it does
not, and the half that does not is the reason there is no evidence row.

**Verified.** Prop. 1 S (2025–2026) Energidepartementet, page 96, kap. 1825 post 50
Klima- og energifondet, in the proposition's own three columns:

| Rekneskap 2024 | Saldert budsjett 2025 | Forslag 2026 |
|---|---|---|
| 1 130 000 | 1 377 000 | 1 300 000 |

Thousands of kroner, so 1 130 / 1 377 / 1 300 mill. The proposed 2026 transfer is
**77 mill. kroner lower than the settled 2025 budget, a cut of 5.6 %**, and the
fund is administered by Enova SF under a steering agreement with the ministry. The
proposition's own sentence:

> Post 50 Klima- og energifondet
> Departementet foreslår ei samla løyving til Klima- og energifondet på 1 300
> millionar kroner, inkludert inntekter frå påslaget på nettariffen på anslagsvis
> 690 millionar kroner […]. Midlane skal forvaltast i samsvar med måla og føringane
> i styringsavtalen mellom Enova SF og Klima- og miljødepartementet.

**Not verified.** That the government presents energy efficiency as a main argument
for Norgespris. Neither budget proposition contains a sentence linking the two:
searched Prop. 1 S (2025–2026), 202 pages, and Prop. 1 S (2026–2027), 204 pages, for
a sentence containing both "Noregspris" and "energieffektivisering" — zero matches
in each. The lead asserted a connection the source does not make. It is recorded
here as wrong rather than argued back to true.

**Why there is no evidence row.** The verified fact is about energy-efficiency
funding. Utsagn-12 asks whether the state should use support and reference-price
schemes rather than a statutory maximum price, and a cut to the efficiency fund does
not bear on that claim. No statement in the dataset mentions energy-efficiency
funding at all (searched the 60 statements for "energieffektivisering": none), so
the finding has no home. Filing it under utsagn-12 would have padded a statement
with a fact it does not speak to, which is the failure mode this project exists to
avoid. It needs a statement, and that is the owner's decision — the same shape as
`budget-vindkraft-kommune`.

## Phase 1 — Demokratisk åpenhet (built and live)


- One dimension: **Demokratisk åpenhet**, statements 1-5 (of the memo's
  42, across 9 dimensions).
- All 9 Storting party groups (2025-2029 period).
- 14 seed evidence objects, taken verbatim from the memo's own seed
  table, covering its three named priority cases: lobbyregister,
  politisk reklame/tredjepartskampanjer, utenlandsk SoMe-påvirkning.
- Three separate, never-merged scores per statement×party cell:
  saksmatch, konsistens, evidensstyrke.
- Server-rendered evidence-chain pages (`/utsagn/{n}`) — the memo's
  "reproduserbar politisk hendelseslogg" — readable with no JavaScript.
- A client-side quiz (`/`) that never sends the user's answers anywhere;
  matching math runs entirely in the browser against server-computed
  cell data.
- Evidence storage: version-controlled YAML in
  `internal/data/content/evidence/`, embedded into the binary at build
  time. This file plus `internal/data/content/corrections/` are the
  correction log the memo asks for ("rettelseslogg,
  versjonshistorikk") — no separate database feature was built for this,
  and the published repository carries one initial commit rather than a
  history to consult.
- Automated regression tests (`internal/data/loader_test.go`,
  `internal/scoring/scoring_test.go`) that fail the Docker build if the
  dataset stops loading, a known evidence item's coded position changes
  unexpectedly, or a party position gets fabricated for a cell that has
  no real source.
- **2026-09-23 update — Rødt/MDG utsagn-01 positions resolved.** Fetched
  and verified the exact motion text for Rødt's forslag nr. 1 and MDG's
  forslag nr. 2-4 (2026-04-21 lobbyregister vote) directly against
  Stortinget's voting-detail pages. All four motions argue FOR a lobby
  register in principle — they only disagree with the presidentskap's
  proposal on scope and process. Rødt is coded `position: 2` (its motion
  demands mandatory registration and sanction authority — stronger than
  the statement itself); MDG is coded `position: 1` (its three motions
  only ask to investigate a scheme, a real but less committed ask) — not
  identical values, on purpose, caught during review's first pass. Both
  positions carry the verbatim source text alongside the claim
  (`internal/data/types.go`'s new `verbatim_excerpt` field). This closed
  what was previously this document's top "next candidate" item.

## Phase 2 (started) — Dimension 2, one real statement (2026-09-24)

**utsagn-06** is the first real statement of the memo's actual Dimension
2 (**Norsk handlefrihet**, statements 6-11, "Beholder Norge reell
beslutnings- og operasjonsevne?") — not the whole dimension, one
statement, added deliberately slowly rather than drafting all six at
once. Topic: whether any future Norwegian nuclear power buildout should
happen under full state ownership/control, or be open to private/
commercial ownership. Source: Innst. 278 S (2025-2026), the Storting's
actual committee treatment of Dokument 8:191 S (FrP's own bill to enable
commercial nuclear power) — read in full (6-page PDF, all merknader and
32 forslag), not just the vote tally, because the tally alone is
actively misleading here: FrP proposed the bill and then voted AGAINST
the final adopted text, but only because that text ("legg frem en
stortingsmelding") was weaker than FrP's own 13 proposals, all of which
fell. A vote-tally-only read would have coded FrP as opposed to nuclear
power; FrP's own proposal text says the opposite.

Coded from each party's own committee-record words, not inference:
- **Rødt** (+2): own forslag 30-31 explicitly demand "fullt statlig
  eierskap, kontroll og styring" plus hjemfallsrett — a direct textual
  match to the statement.
- **FrP, Høyre, Venstre** (-2 each): FrP's own forslag 2/5 and Høyre/
  Venstre's joint merknad ("det er ikke aktuelt med ... eierskap") both
  explicitly want a private/commercial model — the textual opposite of
  the statement.
- **Ap, Sp, MDG** (+1 each): a joint merknad citing the European norm of
  state involvement as a concern about the private model — real, but
  explicitly weaker and different from Rødt's concrete ownership demand,
  coded lower on purpose (same "don't collapse different commitment
  levels into one sign" discipline as MDG/Rødt on utsagn-01).
- **SV and KrF left uncoded, on purpose, for two different reasons**: SV's
  own merknad argues against subsidizing private nuclear and against a
  rushed process — a real but *different* claim than "Norway must have
  full state ownership," so coding SV here would attribute a stronger
  claim than SV's own words make. KrF has no seat on the energi- og
  miljøkomite for this case at all, so there is no committee text to
  read — only a floor vote, which is exactly the ambiguous signal the
  FrP case above warns against trusting alone.

All two cited URLs re-verified live via `scripts/research/verify_evidence.py`
before merge. Tests: `TestKjernekraftStatligKontrollPositionsAreCoded`,
`TestKjernekraftStatligKontrollLeavesAmbiguousPartiesUncoded`.

**Scope reality check, stated plainly:** this one statement took a full
primary-source read (the entire innstilling, not a summary) to code
honestly. Dimension 2 has five more statements (7-11); the memo's other
7 dimensions have none. Filling all of them at this rigor is the
70-90 person-week estimate the "Process lesson" in
`docs/research-notes/dimension-2-3-leads.md` already names — this entry
exists so that estimate stays concrete (one statement, one sitting) and
isn't quietly revised down by finishing a handful more without
re-stating it.

## Phase 2b — the statement-2-5 null result was real but not final (2026-09-24)

The "Next candidate work" item calling for "direct XML API pagination
instead of the JS-rendered web search UI" got tried for real on
utsagn-03 (partifinansiering) — and it worked. Prop. 144 L / Innst. 528 L
(2024–2025) bans anonymous party contributions and contributions
laundered through legal-entity or private-person intermediaries —
exactly utsagn-03's claim — and it went to a real committee vote.

The individual roll-call, pulled directly from
`data.stortinget.no/eksport/voteringsresultat` for votering_id 26057 (169
representatives, cross-checked against the official 89-for/12-mot split:
matches exactly), gives a cleaner signal than utsagn-06's kjernekraft
case: the vote is on the operative legal text itself ("§ 17 a... kan
ikke motta bidrag fra givere som er ukjente for partiet..."), not a vague
"legg frem en stortingsmelding" motion that needs the forslag text to
disambiguate. All 12 "mot" votes across the whole Storting were FrP; 0
FrP representatives voted for; every other party present voted for. That
let all 9 parties get coded on utsagn-03, not just the six with a
committee seat — Ap/H/Sp/SV/Rødt at +2 (committee merknad plus vote),
FrP at -2 (own merknad objecting to the ban's targeting, plus a unanimous
12/12 mot vote), MDG/V/KrF at +1 (no committee seat, no own merknad, but
an unambiguous "for" vote on the operative text — weaker than a written
merknad, still real, still individually verified, unlike the KrF
floor-vote-only situation on utsagn-06 which stayed uncoded because that
vote was on a vague motion).

Tried the same approach on utsagn-02 (political-ad funder disclosure):
found Innst. 275 L (2024–2025), a real, adjacent case, but its actual
subject is candidate-support/in-kind-benefit reporting under partiloven
— a different claim than utsagn-02's "the ad itself must show who paid,"
which is an EU DSA/TTPA-style labeling question. Did not force the
match. Statements 2, 4, and 5 stay without party-level evidence; see
"Explicitly NOT built yet" below for exactly why each one does.

## Phase 2c — utsagn-43 (ACER/energy sovereignty), harvester re-checks (2026-09-24)

`tools/harvester/` (see below) was run against every session 2013–2026
for utsagn-02 and utsagn-04's keyword lists: 0 leads for both, on top of
the earlier manual null result — a systematic re-check, not a repeat of
the same search. A manual (non-keyword) read for utsagn-04 did surface
Innst. 303 S (2021–2022, Høyre's Dokument 8:146 S on combating fake
news/disinformation) with real FrP+V and H+KrF merknader — but its claim
is "combat disinformation generally," not utsagn-04's specific "map,
document AND inform the public about foreign influence targeting
elections." Not coded, for the same reason Innst. 275 L wasn't for
utsagn-02 — see `tools/harvester/topics.yaml`'s utsagn-04 note for the
full reasoning, kept next to the keywords so it doesn't get re-litigated
from scratch next session.

utsagn-43 (Nasjonal kontroll over energimarkedet / ACER, one of the three
addendum statements outside the memo's 1–42) got full 9/9 party coverage
from one case: Innst. 196 S (2024–2025), Sp's Dokument 8:58 S and Rødt's
Dokument 8:75 S on rejecting EU's fourth energy market package and
withdrawing from ACER. Four separate roll-call votes on the same case
(votering_id 24383–24386, each individually pulled from
`data.stortinget.no/eksport/voteringsresultat` and cross-checked
per-party against the official for/mot tallies) gave a genuinely
differentiated picture rather than a flat FOR/MOT split:
- Sp, SV, Rødt at +2 — co-sponsors or joint merknad, voted for on every
  one of the four votes including the most drastic (formal veto +
  starting ACER withdrawal).
- FrP at +1, not +2 — voted FOR the milder asks (assess/don't-incorporate)
  and AGAINST the majority's flat rejection of both representantforslag,
  but AGAINST the most drastic ask (formal veto, votering_id 24383) that
  Sp/SV/Rødt backed. A real "limit further ACER growth, not full exit
  now" position — coding it +2 would have erased that nuance the same
  way a flat vote count erased utsagn-06's FrP nuance.
- Ap, H at −1 — defend the status quo (Ap/V's joint merknad; H's own
  merknad citing Høyesterett's assessment that ACER/ESA can't set
  Norwegian power prices), voted against all four restrictive amendments,
  but don't themselves push for deeper integration.
- V, MDG at −2 — explicit joint merknad wanting the whole fourth package
  incorporated "as fast as possible," the direct opposite of the
  statement's ask.
- KrF at −1 — voting-record only (no own merknad in this case; listed
  only as a co-signer of the majority tilråding), consistent with the
  weaker signal the project already uses for vote-only evidence
  elsewhere (see utsagn-03's MDG/V/KrF +1 entries).

Evidence coverage after this entry (9 statements × 9 parties = 81
possible cells): utsagn-01 9/9, utsagn-03 9/9, utsagn-06 7/9, the rest
0/9 by design or unresolved research — 25/81 cells filled (31%), up from
16/81 (20%) before this entry. Tests:
`TestPartilovenAnonymeBidragPositionsAreCoded`.

## Phase 2d — utsagn-04 resolved (2026-09-24, same day as Phase 2c)

utsagn-04 (utenlandsk påvirkning i sosiale medier rettet mot valg) went
from a twice-confirmed null result to full 9/9 party coverage in the same
session, from a case two separate research passes had missed: Innst.
231 S (2022-2023), Høyre's Dokument 8:51 S "om styrking av rikets
sikkerhet og kontraetterretning." Buried inside a much broader bill
(PST mandate, drone strategy, export control, criminalizing cooperation
with foreign intelligence), the representantforslag's own point 7 asked
the government to "umiddelbart igangsette en kartlegging av uønsket
informasjonspåvirkning fra utenlandske aktører i forbindelse med ...
stortingsvalget i 2025" — a near-verbatim match to the statement's own
"systematisk kartlegge, dokumentere og informere offentligheten." It
surfaced only from a plain web search for the Norwegian phrase itself,
not from the harvester's title/emne keyword sweep (that bill's title has
none of utsagn-04's keywords in it) — see
`tools/harvester/topics.yaml`'s updated utsagn-04 note for the method
lesson this leaves for statements 02/05/44/45.

Read the full 456-line innstilling PDF, not just the vote tally, and
cross-checked the two decisive votes (votering_id 20371, the kartlegging
forslag; committee tilråding, votering_id 20374) per-party against
`data.stortinget.no/eksport/voteringsresultat`:

- **Høyre, Venstre at +2** — co-sponsors of the actual forslag text
  (forslag 6 in the committee's renumbered list); Venstre additionally
  proposed a further, related forslag (7) alone.
- **KrF, MDG, Rødt at +1** — voted for, no own merknad in the case —
  vote-only signal, same weaker-magnitude convention as utsagn-03's
  MDG/V/KrF and utsagn-43's KrF.
- **SV at -1** — a real, substantive own merknad: SV explicitly says it
  shares the underlying concern about foreign election influence, but
  opposes this specific kartlegging mechanism because of a documented
  surveillance-of-legitimate-turnout risk. A genuine objection to the
  method, not to the goal — coded weaker than a party rejecting the aim
  outright.
- **Ap, Sp at -1** — a real, shared own merknad: both argue the
  representantforslag's content (including kartlegging) is already
  substantially addressed by existing government work (Meld. St. 9
  2022-2023, NESS, increased PST/politi/NSM funding) — a redundancy
  argument, not a rejection of the goal, coded the same way as Ap/Høyre's
  -1 on utsagn-43 (defends the status quo, doesn't push for more, but
  doesn't reject the underlying aim).
- **FrP at -1** — voted against, no own merknad addressing this specific
  point in the case (FrP's own merknad here covers a different forslag,
  criminalizing cooperation with foreign intelligence) — vote-only
  signal, same weaker-magnitude convention in the opposite direction.

Tests: `TestUtenlandskPaavirkningPositionsAreCoded`. The pre-existing
regression guard `TestInfluenceOperationTargetsDoNotBecomePartyPositions`
(originally: "EvidenceForCell(utsagn-04, {ap,sv,rodt}) must be empty,"
written when utsagn-04 had zero coverage of any kind, to stop a
TikTok-influence-operation *target* from being coded as if it were an
*actor*) had to be rewritten to check the specific target-evidence items
directly instead of asserting the whole cell stays empty — the cell is
no longer empty, for a real and unrelated reason (these three parties'
own committee-record votes on the actual kartlegging proposal), and the
old assertion would have failed for the right addition and the wrong
reason if left as-is.

Evidence coverage after this entry: utsagn-01 9/9, utsagn-03 9/9,
utsagn-04 9/9, utsagn-06 7/9, utsagn-43 9/9, the rest (02, 05, 44, 45)
0/9 by design or unresolved research — 43/81 cells filled (53%), up from
25/81 (31%) before Phase 2c and 2d.

## Phase 2e — utsagn-02 resolved (2026-09-24, same session as Phase 2d)

utsagn-02 (politisk reklame må vise hvem som betalt, pengenes opprinnelse
offentlig kjent) went from two independent null results to full 9/9
coverage — from a case type this project hadn't looked at yet: a
**plenary floor motion ("løst forslag"), not a Dokument 8 bill**. During
the 2025-10-14 trontaledebatt (the Storting's opening-of-session Address
debate), Rødt's Marie Sneve Martinussen proposed, and the Storting
adopted 54-46 (votering_id 26361, sak p=104588), "å inkludere regulering
av anonym finansiering av politisk reklame og markedsføring i den
helhetlige gjennomgangen av partiloven." Her own plenary speech cites
"flere stygge eksempler" from the 2025 campaign as the motivation.

This surfaced from a plain web search for the statement's own key
phrase — the same method that resolved utsagn-04 two entries above —
not from `tools/harvester/`'s title/emne keyword matching, which only
queries `data.stortinget.no`'s `saker` (case) endpoint and has no
concept of a plenary-only "løst forslag" outside a numbered case at all.
See `tools/harvester/topics.yaml`'s updated utsagn-02 note.

One honest nuance, not a scope mismatch severe enough to withhold
coding (unlike utsagn-44's nordisk/europeisk gap): utsagn-02's text has
two clauses joined by "og" — (1) the ad itself must show who paid
(an on-ad, EU DSA/TTPA-style label) and (2) money's origin must be
publicly known. This forslag's own text ("anonym finansiering av
politisk reklame og markedsføring") tests clause (2) only. Coded anyway,
with the nuance documented in
`evidence/61-politisk-reklame-anonym-finansiering-kontekst.yaml`'s
context_note, because the forslag names "politisk reklame og
markedsføring" specifically — closer to the statement's actual subject
than Innst. 275 L's general candidate-support reporting, the case
correctly rejected for utsagn-02 in Phase 2b above.

Coded from the vote plus the one recorded plenary speech (no committee
merknader exist for a plenary floor motion):
- **Rødt at +2** — own forslag, own plenary speech naming real 2025
  campaign examples.
- **Ap, SV, Sp, MDG, V at +1** — voted for, no individual speech on this
  specific point — vote-only signal.
- **FrP, Høyre, KrF at -1** — voted against, no individual speech on
  this specific point — vote-only signal in the opposite direction.

Tests: `TestPolitiskReklamePositionsAreCoded`.

Evidence coverage after this entry: utsagn-01/02/03/04/43 all 9/9,
utsagn-06 7/9, the rest (05, 44, 45) 0/9 — 52/81 cells filled (64%), up
from 43/81 (53%) before this entry.

## Correction (2026-09-23) — an invented dimension, fixed


A research pass was scoped to draft a "second dimension" without being
handed the memo's own dimension list first, and it invented one —
**"Norsk handlefrihet"**, statements numbered 6-8 — believing at the
time that this name did not exist anywhere in `research-memo.md`. It
was renumbered out of the memo's 1-42 range as an explicit addendum
rather than pretending it belonged there:

- `utsagn-06` → `utsagn-43` (energimarkedssuverenitet / ACER)
- `utsagn-07` → `utsagn-44` (digital suverenitet / datajurisdiksjon)
- `utsagn-08` → `utsagn-45` (forsvars-/beredskapssuverenitet)
- Dimension label changed to `"Tillegg (utenfor de 42)"`, and
  `internal/server/server.go` explicitly excludes it from the memo's
  dimension count (`dimensionNames`/`addendumCount`) so the UI can never
  again claim these three are one of the memo's real 9 dimensions.

The evidence behind the renumbered statements is still real — Meld. St.
13 (2025–2026) and Prop. 109 S (2025–2026) were fetched and verified
directly against stortinget.no — so nothing was deleted, only moved.

Zero party positions are coded on statements 43-45, and that's correct,
not a gap to fill by more searching: both seed cases are proposals the
Storting has received but not yet debated or voted on. A dedicated test
(`TestAddendumPendingCasesHaveNoFabricatedPartyPositions`) asserts every
party shows zero evidence here until that changes.

## Correction 2 (2026-09-23, same day) — the "fix" above was itself wrong

Re-reading `research-memo.md`'s own dimension table (reproduced correctly
in full below) while preparing a research-delegation brief found that
the paragraph above was **also wrong**, in the exact way it accused
itself of fixing:

- **"Norsk handlefrihet" is a real dimension** — it's the memo's actual
  Dimension 2, statements 6–11, question "Beholder Norge reell
  beslutnings- og operasjonsevne?". It was never missing from the memo;
  the mistake was believing it was.
- **"Energi og klimapolitikk" is not a real dimension name** — nowhere
  in `research-memo.md`. It was fabricated during the correction above,
  most likely confused with the memo's actual Dimension 3, **"Forvaltning
  av fellesskapets ressurser / energi"** (statements 12–18, question
  "Hvem kontrollerer ressursene, risikoen og verdiskapingen?").
- **The specific statement-6-10 sub-topics — kjernekraft,
  oljeutvinning-utfasing, havvind, strømpris, klimamål — are not in the
  memo either.** The memo names dimensions, their statement-number
  ranges (some non-contiguous), and one primary question per dimension —
  it does **not** enumerate the 42 statements' literal topics beyond the
  three it seeds in full (statements 1-4/5, `research-memo.md`'s own
  seed table). Those five sub-topics were invented, then a real research
  pass spent real effort (correctly) verifying real Storting cases
  *against those invented topics* — the individual case verifications in
  `docs/research-notes/dimension-2-3-leads.md` are still factually sound
  (real case IDs, real vote tallies, independently re-verified against
  live sources), but the dimension/statement numbers they were filed
  under are not memo fact and must not be treated as such.
- Same problem, same day, for "Dimension 3 (EØS og norsk suverenitet,
  statements 11-15)" in the "Next candidate work" list below at the time —
  no such dimension name or number range exists in the memo. That entry
  is corrected below.

**The actual, verified-against-source dimension table:**

| Dimensjon | Utsagn | Primært spørsmål |
|---|---:|---|
| Demokratisk åpenhet | 1–5 | Kan offentligheten se hvem som påvirker hvem, for hvem og med hvilke penger? |
| Norsk handlefrihet | 6–11 | Beholder Norge reell beslutnings- og operasjonsevne? |
| Forvaltning av fellesskapets ressurser / energi | 12–18 | Hvem kontrollerer ressursene, risikoen og verdiskapingen? |
| Teknologisk kapasitet | 19, 20, 24 | Kan kritiske digitale funksjoner og kompetanse opprettholdes under press? |
| Innovasjon / omstillingsevne | 21–23 | Bygger staten og markedet ny kapasitet eller bare kjøper ferdige løsninger? |
| Universell grunnmur | 25–30 | Fungerer samfunnets grunnleggende tjenester uansett bosted og inntekt? |
| Barns muligheter | 31–36 | I hvilken grad kompenserer samfunnet for forhold barn ikke har valgt? |
| Robusthet / adaptabilitet | 38, 39, 41, 42 | Tåler systemet bortfall, sjokk og feilantakelser? |
| Generasjonsansvar | 37, 40 | Synliggjøres konsekvensene 10–30 år frem? |

**Consequence for the addendum (43-45):** all three of its topics
(energimarked-suverenitet/ACER, digital suverenitet, forsvars-/
beredskapssuverenitet) match Dimension 2's real question — "Beholder
Norge reell beslutnings- og operasjonsevne?" — closely enough that
`internal/data/content/statements/43-*.yaml`'s `primary_question` field
already quotes it near-verbatim, written before this second correction
untangled why that phrasing felt so exact. They are strong candidates
to fold into Dimension 2 (numbers 6-11) once that dimension is drafted
for real, rather than staying a permanent addendum — see "Next candidate
work" below. Not done automatically here: folding them in needs the
same "read the memo, don't invent a sub-topic" discipline as the rest of
Dimension 2, applied once, not twice more.

The lesson, stated plainly because this project's own rule is to never
bury a correction: **two mistakes in the same 2026-09-23 correction, not
one.** The first was inventing a dimension name. The second, made while
writing that correction, was denying a dimension name that was real and
substituting a fabricated one plus five fabricated sub-topics — a
classic overcorrection, and exactly the kind of "I fixed the citation
error, therefore I now trust everything else I wrote in the same
sitting" gap this project's own review discipline exists to catch.

## Brand pass "Civic Ledger" (2026-10-02)

Converged from four independent design proposals (three AI systems plus
our own), then checked by hand rather than trusted.

- **Colour.** Every party owns a hue family (red Ap/Rødt/SV, blue H/FrP,
  green Sp/MDG, teal V, yellow KrF). The old accent `#2c4a5e` sat next
  to the blue parties, and the old A-E tags were green, blue, ochre,
  purple and red, so a reader could see "E = red = Rødt". Now one
  violet at hue ~265, the free window between navy (~210) and magenta
  (~300): `#6a4c93` (links, 6.28:1 on the background), `#4b356b`
  (buttons, white on it 10.39:1). All text pairs are 5.58:1 or better;
  the full table is in the commit message. Red/green stay out entirely.
  Party colour is data, never chrome.
- **Kildegrad.** Letter + five-segment meter + words, one colour. E is an
  unverified OSINT lead, not a weak source, so it has zero filled
  segments rather than one. The proposals invented tier meanings
  ("B = komitéinnstilling", "E = medieuttalelse"); the real scale is the
  `/metode` table, which this pass did not change.
- **Mark.** `[e]`: "ekte" inside citation brackets. A three-line "source
  E" was drawn too and rejected after rendering: at 16-48 px it reads as
  "E:". The mark is SVG strokes, never `<text>` (a favicon cannot load a
  web font).
- **Type.** Headings and source quotes in IBM Plex Serif, body on the
  system font (0 bytes). IBM's own Latin-1 split files, unmodified:
  Plex has a Reserved Font Name, so subsetting it ourselves would be a
  modified version that may not be called Plex. SemiBold (24 KB) is
  preloaded; Regular (23 KB) is only fetched when a "Verifiser" panel
  opens (checked in a real browser). Schibsted Grotesk was proposed and
  rejected: it is the typeface of VG's and Aftenposten's owner, and both
  run their own valgomat.
- **Cost.** HTML +0.2 KB, CSS +1.2 KB (gzip), one 24 KB font, after which
  versioned assets and fonts are `immutable` for a year
  (`TestStaticCacheHeaders`).

## Explicitly NOT built yet

Each item below is a real gap, not an oversight — recorded here so it
can't quietly become "we assumed this was done."

- **Party coverage gaps on 4 statements.** See "Current state" at the
  top for the list. **Rødt's 2025-2029 program** was downloaded by
  hand on 2026-09-28 (roedt.no blocks automated requests with a
  Vercel bot challenge, HTTP 429). It settled R on 55 (-2, file 603)
  but not on 45 (no EU/EØS-to-defence link, same as Sp). To re-run
  `verify_excerpts.py` on file 603 on a new machine, put the PDF in
  `~/.cache/ektevalgomat-sources/` under the sha256 of its
  `source_url`. Filling one means real primary-source evidence per
  party, or leaving the cell honestly empty. Never a guessed position.
  **2026-10-27, the historical-vote search.** `tools/autopilot/history.py`
  walked every gap cell back through earlier sessions (evidence →
  Innst. reference → `henvisning` → sak → emne → earlier saker with
  votes) and settled none of them. What it found, and why each is
  still empty: **utsagn-06** (krf, sv) — the closest case, sak 92789
  (Dokument 8:129 S, 2022-2023, "mer kunnskap om behovet for
  kjernekraft"), is the wrong question: KrF voted for and SV against
  *researching* nuclear power, and the statement is about *state
  ownership and control*, which those voterings never touch.
  Sak 83905 (Innst. 504 S, 2020-2021, national ownership of strategic
  resources) was carried unanimously, so it codes nobody.
  **utsagn-52** (ap) — zero candidate cases in any earlier session.
  **utsagn-57** (ap) — one candidate, a 2010-2011 oversendelsesforslag
  with no voteringer at all. **utsagn-55** (mdg, sp) — no earlier
  session has a case on total tax and fee level. The walk matches on
  emne id plus title words, and only 27 distinct emner are populated
  in the export, so this is a weak net rather than a proof of absence.

- **PostgreSQL / any mutable database.** Deliberate substitution: git +
  embedded YAML instead. Revisit only if evidence volume or an actual
  need for a live-editing UI (not just "it would be more standard")
  makes hand-editing YAML files genuinely painful — see `research-memo.md`'s
  own central evidence table for the fuller schema (beneficiaries,
  cost_bearers, amount_nok, foreign_dependency, time_horizon,
  counterevidence_ids, reviewer_1/reviewer_2, review_status) this
  project's `Evidence` struct deliberately narrowed down from.
- **Automated harvesting from Stortinget's API, SSB, Brønnøysund, FFI,
  NSM, PST, Riksrevisjonen, LOTTE, Meta Ad Library, TikTok Commercial
  Content Library, Google Ads Transparency Center, or the EU political-ad
  register.** Only Stortinget is covered: `tools/harvester/` (built
  2026-09-24, see "Next candidate" item 1) queries its `saker` endpoint
  and proposes YAML for human review. Every other source above is still
  searched by hand.
- **Ownership graph (Brønnøysund entity/ownership_edge/financial_transfer
  tables).** No modeling of company ownership, campaign financing flows,
  or entity relationships at all yet.
- **Double-coding / independent reviewer workflow**
  (`reviewer_1`/`reviewer_2`/`review_status` in the memo's schema). Every
  seed evidence item here was written and reviewed by one person. A real
  second-reviewer workflow needs either a second human or a documented
  single-reviewer caveat surfaced in the UI — currently neither exists.
- **Formal DPIA / personvernvurdering.** Not done, and not needed yet
  because the dataset only names public officials in official capacity
  and state institutions — see `/metode` on the live site. The hard
  rule: the day a case needs to name a private citizen or an unverified
  social media account, work stops until a real DPIA exists. Don't relax
  this by accretion.
- **utsagn-45 for Rødt and Sp.** 44 was resolved 2026-09-26/27 with a
  real vote (Innst. 225 S, 9/9 coded). 45 has Innst. 464 S as tier-A
  context (position null, it does not code a stance on this statement)
  and tier-C party programs for 7 of 9. Rødt and Sp have no program
  quote that links EU/EØS to defence or beredskap (both programs read
  in full), and the planned `consistency_flag` (tier-A context vs tier-C
  program pointing the same or different way) is not built.
  **Status 2026-10-13:** the Rødt programme was re-read against the
  primary PDF. The copy supplied on 2026-10-13 is byte-identical to the
  one already cached under its `source_url` sha256, so it is the same
  document the earlier pass read. Re-confirmed: no single sentence in it
  joins EØS/EU to defence or beredskap. It does carry both halves
  separately, and they point the same way: «Rødts mål forsvarspolitisk er
  et defensivt, sterkt nasjonalt forsvar, uavhengig av stormaktsinteresser»
  and «Med et sterkere nasjonalt forsvar kan vi sørge for at Norge er
  premissleverandør for militær aktivitet på vårt eget territorium»; plus
  «Det er en ensidig avtale som tvinger Norge til å innføre direktiv på
  direktiv som EU vedtar» and «Erstatte EØS-avtalen med en balansert
  handelsavtale som sikrer nasjonal kontroll». A Rødt item is staged at
  `tools/autopilot/staging/rodt-45/` (position +2, tier C, gate PASS);
  its `context_note` states plainly that the join between the two
  chapters is the coder's, and calibrates the magnitude against KrF's +1
  on this same statement, which was demoted because KrF's text had no
  defence component at all. Sp is unchanged: the same read finds no Sp
  passage joining the two, so its cell stays honestly empty.
- **Longitudinal consistency scoring.** `konsistens` correctly reports
  "utilstrekkelig grunnlag" for nearly every cell right now, because the
  dataset is almost entirely single-snapshot evidence. This can only
  genuinely improve by adding more evidence over time for the same
  party×statement pair, not by changing the scoring code.
- **Progressive-disclosure redesign (2026-09-23 design brief).** Three
  info depths (Scan/Understand/Verify), evidence cards as the atomic
  display unit, a claim→proposal→vote→action→outcome stepper as the
  signature component, structural (same-row) party comparison, a
  monochrome-plus-accent visual style with dot-scale evidence strength
  (never traffic-light), one-statement-per-screen quiz flow with "vet
  ikke" kept distinct from "nøytral", importance weighting, trade-off
  questions between conflicting values, and an unranked value-profile
  results page. Staged: design tokens + evidence cards first (foundation
  everything else depends on), then the stepper, then the quiz rework,
  then results, then trade-offs last (highest new-mechanic risk).
  **Status 2026-09-27:** evidence cards (f670ee7) and progressive results
  (5357b25) are built. Not built: the stepper, importance weighting,
  trade-off questions, and "vet ikke" as its own answer (quiz.js still
  has one "Vet ikke / nøytral" option).
  **Status 2026-10-12:** "vet ikke" is now its own answer. The middle
  option (value 0) is "Nøytral" only; a separate off-scale sixth choice,
  "Vet ikke / hopp over", is held out of the party match entirely (see
  web/static/js/scoring.js) and reported as a visible skip count on
  /resultat. The stepper, importance weighting and trade-off questions
  remain unbuilt.
  **Status 2026-10-13:** the one-statement-per-screen quiz flow is built
  (web/static/js/quiz.js): one question in focus, Forrige/Neste with no
  auto-advance, "Spørsmål N av 60", a skippable "Se over svarene dine"
  review step, a "Midlertidig økt" that writes nothing to localStorage,
  and a persisted "Utforskermodus" that restores the full list. The
  thematic heading was removed from above every question and now lives
  only in the neutral progress overview. That decision comes from Kamoen
  et al., PLOS ONE 2019 (N = 27,404), which manipulated party-ASSOCIATED
  headers, not neutral topic labels, and found any header shifted answers,
  only among low-sophistication respondents; so it is a precaution against
  category priming, not a proven requirement for a neutral label. The
  source card on /utsagn/N now leads with reason, quote and a one-click
  original source. Results gained a "Hvorfor dette resultatet?" robustness
  panel (analyse.js) covering contribution, missing or flagged
  documentation, equal-theme weighting and leave-one-out stability.
  Still unbuilt: the claim-to-outcome stepper, importance weighting, and
  trade-off questions.

## Next candidate work, roughly in order

1. ~~A small offline harvester script~~ **Built 2026-09-24:**
   `tools/harvester/` (`uv`-managed, `uv sync` then `uv run harvest.py`)
   queries `data.stortinget.no`'s JSON export API directly (not the
   JS-rendered web search, which the 2026-09-23 research pass hit real
   limits with) across sessions back to 2009, matches case titles/emner
   against `topics.yaml`'s per-statement keywords, and writes leads plus
   evidence-YAML skeletons under `reports/<date>/` — never into
   `internal/data/content/` directly, see the tool's own README for the
   review workflow. 17 unit tests, all pure functions (no network in
   the test suite). Verified against a known positive (kjernekraft,
   2025-2026: finds the real case) and re-ran the existing statement
   2/4 null results across 2020-2026 with zero new title-level matches —
   strengthens, doesn't overturn, the earlier "actively searched, found
   nothing" entries above. Known ceiling, stated in the README: keyword
   matching only sees case titles/emne names, not committee-merknad
   text, so a real case with an off-topic-sounding title (utsagn-03's
   best evidence, Innst. 528 L, is titled about "bakenforliggende
   givere," not "partifinansiering") won't surface this way — this tool
   narrows the search, it doesn't replace reading candidates that a web
   search turns up by other means.
2. The rest of Dimension 2 (**Norsk handlefrihet**, statements 7-11 per
   the memo) and Dimension 3 (**Forvaltning av fellesskapets ressurser /
   energi**, statements 12-18) — corrected names/ranges, see "Correction
   2" above; do not reuse "Energi og klimapolitikk" or "EØS og norsk
   suverenitet", neither is a real dimension name in the memo. **Statement
   6 (kjernekraft / statlig kontroll) is done as of 2026-09-24 — see
   "Phase 2" above — using the kjernekraft lead below.** The addendum's
   ACER/digital-sovereignty/defense-sovereignty research gives Dimension
   2 a further head start for the remaining five. Read the memo's own
   dimension question before drafting statement text; do not invent a
   specific sub-topic and present it as the memo's own the way
   "Correction" and "Correction 2" above both did. **2026-09-23: a first
   research pass ran against invented sub-topics (kjernekraft,
   oljeutvinning, havvind, strømpris, klimamål, sosial dumping, etc.) and
   found real, independently-verified leads for several of them — see
   `docs/research-notes/dimension-2-3-leads.md`, updated to drop the
   fabricated statement-number attributions but keep the verified case
   data, since it's still useful raw material once each topic is
   properly assigned a real statement number under Dimension 2 or 3.
   **Strømpris is shipped as of 2026-09-24 — see "Phase 2d" below.**
   Sosial dumping remains unshipped — three real cases exist (verified)
   but none fits any of the memo's nine dimension questions as written;
   see `docs/research-notes/dimension-2-3-leads.md`'s "Sosial dumping"
   entry, still an open scope question, not a search gap.**

## Phase 2f (2026-09-24): utsagn-12, first statement of Dimension 3

Dimension 3 ("Forvaltning av fellesskapets ressurser / energi", 12-18)
gets its first statement: "Staten bør fastsette en lovbestemt nasjonal
makspris for hva kraftselskaper kan ta for strøm til forbrukere i
Norge." Full 9/9 party coverage from Innst. 535 L (2024–2025),
votering_id 26116 — Rødt's own national-makspris forslag (part of a
merged 4-bill committee report the 2026-09-23 pass had correctly
flagged as too complex to code as a single statement). The fix wasn't
to force the whole bill into one utsagn; it was to isolate the ONE
vote of the 17 that actually tests this claim (Rødt's own forslag
bloc, 12-23, which includes the literal 35-øre/kWt makspris ask) and
code only that. Every other party voted against this specific
mechanism while backing a genuinely different state-interventionist
alternative of their own — Ap/Sp defend Norgespris (their own
referanseprisordning), SV wants a state-owned power-trading company
(arguably a *stronger* intervention), KrF/FrP prefer a 100%-above-
threshold subsidy, MDG rejects any consumption-linked mechanism on
distributional-fairness grounds (wants a flat payout instead), and
only H+V make the actual free-market argument. Coded -1 vs -2
accordingly, with each context_note naming the party's own alternative
rather than flattening nine "mot" votes into one undifferentiated
rejection. The other 16 voteringer on this case (VAT, nettleiemodell,
forbrukstak level, flat-payout redesign, Norgespris's own passage)
remain unshipped, each a distinct claim needing its own statement.

## Phase 2g (2026-09-24): utsagn-13, second statement of Dimension 3

Same session, same nuclear-power case that gave utsagn-06 its evidence
(Dokument 8:191 S / Innst. 278 S, 2025-2026) turns out to carry a
*second*, genuinely distinct axis worth its own statement: not
ownership, but timing and mechanism — should Stortinget legislate NOW
for privately-financed commercial nuclear power, or wait for a
government white paper first. utsagn-13 ("Stortinget bør nå vedta
lovendringer...") codes votering_id 27722 (FrP+H+V's forslag 1-2,
49 for / 46 mot), full 9/9 party coverage, read from the 1352-line
innstilling in full.

Two findings worth carrying forward as method, not just content:
1. FrP is the case's own proposer but voted AGAINST the final adopted
   text (a much weaker "write a white paper" tilråding than what they
   asked for) — the roll-call number alone, read on its own, would
   have said "FrP opposes its own idea," which is backwards. Read the
   committee text before coding a position from a vote tally alone,
   every time, no exceptions.
2. Rødt's own merknad on THIS case states outright that they are not
   opposed to nuclear power as a technology, only to a commercial/
   private ownership framing — the same principle already coded as
   utsagn-06's Rødt position. This is a genuine Konsistens signal (one
   party, one principle, two independent cases, coded and cross-
   referenced in both statements' evidence files) rather than a
   coincidence to let slip past a flat vote count.

The other 9 voteringer on this case (KrF's, Rødt's own hjemfallsrett
forslag, H's stricter private-financing-only forslag, SV+R's IAEA
non-proliferation forslag, etc.) remain unshipped — each is either a
narrower restatement of the same axis already coded, or a genuinely
separate claim (e.g. nuclear waste storage policy) that would need its
own statement rather than a reuse of this one's evidence.

3. **Statement 43 (nasjonal kontroll over energimarkedet / ACER) is done
   as of 2026-09-24 — see "Phase 2c" above — full 9/9 party coverage
   from Innst. 196 S (2024–2025).** Statement 44 (digital suverenitet)
   has a real, decided, on-topic case now (Innst. 225 S, 2025–2026) but
   it's scope-mismatched ("europeisk" jurisdiction tested, "nordisk"
   asked) — see "Explicitly NOT built yet" above and
   `docs/research-notes/dimension-2-3-leads.md` for the full finding and
   the two unresolved options. Not a "watch for a vote" item anymore;
   it's a "decide the statement's own scope" item. Statement 45
   (forsvar/beredskap) was searched for real 2026-09-24 — three
   candidates found and read, none decided-and-differentiated (see
   `docs/research-notes/dimension-2-3-leads.md`) — the concrete next
   step if EU's military mobility regulation ever reaches a Norwegian
   samtykkeproposisjon and vote, not another blind search.
4. Only after several dimensions have real party-level evidence on most
   statements: revisit whether the addendum statements (43-45) belong
   folded into a real dimension instead of standing alone.

## Phase 2h (2026-09-24): utsagn-25, first statement of Dimension 5

Dimension 5 ("Universell grunnmur", 25-30, "Fungerer samfunnets
grunnleggende tjenester uansett bosted og inntekt?") gets its first
statement: "Regjeringen bør fullføre tannhelsereformen og legge frem
en forpliktende plan som sikrer at hele befolkningen er omfattet av en
universell rett til tannbehandling i det offentlige [...] innen 2030."
Full 9/9 party coverage from Innst. 458 L (2024-2025), votering_id
25387 — SV's own forslag for a binding, dated plan to universalize
dental care, read from the 1669-line innstilling in full. The
underlying case merges three documents (the government's own modest
Prop. 124 L, SV's Dokument 8:96 S, Rødt's Dokument 8:157 S) and carries
20 separate voteringer; this statement isolates the one vote that
tests the specific claim (binding plan, whole population, dated) as
opposed to the committee's own much vaguer, adopted alternative (an
undated "gradual" stortingsmelding commitment).

One finding worth flagging rather than smoothing over: Rødt's own
written merknad in this exact case says outright "en tannhelsereform
skal gjelde hele befolkningen" (a dental reform should cover the whole
population) — the same universalist goal SV's forslag operationalizes
— yet Rødt's recorded vote on that specific forslag is "mot". The
most defensible reading (Rødt has its own, separately-submitted,
more mechanism-heavy alternative — a frikort/phasing scheme — and
plausibly preferred it over SV's dated-deadline version) is stated in
the evidence file as an inference, not asserted as settled fact. This
is deliberately left as a documented, unresolved tension: the site's
own standard is to say plainly when a written merknad and a recorded
vote pull in different directions, not to invent a tidy explanation to
make the data look cleaner than it is. A genuine near-miss during this
session's research: an initial pass misattributed which of the two
opposition documents (8:96 S vs 8:157 S) belonged to SV versus Rødt,
based on a web-search result title rather than the innstilling's own
committee-membership line — caught and corrected before anything was
written into an evidence file, precisely because the innstilling text
was read in full rather than trusted from a secondary snippet.

Three parties (FrP, KrF, V) have no committee seat or no documented
merknad on this specific case; their votes are recorded but their
reasoning is not, so their evidence files use lower
`attribution_confidence` (`svakt_dokumentert` or `ukjent`) rather than
inventing a rationale to match their vote.

The other 19 voteringer on this case (Rødt's own frikort/phasing
forslag, Sp/SV/Rødt/Pasientfokus's rural-access forslag, KrF's and
FrP's targeted-group forslag, the government's own modest 25-28-year-
old law text itself) remain unshipped — several are strong candidates
for the dimension's remaining five statement slots (26-30), especially
the rural/geographic-access angle, which tests this dimension's "uansett
bosted" half more directly than this statement's "uansett inntekt"
framing does.

## Phase 2i (2026-09-24, same session): utsagn-31, first statement of Dimension 6

Dimension 6 ("Barns muligheter", 31-36, "I hvilken grad kompenserer
samfunnet for forhold barn ikke har valgt?") gets its first statement:
a lovfestet, årlig prisjustering av barnetrygden (barnetrygdloven §
10), so its real value cannot erode regardless of the annual budget
cycle. Full 9/9 party coverage from Innst. 126 S (2025-2026),
votering_id 26942 — SV and MDG's own forslag, added during committee
treatment (not part of Venstre's original 17-point Dokument 8:11 S),
read from the 24-page innstilling in full.

Two findings worth carrying forward:
1. A real near-miss with an AI web-fetch intermediary: its first
   summary of this innstilling's own vedtak/tilråding text was wrong
   — it returned content from an unrelated document (a different
   sak's committee, different proposers, different topic entirely).
   Caught before anything was written to an evidence file, by the
   same discipline as Phase 2h's misattribution catch: fetch the raw
   HTML directly and read it, never take a secondary AI summary as
   the primary source. Any web_fetch result on a Storting document
   should be treated as a lead to verify, not a citation to use.
2. The same "prefers discretionary control over a binding automatic
   mechanism, without opposing the underlying goal" pattern already
   seen in utsagn-25 (Ap/Sp on tannhelse) recurs here across three
   parties for a different case entirely: Ap (explicit — "barnetrygden
   må forbli en universell ordning", touts its own 10.5 mrd. kroner of
   discretionary increases since 2021), Høyre (own reasoning: prefers
   municipal free-income discretion and ordinary budget-cycle
   treatment over a Storting-mandated automatic rule), and KrF
   (explicitly wants a much larger increase via its own alternative
   budget, just not this specific legal mechanism — coded with lower
   attribution_confidence since KrF's own text argues the amount, not
   the mechanism choice). This is now documented independently in two
   unrelated cases (health, family policy) rather than asserted once
   and assumed to generalize — a real, repeatable methodological
   pattern worth watching for as more dimensions get built, not a
   one-off.

FrP is coded more strongly (-2, not -1) than Ap/H/KrF specifically
because its own written merknad states a flatly ideological objection
("veien ut av fattigdom går primært gjennom arbeid, ikke økt
avhengighet av offentlige overføringer") rather than a mechanism
preference — the same distinction Phase 2h drew for Høyre's dental
merknad ("tannhelsetjenesten... fungerer godt") versus Ap/Sp's shared-
goal-different-tempo framing on that case.

Rødt, Senterpartiet and Venstre have no committee seat and no
documented merknad on this specific forslag; their evidence files use
attribution_confidence `ukjent` rather than inventing reasoning to
match their vote — same discipline as utsagn-25's FrP/KrF/V cells, just
a different trio of parties this time (the two dimensions don't share
a "silent" party set, which is itself worth noting: it is not that
these three parties are always undocumented, only that they were on
this particular case).

One clarification worth stating explicitly in Venstre's own evidence
file: Venstre is the case's original proposer (Dokument 8:11 S, Abid
Raja/Guri Melby/Grunde Almeland) yet voted "mot" the specific forslag
this statement tests. Verified this is NOT a repeat of Phase 2g's
"party opposes its own idea" finding — the indexing forslag was added
by SV and MDG during committee treatment and is absent from Venstre's
own original 17-point list, confirmed by reading that list directly
rather than assuming continuity from authorship of the broader case.

The other 14 voteringer on this case (kontantstøtte/ventestøtte,
gratis barnehage for lavinntektsfamilier, sosialhjelpssatser, MDG's own
separate barnetrygd-taxation forslag, boligpolitikk for enslige
forsørgere, KrF's kontantstøtte-utvidelse, etc.) remain unshipped —
several are strong candidates for this dimension's remaining five
statement slots (32-36), and for Dimension 1/7 crossover statements
about universal-vs-behovsprøvd welfare design more broadly.

## Phase 2j (2026-09-24, same session): utsagn-19, first statement of Dimension 4

Dimension 4 ("Teknologisk kapasitet", 19/20/24, "Kan kritiske digitale
funksjoner og kompetanse opprettholdes under press?") gets its first
statement: should Norway build a national cloud service owned and
operated by the public sector, rather than basing critical societal
functions on foreign-owned commercial cloud platforms. Full 9/9 party
coverage from Innst. 225 S (2025-2026), votering_id 27485 — SV and
Rødt's own forslag, read from the 17-page innstilling in full.

This case reuses the same innstilling already flagged in "Explicitly
NOT built yet" for statement 44 (digital suverenitet, scope-mismatched
on "europeisk" vs "nordisk" jurisdiction) — but tests a completely
different forslag (11, not 4) on a different axis (public ownership of
infrastructure, not jurisdiction of data processing), so the two are
independently coded and the statement-44 gap remains exactly as
documented before this session touched the case.

Three findings worth carrying forward:
1. The government's own concept-value study (KVU) had already
   evaluated and rejected precisely the model this forslag asks for —
   a closed, state-run cloud was found to be the most expensive option
   reviewed, and external quality assurance recommended against it.
   Ap's own merknad cites this directly. This is a case where "mot"
   is grounded in the government's own prior analysis, not just
   philosophy — worth flagging as a stronger evidentiary basis than a
   bare philosophical objection, and coded with that in mind.
2. A genuine three-way split inside the sovereignty-minded bloc, not
   just the by-now-familiar "shares the goal, prefers a softer
   mechanism" pattern: FrP co-sponsors a weaker "investigate first,
   report by 2027" version (forslag 1, with SV+Rødt) instead of this
   forslag; separately, MDG co-sponsors an ownership-model-agnostic
   alternative (forslag 14, with Rødt) that explicitly argues against
   locking the choice to state ownership specifically — a real policy
   disagreement about mechanism, not merely about bindingness or
   timing. Rødt is the only party backing both alternatives (11 and
   14) in addition to voting for this statement's own forslag,
   confirmed by reading each forslag's own sponsor line rather than
   assuming from the vote tally alone.
3. Høyre's own merknad is the most philosophically explicit "mot" in
   this case ("svaret på avhengighet av internasjonale teknologigiganter
   er mangfold og konkurranse — ikke statlige monopoler"), same role
   Høyre played in the tannhelse case (Phase 2h) — a recurring pattern
   worth watching for as more cases get built: Høyre's own written
   merknad is consistently the strongest-worded "mot" among the
   center-right cluster, not merely the median one.

KrF and Venstre have no committee seat and no documented merknad on
this case; their evidence files use attribution_confidence `ukjent`
rather than invented reasoning — the same two parties that were silent
on the digital-cloud case, distinct from utsagn-25's silent trio
(FrP/KrF/V) and utsagn-31's silent trio (Rødt/Sp/V), reinforcing the
earlier note that "which parties go undocumented" varies by committee
membership on the specific case, not by some fixed disengagement
pattern per party.

The other 20 voteringer on this case (forslag 2-10, 12-13, 15-21 —
covering a data-concession regime, a Microsoft 365 exit strategy,
Palantir, open-source procurement mandates, Datatilsynet funding, an
independent digital crisis council, and more) remain unshipped —
several are strong candidates for this dimension's remaining two
statement slots (20, 24) and for statement 44's still-open scope
question once that gets a deliberate decision instead of staying
parked.

## Phase 2k (2026-09-24, same session): utsagn-20, second statement of Dimension 4

Second and, for now, last statement to draw on Innst. 225 S before
moving to a fresh case: should the government build a national
exit-strategy away from large foreign commercial IT platforms such as
Microsoft 365, for rapid migration rather than a gradual one. Full 9/9
party coverage from votering_id 27490 (forslag 6, formally credited to
SV+Rødt in the vote metadata, but the merknad text names SV, Rødt AND
MDG as joint sponsors of the underlying argument — confirmed by
reading the sponsor-attribution sentence directly rather than trusting
the vote system's short title, the same discipline used for utsagn-19's
FrP/MDG alternates).

This is the same three-vote-mot/three-vote-for split pattern as
utsagn-19 in the direction of the parties (SV/Rødt/MDG for; the
remaining six mot), but the "for" bloc composition literally repeats —
worth noting as a data point for whichever statement-similarity or
"parties that vote together" feature comes later, not as a finding
about the case itself, since it's a natural consequence of testing two
forslag from the same three sponsors in the same innstilling.

One methodological finding worth flagging distinctly from the case
content: this is the first statement in the dataset where the
documented "majority" position (six parties voting mot) has *no
shared written merknad at all* — the innstilling contains no
"Flertallet"-style joint statement anywhere, a pattern that turned out
to be true of the whole innstilling, not just this one forslag. Ap in
particular has zero written merknad anywhere in Innst. 225 S despite
being the single largest mot-voting party by seat count (32 of the 86
mot votes) — contrasted directly against utsagn-19's Ap entry, which
DOES have a KVU-citing merknad, on the SAME innstilling, for the OTHER
forslag. This confirms explicitly what earlier session notes only
implied: a party's written-merknad presence is decided per-forslag,
not per-party-per-case, so "no merknad found" can never be treated as
a property of a party in a case — it has to be re-checked for every
individual statement built from that case, even when several
statements share the same source document.

Sp's card is coded with `attribution_confidence: svakt_dokumentert`
rather than treated as equal to Høyre's, despite sharing Høyre's
opening sentence ("Komiteens medlemmer fra Høyre og Senterpartiet
deler intensjonen [...]") — the innstilling narrows to "Komiteens
medlemmer fra Høyre" for the specific, detailed Microsoft-365
critique, so Sp's endorsement of that specific reasoning is not
textually supported, only the shared preceding generality. Reading
each merknad's own subject line at every sentence boundary, not just
at its start, is now confirmed necessary a second time (first case:
utsagn-19's MDG/Rødt boundary inside the same shared paragraph).

Statements 21-23 (Innovasjon/omstillingsevne, completely empty) and 24
(Teknologisk kapasitet's remaining slot) are the natural next targets;
Innst. 225 S's remaining 18 unbuilt voteringer (Palantir-forbud, åpen
kildekode-innkjøp, Datatilsynets ressurser, uavhengig digitalt
kriseråd, konsesjonsordning for persondata, m.fl.) are candidates for
statement 24 specifically, since they're all from the same
already-fully-read source and several test genuinely different
mechanisms (a procurement rule, a named-vendor ban, an oversight
body's funding) than either statement already built from this case.

## Phase 2l (2026-09-24, same session): utsagn-21, first statement of Dimension 5 (Innovasjon / omstillingsevne)

Innovasjon / omstillingsevne (21-23, "Bygger staten og markedet ny
kapasitet eller bare kjøper ferdige løsninger?") was, until this
statement, the only dimension of the nine with literally zero
content. New source: Innst. 424 S (2025-2026), a Sp-led
representantforslag (Dokument 8:267 S) to keep Innovasjon Norge's
regional-office structure intact against internal cost cuts. Full 9/9
party coverage, votering_id 28197.

This case is the most argumentatively differentiated in the dataset so
far — four genuinely distinct positions, not the usual two or three:
(1) Ap+MDG's efficiency framing (organisation should be judged by
measured effect, not by where offices are located, citing FIN's
industry-association input and the minister's own account); (2)
FrP+H+KrF's market-mechanism framing (state subsidy/loan/guarantee
growth "disturbs market mechanisms" and risks taxpayer money), with
FrP alone going further with a costed figure: its own alternative
budget already proposes ~1.6bn NOK in cuts to Innovasjon Norge — coded
-2, the strongest documented "mot" reasoning found in the project so
far, ahead even of Høyre's operational Microsoft-365 critique
(Phase 2k) for the same reason that critique earned -2: a claim
anchored in the party's own numbers, not just a stated preference; (3)
H+KrF's "let the agency choose" framing (agrees Innovasjon Norge
should be decentralized, but Stortinget shouldn't mandate the specific
organisational form) — the fourth confirmed instance of the
"shares the goal, opposes binding the mechanism" pattern first
documented in the dental case (Phase 2h); (4) SV+Sp+R's principled
defence of decentralization as a structural precondition for regional
development "in a country as long as Norway".

Høyre is textually a co-author of all three "mot" framings
simultaneously in this one case — broader coalition participation than
any other party has shown in any case built so far. Worth tracking
forward as a possible pattern (Høyre reading the room per-argument
rather than committing to one fixed coalition partner per case) rather
than concluding anything from a single instance.

MDG voted mot here, alongside Ap — not with SV/Sp/Rødt as in the
digital-sovereignty statements (19, 20). Confirms directly what the
project's evidence model is built to force: MDG's position is decided
per-case on its own merits (here, the efficiency argument outweighs
the regional-equity argument that usually aligns MDG with SV/Sp/Rødt),
not read off a fixed left-bloc assumption.

Venstre goes undocumented a third time (after utsagn-19 and utsagn-20)
for the same structural reason — no seat on the relevant committee —
but this time on a completely different committee (næringskomiteen,
not kommunal- og forvaltningskomiteen), reconfirming that which
parties go undocumented tracks committee membership per case, not a
standing disengagement pattern for any one party.

Forslag 2 in the same innstilling (Vekst- og gründersenteret i Førde,
a more geographically specific version of the same underlying
question) is a ready-made candidate for statement 22 or 23 without
opening a new source.

## Phase 2m (2026-09-24, same session): utsagn-37, first statement of Dimension 9 (Generasjonsansvar)

Generasjonsansvar (37, 40 — "Synliggjøres konsekvensene 10-30 år
frem?") was, until this statement, the second and last dimension with
literally zero content. New source: Innst. 263 S (2025-2026), an
SV-led representantforslag (Dokument 8:206 S) to raise the mandatory
occupational pension (OTP) minimum contribution rate from 2% to 4%.
Full 9/9 party coverage, votering_id 27876. Chosen deliberately for
how directly it tests the dimension's own question: the OTP minimum
rate has been unchanged since the scheme's 2006 introduction, and its
effect on any individual's actual pension is invisible until an entire
working life has passed — exactly the "consequence surfaces decades
out, cost is immediate and visible now" structure the dimension exists
to catch.

Four genuinely distinct positions, not the usual two or three: (1)
Flertallet (Ap, H, Sp, KrF, V — a genuine five-party joint merknad,
the first fully shared "mot" merknad with no per-party splinter text
found anywhere in the dataset so far) shares the goal explicitly
("forståelse for intensjonen") but wants a thorough impact assessment,
dialogue with labour-market social partners, and a gradual phase-in
tied to the budget process — a new variant of the recurring
"shares the goal, opposes the binding mechanism" pattern, this time
about *tempo and process* rather than about bindingness itself; (2)
FrP is textually co-author of SV/Rødt/MDG's own problem diagnosis
("store sosiale ulikheter i pensjon... folketrygden strammet inn
siden 2011") yet votes mot without joining Flertallet's gradualist
alternative and without any own documented remedy -- the most unusual
position found in the dataset yet: agreeing with the proposers'
diagnosis in writing while still rejecting their prescription, with no
stated reason why. Coded with attribution_confidence
svakt_dokumentert specifically because the vote is certain but the
motivation for that specific vote is not; (3) SV and Rødt, with the
strongest documented reasoning in the case (SV as original sponsor,
Rødt citing its own prior 2024-2025 representantforslag on the same
policy as continuity evidence); (4) MDG, a named co-sponsor of the
forslag text itself but explicitly excluded from SV/Rødt's deeper
elaboration (the LO citation and the 2011 reform critique) — the
third confirmed instance of "co-sponsors the ask, not the argument"
after Sp/Høyre in Innst. 424 S.

Methodological note worth carrying forward: this is the first case in
the dataset where the finanskomiteen's full nine-party membership
gave every party, including KrF and Venstre, an actual committee seat
and a documented written position — contrasted directly against
utsagn-19/20 (kommunal- og forvaltningskomiteen, no KrF or V seat) and
utsagn-21 (næringskomiteen, no V seat). Confirms once more, now with a
positive instance rather than only absences, that which parties go
documented tracks the specific committee handling the specific case,
not anything about the parties themselves.

Statements 38, 39, 41 and 42 (Robusthet / adaptabilitet, "Tåler
systemet bortfall, sjokk og feilantakelser?") are now the only fully
empty dimension left in the source memo. Forslag 2 and 3 in this same
innstilling (overtidsbetaling som pensjonsgivende inntekt;
obligatorisk uførepensjon) remain unbuilt and are ready-made
candidates for statement 40, Generasjonsansvar's second and last slot,
without opening a new source.

## Phase 2n (2026-09-24, same session): utsagn-38, first statement of Dimension 8 (Robusthet / adaptabilitet) — last empty dimension closed

Robusthet / adaptabilitet (38, 39, 41, 42 — "Tåler systemet bortfall,
sjokk og feilantakelser?") was the last of the memo's nine dimensions
with zero content. New source: Innst. 392 S (2025-2026), an Sp-led
representantforslag (Dokument 8:266 S) on strengthening Norway's food
preparedness. Full 9/9 party coverage, votering_id 28068 (forslag 1
of six in the case — the seed-grain/grain decentralized on-farm
storage proposal, and the closest vote in the case at 48-53, chosen
over the other five for how directly it names the shock scenarios the
dimension tests: "avlingssvikt, importstans eller andre ekstraordinære
situasjoner"). Two of the case's other forslag (romertall I/II, a
Finnish-model study and grocery-chain agreements) actually PASSED
57-43 — one of very few cases in the dataset where Stortinget adopted
part of what was proposed rather than rejecting all of it; noted in
the kontekst evidence but not built as its own statement since it's an
"utrede" ask rather than a binding mechanism.

The sponsor coalition on the built forslag (FrP, SV, Rødt, MDG, KrF)
crosses the left-right axis in a way not seen in any earlier case in
the dataset — worth flagging as a real pattern rather than
coincidence: food-security stockpiling is apparently one of few policy
areas where FrP and Rødt land on the same vote, for entirely different
reasons (FrP: Finland-model, individual-responsibility-compatible
physical reserves it has "lenge trukket frem som forbilledlig" alongside
the "individual is still responsible for egenberedskap" framing that
tracks FrP's broader anti-statist instinct; Rødt/SV/MDG: market failure
diagnosis, the state must not leave critical inputs to the market
alone).

The single most unusual individual position found in the dataset to
date: Senterpartiet, the case's own original sponsor party and a
co-sponsor of three of the other five forslag in the same innstilling,
voted MOT this one — with no Sp-specific reasoning documented anywhere
in the innstilling for that one choice. Coded attribution_confidence
ukjent rather than svakt_dokumentert specifically because there is no
textual anchor at all to assess, unlike Ap and Høyre in the same case,
both of which have at least a general (if not forslag-specific) stated
position. A plausible but unconfirmed reading is noted in the evidence
file's context_note and explicitly flagged as an interpretation, not a
documented fact — the discipline established across this whole session
of never inventing a party's motive holds even when a gap is this
tempting to fill.

Høyre's position here repeats the "co-authors the general skepticism,
not the specific alternative" pattern already seen for Sp/Høyre in
Innst. 424 S and MDG in Innst. 263 S — now confirmed a fourth time,
different party each time, suggesting this is a structural feature of
how Norwegian committee merknader are built in layers (a shared broad
sentence, then a narrower subset continuing into specifics) rather than
a quirk of any one party.

All nine dimensions now have at least one built statement. The
remaining gap is entirely about breadth and depth within dimensions,
not about any dimension being untouched: 27 of 42 memo statements
still unbuilt, and only Demokratisk åpenhet (1-5) is fully complete.

## Phase 2o (2026-09-24, same session): utsagn-26, second statement of Dimension 6 (Universell grunnmur)

Second statement in Universell grunnmur (25-30 — "Fungerer samfunnets
grunnleggende tjenester uansett bosted og inntekt?"), which already
had utsagn-25 (tannhelsereform) from an earlier session. New source:
Innst. 239 S (2025-2026), an Sp-led representantforslag (Dokument
8:194 S) on mobile coverage on all public roads by 2030. Full 9/9
party coverage, votering_id 27610 (the "alternativ votering" count for
Sp's own forslag text against the committee's "vedtas ikke"
recommendation — 9 for (Sp+KrF) against 90 mot, the most lopsided
vote in the dataset so far). The committee itself frames mobile
coverage as "en viktig del av samfunnets grunnleggende infrastruktur"
— a near-verbatim match to the dimension's own question.

Process note, worth remembering: this statement was originally drafted
as utsagn-25 before discovering — only via a browser screenshot, not
via `go test`, which stayed green throughout — that utsagn-25 already
existed (tannhelsereform, from an earlier session) and the loader does
not detect duplicate statement IDs. Renumbered to utsagn-26 (file,
`id`, `number`, and all ten evidence files' `statement_id`) before
committing. See project MEMORY.md for the operational lesson (always
`ls` the target statement-number file before creating one) — this is
the same "the artifact is the evidence, not the green test" pattern as
earlier drift-detection findings in this project, just one layer
closer to authoring than to CI.

Richest individual finding: Fremskrittspartiet's merknad is the
longest, most sympathetic text on the entire underlying problem
(trafikksikkerhet, beredskap, distriktene, nødetatenes responstid) of
any party in the case — longer than even Sp's own forslag text — yet
FrP voted mot with zero documented reason for opposing specifically
Sp's binding 2030 mechanism. Coded -1 (not -2) and flagged as the
clearest "shares the diagnosis in full, no stated reason for the vote"
case in the dataset to date. Høyre, by contrast, is the only party
with a complete, forslag-specific "mot" argument (the minister's
letter on cost/feasibility, explicit wait for Nkom's 2025 analysis) —
coded -2. KrF and Venstre both lack committee seats and both go
undocumented, but land on OPPOSITE sides of the vote (KrF for, Venstre
mot) — the first case in the dataset where two undocumented parties
diverge rather than defaulting to the same side.

Verified: go build/vet/test all green, gofmt-clean, pre-commit grep
for English-language slips clean. Coverage now 18 real statements ×
9 parties (161 items across the 18 non-tannhelse-duplicate-affected
statements; tannhelse's own 10 items untouched). Browser-verified via
playwright-cli at desktop (1280x900), iPad (810x1080, scrolled to
FrP/Høyre/KrF cards) and mobile (390x844) viewports: /utsagn/26
renders correctly at all three, zero console errors, and a second
check confirmed /utsagn/25 (tannhelse) still renders its own correct
content unaffected by the renumbering.


## Phase 2p (2026-09-24, same session): utsagn-39, second statement of Dimension 8 (Robusthet / adaptabilitet)

Second statement in Robusthet / adaptabilitet (38, 39, 41, 42 — "Tåler
systemet bortfall, sjokk og feilantakelser?"). Reuses the same
underlying case as utsagn-38 (Dokument 8:266 S, Innst. 392 S,
matberedskap) — a different one of the case's six numbered forslag,
so no new source document needed, only a second read of the same
innstilling plus a second votering export. Forslag 2 (votering_id
28067): a mandated, comprehensive national risk-and-vulnerability
analysis (ROS) for food security and food supply.

This is the broadest coalition split in the dataset to date: SIX
parties for (SV, Sp, Rødt, MDG, KrF, Venstre) against only THREE mot
(Ap, FrP, Høyre) — 26 for / 74 mot. The five sponsor parties share one
merknad citing Riksrevisjonen and Totalberedskapskommisjonen by name,
and stating as fact that the government itself has "varsla" (announced)
a national ROS-analysis that "har så langt ikkje materialisert seg" —
a direct, checkable claim that the government's own promised work is
overdue, coded +2 for all five sponsors.

Two findings worth surfacing for future development:

1. **FrP's clean split across two forslag in the same case.** FrP voted
   +2 FOR forslag 1 (utsagn-38, physical decentralized stockpiles, own
   named Finnish-model reasoning) but -1 MOT forslag 2 (this statement,
   formal risk mapping) with zero documented reason. A plausible reading
   — FrP already believes it knows the answer (build stockpiles) and
   doesn't want to wait on an analysis — is written into the evidence
   file's context_note explicitly flagged as interpretation, not fact,
   per this project's standing discipline against inventing motive.
2. **Venstre flips sides between the two forslag.** Venstre lacks a
   næringskomite seat in both cases (undocumented both times) but voted
   MOT forslag 1 and FOR forslag 2 — the first case in the dataset where
   the SAME undocumented party lands on opposite sides of two different
   forslag within one innstilling, rather than just diverging from
   another undocumented party (as utsagn-26's KrF/Venstre split did).

Structural observation flagged in the kontekst file: the three mot
parties (Ap, Høyre, FrP) are exactly the three with government
experience or proximity to it, and none of them write a forslag-specific
reason for opposing a risk analysis of their own food-security posture —
worth testing again the next time a "who has to explain themselves"
pattern shows up in this dataset.

Verified: go build/vet/test all green, gofmt-clean, no duplicate
statement IDs or numbers (grepped explicitly this time, given the
utsagn-25/26 collision two statements ago). Coverage now 20 real
statements × 9 parties (180 evidence items total). Browser-verified via
playwright-cli at desktop (1280x900), iPad (810x1080) and mobile
(390x844) viewports: /utsagn/39 renders correctly at all three, zero
console errors, 9/9 party coverage confirmed via curl before and after
the deploy.

## Phase 2q (2026-09-24, same session): utsagn-32, second statement of Dimension 7 (Barns muligheter)

Second statement in Barns muligheter (31-36 — "I hvilken grad
kompenserer samfunnet for forhold barn ikke har valgt?"). New source,
found by searching for a specific known-live policy area
("gratis SFO") rather than reusing an existing case: Prop. 160 L
(2024-2025), Innst. 78 L (2025-2026) — a government bill amending
opplæringslova to give pupils in grades 1-3 a right to 12 hours of
free skolefritidsordning (SFO) per week.

Genuinely rare shape for this dataset: Stortinget actually STRENGTHENED
the right beyond what the committee majority (H, FrP, KrF, V) wanted —
an opposition-led forslag (Ap, SV, Sp, Rødt, joined by MDG in the vote)
won an alternativ votering 52-49 to make the right unconditional and
statutory, rather than left to later regulation on staffing/financing.
The full amended law then passed 53-48. First statement in the dataset
where the "for" side's own merknad explicitly says what lovfesting
changes in practice: turns a reversible yearly budget line into a
right parents can invoke regardless of future budget fights.

The richest single find: FOUR distinct, independently documented mot
reasons exist for this one vote, rather than the usual pattern of one
shared reason or an undocumented majority:
- **KrF** (-2): explicit resource-prioritization ("ikke riktig
  prioritering av ressursene når skolen sårt trenger et løft").
- **Venstre** (-2): a principled unfunded-mandate objection to the
  state imposing costs on municipalities without matching funding —
  notable because Venstre actually HAS a seat on this committee
  (utdannings- og forskningskomiteen), unlike the næringskomiteen
  cases (utsagn-21, 38, 39) where Venstre went undocumented for lack
  of a seat.
- **FrP** (-1, weaker): its own forslag 2 (fell 30-71) asked for an
  evaluation of staffing/quality/financing FIRST, not a rejection of
  free SFO as a policy — the same "wants a pause, not a no" shape as
  its evaluation-first coding elsewhere in this dataset.
- **Høyre** (-2): a municipal-autonomy argument that directly quotes
  the government's OWN admission that free-SFO-so-far has already
  made things worse for pupils needing extra support — used as
  ammunition against extending the same policy further before that
  problem is fixed.

MDG (+1, weakly documented — voted with the sponsor bloc but has no
own or shared merknad in the text) is the only party without a
documented reason on either side of this vote.

Verified: go build/vet/test all green, gofmt-clean, explicit grep for
duplicate statement IDs/numbers (clean). Coverage now 21 real
statements x 9 parties (190 evidence items total). Browser-verified
via playwright-cli at desktop (1280x900), iPad (810x1080) and mobile
(390x844) viewports: /utsagn/32 renders correctly at all three, zero
console errors, 9/9 party coverage confirmed via curl.

## Phase 2r (2026-09-24, same session): utsagn-43 renumbered back to utsagn-07 — into Dimension 2 (Norsk handlefrihet)

Full circle from the "Correction" narrative above (line ~307): this
statement was originally drafted as utsagn-06/07/08 under an invented
dimension name, discovered not to match the memo, and moved OUT to
43/44/45 as an addendum. `docs/research-briefs/TASKS.md`'s Task 1
explicitly flagged all three as strong candidates to fold back into
the real Dimension 2 (Norsk handlefrihet, statements 6-11,
"Beholder Norge reell beslutnings- og operasjonsevne?") once that
dimension was actually drafted from the memo's own question — which
utsagn-06 (kjernekraft-statlig-kontroll) already did.

utsagn-43 (energimarked-suverenitet / ACER) is a clean fold-in: its
own `primary_question` already read verbatim as the dimension's real
question (just with an explanatory parenthetical for the addendum
framing), it already has full 9/9 party evidence, and its topic — can
Norway keep real control over electricity pricing against EU/EEA
energy-market integration (ACER) — is squarely "beholder Norge reell
beslutnings- og operasjonsevne." No new research needed: renamed the
statement file, changed `id`/`number`/`dimension`/`primary_question`,
and updated `statement_id` in its 10 evidence files (41-50) plus six
other evidence files that cross-reference it by number for Konsistens
comparisons (54, 55, 56, 57, 58, 63) so those citations still resolve.

utsagn-44 (digital-suverenitet) and utsagn-45 (forsvar-beredskap-
suverenitet) were NOT folded in this pass: utsagn-44 has a documented
scope mismatch (nordisk vs. europeisk jurisdiksjon, see
`docs/research-notes/dimension-2-3-leads.md`) and only 2 evidence
items; utsagn-45 has zero evidence (no matching decided case found).
Both remain addendum statements pending either a text revision or a
newly decided case.

Dimension 2 (Norsk handlefrihet) is now 2 of 6 real (06, 07). The
Tillegg addendum shrinks from three statements to two (44, 45).

Verified: go build/vet/test all green after the renumber (drift-
detection counts unchanged — same statement/evidence totals, 21
statements, 190 evidence, since this was a rename not an addition),
gofmt-clean, no duplicate statement IDs/numbers, `grep -rl
"utsagn-43"` returns nothing outside historical narrative docs
(roadmap.md's own "Correction" section and design-review-2026-09-23.md,
left as historical record of the original mistake, not live
cross-references). Browser-verified live at ektevalgomat.no:
/utsagn/07 renders the same content that was previously at /utsagn/43,
9/9 party coverage, zero console errors, at desktop/iPad/mobile.

## Phase 2s (2026-09-24, same session): utsagn-40 — Dimension 9 (Generasjonsansvar) now COMPLETE (2/2)

Second and final statement in Generasjonsansvar (37, 40 —
"Synliggjøres konsekvensene 10–30 år frem?"). New case: Dokument 8:19
L (2025-2026), Innst. 54 L (2025-2026), FrP's representantforslag to
abolish "samordningsfellen" — a mechanism from the 2011 pension reform
that reduces public-sector service pension the longer someone works
past age 67, specifically hitting the cohorts born 1944-1962. A
textbook test of the dimension's question: a 14-year-old reform's
side effect on one named, bounded generation, still unresolved.

Three floor amendments were put during the 2025-12-16 debate, not
printed in the innstilling itself (only visible via the vote list and
the vedtak-i-korthet page, not the standard innstilling text — a new
retrieval pattern for this dataset, worth remembering for future
cases where "Forslag N" appears in votering_tema but not in the
printed committee report):
- **Forslag 1** (FrP, full lovendring): fell 29-73, FrP alone.
- **Forslag 3** (FrP, SV, Rødt, a softer alternative in the same
  spirit as SV's own merknad): fell 40-60.
- **Forslag 2** (MDG, alone, an anmodningsvedtak asking government to
  investigate and report back on cost/economic effects of a fix by
  1.1.2027): WON 51-50 — the actual adopted text, Vedtak 164.

The winning coalition on forslag 2 is genuinely cross-cutting: FrP,
MDG, Rødt, SV, Sp and Venstre for; Ap, Høyre and KrF against — neither
a left-right split nor a government-vs-opposition split. Ap's own
merknad makes an explicit consistency argument against FrP itself
("kun har fremmet forslag av denne typen når de har vært i
opposisjon... unnlot å gjennomføre slike endringer da de selv var
med i en flertallskonstellasjon fra 2013 til 2021") rather than
rebutting the specific investigate-request — coded -1, the same
"motive/process objection, not a substantive rebuttal" shape seen
elsewhere in this dataset. Høyre's own merknad is a bare citation of
five prior committee reports without restating their content — real
text, but thin (-1). SV's own merknad is the most substantive on
either side: names a specific state minister's reply letter (17
January 2024) and defends the 384 million kroner cost of full
abolition explicitly.

MDG (+2) is coded on the strength of being the sole sponsor of the
literal adopted text itself — no prose merknad exists (it was a floor
amendment, not part of the printed committee report), but the ask's
own wording is the documentation. Sp and Venstre (+1 each) were the
two parties whose votes were structurally decisive in a 51-50 result,
yet neither left any documented reasoning — flagged explicitly in
both context_notes as "undocumented does not mean unimportant" in
this dataset's vote-counting logic.

Verified: go build/vet/test all green, gofmt-clean, no duplicate
statement IDs/numbers. Coverage now 22 real statements x 9 parties
(200 evidence items total, a round number). Browser-verified via
playwright-cli at desktop (1280x900), iPad (810x1080) and mobile
(390x844) viewports: /utsagn/40 renders correctly at all three, zero
console errors, 9/9 party coverage confirmed via curl.

## Phase 2t (2026-09-24, same session): utsagn-24 — Dimension 4 (Teknologisk kapasitet) now COMPLETE (3/3)

Third statement in Teknologisk kapasitet (19, 20, 24 — "Kan kritiske
digitale funksjoner og kompetanse opprettholdes under press?"). Same
underlying case as utsagn-19 and utsagn-20 — Dokument 8:119 S
(2025-2026), Innst. 225 S (2025-2026), SV's representantforslag "om
digital suverenitet i en urolig tid" — but a THIRD, distinct forslag
from that same innstilling (forslag 8, SV+Rødt+MDG): a ban on Norwegian
public-sector use of Palantir's systems, citing Palantir's ties to
American intelligence and the US Department of Defense, and naming
Tolletaten's (Norwegian Customs) existing use of it directly. Avvist
17-84 (votering_id 27488).

Retrieval note worth keeping: this innstilling has 22 total votes
(21 mindretall forslag + komiteens tilråding) for sakid 200036, and
the printed innstilling text does not surface individual votering_id
per forslag. Two calibration points from prior work (forslag 11 →
votering_id 27485, forslag 6 → votering_id 27490) revealed an exact
formula: votering_id = 27496 − forslagsnummer (27496 being the
tilråding's own vote). Verified against a third point (forslag 9,
Høyre-only sponsor → predicted 27487 → confirmed Høyre-only 14 "for"
via data.stortinget.no's per-representative export) before trusting
it for forslag 8 → 27488. This formula is LOCAL to this specific
sak's floor-voting order and must be re-derived (not assumed) for any
other sak — it happened to hold here because Stortinget voted this
particular innstilling's forslag in strict reverse-numeric order.

Genuinely interesting substantive finding: FrP and Sp both have a
committee merknad (shared with Rødt) explicitly agreeing "at det må
tas grep for å sikre større digital suverenitet, og at dagens
sikkerhetspolitiske situasjon tilsier en dreining mot større egenevne
og fokus på beredskap" — yet both voted MOT the one concrete,
named-vendor security measure on the table. A rhetoric/action gap,
coded -1 (documented general sympathy, not a rebuttal of the specific
ask) rather than -2. Contrast with FrP's utsagn-40 record (evidence/202),
where the party's rhetoric and vote matched all the way through — this
dataset now has a clean pair for "does a party's general position
survive contact with the specific, named mechanism" as an analytical
angle for later feature work. Høyre's own "smart suverenitet"
merknad (diversification and strict security requirements instead of
bans/isolation) is the most developed reasoning on the mot side but
never names Palantir, so it stays at -1/svakt_dokumentert rather than
-2. Venstre and KrF are both outside kommunal- og forvaltningskomiteen
for this sak (unlike Sp, which IS seated via Bengt Fasteraune) — and
landed on OPPOSITE sides of this specific vote (V for, KrF mot),
neither with any documented reasoning.

Verified: go build/vet/test all green (23 statements, 210 evidence —
another round number), gofmt-clean, no duplicate statement IDs/numbers.
Browser-verified via playwright-cli at desktop (1280x900), iPad
(810x1080) and mobile (390x844): /utsagn/24 renders correctly at all
three, zero console errors, 9/9 party coverage confirmed via curl.

## Phase 2u (2026-09-24, same session): utsagn-22 — Dimension 5 (Innovasjon/omstillingsevne) now 2/3

Fourth statement mined from the same rich source as utsagn-19/20/24:
Dokument 8:119 S / Innst. 225 S (2025-2026), digital suverenitet.
This time forslag 10 (Høyre alone): utrede felleskommunale
innkjøpsordninger for KI etter modell av den danske selskapsmodellen
KOMBIT, so municipalities buy AI capability jointly instead of each
negotiating separately. Avvist 23-77 (votering_id 27486, confirmed by
the same 27496-minus-forslagsnummer formula documented in Phase 2t).

Genuinely new "kjøtt": this is the only forslag in the entire Innst.
225 S sak where a party's own floor vote splits from its own
sponsored text — 12 of 13 present Høyre representatives voted for
their own party's forslag, one voted mot. No written explanation
exists in the innstilling; flagged explicitly in the context evidence
as future analytical material (either a genuine individual dissent or
a data artifact, both worth chasing later). This dataset's dissent
mechanic (intra-party vote splits) has appeared before at the
representative level but this is the cleanest single-file example of
a party voting against its own named forslag.

Weaker evidence bar than utsagn-24: none of the five mot-side parties
(Ap, FrP, Rødt, SV, Sp) have a merknad specific to KOMBIT/AI
procurement anywhere in the innstilling — the general merknader in
this sak are all about cloud sovereignty, Microsoft 365, Palantir, or
open source, not municipal joint purchasing. Coded from vote count
alone (uklart/ukjent) for all five, an honest gap rather than a forced
connection to adjacent merknader. KrF, MDG and Venstre (all +1, voted
for) are similarly undocumented on this specific ask, following
Høyre's forslag without stated reasoning of their own.

Verified: go build/vet/test all green (24 statements, 220 evidence),
gofmt-clean, no duplicate statement IDs/numbers. Browser-verified via
playwright-cli at desktop (1280x900), iPad (810x1080) and mobile
(390x844): /utsagn/22 renders correctly at all three, zero console
errors, 9/9 party coverage confirmed via curl.

## Phase 2v (2026-09-24, same session): utsagn-23 — Dimension 5 (Innovasjon/omstillingsevne) now COMPLETE (3/3)

Fifth and final statement mined from Innst. 225 S (2025-2026, digital
suverenitet) — the same source now used for utsagn-19, 20, 22, 23 and
24. This time forslag 7 (SV, Rødt, MDG): opprette et nasjonalt
program for åpen kildekode i offentlig sektor med egne bevilgninger
til kompetanseheving. Avvist 16-85 (votering_id 27489, same
27496-minus-forslagsnummer formula).

Of the five statements sourced from this one sak, this is the
cleanest fit to its dimension's actual question ("bygger staten ny
kapasitet eller bare kjøper ferdige løsninger?") — a program with its
own funding line for competence-building and open tooling, tested
directly against continuing to buy proprietary commercial software.

Second instance in this same sak of an intra-party floor-vote split:
4 of 5 present MDG representatives voted for their own party's
co-sponsored forslag, one voted mot — the mirror of utsagn-22's Høyre
split (12 of 13 for their own forslag, one mot). Two independent,
undocumented splits in one innstilling is itself a finding worth
flagging as a candidate feature: an automatic "internal party
discipline" flag comparing a party's own sponsored-forslag vote
tally against its full floor strength.

Verified: go build/vet/test all green (25 statements, 230 evidence —
another round number), gofmt-clean, no duplicate statement IDs/numbers.
Browser-verified via playwright-cli at desktop (1280x900), iPad
(810x1080) and mobile (390x844): /utsagn/23 renders correctly at all
three, zero console errors, 9/9 party coverage confirmed via curl.

## Phase 2w (2026-09-24, same session): utsagn-41 — Dimension 8 (Robusthet/adaptabilitet) now 3/4

Sixth statement mined from Innst. 225 S (2025-2026, digital
suverenitet), used here for a different dimension: forslag 21 (MDG
alone): utrede etablering av et uavhengig nasjonalt digitalt
kriseråd. Avvist 8-91 (votering_id 27475, same formula).

This is the most dramatic intra-party finding across all six
statements sourced from this sak. Rødt — which votes in lockstep with
MDG on every other digital-sovereignty forslag tested (Palantir ban,
national cloud, KOMBIT, open source) — split 4 mot / 1 for here,
nearly the OPPOSITE of its usual alignment. SV split exactly down the
middle among attending members (2 for / 2 mot / 5 absent), coded
position 0 (nøytral/delt) rather than forcing a +1/-1 call the vote
tally does not support. Venstre, which had voted for all four prior
statements from this sak, voted mot here — the only break in its
otherwise consistent pattern.

Confirms position:0 is a legitimate, renderable value end to end: the
scoring/loader layer accepted it without special-casing, and the
frontend renders it as "NØYTRALT / DELT" — verified live via
playwright-cli screenshot, not just asserted from the YAML.

Verified: go build/vet/test all green (26 statements, 240 evidence),
gofmt-clean, no duplicate statement IDs/numbers. Browser-verified via
playwright-cli at desktop (1280x900, including a dedicated screenshot
of the SV "NØYTRALT / DELT" card), iPad (810x1080) and mobile
(390x844): /utsagn/41 renders correctly at all three, zero console
errors, 9/9 party coverage confirmed via curl.

## Phase 2x (2026-09-24, same session): utsagn-08/09 -- Norsk handlefrihet 2/6 -> 4/6

New source: Innst. 374 S (2025-2026), energi- og miljokomiteen, on
Dok. 8:235 S from Rodt (Sofie Marhaug m.fl.) about kraftEKSPORT
control. Confirmed endpoint name is `voteringer?sakid=X` (not
`voteringliste` -- that started 404ing mid-session on the same host
that served it fine earlier; re-verified identical data via the
working endpoint before trusting anything new).

Two statements from one case:
- utsagn-08: Statnett export-limiting mechanism + cable renegotiation
  + hybrid-cable ban (forslag 1-3, voted as ONE bloc, votering_id
  28137). FrP+SV+Sp+Rodt+KrF for (47), Ap+H+MDG+V+1 FrP mot (54).
  A genuine cross-bloc coalition: FrP on the right voting with
  SV/Sp/Rodt on the left, against the Ap+H+MDG+V governing-adjacent
  majority -- neither the usual left-right axis nor the usual
  government/opposition axis.
- utsagn-09: stop NVE's Skagerrak 1/2 reinvestment study (forslag 4,
  FrP+Sp+Rodt only, votering_id 28136). SV is NOT a co-sponsor here
  and votes MOT, the opposite of utsagn-08 -- a real, vote-documented
  nuance in SV's position (broad export-control principle yes,
  this specific concrete cable study no), not just absence noise.
  KrF also flips (for on 08, mot on 09). Rodt's own merknad accuses
  Ap of a documented campaign-promise reversal on these cables; Ap's
  evidence file flags this as *unanswered* in the innstilling rather
  than coding a fabricated rebuttal.

Both use partiets_begrunnelse where a real quoted merknad exists
(FrP, SV, Sp, Rodt, Ap, Hoyre, MDG, Venstre all have one on utsagn-08;
fewer on utsagn-09 since it's a narrower forslag) and uklart where the
vote is the only evidence.

Verified: go build/vet/test all green (29 statements, 270 evidence),
gofmt-clean, no duplicate statement or evidence IDs. Browser-verified
via playwright-cli: /utsagn/9 desktop (1280x900, zero console errors)
and mobile (390x844, screenshot confirmed clean render); both new
pages return 9/9 party coverage via curl.

Remaining after this: Norsk handlefrihet needs 2 more (10, 11).
Forvaltning ressurser needs 5 (14-18) -- NOTE: forslag 12
("leverandorlasing apen kildekode") and reduserer/interoperabilitet
themes from Innst 225 S don't fit this dimension (that's tech
sovereignty, already used); need genuinely resource/energy-control
cases. Universell grunnmur needs 4 (27-30). Barns muligheter needs 4
(33-36). 15 statements still missing across 4 dimensions.

## Phase 2y (2026-09-24, same session): utsagn-14/15 -- Forvaltning ressurser 2/7 -> 4/7

New source: Innst. 354 S (2024-2025), energi- og miljokomiteen, on
Dok. 8:206 S from FrP (Terje Halleland m.fl.) about havbunnsmineraler
(seabed minerals). sakid=103035, four mindretallsforslag plus an
adopted main tilraading (romertall I/II, gradual approach).

utsagn-14: start 1st concession round for seabed minerals early 2026
(forslag 1, H+FrP, votering_id 25349, for=35/mot=70). utsagn-15: work
internationally (ISA) for a moratorium until proven safe (forslag 2,
SV+Rodt+Venstre+MDG+KrF, votering_id 25348, for=23/mot=82) -- almost
the exact mirror-image coalition of utsagn-14.

This is a clean environment-vs-growth axis, NOT left-right: Ap+FrP+
Hoyre+Sp together for extraction-with-conditions (all four voted for
the adopted main tilraading, 82-23), against SV+Rodt+Venstre+MDG+KrF
together for caution. Notable: FrP and MDG cite the *same* fact
(China would benefit most from international seabed opening) to reach
opposite conclusions (FrP: reduce dependency by extracting ourselves;
MDG: an international opening mainly helps China). Ap and Sp both get
position -1 (not -2) on utsagn-14: they oppose a fixed timeline but
voted FOR the more gradual adopted tilraading -- "yes in principle,
not this timeline" is documented as a real nuance, not manufactured.

Verified: go build/vet/test all green (31 statements, 290 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/14 desktop (1280x900) and /utsagn/15 iPad
(810x1080, screenshot confirmed), zero console errors on both; 9/9
party coverage confirmed via curl.

Remaining: Forvaltning ressurser needs 3 more (16,17,18). Universell
grunnmur needs 4 (27-30). Barns muligheter needs 4 (33-36). Norsk
handlefrihet needs 2 (10,11). 13 statements still missing across 4
dimensions.

## Phase 2z (2026-09-24, same session): utsagn-16/17 -- Forvaltning ressurser 4/7 -> 6/7

Two more cases: Innst. 377 S (2025-2026, finanskomiteen) on avvikling
of Prisradet for havbruk (Hoyre's Dok. 8:185 S) -- utsagn-16, VEDTATT
54-47 (FrP+H+KrF+Sp+V for, Ap+MDG+Rodt+SV mot, votering_id 28077).
Rich quotes on both sides about who the normprice actually protects
(FrP: it hurts small independent operators via internal-pricing
loopholes for big vertically-integrated firms; Ap/SV/Rodt: removing
it is what opens that same loophole -- same mechanism, opposite
read).

And Innst. 388 S (2023-2024, naeringskomiteen) on landbasert
mineralutvinning (31 mindretallsforslag, Dok. 8:125 S from H+V+FrP+
MDG-representatives) -- utsagn-17, forslag 28 (SV+Rodt: opprette
statlig mineralselskap), avvist 12-89 (votering_id 22966). Rodt names
a real company (Norge Mining, Swiss-owned) as the case example.
FrP's mot vote is nuanced (-1 not -2): FrP wants a STATE CAPITAL FUND
for private operators (per Innst. 74 S), not a state-run extraction
company -- state involvement yes, state operation no.

Verified: go build/vet/test all green (33 statements, 310 evidence --
loader_test.go had an arithmetic slip mid-session, caught by the test
itself failing with the wrong expected count, fixed before commit).
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/16 mobile (390x844) and /utsagn/17 desktop
(1280x900), zero console errors on both; 9/9 party coverage confirmed
via curl.

Remaining: Forvaltning ressurser needs 1 more (18). Universell
grunnmur needs 4 (27-30). Barns muligheter needs 4 (33-36). Norsk
handlefrihet needs 2 (10,11). 11 statements still missing across 4
dimensions.

## Phase 3a (2026-09-24, same session): utsagn-27/28 -- Universell grunnmur 2/6 -> 4/6

New source: Innst. 252 S (2025-2026, helse- og omsorgskomiteen), on
MDG's Dok. 8:133 S about smaller hospitals, decentralized outpatient
services, and shorter travel distance for care. sakid=200056.

utsagn-27: ambulance response-time targets in ALL municipalities
(forslag 1, FrP+Sp+Rodt, votering_id 27581, for=45/mot=54). Hoyre
co-authors a majority merknad WITH the sponsors saying capacity must
be secured in rural areas, then votes mot the binding text anyway --
same "yes in principle, no to this specific ask" pattern as
utsagn-08/14. MDG (the case's OWN original sponsor) votes mot this
forslag because FrP/SV/Sp/Rodt's shared merknad explicitly names and
criticizes MDG's own alternative proposals (meeting points, home-duty
combos) as a disguised weakening.

utsagn-28: use local hospitals actively to cut wait times (forslag 3,
SV+Sp+Rodt, votering_id 27579, for=22/mot=77). Same 9 parties, same
sak, but FrP and MDG SWAP sides relative to utsagn-27 (FrP for->mot,
MDG mot->for) -- documented proof that coalitions in this dataset are
per-forslag, not per-sak. Evidence files cross-reference both to make
the swap visible rather than treating each statement in isolation.

Verified: go build/vet/test all green (35 statements, 330 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/27 desktop (1280x900) and /utsagn/28 mobile
(390x844), zero console errors on both; 9/9 party coverage confirmed
via curl.

Remaining: Universell grunnmur needs 2 more (29,30). Barns muligheter
needs 4 (33-36). Norsk handlefrihet needs 2 (10,11). Forvaltning
ressurser needs 1 (18). 9 statements still missing across 4
dimensions.

## Phase 3b (2026-09-24, same session): utsagn-33/34 -- Barns muligheter 2/6 -> 4/6

New source: Innst. 126 S (2025-2026, familie- og kulturkomiteen), on
Venstre's (Abid Raja, Guri Melby, Grunde Almeland) Dok. 8:11 S about
a comprehensive package against social inequality. sakid=104872, 22
mindretallsforslag.

utsagn-33: increased barnehage staffing in disadvantaged areas +
funding for minority-language support (forslag 1+2 bundled,
SV+MDG+KrF, votering_id 26946, for=16/mot=85). Venstre votes for its
OWN underlying case even though SV/MDG/KrF are the formal sponsors of
this specific text. Ap writes an explicit "we share this commitment"
merknad then votes mot anyway -- same yes-in-principle pattern as
Hoyre on digital suverenitet (evidence/259). Rodt votes mot despite
ideological proximity to SV, undocumented reason.

utsagn-34: free kindergarten place (not just free core hours) for
low-income families (forslag 5+6+8 bundled, SV+MDG, votering_id
26943, for=12/mot=89). KrF FLIPS from for (utsagn-33) to mot here --
not a co-sponsor on this one, and KrF's own preferred mechanism is a
higher cash child benefit (barnetrygd), not a free public service --
same goal, different instrument, both documented via merknader in
this same case.

Checked forslag 4 (barnetrygd prisjustering) against existing
statement 31 before starting -- same theme, correctly NOT duplicated.

Verified: go build/vet/test all green (37 statements, 350 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/33 iPad (810x1080) and /utsagn/34 desktop
(1280x900), zero console errors on both; 9/9 party coverage confirmed
via curl.

Remaining: Universell grunnmur needs 2 more (29,30). Barns muligheter
needs 2 (35,36). Norsk handlefrihet needs 2 (10,11). Forvaltning
ressurser needs 1 (18). 7 statements still missing across 4
dimensions.

## Phase 3c (2026-09-24, same session): utsagn-29/30 -- Universell grunnmur COMPLETE (6/6)

New source: Innst. 280 S (2025-2026, naeringskomiteen), on Rodt's
(Remi Solvberg, Marie Sneve Martinussen) Dok. 8:229 S about postal
service for the future. sakid=200183.

utsagn-29: nationwide dorterskeltjeneste (doorstep delivery for
non-digital/vulnerable recipients) (forslag 1, Sp+Rodt, votering_id
27675, for=23/mot=77). First documented INTERNAL KrF split in this
dataset: 3 of 4 attending for, 1 mot (KrF had until now always voted
unanimously in either direction across every prior statement).

utsagn-30: statutory duty to ensure newspaper distribution in areas
without commercial networks (forslag 2, Sp+Rodt, votering_id 27674,
for=16/mot=84). MDG splits 1 for / 4 mot here -- opposite of its
unanimous FOR on utsagn-29 in the same case. Two internal splits
(KrF, MDG) in one innstilling is unusual in this dataset and flagged
explicitly in both context files rather than smoothed over.

Verified: go build/vet/test all green (39 statements, 370 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/29 desktop (1280x900) and /utsagn/30 mobile
(390x844), zero console errors on both; 9/9 party coverage confirmed
via curl.

Remaining: Barns muligheter needs 2 (35,36). Norsk handlefrihet needs
2 (10,11). Forvaltning ressurser needs 1 (18). 5 statements still
missing across 3 dimensions.

## Phase 3d (2026-09-24, same session): utsagn-35 -- Barns muligheter 4/6 -> 5/6

New source: Innst. 55 S (2024-2025, utdannings- og forsknings-
komiteen), treatment of Meld. St. 34 (2023-2024) "En mer praktisk
skole". A large omnibus meld-innstilling (66+ mindretallsforslag);
this statement uses one of the VEDTATT romertall, not a rejected
mindretallsforslag.

utsagn-35: government to present a phase-in plan for a daily healthy
school meal for all primary-school pupils by budget year 2026
(romertall IV, VEDTATT 61-37, votering_id 23652). Ap+KrF+MDG+Rodt+
SV+Sp for, FrP+Hoyre+Venstre mot. SV's education spokesperson Grete
Wold's post-vote quote ("Endelig er skolemat pa vei!") captured
verbatim from news coverage (frifagbevegelse/utdanningsnytt), since
the innstilling itself truncates before the full merknad text at
50KB fetch limit -- primary-source vote data still came from
data.stortinget.no, only the quote is secondary-sourced, flagged as
such in source_tier/source_name.

Verified: go build/vet/test all green (40 statements, 380 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/35 desktop (1280x900), zero console errors;
9/9 party coverage confirmed via curl.

Remaining: Barns muligheter needs 1 more (36). Norsk handlefrihet
needs 2 (10,11). Forvaltning ressurser needs 1 (18). 4 statements
still missing across 3 dimensions.

## Phase 3e (2026-09-24, same session): utsagn-36 -- Barns muligheter COMPLETE (6/6)

New source: Innst. 136 S (2025-2026, familie- og kulturkomiteen),
treatment of Meld. St. 28 (2024-2025) "Tro pa framtida - uansett
bakgrunn" (same underlying meld as the Innst. 126 S source for
utsagn-33/34, but a different parliamentary item). sakid=103839.

utsagn-36: lovfeste barnets beste as a fundamental consideration in
the sosialtjenesteloven (forslag 4, SV+KrF, votering_id 26988,
for=21/mot=80).

Data-quality note worth carrying forward: data.stortinget.no's
voteringer list for this sakid names two additional votes as
"Forslag nr. 10/11 pa vegne av Sp", but the published innstilling
text (fetched via direct curl + HTML tag-stripping to bypass
web_fetch's 50KB truncation limit -- see bash technique in this
phase) contains only forslag 1-9, none sponsored by Sp. Rather than
build on unverifiable data, only forslag 4 was used here, since its
sponsor list (SV og KrF) is independently confirmed against both the
innstilling text and the vote breakdown sum (21 for matches exactly).
The two Sp-labeled votes were deliberately left unused.

This completes Barns muligheter (31,32,33,34,35,36 = 6/6).

Verified: go build/vet/test all green (41 statements, 390 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/36 iPad (810x1080), zero console errors; 9/9
party coverage confirmed via curl.

Remaining: Norsk handlefrihet needs 2 (10,11). Forvaltning ressurser
needs 1 (18). 3 statements still missing across 2 dimensions --
closest to full completion (41/42) yet in this project.

## Phase 3f (2026-09-24, same session): utsagn-10/18 -- Forvaltning ressurser COMPLETE (7/7), Norsk handlefrihet 5/6

New sources:
- Innst. 266 S (2025-2026, kommunal- og forvaltningskomiteen), FrP's
  Dok. 8:162 S about a fast-track building law for defense
  infrastructure in war/crisis (Danish-model exemption from plan- og
  bygningslov). sakid=200087. utsagn-10: FrP+Hoyre's forslag 1
  (special legal exemption), decided via an "alternativ votering"
  (mutually exclusive choice between the tilraading and forslag 1),
  votering_id 27597: 57 chose the tilraading/rejection
  (Ap+KrF+MDG+Rodt+SV+Sp+V+1 FrP), 41 chose forslag 1 (FrP+Hoyre). Sp's
  merknad names the actual reason: existing tools (statlig
  reguleringsplan + forskrift-based exemptions) already cover the
  need per the statsraad's own answer -- coded -1 not -2, since this
  is "already solved" not "opposed in principle".

- Innst. 264 S (2023-2024, naeringskomiteen), the fisheries quota
  white paper "Folk, fisk og fellesskap". sakid=97159. utsagn-18:
  SV+Rodt's forslag 15, reinstating an actual 28-metre length
  boundary between coastal and ocean fishing fleets (votering_id
  22561, for=14/mot=87). Sp splits internally 1 for / 16 mot -- the
  third documented internal split in this dataset (after KrF and
  MDG on the postal-service case). The majority (Ap+Hoyre+Sp+FrP+V)
  countered with their OWN alternative boundary
  (hjemmelslengde-based, less strict) rather than rejecting the
  concept of a boundary outright -- coded Ap/H/V as -1 (nuanced
  counter-proposal), FrP as -2 (the only party arguing against any
  length-based boundary on principle).

This completes Forvaltning av fellesskapets ressurser/energi
(12,13,14,15,16,17,18 = 7/7).

Verified: go build/vet/test all green (43 statements, 410 evidence),
gofmt-clean, no duplicate statement/evidence IDs. Browser-verified via
playwright-cli: /utsagn/10 desktop (1280x900) and /utsagn/18 mobile
(390x844), zero console errors on both; oversikt page screenshot
confirms live per-dimension counts (Forvaltning ressurser 7/7,
Universell grunnmur 6/6, Barns muligheter 6/6, Robusthet 4/4, all
correct); 9/9 party coverage on both new pages via curl.

Remaining: Norsk handlefrihet needs 1 more (11). Then all 42
statements + the 2 extras (44,45) are complete.

## Phase 3g (2026-09-24, same session): utsagn-11 -- Norsk handlefrihet COMPLETE (6/6); ALL 42 STATEMENTS NOW BUILT

New source for utsagn-11: Innst. 140 S (2025-2026, energi- og
miljokomiteen), SV's (Kirsti Bergstø m.fl.) Dok. 8:60 S about a
national concession system for data centers. sakid=105697. Forslag 1
(Hoyre, SV, Sp, Rodt og MDG -- a rare cross-bloc coalition spanning
the whole left-right axis), votering_id 27047, for=40/mot=62.
Venstre splits exactly 1-1. Ap and FrP are the only two parties
voting together against it -- the broadest for-coalition against the
narrowest mot-coalition anywhere in this dataset. Hoyre and Sp share
a "national strategic thinking" merknad; SV/Sp/Rodt/MDG have a
sharper ownership-and-control merknad; MDG's own addition
("norsk eller europeisk eierskap" as an explicit requirement) is the
most specific of all.

Completes Norsk handlefrihet (6,7,8,9,10,11 = 6/6). This is the last
of the 8 workable dimensions -- Robusthet/adaptabilitet, Universell
grunnmur, Barns muligheter, Forvaltning ressurser, Norsk handlefrihet
are now all fully built alongside the already-complete Demokratisk
apenhet, Teknologisk kapasitet, Innovasjon/omstillingsevne, and
Generasjonsansvar.

## Gap-fill pass (same session): utsagn-05 filled from 0/9 to 9/9

A full sweep of all 42 statements (not just the ones built this
session) found utsagn-05 ("Etterprøvbarhet i åpenhetssaker") had
ZERO evidence files -- a pre-existing gap from before this session,
undocumented in the roadmap. Filled via Innst. 140 S (2023-2024),
originally Rødt's Dok. 8:1 S about strengthening habilitet rules;
MDG took up one of its points as its own committee forslag 3:
"utrede og vurdere sine muligheter til å sanksjonere mot brudd på
habilitetsreglene og overprøve statsråders habilitetsvurderinger" --
a literal etterprøvbarhet (external review) mechanism for ministers'
own conflict-of-interest self-assessments. votering_id 22102,
for=7 (MDG+Rodt+1 independent)/mot=93.

Also attempted a similar gap-fill for utsagn-06 (7/9, missing SV and
KrF) using a real vote (votering_id 27715, Rodt's forslag 30-32 on
full state ownership of nuclear power) -- but internal/data/
loader_test.go's TestKjernekraftStatligKontrollLeavesAmbiguousParties
Uncoded, written in a PRIOR session, deliberately asserts SV and KrF
stay uncoded here: SV's own merknad argues a narrower claim (against
subsidizing private nuclear, against a rushed process) than the
statement's actual text ("Norway must have full state ownership"),
and KrF has no committee seat for this case so only an ambiguous
floor vote exists. Coding SV/KrF from the vote alone would have
attributed a stronger, more specific claim than either party's own
words support -- exactly the discipline this project's tests exist
to enforce. Reverted the addition; utsagn-06 stays at 7/9
DELIBERATELY, not as an outstanding gap. This is the one statement
in the whole 42 that is correctly, honestly incomplete.

Also fixed internal/scoring/scoring_test.go's
TestComputeCellEmptyIsHonestlyEmpty, which had hard-coded utsagn-05
as its "known-empty" fixture -- it broke the moment utsagn-05 got
real evidence. Rewrote it to construct an isolated `&data.Dataset{}`
with a nonexistent statement/party ID, so the empty-cell behavior is
tested against a fixture that can never accidentally get filled in
by future evidence-gathering, rather than against production data
whose completeness is exactly what the project is trying to improve.

FINAL STATE: 44 statement files (all 42 canonical + 2 "Tillegg"
extras), 430 evidence files, 41/42 statements at full 9/9 party
coverage, 1/42 (utsagn-06) at a deliberate, tested 7/9. Every
dimension in the memo's own table is now fully built.

Verified: go build/vet/test all green (44 statements, 430 evidence).
gofmt-clean, no duplicate statement/evidence IDs. Full sweep of all
42 /utsagn/N pages: HTTP 200 on every one, correct party-coverage
count on every one, checked via curl. Browser-verified via
playwright-cli across this whole session's work: desktop (1280x900),
iPad (810x1080), and mobile (390x844) viewports, zero console errors
throughout.

## Phase 4 (2026-09-25 to 2026-09-27): two new dimensions, 44/45, bias, statement form

From git, oldest first:

- `0b94514` review pass: a11y, footer parity, stale README.
- `b7b5a0c` evidensstyrke capped for the 5 cells resting only on a
  party's own merknad.
- `b65d401`, `32a18d8`, `68c02fe` utsagn-44: real tier-A vote evidence
  (Innst. 225 S), vote decoding and quotes corrected. utsagn-45: tier-A
  context (Innst. 464 S, position null) plus tier-C party programs for
  7 of 9.
- `6bc48a1`, `5a1c146` dimension 11 "Tilgang, integrering og
  rettsstat" (46-51), utsagn-51 KrF coding fixed, 46/49 filled.
- `09bdb7e`, `67f099e` dimension 10 "Finansiering av fellesskapet"
  (52-56).
- `88d070f` ratchet test for acquiescence bias (measured 41 points).
- `a9887cf`, `8342ebb`, `036af90` 13 statements flipped to the opposite
  direction, party positions negated to match, plain wording, stored
  answers migrated in quiz.js. Spread 41 -> 14.
- `c6479dd`, `2d33124`, `eaab894` all 55 statements made atomic, practice
  box on all 55, form test added, stale dates (14, 35) removed.
- This entry: docs and pages brought in sync with the data. `/om` said
  no party had a position on 44/45 (stale since 2026-09-26) and pointed
  to build status on the front page (gone since `88e836c`). The
  statement headline said "9 av 9 standpunkter" on 45 while two rows
  said "Ikke vurdert", now it counts coded positions. The badge said
  "låst", which read as "data frozen" while 497 of 547 sources were
  retrieved after the cutoff. It now says what is true and tested. Footer
  "fase 1" removed. `/metode` explains the statement form and shows the
  flip count and the spread, both tied to the code by tests.


## Phase 4b (2026-09-28): every source claim verified

Two checkers, then six parallel workers, one git worktree each, owning
disjoint statements. Every branch was reviewed from the artefact
(`review_branch.sh`: owned files only, both checkers, go test, every
position/status diff read by hand), never from the worker's report.

- First run: 252 excerpts, **109 failed** (56 EDITED: the passage is
  there but words were changed, e.g. "Dette medlem" for "Komiteens
  medlem fra Rødt", labels inside quotes, two passages joined without
  `[…]`; 50 MISS: not in the cited source; 3 dead links). Votes: 354
  checked, 0 real mismatches, 30 without a vote id.
- Checker bugs found and fixed on the way: Stortinget serves long
  innstillinger in pages (only `?all=true` is the whole text);
  "tilråding … vedtas ikke" votes read inverted; "votering 27523" was
  not recognised (W6); line-break dehyphenation belongs to PDFs only,
  in HTML it turned "bort- eller" into "borteller" (W2).
- W1 (1-7) 38 fixed, 19 moved to the page that holds the text. W2
  (8-20) 22 fixed. W5 (27-56) 24 fixed, vote ids for 446-451. W6 (46,
  48-50) vote ids. W4 (47, 51) 47 3/9 -> 9/9, 51 6/9 -> 9/9. W3 (45,
  52, 53, 55) 24 fixed incl. 3 dead Ap links, R on 52 (-2) and 53 (+2)
  from Rødt's own merknad in Innst. 2 S.
- One worker correction: W5 moved SV's skolemat quote (382) to an NTB
  story in Nationen as tier B. A newspaper is D, and the file also held
  SV's tier-A vote. Split into 382 (vote, A) and 700 (quote, D). The
  old excerpt had even changed "utdanningsministeren" to
  "kunnskapsministeren".
- No existing position changed. Coverage 48 -> 51 of 55 at 9/9. Bias
  spread 14 -> 13, ratchet lowered to match.
- Final: 263 excerpts, 0 failed. 393 vote files, 0 mismatch, 0 without
  a vote id. 559 evidence files.
- Follow-up the same day: the owner downloaded Rødt's program in a
  browser (roedt.no blocks automated requests). R on 55 coded -2 (file
  603, total-level sentence), 55 6/9 -> 7/9. R on 45 stays empty, same
  reason as Sp. The program confirms R on 52 and 53 (601/602). Three
  context files (516, 524, 543) still said R was "Ikke vurdert", which
  was already wrong for 52 and 53. Fixed. 560 evidence files.

## Tax fairness: utsagn 57 and 58 (2026-10-02)

Owner asked for statements on inheritance and labour vs capital tax.
Two statements, a direction-balanced pair, 57 at 8/9 and 58 at 9/9:

- **57 «Det bør innføres skatt på store arv.»** R, SV, MDG +2 and H,
  FrP -2 (programs). Sp -2 (finance minister Vedum in the Storting,
  12.06.2024, capped moderat). Ap not coded: the same sentence is
  Vedum's words about Ap, and Ap's own speakers in that debate did not
  mention arv (file 710, position null, the utsagn-52 rule). V, KrF -1 (fellesmerknad with H and
  FrP in Innst. 410 S (2023-2024): tax at final realisation, not at
  transfer, capped moderat).
- **58 «Utflytterskatten … bør lempes.»** H, V +2, KrF +1 (programs,
  V also forslag 8 in Innst. 3 S, votering 26616). FrP +1 (merknad
  against the 2022 tightening). Ap, Sp, MDG -2 (majority merknad in
  Innst. 4 L (2022-2023) behind the tightening, capped moderat). SV,
  R, MDG also votering 18780 (2022, stricter rule), SV and R programs.
- Context, position null: SSB analyse 2020/13 (top 1 % paid about 220
  kr per 1 000 kr vs 330 for the next 9 %), skatteloven § 9-7 and
  § 10-70, Skatteetaten's 2026 top marginal rates (53,9 % on wages
  incl. arbeidsgiveravgift, 51,5 % on dividends incl. selskapsskatt).
- Arveavgiftsloven was repealed by LOV-2013-12-13-110 (Lovdata). The
  Storting vote behind that law was never fetched or assessed. An
  earlier version of this entry said it was found and left out on
  purpose. That was wrong, and was corrected on 2026-10-02.
- Left out: holding companies and fritaksmetoden (most parties only
  through one vote on 9 bundled proposals, votering 23104).
- Research workers failed in two of three threads (wrong Lovdata URLs;
  a read-only sandbox) and the third invented 15 of 16 quotes,
  including SSB numbers that are on no page. Everything here was
  re-fetched and substring-checked before it was written.
- Spread 13 -> 12, ratchet lowered to match. verify_excerpts: 293
  checked, 0 failed. verify_votes: 400 checked, 0 mismatch, 0 NOID.

## Uncoded cells no longer get a strength rating (2026-10-02)

Three cells list a source but code no position on purpose: Ap on 57
(file 710), Sp and Rødt on 45 (files 458, 459). The strength label
used to fall back to rating that uncoded source, so each card read
"Ikke vurdert" next to "●●● Sterk". The label now rates only coded
evidence. With nothing coded, the card says "Ingen kodet posisjon
(1 kilde(r), vist som kontekst)", with no dots and no second
"Ikke vurdert". The result page said «Ikke vurdert» means no evidence
was gathered. That was false for these three cells, so it now says no
gathered source shows the party's stance on that statement. "1 partier"
is now "1 parti". New tests check every cell and every statement page,
and both go red when either fix is reverted.

A GitHub Actions job (`.github/workflows/ci.yml`) now runs
gofmt, go vet and go test on every PR that touches this stack, in the
Go image the Dockerfile pins. Before this, the only gate was the
`go test` inside the Docker build on the deploy host, after merge. The
excerpt and vote verifiers fetch live sources and stay a local step.

## Digital dependencies: utsagn 59 (2026-10-03)

Owner asked for more statements from the digital sovereignty case
(Dokument 8:119 S, Innst. 225 S (2025-2026), sak 200036). Three drafts
were checked against the dataset first. Two were dropped: «nasjonalt
program for åpen kildekode» is utsagn-23 and «felleskommunal innkjøp
av KI» is utsagn-22. One was built, 9/9:

- **59 «Regjeringen bør jevnlig rapportere til Stortinget om hvilke
  digitale avhengigheter Norge har til andre land.»** The committee's
  tilråding (proposed by H, co-signed by FrP and SV) failed 47-54 in
  votering 27496 on 07.05.2026. H, FrP +2 (own proposal, vote, plenum;
  H also Dokument 8:321 S, 19.06.2026, consistent). KrF +2 (plenum,
  vote 4/0). V +1 (vote only). SV 0: co-signed in committee (+1), then
  voted against citing budget consequences (-1), `changed_position`;
  Fiskaa (SV) told Altinget the proposal "er ikke viktig for oss"
  (tier D, position null). Ap -1: minister Tung said the government
  already maps dependencies (Lysne, Nkom, NSM, DFØ), so the wording
  asks about the fixed report to Stortinget, not about mapping as
  such. MDG -1 (plenum: acceptable, but budget). Sp, R -1 (vote only).
  Sp's programme (risk assessments by infrastructure owners) is shown
  with `not_comparable`, position null.
- Context, position null: NSM Risiko 2026 (tier B), NITO's hearing
  input (the only one of 19 published that asks for mapping; the
  hearing predates the tilråding).
- Programmes of all nine were searched with the same terms. Only Sp
  (above) and R ("digital suverenitet er et mål", not about reporting)
  mention the area. None of the others do.
- Spread stays at 12 in both directions, so no balancing statement was
  added. verify_excerpts: 309 checked, 0 failed. verify_votes: 410
  checked, 0 mismatch, 0 NOID.
- Fixed in passing: file 121 said the tilråding was "vedtatt av FrP, H
  og SV". It was proposed by them and failed. Not fixed: files 211,
  221, 231 and 241 describe the formula votering_id = 27496 -
  forslagsnummer as exact. It holds for the votes those files use, but
  not for forslag 18-20 (27476, 27478, 27477). Resolve ids from
  `voteringer?sakid=200036`, never from the formula.

CURRENT STATE: see "Current state (2026-10-03)" at the top.

## Corrections log published on /om (2026-10-23)

`/om`'s "Rettelseslogg" tier used to say the corrections stood in this
file. That is not a correction record a reader can check while the repo
is private, and the footer already says the source is not public yet. The
tier now renders a real, dated log from
`internal/data/content/corrections/`, one YAML file per entry, loaded and
validated like every other content type (date, kind, title, what, reason,
effect; an empty entry fails the build).

Seven entries, taken from this file's own phase entries, not written
fresh: 2026-09-23 the invented dimension and the correction OF that
correction, 2026-09-24 the AI summary that returned another document
(`nesten`) and utsagn-43 folded back into utsagn-07, 2026-10-02 the
arveavgift claim that was never researched, the 15 invented quotes
(`nesten`), and the "Ikke vurdert" label that rated an uncoded source.

`kind` separates `rettelse` (a reader could see it and it changed) from
`nesten` (stopped before it reached the site). Both are listed: a log
holding only the mistakes that escaped is a curated list, not a log.

Newest first across days; within one day, the order the events happened,
so "Rettelsen over var også feil" cannot appear before the rettelse it
refers to. Gates: `internal/data/corrections_test.go` (non-empty, every
field, ordering, ids match filenames, every file loaded) and
`TestOmPublishesEveryCorrection` (every entry rendered, newest first, no
`docs/roadmap.md` pointer left). All four were mutation-checked: blank an
`effect`, render only the first entry, drop the sort, restore the roadmap
pointer — each one turns a test red.

Not done, by design: the log still does not estimate how many readers saw
each wrong claim, because that number was never measured.

