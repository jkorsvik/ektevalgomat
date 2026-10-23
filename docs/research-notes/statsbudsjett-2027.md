# Statsbudsjettet 2027 — collected and analysed

Collected and analysed 2026-10-12. Every number here comes from a source
fetched in this session: the Gul bok data foundation (regjeringen.no) for two
years, and the 20 budget documents for 2027. Where a claim is quoted, the quote
is verbatim. Where a lead turned out wrong or absent, that is said plainly.

## What "the entire statsbudsjett" is

The government presented the 2027 budget on **7 October 2026**. It is four
document families plus one per department:

| Document | Content | Collected |
|---|---|---|
| **Meld. St. 1 (2026–2027)** | Nasjonalbudsjettet: the macro picture | 186 pp |
| **Prop. 1 S (2026–2027)** | the budget (Gul bok) + one per department | 255 pp + 17 files |
| **Prop. 1 LS (2026–2027)** | skatter og avgifter (needs law) | 344 pp |
| **Prop. 2 S (2026–2027)** | Svalbardbudsjettet | 71 pp |
| **Fagproposisjoner** | Prop. 1 S per departement | 17 documents |

20 documents, **4 640 pages**, all downloaded and text-extracted.

Plus the machine-readable core: **`2027_gulbok_datagrunnlag.xlsx`**, 1 595 rows,
every kapittel/post with its amount. The same file for 2026 (1 600 rows) gives
the year-on-year change. This is the whole budget as a table, and it is the
backbone of the analysis below.

## The headline numbers

From Meld. St. 1, table 3.1 (verbatim): the structural, oil-corrected deficit —
"underliggende bruken av fondsmidler" — is the figure the handlingsregel binds:

| Mrd. kroner | 2025 | 2026 | 2027 |
|---|---|---|---|
| Strukturelt oljekorrigert budsjettunderskudd | 507,4 | 583,4 | **608,4** |
| … i prosent av trend-BNP Fastlands-Norge | 11,6 | 12,7 | **12,6** |
| … budsjettindikator (endring i pst.poeng) | 1,8 | 1,1 | **0,02** |
| … i prosent av Statens pensjonsfond utland | 2,6 | 2,7 | 2,7 |

So 2027 spends 608,4 mrd of the fund, 0,02 points more than 2026 — a
**neutral** budget by the indicator, not an expansion. Fastlands-BNP effect:
+0,1–0,2 %.

From the Gul bok table (all figures MNOK, chapters 1–2999 utgifter / 3000–5999
inntekter, excluding the fond-overføring):

| | 2026 | 2027 | change |
|---|---|---|---|
| Utgifter | 2 433 924 | 2 550 023 | +4,8 % |
| Inntekter | — | 2 463 100 | — |
| Overføring til SPU (kap. 2800) | 521 306 | 568 427 | +47 121 |

## Where the money goes (utgifter per departement, 2027)

| Departement | Utgifter MNOK | 2026→2027 |
|---|---|---|
| Arbeids- og inkluderingsdepartementet | 779 424 | +5,5 % |
| Finansdepartementet | 719 350 | +6,9 % |
| Kommunal- og distriktsdepartementet | 330 402 | +4,9 % |
| Helse- og omsorgsdepartementet | 319 562 | +7,7 % |
| Forsvarsdepartementet | 191 681 | +6,2 % |
| Kunnskapsdepartementet | 162 746 | +6,2 % |
| Nærings- og fiskeridepartementet | 133 788 | −4,1 % |
| Samferdselsdepartementet | 104 978 | +4,2 % |
| Barne- og familiedepartementet | 88 802 | +8,1 % |
| Justis- og beredskapsdepartementet | 65 111 | +7,2 % |
| Utenriksdepartementet | 63 015 | +6,6 % |
| Landbruks- og matdepartementet | 38 496 | +10,3 % |
| Digitaliserings- og forvaltningsdepartementet | 37 213 | +11,2 % |
| Kultur- og likestillingsdepartementet | 28 059 | +3,8 % |
| Klima- og miljødepartementet | 25 364 | −1,6 % |
| Energidepartementet | 24 160 | +17,1 % |

Relative winners: Digitaliserings- (+11,2 %), Energi (+17,1 %), Landbruk
(+10,3 %), Barne- og familie (+8,1 %), Helse (+7,7 %). Relative losers:
Nærings- og fiskeri (−4,1 %), Klima og miljø (−1,6 %).

## The owner's leads, checked

Each line from the owner's own reading, against the document:

| Lead | Verdict | Evidence |
|---|---|---|
| Nye Nødnett gets ~200 MNOK more, 324,5 MNOK total | **Confirmed** | kap. 1543/22 "Nytt nødnett" = 324 MNOK, **+224** (2026: 100) |
| Enova is cut | **Confirmed** | Klima- og energifondet (kap. 1428/50) 5 905 → 3 002 (−2 903); Enova chapter −53,9 % |
| Bredbånd subsidy wound down from 2027 | **Confirmed** | kap. 1541/60 "Bredbåndsutbygging" 34 MNOK in 2026, **absent from 2027** |
| Kommuner get vindkraft income | **Partly** | kap. 1820/62 "Fordeling av inntekt fra avgift på vindkraft" 390 (+30); "Avgift på vindkraft" 395 (+10) |
| NVE gets more, faster saksbehandling | **Confirmed** | kap. 1820/01 NVE driftsutgifter 1 112 (+80) |
| A "beredskapspakke" that omits energy and digital | **Not found** | no budget document for 2027 uses the term "beredskapspakke" |
| Nødnett cost, can brann take part (vikar) | **Partly new** | Justis proposes **85 MNOK to shield municipal and voluntary Nødnett users from higher cost**; state contract runs to 2031 (+2 yr option). No brann/vikar provision found |

## Beredskap, digital and energi — the detail

**Nødnett.** Two things moved. The **project for the new Nødnett moved from
Justis to the Digitaliserings- og forvaltningsdepartementet** in autumn 2025
(verbatim: "Digitaliserings- og forvaltningsdepartementet tok hausten 2025 over
ansvaret for prosjektet for nytt nødnett"), to be built on "5G-teknologi og
skivedeling". Funding: kap. 1543/22 = **324 MNOK (+224)**. The **current**
Nødnett stays under Justis/DSB: kap. 451/22 = 838 MNOK (+214). And the
government proposes **85 MNOK to shield municipal and voluntary users** from
higher fees — which is the direct answer to "is it enough to spare kommunene".
There is no sign of funding for a brann representative to take part (vikar).

**Bredbånd.** The government's own reason for removing the subsidy (verbatim,
Digitaliseringsdepartementet): "99,3 pst. av husstandane og 98,9 pst. av
verksemdene hadde tilbod om høghastigheits breiband med minst 100 Mbit/s … ved
utgangen av 2025." The remaining gap is said to be covered by "allereie løyvde
tilskotsmidlar" and "tilbod over lågbanesatellittsystem", plus the scheme
"funksjonell internettilgang". So the value question (does 100 % coverage still
need the bevilgning?) is a real disagreement, not an omission.

**Enova.** The cut is real and large. The government still claims a continued
push: "ei vidareført satsing på målretta energitiltak i bygg gjennom Enova".
Both are true at once: the fund transfer falls by half while efficiency
targets are kept in words.

**Norgespris.** kap. 1820/77 "Norgespris for strøm" = 12 000 MNOK (+2 900);
"Strømstønadsordning" (1820/75) 3 500 (+1 600); "Norgespris for fjernvarme"
(1820/78) 760 (+260). This is the largest single energy item, and it is what
the owner's note says reduces the household incentive to cut consumption.

**Brann.** No dedicated bevilgning. The structural facts: 110-sentralane
(brann) are **co-located** with police 112 and health 113 in several
politidistrikt, and a Storting vedtak asks the government to study "ei felles
satsing på beredskap i Kirkenes for blant annet spesialisthelsetjenesten, brann
og redning, DSB og politiet".

## What this gives ektevalgomat

These are tier-A facts: budget decisions, not promises. They bear on existing
statements and suggest new ones:

- **utsagn-26** (mobildekning/bredbånd): the bredbånd subsidy removal is direct
  evidence.
- **utsagn-12** (referansepris/makspris): Norgespris 12 mrd is the government's
  instrument in the debate.
- **utsagn-42 / utsagn-45** (digital/nasjonal beredskap): Nødnett financing and
  the department move.
- **Candidate new statements**: "Staten skal fullfinansiere overgangen til nytt
  Nødnett, ikke kommunene" and "Bredbåndsstøtten bør videreføres til 100 %
  dekning er nådd". These fit no existing statement and should be raised, not
  forced.

## Method and caveats

- The structural deficit comes from Meld. St. 1; the per-post numbers from the
  Gul bok xlsx. The two have different definitions (the deficit is
  oil-corrected; the table is gross), so they are reported separately, never
  reconciled into one fake total.
- "Absent from 2027" is a claim about the Gul bok table only. A post can be
  renamed or moved; the broadband item (1541/60) is genuinely gone, checked in
  both years' files.
- Machine-readable index: `statsbudsjett-2027-index.jsonl`, 1 595 rows, one per
  kapittel/post, with `belop_2026` and `endring` beside the 2027 amount.
