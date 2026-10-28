# Method: what a position number means

This is the scale every evidence row's `position` is coded on, and the one rule
that decides between the two magnitudes. It exists because the definition lived
only in the label strings of `internal/scoring` and in the `context_note` of
individual rows, so a coder had to infer it. Three queue items stalled on exactly
that: a party's magnitude, a school-meal row, and whether a group-scoped sentence
answers a question about the national total.

## The five values

| `position` | `PositionLabel` renders | The party's relation to the statement |
|---|---|---|
| `2` | `+2 (sterkt for)` | The party's own instrument states the claim directly, with a figure |
| `1` | `+1 (for)` | The direction is clear, but no own instrument states it |
| `0` | `0 (nøytral/delt)` | Balanced, or explicitly for and against in the same breath |
| `-1` | `−1 (mot)` | The direction is clear against, but no own instrument states it |
| `-2` | `−2 (sterkt mot)` | The party's own instrument states the opposite directly, with a figure |
| `null` | nothing | Not coded. No source answers the statement |

`null` is a real value, not a placeholder. It is how the dataset says a party is
unmeasured rather than neutral, and the pages print it as absent.

## The rule that decides 1 against 2

**An own instrument, with a number, earns magnitude 2. Everything else earns 1.**

An *own instrument* is a document the party put its own name on and that states
the claim in its own terms: its alternative budget with a figure against the
government's proposal, its own legislative proposal, or a programme point written
as a goal. A *vote*, a *co-sponsorship*, or a *committee remark that states only
a direction* is not an instrument. It tells you which way the party pushed, not
how far it committed.

Two rows on the same statement show the cut. FrP is coded `2` on utsagn-55
because its own alternative budget lowers the total level by a stated 51,4 mrd.
kroner; its programme point says the same in words. Sp is coded `1` on the same
statement from a remark that says it will reduce taxes and duties for ordinary
people and then rules out large cuts for the highest earners: a direction, no
instrument, no figure for the total.

## What the existing rows show

Derived from the 543 coded rows in `internal/data/content/evidence`:

| | rows | carrying `partiets_begrunnelse` | carrying `uklart` |
|---|---|---|---|
| `±2` | 235 | 228 (97 %) | 6 |
| `±1` | 293 | 125 (42 %) | 122 |

Of the 131 rows classed `uklart`, 122 sit at magnitude 1. The practice
already follows the rule: magnitude 2 rests on the party's own reasoning, and
magnitude 1 is where attribution is uncertain. The full spread is
`2` 146, `1` 128, `0` 15, `-1` 165, `-2` 89.

## What this rule does not settle

Magnitude and answerability are different questions, and this document only
answers the first.

Row 774 codes Sp at `1` for utsagn-55 from a sentence that is *group-scoped*:
lower taxes for ordinary people, explicitly not large cuts for the highest
earners. Utsagn-55 asks whether the **samlede** level should be lowered. A
group-scoped sentence may not answer that question at all, in which case the row
is not a magnitude problem but a `null`, and utsagn-55 covers 8 of 9 parties
rather than 9 of 9. `internal/data/loader_test.go` states this objection and
`queue.yaml` carries it as `skatteniva-sp-gruppeavgrenset-regel`. It needs the
owner, because it changes whether a cell exists.

Related, and measured rather than assumed: in Innst. 2 S (2025-2026) the
rammevedtak overview gives Ap, Sp and Rødt **one joint column**, so the three
front one alternative budget. That column reports spending deviation from the
government's proposal, not the tax level, so it cannot settle the group-scoped
question either way. Any future row for Sp that cites the joint budget must say
which of the three it speaks for.

## One party, one position

A party usually leaves several documents behind. Read together they can disagree,
and a cell is a mean, so adding a weaker row silently drags it. The rule:

**The strongest unambiguous own instrument decides the magnitude. A weaker or
self-contradicting document may not downgrade it.**

Applied to the KrF item (`krf-skatteniva-polarity`): KrF's programme lists "Holde
skatte- og avgiftsnivået om lag på dagens nivå" as its own point, which is an
instrument written as a goal, so the party's position on utsagn-55 is `-2`. Its
budget draft moves the total by 3,4 mrd. kroner more in revenue on the netto
rammevedtak and by 1,8 mrd. LESS in its own påløpt column in the same merknad. A
figure that changes sign between two measurement bases in one document does not
state the claim directly, so it cannot earn magnitude 2, and it may not pull the
programme row's `-2` down to `-1`. Under this rule nothing moves: the programme
row stands and the budget row is not added.

## What still needs the owner

This document settles magnitude and it settles which row wins inside one party.
Three questions are not magnitude questions, and no reading of the rows answers
them:

1. **Can a group-scoped sentence answer a total-level statement?** Row 774 codes
   Sp at `1` from a sentence scoped to ordinary earners on a statement about the
   samlede level. `loader_test.go` says no. If the rule wins, 774 is withdrawn and
   utsagn-55 covers 8 of 9. This changes whether a cell exists.
2. **Is a very small but real figure a position at all?** If a party's only
   instrument moves the total by 0,18 pst., the rule above gives it magnitude `-1`
   once the row exists. Whether such a row should exist, and whether the same
   threshold applies to Venstre's 0,37 pst., is a threshold decision.
3. **May a numeric finding ship without an excerpt?** The gate permits it. The
   precedent in `768-h-skatteniva-innst2s.yaml` is the standard to follow if it
   does.

Rulings belong in this file once made, so the next worker does not re-derive them.

## Where a row may have no excerpt

The gate permits a row without `verbatim_excerpt`: `excerpt_error` only runs when
an excerpt is present. That is deliberate, and it is a real weakening. A row with
a figure only in `claim_action` cannot be checked by substring against its
source, so it rests on the reader trusting the transcription. Precedent already
exists where no quotable sentence carries the number: the digit-only excerpt in
`768-h-skatteniva-innst2s.yaml`, which declares its column arithmetic and flags
the one label that is an inference. A bevilgningstall row should follow that
precedent or not ship.
