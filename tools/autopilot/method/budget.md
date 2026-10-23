# Method: reading the statsbudsjett for real party priorities

This is the method a worker agent follows when a queue item points at the
national budget. It exists because the budget is where a party's *promise*
(programme) meets its *action* (bevilgning), and because the budget is written
so that only someone who knows the map can read it in time. Everything here is
a search technique, not a claim. A claim is only made when a verbatim quote
from a fetched document is on file.

## The four documents

The government presents its budget on **statsbudsjettsdagen** (usually a
Thursday in early October). Four document families matter:

| Document | What it is |
|---|---|
| **Meld. St. 1** | Nasjonalbudsjettet: the macro picture, key figures, the oil fund link |
| **Prop. 1 S** | "Gul bok": the budget proposition itself, the numbers per post |
| **Prop. 1 LS** | The legislative part: tax and fee changes that need a law |
| **Prop. 1 (one per departement)** | Fagproposisjoner: each ministry's own detailed budget |

Stortinget then treats it: **Innst. 1 S** and the faginnstillinger follow, the
parties file **alternative budgets** and **merknader**, and the whole thing is
voted in December. A party's real priorities show up in three places, in
increasing order of strength:

1. its programme (tier C, a promise),
2. its **alternative budget** or **merknad** to a faginnstilling (tier A, its
   own written priority when money had to be divided),
3. the **votering** on the budget item (tier A, what it actually did).

## How to read it fast

The budget is deliberately hard to search. The text uses institutional terms,
not topic words. Follow the map:

- **Table of contents first.** The ministry→kapitel→post tree is the index.
  A topic lives under a ministry, not under its own name: **brann og redning
  lives under DSB** (Direktoratet for samfunnssikkerhet og beredskap); digital
  infrastructure under Digitaliseringsdepartementet / NSM / NKOM; energy under
  OED, with **NVE and Enova** as under-posts; sivil beredskap under
  Justis- og beredskapsdepartementet.
- **Search few, exact terms.** "Kan videreføres" means a post may carry over to
  next year's budget. "Kan økes", "engangsbevilgning", "tilskuddsordning" are
  load-bearing words.
- **Compare years, not just last year.** A programme that the government claims
  to continue must be visible across two or three years. A single-year number
  with no history is usually a one-off.
- **Read down to the detail.** Under a big post there is a driftsbudsjett and a
  særskilt driftsbudsjett with the explanation of where the money actually
  goes. The headline number is not the claim; the explanation under it is.
- **Check "Gul bok" for the macro claim, the fagproposisjon for the measure.**

## What a finding looks like

A budget finding becomes an ektevalgomat evidence item only when it satisfies
all of:

- a **verbatim quote** from the fetched Prop. 1 / Meld. St. / Innst., with the
  kapitel/post number where the quote sits;
- the **year**, and the direction of change against a named earlier year;
- the **responsible ministry** and, when relevant, the agency (DSB, NVE, NSM,
  NKOM);
- a **statement_id** it bears on, and a **party_id** only when a party's own
  budget document or vote is the source. A government proposition is evidence
  about the government, not about the opposition — and "the government" is
  time-dependent, never a fixed pair. Senterpartiet left the government on
  2025-02-04, so a 2026 or 2027 proposition is an Ap-led government's, and
  must not be filed as "Ap/Sp". Government membership is also not identical
  to a party's votes in Stortinget: a party in government can vote for, and
  against, the cabinet's own proposals.

## Worked shape (from the owner's 2026 notes, to be verified, not trusted)

The owner's notes claim: Nye Nødnett gets 200 MNOK more (324.5 MNOK total) for
testing a 5G-based emergency network toward 2031, and the notes question
whether it covers the municipalities' cost and whether brann can afford to take
part; Enova is cut while the government calls energy efficiency a main argument
for Norgespris; the bredbåndssubsidy is proposed wound down from 2027; the
beredskapspakke does not mention the energy sector or digital infrastructure.
**None of that is evidence yet.** Each line is a lead to open in the real Prop.
1 S / fagproposisjon and quote. A lead that turns out wrong is a valuable
result; say so plainly, do not argue it back to true.
