# Delegation task list — ektevalgomat evidence research

Read `CONTEXT.md` first (paste it into the AI tool before any task below —
it has the schema, verification discipline, party list, and source ladder
every task depends on). Each task here is sized for one AI-tool session.
Paste `CONTEXT.md` + the one task's block as a single message.

Work in roughly this order: Task 0 is the most bounded and highest-value
(deepening what's already live). Tasks 1-8 are the eight undrafted
dimensions, ordered as they appear in the memo.

## How to feed results back

Paste the AI tool's output back into this project (chat, or a file) and it
gets reviewed against its cited sources before merging into
`internal/data/content/`. Nothing is merged unverified — but do this
review pass yourself first if you can: open a handful of the cited URLs
and confirm they say what's claimed, exactly like this project caught its
own two citation mistakes (see `CONTEXT.md`'s "two real incidents"). A
batch with obviously-real, checkable URLs and honest "not found" results is
far faster to merge than one that reads too smoothly to be true.

---

## Task 0 — Deepen statements 1-5 (Demokratisk åpenhet)

**Scope:** all nine parties, on all five existing statements, back to the
2011-2012 Storting session (not just 2023-2026).

**Current state:** Venstre, Rødt, and MDG are coded on utsagn-01 only.
Ap, FrP, H, SV, Sp, and KrF have zero evidence on utsagn-01. All nine
parties have zero evidence on utsagn-02 (politisk reklame), utsagn-03
(partifinansiering/tredjepartskampanjer), utsagn-04 (utenlandsk
påvirkning), and utsagn-05 (etterprøvbarhet).

**Why redo this:** a 2026-09-23 pass searched but came back empty for all
nine parties on statements 2-5, using Stortinget's JS-rendered web search
UI, not exhaustively, and only 2023-2026. Use the direct
`data.stortinget.no/eksport/...` API endpoints from `CONTEXT.md` instead,
and go back to the earliest session the API covers (2011-2012). A null
result from a genuinely thorough search is still valuable — but this one
wasn't thorough enough to trust yet.

**Search terms** (from the source memo, expand as needed): `lobby`,
`partifinans*`, `politisk reklame`, `valgkamp`, `bidrag`, `bakenforliggende
giver`, `utenlandsk påvirkning`, `desinformasjon`, `sosiale medier`,
`habilitet`, `karantene`, `interessekonflikt`, `åpenhet`.

**Statement texts:** see `CONTEXT.md`'s "Existing five statements" section
— use these verbatim, do not redraft them.

---

## Task A — Finish three partial leads (candidates for Task 1 or 2 below)

**Scope:** verify and complete three specific candidate topics that
already have a first-pass lead, but were never independently re-verified
and never assigned a real statement number. Full detail in
`../research-notes/dimension-2-3-leads.md`.

1. **Kjernekraft** — Dok 8:191 S (2025-2026), sak `p=200121`, Innst. 278 S.
   A per-party vote tally on the final adopted text already exists in the
   leads file (computed by direct XML parse, trustworthy) — but WHY FrP
   voted against its own case's final text is unresolved. Read the actual
   adopted "Innstillingens tilråding" text and FrP's own fallen proposals
   (voteringer 27719-27722) before coding any party's position.
2. **Strømpris makspris** — Rødt's Dok 8, sak `102973`. Not yet
   independently re-verified — do that first (open the case page, confirm
   title/date/proposers match what's claimed), then find the vote result
   and every other party's position.
3. **Sosial dumping** — three claimed cases (Wizz Air `p=82607`, taxi
   `p=93672`, courier `p=92023`). None independently re-verified — do that
   first for all three, then complete the per-party research.

**Then decide, and say explicitly which:** do these three fit Dimension 2
(Norsk handlefrihet, "Beholder Norge reell beslutnings- og
operasjonsevne?") or Dimension 3 (Forvaltning av fellesskapets ressurser /
energi, "Hvem kontrollerer ressursene, risikoen og verdiskapingen?")? Sosial
dumping likely fits neither cleanly — say so if that's what you find, don't
force it. Propose specific statement numbers within the dimension's real
range (6-11 or 12-18) once you've drafted the statement text per
`CONTEXT.md`'s drafting process.

---

## Task 1 — Dimension 2: Norsk handlefrihet (statements 6-11, six statements)

**Question:** Beholder Norge reell beslutnings- og operasjonsevne?

**Known starting material:** statements 43-45 (energimarked-suverenitet/
ACER, digital suverenitet, forsvars-/beredskapssuverenitet) already exist
as an addendum outside the 1-42 range and closely match this dimension's
question — read them (`internal/data/content/statements/43-*.yaml` through
`45-*.yaml`) as strong candidates for three of this dimension's six slots.
Recommend explicitly whether to fold them in (with new numbers 6-11) or
keep them separate, but don't renumber unilaterally.

**Real topics to investigate for the other slots** (verify currency before
drafting — these are suggestions, not confirmed current controversies):
EU/EØS directive implementation timelines vs. Norwegian sovereignty debates,
Schengen/border control autonomy, NATO/defense procurement sovereignty,
Norway's veto rights (or lack of them) in EØS decision processes, currency/
monetary policy independence. Find what's actually live in current
Norwegian political debate, not just what sounds plausible.

**Also apply Task A's kjernekraft/strømpris findings here if they fit
better under this dimension's question than Dimension 3's.**

---

## Task 2 — Dimension 3: Forvaltning av fellesskapets ressurser / energi (statements 12-18, seven statements)

**Question:** Hvem kontrollerer ressursene, risikoen og verdiskapingen?

**Real topics to investigate** (verify currency first): kjernekraft (see
Task A — probably belongs here), oljeutvinning and phase-out timelines,
havvind (offshore wind) siting and ownership, strømpris/makspris (see Task
A), kraftkabler/power export to the EU, statlig eierskap in energy
companies (Equinor, Statkraft), grunnrenteskatt (resource rent tax) on
energy or aquaculture, mineral/sjeldne jordarter extraction rights.

---

## Task 3 — Dimension 4: Teknologisk kapasitet (statements 19, 20, 24 — three statements, non-contiguous)

**Question:** Kan kritiske digitale funksjoner og kompetanse opprettholdes
under press?

**Real topics to investigate:** datasenterkapasitet and foreign cloud
dependency (note: statement 44 already covers a closely related "digital
suverenitet" angle in the addendum — check for overlap before drafting a
near-duplicate), skycomputing/AI compute sovereignty, cybersecurity
workforce/competence retention, critical infrastructure (power grid,
telecom) digital control systems and foreign vendor dependency (NSM's
Risiko 2026 report is a named primary source for this in the memo).

---

## Task 4 — Dimension 5: Innovasjon / omstillingsevne (statements 21-23, three statements)

**Question:** Bygger staten og markedet ny kapasitet eller bare kjøper
ferdige løsninger?

**Real topics to investigate:** statlig innovasjonsstøtte (Innovasjon
Norge, Forskningsrådet funding allocation and outcomes), offentlige
anskaffelser (public procurement) — buy vs. build vs. domestic-develop
policy, gründerskap/oppstartsselskap conditions (options taxation, exit
tax debates), grønn omstilling industrial policy (batteries, hydrogen,
carbon capture) — state funding vs. private capacity building.

---

## Task 5 — Dimension 6: Universell grunnmur (statements 25-30, six statements)

**Question:** Fungerer samfunnets grunnleggende tjenester uansett bosted og
inntekt?

**Real topics to investigate:** sykehusstruktur og reisetid (hospital
structure and distance-to-care), distriktspolitikk og statlige
arbeidsplasser, bredbånds-/mobildekning i distriktene, kollektivtransport
utenfor storbyene, eldreomsorg bemanningsnorm, fastlegeordningen (GP
shortage crisis — a genuinely live, well-documented current issue worth
checking first).

---

## Task 6 — Dimension 7: Barns muligheter (statements 31-36, six statements)

**Question:** I hvilken grad kompenserer samfunnet for forhold barn ikke
har valgt?

**Real topics to investigate:** barnefattigdom/lavinntektsfamilier tiltak,
gratis SFO/barnehage, skolemåltid, barnevernsreformen, mobbing/
skolemiljø lovgivning, fritidskortordning (leisure-activity subsidy for
low-income families — check current status, this has changed hands
between governments).

---

## Task 7 — Dimension 8: Robusthet / adaptabilitet (statements 38, 39, 41, 42 — four statements, non-contiguous)

**Question:** Tåler systemet bortfall, sjokk og feilantakelser?

**Real topics to investigate:** beredskapslagre (strategic reserves — food,
medicine, fuel — check status post-2022 European security reassessment),
totalforsvaret/sivilt beredskap funding, kritisk infrastruktur
redundans requirements, klimatilpasning (flood/extreme-weather
infrastructure resilience funding).

---

## Task 8 — Dimension 9: Generasjonsansvar (statements 37, 40 — two statements)

**Question:** Synliggjøres konsekvensene 10–30 år frem?

**Real topics to investigate:** oljefondets handlingsregel (sovereign
wealth fund spending rule and long-term fiscal sustainability debate),
pensjonsreform/bærekraft i folketrygden, statsgjeld/fremtidige
generasjoners skattebyrde, klimabudsjettering med langsiktige
karbonbudsjett-forpliktelser.

---

## After all tasks: what happens next

Once dimensions have real, verified statement text and evidence, the
addendum question (fold 43-45 into Dimension 2?) gets revisited — see
`../roadmap.md`'s "Next candidate work." Longitudinal consistency scoring
(the `konsistens` field) also only becomes meaningful once the same
party×statement pair has more than one time-separated evidence item — that
improves automatically as more of these tasks complete, not by changing
any scoring code.
