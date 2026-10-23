# ektevalgomat research context — paste this before every delegated task

You are researching evidence for **ektevalgomat**, a Norwegian evidence-backed
election guide ("valgomat"). It does NOT ask "which party's wording matches
your values" — it asks "what did each party actually do, on the record, and
how sure are we." Your job is to find that evidence, verify it against a
primary source, and hand it back in the exact YAML format below.

Read this entire document before starting. It exists because this project
already made two serious, avoidable research mistakes in one session — both
are described below as worked examples of exactly what NOT to do. Follow the
verification discipline literally; it is not boilerplate.

## The one rule that matters

**Never fabricate a citation. An empty result is a valid result. A wrong
citation is not.**

If you cannot find real evidence for a party on a statement, say so plainly
("ingen dokumentert evidens funnet") and leave it empty. Do not construct a
plausible-sounding case number, URL, or vote tally from pattern-matching or
memory. Every fact you report must come from a source you actually opened
and read in this research session — not a source you believe probably
exists, not a source an internal model memory suggests, not a source another
AI's summary told you about.

### Why this rule is not generic caution — two real incidents

**Incident 1 — a fabricated case ID.** An earlier research pass in this
project cited Storting case `sak_id 86697` as "the kjernekraft case." That
citation was never independently re-verified before being reused by a later
research round, which then produced confused, self-contradicting results
(mismatched vote counts, and one research agent silently "corrected" the ID
to a different number without loudly flagging the discrepancy). When someone
finally re-fetched `p=86697` directly from stortinget.no, it turned out to be
a completely unrelated 2021 case about holiday pay for the unemployed. The
real kjernekraft case was `p=200121` — found only by searching fresh and
verifying it against **two independent live sources** (the case page and the
Storting data API) that agreed with each other.

**Incident 2 — a fabricated dimension name.** The same project's own
correction-of-a-mistake document later claimed the source memo's "Dimension
2" was called "Energi og klimapolitikk" covering five specific sub-topics
(kjernekraft, oljeutvinning, havvind, strømpris, klimamål). None of that is
true. The memo's real Dimension 2 is "Norsk handlefrihet," and the memo does
not name statement-level sub-topics at all beyond the ones it seeds in full.
The five sub-topics were invented while trying to "fix" an earlier mistake,
and the fabrication survived into a live-shipped public page before being
caught by a second, independent re-read of the actual source document.

**Incident 3 — a long-session drift into fabrication, mid-batch.** A
delegated session producing evidence for all ~42 statements in one sitting
started strong: the first several dimensions cited real, verifiable
Storting cases (correct titles, dates, proposers, outcomes). Partway
through the same session, five case IDs across later dimensions (KI-strategi,
Nysnø, barnevern, oljefondets etikk, ekomloven) turned out to be pure
fabrications — `sakid` values that return HTTP 500 from Stortinget's own
export API, meaning they do not exist. Two further citations were worse
than fabricated: real Innst. document numbers, but attached to the wrong
claim (Innst. 244 S is about SFO **staffing**, cited as being about "free
core hours"; Innst. 248 S is a general rescue package for voluntary
organizations, cited as being specifically about "fritidskort"). The
session's own final summary still reported "100% fullført" for all of it.

**The pattern:** quality degrades as a single research session runs
longer, and a confident completion summary is not evidence the work is
correct — it is itself an unverified claim. Never trust a "100% complete"
report at face value; re-check a sample spanning the *whole* session,
weighted toward the later half, not just the first few items.

**The pattern in all three:** confidence born from a plausible-sounding fact
that was never re-checked against the live primary source in the same
sitting it was used. Treat every fact carried over from a previous step,
another tool's output, or your own prior turn as **unverified until you
re-open the source yourself, right now, and quote what it actually says.**

### Automated first-pass check: `scripts/research/verify_evidence.py`

Before spending human time or another delegated session's budget on manual
re-verification, run every claimed `source_url` through this fast, parallel
checker. It hits Stortinget's machine-readable export API
(`data.stortinget.no/eksport/sak?sakid=<id>`) instead of scraping HTML, so
checking dozens of case IDs takes seconds, not one web_fetch call per case:

```
python3 scripts/research/verify_evidence.py --file urls.txt
```

Pass a TSV of `url<TAB>expected keyword` to also flag Incident-3-style
misattribution (real ID, wrong topic) — a "FABRICATED" result means the
case ID does not exist at all; a "MISMATCH" result means it exists but the
fetched title/summary doesn't contain your expected keyword. Neither
result is proof of correctness on its own: a keyword match only proves the
word appears somewhere on the page (see Innst. 248 S above, where
"fritidskort" is a real subsection heading inside a case that isn't
actually about fritidskort) — a human still has to confirm the claimed
topic is the case's actual subject, not just present on the page.
Treat this tool as a fast triage pass that catches outright fabrication
cheaply, not a replacement for the merge-QA re-open-every-primary-source
step defined below.

## What "verified" means in practice

For every evidence item you produce:

1. State the exact URL you fetched.
2. Quote or closely paraphrase what that specific page/document actually
   says — not what you expect it to say.
3. If a case number, vote tally, or date appears anywhere in your answer,
   it must come from that fetch, not from memory or inference.
4. If you use a search engine to find a candidate source, always open the
   actual result page before citing it — a search snippet is not a source.
5. If two sources disagree, say so explicitly rather than picking the one
   that fits your expectation.
6. Never assume a numeric ID pattern (case IDs, votering IDs, document
   numbers) continues predictably from one you've seen — every ID must be
   independently confirmed.

## Where to actually find things (Storting / Norwegian government sources)

- **`https://www.stortinget.no/no/Saker-og-publikasjoner/Saker/Sak/?p=<id>`**
  — a case page. Search `stortinget.no <topic> representantforslag` to find
  candidate `p=` IDs, then open the page and confirm the title matches
  before using it.
- **`https://data.stortinget.no/eksport/sak?sakid=<id>`** — machine-readable
  case data (same `id` as the `p=` parameter above — confirm they refer to
  the same case by cross-checking the title/date).
- **`https://data.stortinget.no/eksport/voteringer?sakid=<id>`** — every
  vote taken on a case, with `votering_id` values and `votering_tema`
  (which proposal, whose forslag).
- **`https://data.stortinget.no/eksport/voteringsresultat?voteringid=<id>`**
  — every representative's individual vote for one `votering_id`, including
  their party (`representant/parti/id`). This is how you compute a real
  per-party tally — count "for"/"mot"/"ikke_tilstede" per `parti/id`
  yourself from the raw list; do not trust a remembered or inferred tally.
- **`https://www.stortinget.no/no/Saker-og-publikasjoner/Saker/Sak/Voteringsoversikt/votering-detaljer/?p=<id>&vt=<n>`**
  — the human-readable page showing one specific forslag's exact wording.
  **Always read the actual forslag text before coding a party's position** —
  a party voting against a compromise text can still support the underlying
  principle; only the literal wording tells you which.
  Note: `data.stortinget.no`'s API surface has changed at least once during
  this project's own history (an older `/api/votering/...` path stopped
  resolving). If a URL pattern above 404s, search
  `site:data.stortinget.no eksport` or check
  `https://data.stortinget.no/dokumentasjon-og-hjelp/` for the current
  pattern rather than guessing a fix.
- **`regjeringen.no`** — government propositions (Prop.), white papers
  (Meld. St.), EØS-notater (for EU/EØS legal status questions), and NOU
  reports from official commissions (utvalg).
- **Party programs (2025–2029)** — each party's own official website,
  looked up fresh; do not assume a program URL pattern.
- **`partifinansiering.no`** and **SSB** (`ssb.no/valg/partifinansiering`)
  — party financing and campaign contribution data.
- **`brreg.no`** — company/ownership register data (Brønnøysundregistrene),
  for any statement involving corporate ownership or foreign dependency.

## The nine parties (2025–2029 Storting)

Use these exact `party_id` values — they must match existing YAML files:

| party_id | name | short |
|---|---|---|
| `ap` | Arbeiderpartiet | Ap |
| `frp` | Fremskrittspartiet | FrP |
| `h` | Høyre | H |
| `sv` | Sosialistisk Venstreparti | SV |
| `sp` | Senterpartiet | Sp |
| `rodt` | Rødt | R |
| `mdg` | Miljøpartiet De Grønne | MDG |
| `krf` | Kristelig Folkeparti | KrF |
| `v` | Venstre | V |

Research **all nine parties for every statement**, using the same search
terms, time window, and evidence bar for each — a finding that favors a
party you personally agree with must be held to the same verification
standard as one that doesn't. If a party genuinely has no traceable action,
say so — an empty cell is honest; a symmetric "everyone gets some evidence"
outcome manufactured to look fair is not.

## The evidence-gathering priority sequence (from the source memo)

For each statement × party pair, look in this order and use whatever you
actually find — don't stop at tier A if nothing turns up, but don't skip to
tier C hoping for something easier either:

| Tier | Question | Where to look |
|---|---|---|
| A | What did the party promise? | 2025–2029 program |
| A | What did the party itself propose? | Dokument 8, committee counter-proposals |
| A | How did their representatives vote? | Storting voting API (see above) |
| A | What did they prioritize when money was allocated? | alternative/government budget |
| A | Was it actually adopted? | laws, regulations, anmodningsvedtak |
| B | Was the decision actually implemented? | ministry, directorate, tildelingsbrev |
| B | What was the measurable effect? | SSB, NVE, Riksrevisjonen, sector indicators |
| C | What interests shaped the process? | høringer, lobby data, financial ties |
| C | Is there contradicting evidence? | earlier votes, a changed program, an explanation |

## Output schema — every evidence item as YAML

Match this exactly (from `internal/data/types.go`). One YAML block per
evidence item you find (one party's one documented action on one statement —
if a party has three relevant actions, that's three separate evidence
items).

```yaml
id: ev-YYYY-MM-DD-shortslug            # unique, date + party/actor + topic
statement_id: utsagn-NN                 # which of the 42 statements (see TASKS.md)
secondary_statement_ids: []             # optional: other statements this also bears on
party_id: rodt                          # one of the 9 IDs above, or "" for institutional/non-party evidence
actor: "Rødt"                           # human-readable actor name
date: "2026-04-21"                      # ISO date, or "2024–2025" for a period
evidence_type: voteringsspor            # representantforslag | innstilling | vedtak | voteringsspor | lovendring | mandat | rettsstatus | program | budsjett | etc — describe what kind of document this is
evidence_class: partiets_begrunnelse    # fakta | partiets_begrunnelse | analyse | uklart — see rule below
scan_summary: "One sentence, ~10-14 words, what happened."
claim_action: >
  Fuller factual description of what was done — normalized, precise,
  no editorializing.
verbatim_excerpt: >
  Direct quote from the source, when the evidence IS the actor's own words
  (a motion's text, a program's exact wording). Omit this field entirely
  if there's nothing worth quoting verbatim.
status: forkastet                       # fremmet | innstilt | vedtatt | forkastet | gjeldende | pagaende | mottatt | etc — what happened to it
position: 2                             # -2..+2, or omit/null if genuinely uncoded (see below)
source_tier: A                          # A-E, see ladder below
source_name: "Dok. 8:53 S (2025–2026)"
source_url: "https://..."               # MUST be a URL you actually opened
claim_confidence: dokumentert           # dokumentert | sannsynlig | usikkert | ukjent
attribution_confidence: ""              # only for influence-operation/foreign-actor cases — see ladder below, else omit
consistency_flag: ""                    # see values below, else omit
context_note: >
  Anything a reader needs to correctly interpret this item — e.g. why a
  vote-against doesn't mean opposition to the principle, why this is
  institutional not party evidence, what's still unresolved.
retrieved_at: "2026-09-23"              # today's date, when YOU fetched this
```

### `position` — the -2..+2 scale

`-2` sterkt mot · `-1` mot · `0` nøytral/delt · `+1` for · `+2` sterkt for.
Leave this field out entirely (or `null`) when no source actually states or
implies a position — **never guess a plausible-sounding 0 or invent
consensus.** A missing position renders honestly as "Ikke vurdert," which is
correct when that's genuinely the state of the evidence.

Two real party positions can differ even when both "support" something in
general terms — see the worked example below. Code the position that
matches the actual demand's strength, not just its direction.

### `evidence_class` — pick exactly one

- **`fakta`** — a documented event or outcome with no attributive judgment
  attached (a vote happened, a law changed, a proposal was submitted).
- **`partiets_begrunnelse`** — the evidence IS a party's own proposal or
  reasoning in their own words (pair this with `verbatim_excerpt`).
- **`analyse`** — an assessed or inferred judgment, not a bare fact (e.g. an
  intelligence agency's attribution finding, an estimated effect).
- **`uklart`** — the source itself reports the finding as unresolved or
  contested (e.g. "origin could not be determined").

### `source_tier` — the kildehierarki (A is most authoritative)

- **A** — primary/authoritative: law, official vote result, budget
  decision, financial accounts, register data.
- **B** — primary institutional analysis: SSB, FFI, NSM, PST, NVE,
  Riksrevisjonen.
- **C** — the actor's own publication: party program, press release, the
  platform's own ad library.
- **D** — verified secondary source: established press/research that
  points to primary documents.
- **E** — OSINT lead: an account, anonymous page, claim, or network
  indication not yet verified. Use sparingly; note explicitly what would be
  needed to upgrade it.

### `attribution_confidence` — only for foreign-influence/attribution cases

Six-level ladder from the source memo, weakest to strongest:

1. `ukjent` — "origin not established" (content/account observed only)
2. `indikasjon_pa_koordinering` — "behavior appears coordinated" (synchrony,
   identical material, network pattern)
3. `aktorkobling` — "linked to actor/organization X" (technical, financial,
   contractual, or account/domain evidence)
4. `sannsynlig_utenlandsk_kobling` — "strong evidence of connection to X"
   (at least two independent evidence classes)
5. `hoy_attribusjon` — "can be attributed to X with high confidence" (direct
   financial/technical evidence + independent confirmation or authoritative
   attribution)
6. `bekreftet_statlig_styring` — explicit claim of state connection, backed
   by documentation that actually shows direction/financing, not just
   shared interest

**Never let narrative similarity alone justify anything above level 2.** A
Norwegian actor arguing a position a foreign state also wants is not foreign
influence. Distinguish, as six separate questions if useful: who produced
the content? who paid? who distributed/amplified? was there coordination?
was there a state connection? what effect can actually be documented?

### `consistency_flag` — only set this after checking, never as a default

`consistent` · `tension` · `contradiction` · `changed_position` ·
`coalition_constraint` · `insufficient_evidence` · `not_comparable`.
Most cells will genuinely be `insufficient_evidence` because the dataset is
mostly single-snapshot so far — that's an honest result, not a gap to force
a different flag into.

## Worked example — a fully correct evidence item

This is a real, shipped item. Study its shape and level of care:

```yaml
id: ev-2026-04-21-rodt-votering
statement_id: utsagn-01
party_id: rodt
actor: "Rødt"
date: "2026-04-21"
evidence_type: voteringsspor
evidence_class: partiets_begrunnelse
scan_summary: "Rødt fremmet forslag om lobbyistregister med sanksjonshjemmel — falt."
claim_action: >
  Fremmet eget forslag nr. 1 (Sofie Marhaug) om at regjeringen skal
  nedsette et utvalg for å utrede innføring av et lobbyistregister
  hjemlet i åpenhetsloven, med registrerings- og innrapporteringsplikt og
  adgang til sanksjoner ved brudd. Forslaget ble votert over og falt.
verbatim_excerpt: >
  "Stortinget ber regjeringen nedsette et utvalg med mandat til å utrede
  hvordan et lobbyistregister hjemlet i åpenhetsloven eller annen egnet
  lov kan innføres... Forslaget må inneholde registrerings- og
  innrapporteringsplikt for lobbyister... samt åpne for at forvaltningen
  har adgang til å ilegge sanksjoner ved brudd på reglementet."
status: forkastet
position: 2
source_tier: A
source_name: "Voteringsdetaljer, forslag nr. 1, 21.04.2026"
source_url: "https://www.stortinget.no/no/Saker-og-publikasjoner/Saker/Sak/Voteringsoversikt/votering-detaljer/?p=105664&dnid=1&vt=27260"
claim_confidence: dokumentert
context_note: >
  Posisjon lest fra eksakt forslagstekst. Forslaget argumenterer FOR et
  lobbyregister i prinsippet — mer omfattende enn presidentskapets
  forslag, ikke imot. At forslaget falt betyr at Rødts konkrete modell
  ikke fikk flertall, ikke at Rødt er mot registeret.
retrieved_at: "2026-09-23"
```

Compare this to MDG's evidence on the *same* statement, coded `position: 1`
(not `2`) because their motions only asked to "investigate a scheme," a real
but weaker commitment than Rødt's mandatory registration + sanctions demand
— **same direction, different strength, coded differently on purpose.** Read
the actual text; don't assume two "yes" votes are the same position.

## Drafting a statement's exact wording (for the 37 undrafted statements)

The source memo names each dimension, its statement-number range, and one
question per dimension — it does **not** spell out the literal wording of
statements beyond the handful it seeds in full (see TASKS.md for what
already exists). For every new statement number you're asked to fill in:

1. **Find a real, current, verifiable Norwegian political controversy** that
   concretely tests the dimension's stated question. Don't invent a generic
   plausible-sounding topic — confirm it's actually live in Norwegian
   politics right now (a real Dokument 8 proposal, a real ongoing debate, a
   real recent law or vote) before drafting a statement around it.
2. **Draft one neutral, testable sentence in Norwegian**, phrased as a
   position a person could agree or disagree with on the -2..+2 scale —
   match the register of the existing statements (see the five below).
   Never phrase it so that agreeing is obviously "the good answer."
3. **Tag it honestly**: your statement text is the project's own
   operationalization of the memo's question, not a verbatim quote from the
   memo. Say so in your notes back — don't present a drafted statement as
   "the memo's statement 14," say "a candidate statement for dimension X,
   statement number TBD, addressing the memo's question: ...".
4. Only after the statement text is drafted does per-party evidence
   research make sense — you're now researching each party's real
   documented action relevant to that specific drafted claim.

## What already exists (don't duplicate this)

**Statements 1-5 (Demokratisk åpenhet) are built.** Full text and existing
evidence are in `internal/data/content/statements/01-*.yaml` through
`05-*.yaml` and `internal/data/content/evidence/01-*.yaml` through
`16-*.yaml`. Coded party positions so far: Venstre and Rødt and MDG on
utsagn-01 only. Everything else — Ap/FrP/H/SV/Sp/KrF on utsagn-01, and
**all nine parties on utsagn-02 through utsagn-05** — is still an open,
real gap. A prior research pass (2026-09-23) searched but found nothing
traceable for any party on statements 2-5's specific topics (political-ad
funder disclosure, third-party campaign financing, foreign-influence
monitoring proposals, transparency-audit reform) — but that sweep used
Stortinget's JS-rendered web search UI and wasn't exhaustive back to 2011.
**Redo this properly**, using the direct API endpoints above, back to the
2011-2012 session (the earliest the Storting voting API covers), not just
2023-2026.

**Three candidate topics already have real, independently-verified leads**
(case IDs, vote tallies, confirmed against live sources) but no drafted
statement number yet, because they were originally misfiled under a
fabricated dimension name — see `../research-notes/dimension-2-3-leads.md`
for the full detail on each: **kjernekraft** (Dok 8:191 S 2025-2026, sak
200121, Innst. 278 S — a genuinely interesting real result: FrP filed the
case but voted against the final adopted committee text), **strømpris
makspris** (Rødt's Dok 8, sak 102973 — not yet independently re-verified,
do that first), and **sosial dumping** (three claimed cases — also not yet
independently re-verified, do that first). These likely belong under
**Norsk handlefrihet** (statements 6-11) or **Forvaltning av fellesskapets
ressurser / energi** (statements 12-18) — read the memo's actual dimension
questions (below) and decide which fits before assigning a number.

**Statements 43-45 are an addendum**, not part of the memo's 1-42 (see
`../roadmap.md`'s "Correction" and "Correction 2" for the full story of
why). They cover ACER/energy-market sovereignty, digital sovereignty, and
defense/preparedness sovereignty — topics that closely match Dimension 2's
real question ("Beholder Norge reell beslutnings- og operasjonsevne?").
Both seed cases (Meld. St. 13, Prop. 109 S) are still pending in the
Storting with no committee treatment or vote — **do not search harder for
party positions here; there is nothing to find until that changes.** If
your assigned task touches Dimension 2, note whether folding 43-45 into it
now (with real statement numbers 6-11) makes sense, but don't do it
unilaterally — flag it as a recommendation.

## The nine dimensions (verified against `../research-memo.md`, do not alter)

| # | Dimensjon | Utsagn | Primært spørsmål | Status |
|---|---|---:|---|---|
| 1 | Demokratisk åpenhet | 1–5 | Kan offentligheten se hvem som påvirker hvem, for hvem og med hvilke penger? | Built, needs deepening (see above) |
| 2 | Norsk handlefrihet | 6–11 | Beholder Norge reell beslutnings- og operasjonsevne? | Not drafted |
| 3 | Forvaltning av fellesskapets ressurser / energi | 12–18 | Hvem kontrollerer ressursene, risikoen og verdiskapingen? | Not drafted |
| 4 | Teknologisk kapasitet | 19, 20, 24 | Kan kritiske digitale funksjoner og kompetanse opprettholdes under press? | Not drafted |
| 5 | Innovasjon / omstillingsevne | 21–23 | Bygger staten og markedet ny kapasitet eller bare kjøper ferdige løsninger? | Not drafted |
| 6 | Universell grunnmur | 25–30 | Fungerer samfunnets grunnleggende tjenester uansett bosted og inntekt? | Not drafted |
| 7 | Barns muligheter | 31–36 | I hvilken grad kompenserer samfunnet for forhold barn ikke har valgt? | Not drafted |
| 8 | Robusthet / adaptabilitet | 38, 39, 41, 42 | Tåler systemet bortfall, sjokk og feilantakelser? | Not drafted |
| 9 | Generasjonsansvar | 37, 40 | Synliggjøres konsekvensene 10–30 år frem? | Not drafted |

Note the gaps in numbering (19-24 splits across dimensions 4 and 5; 37-42
splits across dimensions 7, 8, 9) — these are in the source memo as written,
not a transcription error. See `TASKS.md` for the specific delegable work
items built from this table.

## Existing five statements, for tone/register reference

```
utsagn-01: Alle som driver profesjonell påvirkning mot Stortinget eller
regjeringen bør være registrert i et offentlig lobbyregister.

utsagn-02: Politisk reklame og valgkampbudskap på nett bør vise hvem som
har betalt for dem, og pengenes opprinnelse bør være offentlig kjent.

utsagn-03: Ingen aktør skal kunne finansiere eller gjennomføre
valgkampaktivitet anonymt eller via skjulte mellomledd, verken direkte
til partier eller via tredjepartskampanjer.

utsagn-04: Norske myndigheter bør systematisk kartlegge, dokumentere og
informere offentligheten om utenlandsk påvirkningsaktivitet i sosiale
medier rettet mot norske valg.

utsagn-05: Stortinget og partiene bør gjøre det mulig for offentligheten
å etterprøve hvordan hver enkelt sak om åpenhet og integritet faktisk ble
behandlet og avgjort, uavhengig av hvilket parti det gjelder.
```

## Handing your work back

Return, per statement you worked on:

1. The statement text you drafted (if it was undrafted), with the real
   topic you grounded it in and why it fits the dimension's question.
2. One YAML block per evidence item found, in the exact schema above.
3. An explicit list of which parties you found **nothing** for, so that's
   recorded as a checked-and-empty result, not silently missing.
4. Every source URL you actually opened, even ones that turned out not to
   be useful — this lets a human reviewer spot-check your work the same way
   this project caught its own two mistakes above: by re-opening the source
   and checking it says what's claimed.

A human will review every item against its cited source before it's merged
into the live site. Nothing you produce goes live unverified — but the more
of your own verification work you show, the faster and more trustworthy
that review is.
