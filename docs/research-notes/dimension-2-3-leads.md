# Dimension 2 & 3 research leads (2026-09-23) — verified case data, NOT shippable evidence yet

**Correction, same day, added when preparing a research-delegation
brief:** this file originally filed every lead below under "Dimension 2
— Energi og klimapolitikk" and "Dimension 3 — EØS og norsk
suverenitet". Neither name exists in `docs/research-memo.md`. The
memo's real Dimension 2 is **Norsk handlefrihet** (statements 6-11,
"Beholder Norge reell beslutnings- og operasjonsevne?") and Dimension 3
is **Forvaltning av fellesskapets ressurser / energi** (statements
12-18, "Hvem kontrollerer ressursene, risikoen og verdiskapingen?") —
see `docs/roadmap.md`'s "Correction 2" for the full story of how that
happened. The specific topics below (kjernekraft, strømpris, sosial
dumping, etc.) were never in the memo either — they were this session's
own guesses at what a statement in the relevant dimension might cover.
**The case verification work itself is still sound** (real case IDs,
real vote tallies, independently re-verified against live
data.stortinget.no/stortinget.no, not relayed from an agent's summary)
— what's struck is only the dimension/statement-number filing, which is
undecided until each topic is deliberately assigned a real statement
number during Dimension 2/3 drafting.

Not committed as statement/evidence YAML. Every item below needs the
same treatment Phase 1 got (read the actual innstilling/forslag text,
not just a vote tally) before it can be coded, IN ADDITION TO getting a
real statement number once Dimension 2 or 3 is actually drafted from the
memo's own question, not from these guessed topics alone.

## Kjernekraft (candidate topic for Dimension 2 or 3 — number TBD)

**SHIPPED 2026-09-24 as utsagn-06** (Dimension 2, Norsk handlefrihet,
"statlig eierskap og kontroll over kjernekraft"), see
`docs/roadmap.md`'s "Phase 2" entry and
`internal/data/content/statements/06-kjernekraft-statlig-kontroll.yaml`.
The close-reading warning below was correct and necessary: the actual
statement text ended up being about state-vs-private ownership, not a
generic "should Norway have nuclear power" framing, precisely because
that's the axis the innstilling's own merknader actually argued over.

**SHIPPED 2026-09-24 (later, same case) as utsagn-13** (Dimension 3,
"Kommersiell kjernekraft nå, ikke senere"), see
`internal/data/content/statements/13-kjernekraft-kommersiell-hastighet.yaml`.
This is the second axis the same case exposed: not ownership, but
whether to legislate NOW for commercial nuclear (forslag 1-2,
votering_id 27722, 49-46) versus the committee majority's much weaker
adopted text (just a white paper). Read the full 1352-line innstilling
text before coding — the roll-call tallies alone would have missed
that FrP is the case's own proposer yet voted against the final
tilråding (because it was weaker than what they asked for, the exact
"voted against ≠ opposes" trap /metode warns about), and would have
missed that Rødt's own merknad explicitly says they are NOT opposed to
nuclear power as a technology — only to a *commercial/private*
ownership framing, a genuine Konsistens link back to utsagn-06's own
Rødt position, not a coincidence worth losing to a flat vote count.


**Real case, independently re-verified 2026-09-23 by direct XML parse
(not by an agent summary):** Dokument 8:191 S (2025-2026), "Representantforslag
om å tilrettelegge for kjernekraft i Norge", proposed by FrP (Tor Mikkel
Wara, Morten Stordalen, Kristoffer Sivertsen, Erlend Wiborg, Bjørn Larsen,
Rikard Spets). Innst. 278 S (2025-2026). Case page:
https://www.stortinget.no/no/Saker-og-publikasjoner/Saker/Sak/?p=200121
(sak_id=200121 in the data API — NOT 86697, which is a different, unrelated
2021 case; that wrong ID was cited earlier in this project's own session
history and never re-verified before now).

11 separate voteringer on 2026-05-26, most FrP/H/V-proposed forslag fell;
the final "Innstillingens tilråding" (committee majority recommendation)
passed. Per-party tally on that final vote, computed by parsing
https://data.stortinget.no/eksport/voteringsresultat?voteringid=27723
directly (XML, `representant/parti/id` cross-tab against `votering`):

| Parti | For | Mot | Ikke til stede |
|---|---|---|---|
| Ap | 31 | 0 | 22 |
| FrP | 0 | 29 | 18 |
| H | 14 | 0 | 10 |
| KrF | 0 | 4 | 3 |
| MDG | 5 | 0 | 3 |
| Rødt | 5 | 0 | 4 |
| SV | 5 | 0 | 4 |
| Sp | 5 | 0 | 4 |
| V | 2 | 0 | 1 |

**The non-obvious part, unresolved:** FrP proposed this case and voted
AGAINST the final adopted text. The other 10 voteringer (mostly FrP's own
forslag nr. 1-2, 3-5, 6-15, 16, plus H+V's 17-26, H's 27-28, SV+R's 29,
R's 30-32, KrF's 33-34) all fell. This strongly suggests the adopted
committee text was weaker than what FrP actually wanted (e.g. "utrede"
rather than a firmer commitment), and FrP voted against the compromise —
exactly the "voted against ≠ opposes the principle" case /metode already
warns about. **Before coding a position for FrP (or anyone) on this
statement, read the actual "Innstillingens tilråding" text and at least
FrP's own fallen forslag 1-2/3-5/6-15/16** — vote tallies alone are not
enough here. Votering IDs for reference: 27714 (KrF 33-34), 27715 (R
30-32), 27716 (SV+R 29), 27717 (H 27-28), 27718 (H+V 17-26), 27719 (FrP
16), 27720 (FrP 6-15), 27721 (FrP+H 3-5), 27722 (FrP+H+V 1-2), 27723
(tilråding, vedtatt).

## Oljeutvinning-utfasing 2050 (candidate topic — number TBD): NOT FOUND

No real Storting vote or Dokument 8 with an explicit 2050 phase-out date
found. Two independent research passes searched 2021-2026 sessions.

## Havvind vs. fiskere/naturvern (candidate topic — number TBD): NOT FOUND

Two havvind cases exist in 2024-2026 but neither is framed as a
fisheries/nature-conservation conflict — they're about CO2 reduction
subsidies. Statement itself may need reframing to match what real cases
actually exist, rather than assuming the memo's framing has a matching
case.

## Strømpris makspris — SHIPPED as utsagn-12 (2026-09-24)

**Built.** Rather than trying to code the whole merged 4-bill Innst.
535 L in one statement (the complexity flagged below), utsagn-12
("Staten bør fastsette en lovbestemt nasjonal makspris...") codes only
the ONE clean, single-topic vote out of the case's 17: votering_id
26116, Rødt's own bloc of forslag 12-23, which includes forslag 14 —
the literal national-makspris-at-35-øre/kWt ask from Dokument 8:166 S.
Full 9/9 party coverage, first statement of Dimension 3 ("Forvaltning
av fellesskapets ressurser / energi", 12-18). Every party's
context_note traces to a specific merknad passage in the 3554-line
innstilling (read in full), not just the roll-call number — several
parties (Ap, Sp, SV, KrF, FrP) voted against this SPECIFIC mechanism
while backing a *different* state-interventionist alternative of their
own (Norgespris, a state trading company, or a 100%-above-threshold
subsidy), which the -1 vs -2 coding and each context_note make explicit
rather than flattening into one undifferentiated "mot". FrP's own
committee-record history of once supporting a similar cap is called
out as a Konsistens-relevant fact, not folded into the Saksmatch score.
The other 16 voteringer on this same case (VAT cuts, nettleiemodell,
forbrukstak levels, flat-payout redesign, Norgespris itself) remain
unshipped — each tests a different, distinct claim and would need its
own statement, not a reuse of this one's evidence.

## Strømpris makspris (historical notes, superseded by the above): re-verified 2026-09-24, real but complex

**Re-verified live 2026-09-24** (was previously "not independently
re-verified" above): sak_id 102973, Rødt's Dokument 8:166 S (2024-2025)
about a national maksprispå strøm, is real and confirmed via
`data.stortinget.no/eksport/sak?sakid=102973` — status `behandlet`,
`henvisning: Dokument 8:166 S (2024-2025), Innst. 535 L (2024-2025),
Lovvedtak 135 (2024-2025)`.

**The complication, found only by opening the actual innstilling:** Dok
8:166 S was NOT decided alone. Innst. 535 L (2024-2025, 3555-line PDF)
merges FOUR documents into one committee report: the government's own
Prop. 148 L (a new law establishing "Norgespris" — a state-administered
reference-price/subsidy scheme, not a price cap) plus three separate
opposition representantforslag (Dok 8:82 S, 8:136 S, and Rødt's 8:166 S)
about VAT cuts, a flat payout redesign, and the makspris ask
respectively. The committee tilråding (the actual adopted law) is
co-signed by Ap and Rødt only — an unusual coalition, since Rødt's
own separate makspris proposal was itself voted down (part of the 12
rejected amendment votes) while Rødt still backed the government's
different Norgespris mechanism. 17 separate roll-call votes exist on
this one case (votering_id 26113-26129,
`data.stortinget.no/eksport/voteringsresultat`), crossing coalitions
differently on nearly every one (e.g. H+V joint forslag nr. 2-4: 27 for;
FrP+KrF forslag nr. 1: 15 for; Ap+Sp forslag nr. 31/33/34/32/35-37 all
passed as part of the government's own bill).

**Why this isn't shipped as a statement yet:** this bill mixes at least
four genuinely distinct policy questions (state price-fixing vs. a pure
price cap vs. VAT reduction vs. flat-payout redesign), and forcing them
into one utsagn would either (a) conflate positions parties took on
different sub-mechanisms, or (b) require splitting this into 3-4
separate statements — a real option, but each needs its own careful
read of which of the 17 votes actually tests that specific claim, not a
single vote count. Thematically this is a strong fit for Dimension 3
("Forvaltning av fellesskapets ressurser / energi", 12-18, "Hvem
kontrollerer ressursene, risikoen og verdiskapingen?") — state control
over electricity pricing is exactly that question — but it needs a
dedicated session, not a rushed coding pass appended to an unrelated
turn's work.

## Sosial dumping (candidate topic — number TBD): all three re-verified 2026-09-24, dimension fit still unresolved

All three case IDs confirmed real and `behandlet` via
`data.stortinget.no/eksport/sak?sakid=<id>` (2026-09-24, replacing the
"NOT independently re-verified" caveat above):
- p=82607: Dokument 8:83 S (2020-2021), Innst. 194 S (2020-2021), Rødt
  (Bjørnar Moxnes) — Wizz Air / union organizing.
- p=93672: Dokument 8:206 S (2022-2023), Innst. 395 S (2022-2023), Rødt
  (Geir Jørgensen, Mímir Kristjánsson) — taxi industry.
- p=92023: Dokument 8:74 S (2022-2023), Innst. 232 S (2022-2023), SV
  (Mona Fagerås, Kirsti Bergstø) — courier industry.

**Still not shipped, and the reason changed:** these are no longer an
unverified-URL risk, but a genuine dimension-fit question. "Sosial
dumping" (labor-market undercutting via imported labor/competition)
doesn't cleanly match any of the memo's nine dimension questions as
read today — not Dimension 2's "Beholder Norge reell beslutnings- og
operasjonsevne?" (this is about labor standards, not sovereignty), not
Dimension 6's "Fungerer samfunnets grunnleggende tjenester uansett
bosted og inntekt?" (this is about labor-market protection, not service
access). Forcing a fit here would repeat the exact mistake
`docs/roadmap.md`'s "Correction" and "Correction 2" already document
fixing once each — invent a sub-topic, then a dimension to hold it,
rather than starting from the memo's own question. Needs either a
tenth dimension the memo doesn't have, or a considered decision that
this is genuinely out of scope for the current nine.

## Digital suverenitet (utsagn-44's own topic — number already assigned): Innst. 225 S found, scope mismatch, NOT coded (2026-09-24)

Real, decided case found and read in full (955-line PDF): Innst. 225 S
(2025–2026), SV's Dokument 8:119 S ("digital suverenitet i en urolig
tid"), kommunal- og forvaltningskomiteen. Confirmed live via
`data.stortinget.no/eksport/sak?sakid=200036`. This is squarely
utsagn-44's topic — the case's own Forslag 4 (SV+Rødt) asks that
"sensitiv og samfunnskritisk informasjon om innbyggere utelukkende skal
behandles i løsninger som er under norsk eller europeisk jurisdiksjon."

**Why it's not coded despite the topical match: a real scope mismatch.**
`internal/data/content/statements/44-digital-suverenitet.yaml`'s text
asks for "norsk eller **nordisk** jurisdiksjon" — narrower than the
case's "norsk eller **europeisk** jurisdiksjon." A party voting FOR the
case's European-jurisdiction ask does not thereby support the stricter
Nordic-only version utsagn-44 actually asks about (a party could think
all-of-Europe is fine but Nordic-only unrealistic) — coding those "for"
votes as evidence for utsagn-44 would attribute a narrower, stronger
claim than the party actually took a position on, the same trap
Innst. 275 L was for utsagn-02 and Innst. 303 S was for utsagn-04. The
"mot" votes are logically safer (opposing the weaker European-only ask
implies opposing the stricter Nordic-only one too), but relying on that
inference instead of a party's own words is exactly the "vote tally
alone, without reading whether the text actually applies" shortcut this
project's discipline exists to catch — not done here either.

Vote result for the record, not as coded evidence: votering_id 27492
("Forslag nr. 4 på vegne av SV og R"), 15 for (SV 5, Rødt 5, MDG 5), 86
mot (Ap 32, FrP 29, H 14, KrF 4, Sp 5, V 2) — cross-checked per-party via
`data.stortinget.no/eksport/voteringsresultat`. The whole case (22
separate votes across many crossing coalitions — FrP+SV+Rødt on public
cloud ownership, H's own "smart suverenitet" diversification framing,
Ap+H jointly calling the asks "too detailed without further study") ended
with even the committee's own toned-down tilråding (FrP+H+SV's compromise
text) failing 47–54 on the floor — nothing from this case was actually
adopted.

**Two honest options for the future, neither done now:** (a) revise
`44-digital-suverenitet.yaml`'s text from "nordisk" to "europeisk" to
match what a real case actually tested — but that's curve-fitting the
question to the data, backwards from how every other statement here was
built (text first, from the memo's own question, then evidence found to
match); or (b) keep the statement as written and wait for a case that
actually tests the Nordic-only claim specifically. Recorded here so this
955-line read doesn't need repeating, and so nobody mistakes "found a
real case on this exact topic" for "therefore it must be codeable."

## Klimamål 65% by 2030 (candidate topic — number TBD): confused, needs rework

Found sak_id 103136, "Endringer i klimaloven" — but it's about a 2035
target, not 2030. Norway's actual current statutory 2030 target was not
confirmed by this research pass. The statement itself may need
reformulating to match whatever the real current legal target actually
is before any case can meaningfully test "should it be tightened."

## Forsvar/beredskap suverenitet (utsagn-45's own topic — number already assigned): searched, no matching decided case (2026-09-24)

utsagn-45's claim: "Norsk forsvars- og beredskapslovgivning skal vedtas
av Stortinget og regjeringen uavhengig av EØS-koordinering eller
EU-mandater som kan begrense norsk operativt handlingsrom." Three real
candidates found and read, none usable:

- **Innst. 440 S (2025–2026), "Forsvarsløftet"** (the actual Long-Term
  Defense Plan increase) — mentions Norway's real participation in EU's
  EDF (forsvarsfond), PESCO projects, SAFE, and EDIP, and even has
  committee language about using "handlingsrommet i EØS-avtalen" for
  defense procurement to protect national control. Read in full
  (1579-line PDF) — but every passage is unanimous "Komiteen mener/ber
  om," with zero "disse medlemmer" dissents and no forslag list at all.
  A valgomat cell needs a real difference between parties; a unanimous
  committee report gives none, regardless of how on-topic the content is.
- **PESCO/EDF association itself** — confirmed real (Norway is a member
  of specific PESCO projects and EDF since ~2021, per a 2021-2022
  plenary EU/EEA debate, refs-202122-05-05.pdf) but that debate is a
  general redegjørelse-and-interpellation exchange about EU/EEA
  relations broadly (fisheries, whaling, oil, national self-determination)
  with individual MPs' personal speeches, not a recorded vote on a
  defense-specific text — no clean party-level "for/mot" signal to code
  from.
- **EU's "military mobility" regulation package** — a live EU Commission
  proposal as of late 2025 that would require Norwegian infrastructure/
  customs alignment for military transport, the most concrete real-world
  match to "EU mandates that could limit Norwegian operational freedom of
  action." Still at the Commission-proposal stage as of this research
  pass (EU/EØS-nytt bulletins Sept/Nov 2025) — no Norwegian
  samtykkeproposisjon or Storting vote exists yet. Same shape as
  utsagn-44's original seed cases before Innst. 225 S: real, on-topic,
  but not yet decided.

Not coded, for a different reason than any of this project's other
statements: not a scope mismatch (utsagn-44), not "already tried and
found genuinely nothing" (utsagn-04/02 before their resolutions), but "the
real cases that exist are either unanimous (no differentiation possible)
or still pending (nothing to code yet)." Revisit if/when the EU military
mobility regulation reaches an actual Norwegian samtykkeproposisjon and
vote — that would be the cleanest possible test of this exact claim.

## EØS/suverenitet candidate topics (EØS-avtalens videreføring, Storting-veto over EØS-regelverk, suverenitetskonsekvensutredning, EU-program trygdeeksport — numbers TBD): NOT FOUND

No real Storting vote or Dokument 8 proposal found specifically about:
EØS-avtalens videreføring (11), et Storting-vetomekanisme over
EØS-regelverk (12), en formalisert suverenitetskonsekvensutredning for
EU-regelverk (14), or EU-program-basert trygdeeksport-begrensning (15).
Two independent research passes. This may mean these specific statement
framings don't have matching real cases yet, not that the topics
themselves are undiscussed — party PROGRAMS clearly address EØS/suverenitet,
just not via a traceable Storting action matching this exact framing.

## Process lesson

Two full research fan-outs (4 agents Phase-2-round, 4 agents this round)
have now shown a consistent pattern: Dimension 1 (Demokratisk åpenhet)
had unusually rich, easy-to-find seed evidence because the source memo's
own authors had already done that research. Every dimension beyond it
requires this project's own from-scratch research, and the honest yield
is much lower — most statement/party combinations return "no evidence
found," and the ones that do return something need the same close
reading (actual bill text, not just vote tallies) that the kjernekraft
case above demonstrates. Budget accordingly: this is 70-90 person-weeks
of work in the original memo's own estimate, and nothing in these two
research rounds contradicts that.
