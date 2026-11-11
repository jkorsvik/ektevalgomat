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

## 2. Extract with the production extractor, never with pdftotext

The gate that decides whether a quote may be archived reads the document through
`tools/verify/verify_excerpts.py:to_text`, which uses PyMuPDF and then removes
line-break hyphenation (`-\n` becomes nothing). A quote chosen from any other
extraction is not a substring of what the gate reads, so `build_archive.locate`
returns None and the row cannot be archived. Measured on three exact quotes: all
three failed to locate despite being faithful to the document.

So extract with that same function. It is also better text than `pdftotext`:

| extractor | spaced-letter runs on inns-202526-002s | headings | numbers occurring twice |
|---|---|---|---|
| `pdftotext -layout` | 932 | 36 | not separable, columns interleaved |
| `pdftotext -raw` | 0 | 36 | 18 of 18 |
| `to_text` (PyMuPDF) | 1 | 36 | 18 of 18 |

`to_text` also resolves the hyphen wrap for you: `Sen-\nterpartiet` arrives as
`Senterpartiet`. Measured on inns-202324-002s the coalition heading comes through
intact as «Arbeiderpartiet og Senterpartiet».

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
- a heading whose party cannot be resolved to the nine canonical ids
- any document whose bytes were not hashed at fetch time

## 4b. A heading can name two parties, and can be wrapped mid-word

Measured on `inns-202324-002s.pdf`: a section heading reads «Arbeiderpartiet og
Senterpartiet» and the label is hyphen-wrapped across the printed line as
crosses a printed line as `Sen-\nterpartiet`. Spacing inside a label also
collapses, so `Miljøpartiet De Grønne` may arrive as `Miljøpartiet DeGrønne`.

So two normalisations are required before matching:

- join a word broken by a trailing hyphen and a line break;
- split the label on ` og ` and resolve each part to a canonical id.

A coalition section is evidence for every party it names. Requiring exactly one
canonical id rejects a real, common shape and reports it as a missing source.

## 4c. The date comes from the API, not from the PDF's signature line

The PDF ends with a line like «Oslo, i finanskomiteen, den 27. november 2023».
That is a harder parse than necessary, because `sak?sakid=<id>` already carries
the same fact as data:

```
saksgang.saksgang_steg_liste[*].saksgang_hendelse_liste[*]
    dato                    "27.11.2023 00:00:00"
    publikasjonsReferanse   "http://www.stortinget.no/inns/2023/inns-202324-002.html"
```

Measured on sak 95557 (Innst. 2 S, 2023-2024): the komitebehandling event is
dated 27.11.2023 and its reference file name `inns-202324-002` agrees with the
PDF the locator builds. Use the API date as the source date and keep the PDF
line as a cross-check.

Two traps in that payload. `01.01.0001 00:00:00` is a null sentinel that appears
on empty events; treat it as absent, never as year 1. And several saker rows can
share one `innstilling_id` — four rows do for Innst. 2 S — so key on the
innstilling reference, not on a single sak id.


Measured on `inns-202526-002s.pdf`: 18 numbers, each exactly twice, second
occurrences at offsets 36,396 (2.1.2.1) through 322,865 (3.1.2.9).

## 4d. The search interval and the excerpt are two different limits

A section slice can run to the end of the document: measured via `to_text`,
slices run 1,006 to 234,520 characters on inns-202526-002s and 215 to 88,773 on
inns-202324-002s. The long one is the last section, which continues into the rest
of the document because no further «Merknader fra» heading stops it. End the
search interval at the next heading of any kind, and zero sections will contain
another heading; that isolation was verified on both documents.

The **6,000-character limit applies to the excerpt you emit and to its padded
archive window**, never to how far you may look. Capping the party section itself
at 6,000 characters is wrong: the three party sections the source feeder read on
inns-202324-002s are 19,876, 17,622 and 23,995 characters, and every one of them
starts below the first 6,000.

Which leads to the other half: a whole section is not an excerpt. At 17,622 to
23,995 characters it cannot be archived under a 6,000-character window at any
extractor. Pick a bounded passage inside the verified party interval — a few
sentences that carry the position — and prove its archive window is under the
limit.

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
