# Method: adding a verbatim excerpt to an evidence file that has none

This method exists for the ~18 files that can be quoted but were never
quoted. It is not a licence to fill all 299: 281 of them are votes, and for a
vote the claim *is* the action. Quoting an innstilling sentence next to "the
party voted against" adds decoration, not proof, and would make the quote
coverage number mean less than it says.

So: quote only where a sentence in the source carries the claim.

## What counts

An excerpt is proof, not illustration. It must satisfy all of:

1. **Exact.** The excerpt, after normalisation, must be a literal substring of
   the fetched source. Normalisation (tools/verify/verify_excerpts.py `norm`) is
   only: NFKC, soft hyphen removed, the quote characters «»""'' → `'`, the dash
   family –—‐‑− → `-`, whitespace collapsed, lowercased. Nothing else. Word
   order, wording and punctuation are not yours to change. A paraphrase is
   rejected as `EDITED`.
2. **It carries the claim.** A reader who sees only the quote and the claim must
   see why the claim follows. If the quote is true but says something adjacent
   to the claim, it is not proof of that claim.
3. **One claim, one quote.** Where a claim has two halves, either find a sentence
   with both, or quote the half the source actually tests and say in
   `context_note` which half is untested. Do not stitch two distant sentences
   with "…" to manufacture a match: `fragments()` checks each fragment
   separately, so a stitch is only accepted when every fragment is really there.
4. **It is in the document at `source_url`.** Fetch it now. A PDF is extracted
   with pymupdf, and PDF line-break hyphenation ("utenlands-\nkabler") is
   removed before comparison, so a hyphen across a line break is safe. In HTML a
   hyphen before a newline is real text, not layout.

## Where the source text comes from

`verify_excerpts.py` caches every fetched document under
`~/.cache/ektevalgomat-sources/` by the sha256 of its `source_url`. Fetching
twice is free after the first time. To read the whole source by hand:

```sh
uv run tools/verify/verify_excerpts.py internal/data/content/evidence/<file>.yaml
```

A PASS there is the only evidence that the quote is right. Your own reading of
the document is a lead.

## When there is no quote

Frequently the honest answer. Some documents state an outcome without a sentence
you can lift: a vedtak is a decision, a mandat is a mandate, a status is a
register entry. In that case:

- leave `verbatim_excerpt` out entirely (do not write `""` and do not write a
  summary of the document),
- write in `context_note` what you looked for and why no sentence carries the
  claim, naming the document and the section you read,
- and say it in your report. "No quotable sentence in §4" is a result.

Never invent a quote to raise a coverage number. The number is published, and a
fabricated quote is the one defect this project's whole design exists to prevent.
