# Source archive

The project stores the claim and the citation, but not the document. That
leaves one hole: when a source moves or deletes a page, a claim becomes
uncheckable, and the honest answer to "show me" becomes a dead link.

This directory closes that hole as far as it can be closed without
republishing documents.

## What is stored

For each claim that carries a `verbatim_excerpt`, roughly 400 characters
either side of every quoted fragment, taken from the text the verifier
already read. Quoted fragments that sit far apart in one document are stored
as separate windows joined by ` […] `, so a reader can see that the claim
rests on two places without being handed the pages between them.

Recorded with every window, because a window without provenance is just an
anonymous quotation:

| field | why |
|---|---|
| `source_url` | as the data wrote it, so the window can be matched to its claim |
| `fetched_url` | after widening (`?all=true` for Stortinget publications) |
| `fetched_at` | when the document was retrieved |
| `document_sha256` | of the bytes actually read, tying a window to one retrieval |
| `window` | the source's own text, original casing and punctuation |

The text is raw extracted text, so a page that renders a menu above its
content has that menu in its window. That is what the extractor saw and what
the quote was verified against; cleaning it would make the archive a second,
disagreeing version of the source.

## Build

Fetches each distinct `source_url` once, cached under
`~/.cache/ektevalgomat-sources`, so a rebuild is offline after the first run.

```sh
uv run tools/archive/build_archive.py --out <dir>
```

## Reproducing it yourself

The build is self-sufficient: it fetches every distinct `source_url` itself and
caches the bytes, so a fresh clone needs nothing but network access and `uv`.

Measured on a **clean clone and an empty cache** on 2026-10-24:

```
archived 312 of 615 evidence items      # 299 rows have no excerpt to archive
fetch failed: 4                          # all four: roedt.no, HTTP 429
```

That is the honest cold-start result, and it is not a bug in this tool. The
four failures are Rødt's programme, and `roedt.no` answered 429 to every
attempt: the build retried with backoff (2s, 6s, 18s, honouring `Retry-After`
when sent), and `curl` with both a default and a browser user agent got 429
directly from the host, so the block is on the requesting address rather than
in the request. The 316 entries measured below came from documents fetched
before that block, which the cache still holds.

So: a first run can come back short on one host, exits 1, and names what it
could not fetch. The cache keeps everything that succeeded, so re-running
later retries only the failures. `check_archive.py` is what decides whether an
archive is complete: it fails when a quoted claim has no window, so a partial
build can never be mistaken for a whole one.

## Gate

Re-proves the archive against the data with no network:

```sh
uv run tools/archive/check_archive.py --archive <dir>
```

It fails when a quoted claim has no window, when a window does not hold its
quote, when a fragment is joined without an elision marker, when provenance
is incomplete, and when it checked nothing at all. Each of those was
mutation-tested: deleting a window, cutting a quote out of a window, dropping
an entry, drifting a `source_url`, removing an elision, and emptying the
archive all fail the gate.

## Measured, 2026-10-23

```
archived 316 of 615 evidence items (the other 299 have no excerpt)
unlocatable: 0   fetch failed: 0
windows: 316   chars: 490941   median: 1214   max: 5583
distinct sources: 69
```

About 500 KB of text, 1.4 MB on disk with the per-entry JSONL. This replaces
an earlier spanning design that produced a single 358 KB window for one claim
and quietly republished most of a Stortinget report.

## How it is wired in

Built into `internal/data/archive/` and embedded next to the YAML content,
because the runtime container is `FROM scratch` with no filesystem to read:

| piece | what it does |
|---|---|
| `internal/data/archive.go` | embeds `entries.jsonl` and `windows/`, loaded once at start |
| `/utsagn/{n}` | shows the window per claim under "Kildeteksten rundt sitatet", with the date read and the document's SHA-256 |
| `/data` | states what is stored, what is not, and how to ask for a window to be removed |
| `.github/workflows/ci.yml` | runs `check_archive.py` on every push, offline |
| `internal/data/archive_test.go` | coverage and provenance against the dataset, not against the manifest |
| `internal/server/source_window_test.go` | the rendered page actually shows a window, with its provenance |

The key is the evidence **`id`** field, not the filename. Keying on the
filename produced 316 windows no page could find, which the coverage test
caught before anything shipped.

Deliberately not checked in Go: whether the quote appears in the window.
`check_archive.py` owns that, because it owns the normalisation that makes a
verbatim comparison fair. A second normaliser would eventually disagree with
the first, and the disagreement would look like a bad claim instead of a bad
test.
