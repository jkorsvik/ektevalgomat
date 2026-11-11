# Method: reading a party's own merknad from an innstilling

An innstilling carries a heading per party — «Merknader fra Arbeiderpartiet»,
«Merknader fra Venstre» — and the prose under it is that party's own position,
labelled by the document itself. It is the cleanest second date available for a
cell whose only observation is a vote, because the party states the position in
words instead of being inferred from a division tally.

Everything below is measured, not assumed. The errors it prevents are the ones
that make a broken reading look plausible.

## 1. Locate the document

`saker?sesjonid=<session>` has 664 rows for 2025-2026. Two fields matter and
they do different jobs.

- `innstilling_kode` is an **enum**, not a slug. Measured distribution: `"1"` on
  429 rows, `"0"` on 139, `"2"` on 95. Every row with `"1"` names an innstilling
  in `henvisning`; no row with `"0"` or `"2"` does. It answers WHETHER.
- `henvisning` is the only field that names WHICH. It answers with
  `Innst. <number> S (<from>-<to>)`.

The session comes from the parentheses, never from the sak's own session. A
document raised in one session is handled in another: `Dokument 12:11
(2023-2024), Innst. 468 S (2025-2026)`. Using the sak's session builds the wrong
directory and the fetch 404s.

`tools/autopilot/innstilling_locate.py` does this, and
`tools/autopilot/test_innstilling_locate.py` pins seven URLs that were fetched
and answered 200 with a byte length on record.

```
https://www.stortinget.no/globalassets/pdf/innstillinger/stortinget/<y1>-<y2>/inns-<y1><y2[2:]>-<nr:03d>s.pdf
```

Three-digit padding is measured at 005, 007 and 010 in three different sessions.
The letter is the chamber; only `S` is in range for a corpus that starts in 2011.

The innstilling HTML page is a table of contents only, about 7,900 characters of
text. The document is the PDF. Do not build an extractor against the HTML.

## 2. Extract with -raw, never -layout

Storting PDFs are printed in two columns. `pdftotext -layout` reproduces the page
geometry, so it interleaves the columns: left-column prose and right-column
headings land on one line. Measured on `inns-202526-002s.pdf` (3,563,108 bytes):
`-layout` produces **932** runs of spaced-out letters («Ko m i t e e n s m e d l e m m e r») and mixes the columns; `-raw` produces **0** spaced-letter runs and
correct reading order.

A reader that uses `-layout` will attach one party's prose to another party's
heading. Nothing errors. The excerpt reads plausibly and is wrong.

## 3. Split digits from letters before matching

Headings often lose the space between the number and the label:
`3.1.2.5Senterpartiet`. Normalise first or the heading is invisible.

```python
text = re.sub(r"(?<=\d)(?=[A-ZÆØÅM])", " ", text)
```

## 4. Find the headings, then take the second occurrence

```python
HEADING = re.compile(r"(\d\.\d\.\d\.\d)\s+Merknader\s*fra\s*([A-ZÆØÅ][A-Za-zæøåÆØÅ .\-]{2,40})")
```

Dotted leaders do **not** separate the table of contents from the body. Measured:
17 headings carry leaders, not 18, because one TOC entry loses them and reads
exactly like a body heading, and the undotted count is 19 against 18 real bodies.
A dots-based rule slices prose out of the table of contents.

The rule that holds: **every section number occurs exactly twice, and the second
occurrence is the body.** Split the body on second occurrences only.

Fail closed on all of these, and emit nothing for the affected party rather than
falling back to the neighbouring text:

- a number occurring once, or three or more times
- a body count that does not equal the expected section count
- a heading whose party is not one of the nine canonical ids
- any document whose bytes were not hashed at fetch time

Measured on `inns-202526-002s.pdf`: 18 numbers, each exactly twice, second
occurrences at offsets 36,396 (2.1.2.1) through 322,865 (3.1.2.9).

## 5. What the section does and does not prove

The excerpt is the party's own words about the question the innstilling addresses.
A title match is a lead, never a coded position. The text must address the cell's
actual question; an adjacent mechanism stays uncoded even when the words overlap.
Votes and merknader are different evidence types and are coded separately.

## 6. Personal names can sit in the prose

The heading is a party, the body is not guaranteed to be about parties only.
9.3% of vote proposal texts name individuals (2,388 of 25,773). Official is not
the same as impersonal. Screen the excerpt before it reaches a model that may not
see personal data, and keep the URL, the content hash and the exact substring
mapping in a separate provenance record so the packet can be re-taken from the
document rather than trusted.

An excerpt that cannot be screened is withheld and reported as withheld. Withheld
is a result; a paraphrase dressed as a quote is not.
