# Method: judging a MIXED evidence file

`tools/verify/verify_votes.py` reports `MIXED` when one evidence file cites
several voterings and the party did not vote the same way in all of them.
`MIXED` is not a failure; it exits 0. It is a question: does the coded
`position` still follow from what the party actually did, and does the file cite
the vote that carries the claim?

23 files are MIXED today. This method is how to settle them.

## Why it happens

Almost always because the file cites two different things:

- **the party's own proposal** (which it voted for) *and* **the committee's
  tilråding on that proposal** (which it voted against, or the reverse), or
- **two related forslag** voted separately in the same case, where the party
  supported one mechanism and rejected another.

Neither is an error by itself. The error is when `claim_action` and `status`
describe one of them while the coded `position` was reasoned from the other.

## Read inverted votes correctly

A vote on "Komiteens tilråding: <Dokument 8> – vedtas ikke" is read inverted:
voting *for* the tilråding means voting *against* the proposal. `verify_votes.py`
already applies this (`rejects_proposal`). Apply it yourself before you conclude
anything, or you will read a party as opposing what it supported.

## What to decide, per file

Exactly one of these, with the evidence for it:

1. **KEEP.** The coded position follows from the vote the claim actually rests
   on, and `claim_action` names that vote. State which votering id carries it and
   the party's tally on it. No file change needed beyond, at most, a
   `context_note` line naming the second vote and why it does not change the
   reading.
2. **RE-CITE.** The position is right but the file points at the wrong vote.
   Report the votering id that should be cited, its forslag text, and the tally.
   Do not edit the file: report it, and a second pass applies it.
3. **RE-CODE.** The position does not survive the second vote. Report the
   position you would code, the vote that justifies it, and quote the forslag
   text. A position change is the owner's call, never a worker's: state the
   recommendation and stop.
4. **SPLIT.** The party genuinely split (some for, some against). The file must
   say so; a split is not a position. Report the counts.

## Rules

- Never lower a position to make a mismatch disappear, and never raise one to
  make a story tidier. The vote record is the fact; the coded number is a reading
  of it.
- Every claim you make must name a votering id and the per-party tally from
  `data.stortinget.no`. A reading without the tally is an opinion.
- Read both sides of the case before deciding. The forslag text behind each
  votering is what makes them related or not: two votes on the same paragraph are
  one question; two votes on different paragraphs may be two.
- Output is a report per file, not an edit. The files stay untouched.
