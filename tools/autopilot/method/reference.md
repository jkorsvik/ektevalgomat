# Method: reference data (who, when, where)

Some of the most load-bearing facts in this project are not positions at all.
They are reference rows: who governed when, who sat on which committee, which
party held which ministry, what a constituency returned. They are cheap to
collect, they are tier A, and they silently decide whether every position built
on top of them is attributed correctly.

The failure this prevents: a 2026 proposition filed as "Ap/Sp" because the
government was called that in 2022. Senterpartiet left on 2025-02-04. A table
would have caught it; prose did not.

## Rules

1. **One row, one fact, dated.** A government row has `from` and `to`. A
   committee row has a session. A constituency row has an election year. A row
   with no date cannot be used to attribute anything.
2. **The authority is the institution, not a summary.** regjeringen.no for
   cabinets and ministers, stortinget.no for committees and representatives,
   valgdirektoratet.no for results, ssb.no for party finance. Never a Wikipedia
   table, never a news article, and never a model's recollection of who was
   minister when.
3. **Names as the source writes them.** A minister's name, a party's name and a
   post's title are quoted, not normalised. If the source says «Digitaliserings-
   og forvaltningsdepartementet», the row says that, even when the older
   documents say «Kommunal- og moderniseringsdepartementet» for the same area.
4. **Say which source version you read.** A ministry is renamed, split and
   merged. Record the page you read and the date you read it, so a later reader
   can see why two rows disagree about the same year.
5. **A reference row carries no `position`.** It is not evidence about what a
   party wants. If the same document also states a position, that is a separate
   evidence row with its own `statement_id`.
6. **Ambiguity is recorded, not resolved by guessing.** Two sources that
   disagree on a date, or a minister who served twice, are reported as they are,
   with both sources named.

## What it looks like

```yaml
- period: "2021-10-14..2025-02-04"
  cabinet: "Jonas Gahr Støres regjering"
  parties: [ap, sp]
  source_name: "regjeringen.no, regjeringens sammensetning"
  source_url: "https://www.regjeringen.no/..."
  retrieved_at: "2026-10-27"
```

Keep the shape minimal and let the consumer derive: a page that needs "which
parties were in government on this date" ranges over the periods; it does not
store the answer per evidence row.

## Where it lands

Reference data is not `evidence/`. It goes under `internal/data/content/` in its
own directory, with a loader test that checks the periods do not overlap and do
not leave a gap inside the range they claim to cover. A reference table that
asserts complete coverage is exactly the kind of claim that must be checked
rather than believed.
