# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml", "pymupdf", "rapidfuzz"]
# ///
"""Build a bounded source archive: for each quoted claim, the text around the quote.

Usage:
  uv run tools/archive/build_archive.py --out /tmp/archive          # dry run
  uv run tools/archive/build_archive.py --out internal/data/archive   # publish

Why not the whole document: the project's claim is that a reader can check a
statement against its source. Keeping the whole document would mean
republishing news and public-sector text at a size and under a licence the
project cannot carry. Keeping a few hundred characters either side of the
quote carries the one thing that cannot be recovered when a link dies: the
words the claim rests on, in the context that shows what they meant.

What is recorded for each entry, because a window with no provenance is just
an anonymous quotation:
  source_url        as the data wrote it
  fetched_url       after widening (Stortinget serves only ?all=true in full)
  fetched_at        when the document was retrieved, from the cache file
  document_sha256   of the bytes we actually read, so the window can be tied
                    to one specific retrieval rather than to "the web"
  window            the extracted text, original casing and punctuation

Reuses the verifier's fetch, to_text, norm and fragments rather than copying
them: two normalisers that must agree are two normalisers that will not.
"""

import argparse
import hashlib
import json
import pathlib
import sys
import time
import unicodedata
import urllib.error
from urllib.parse import urlparse

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "tools/verify"))
import verify_excerpts as ve  # noqa: E402  (path set above on purpose)

EVIDENCE = ROOT / "internal/data/content/evidence"
PAD = 400  # characters of context on each side


def fold(c: str) -> str:
    if c in "«»“”„\"‘’‚'":
        return "'"
    if c in "–—‐‑−":
        return "-"
    return c


def norm_with_map(s: str) -> tuple[str, list[int]]:
    """norm() from the verifier, plus the source offset of every output char.

    Whole-string NFKC can change a string's length in ways that are hard to
    map back, so this normalises per character. The result is the same text
    for every character sequence the verifier has been run against; where the
    two could disagree, the containment check in locate() fails loudly rather
    than storing a window that does not hold the quote.
    """
    chars: list[str] = []
    idx: list[int] = []
    for i, ch in enumerate(s):
        for n in unicodedata.normalize("NFKC", ch).replace("\xad", ""):
            for f in fold(n).lower():
                chars.append(f)
                idx.append(i)
    out: list[str] = []
    out_idx: list[int] = []
    prev_space = True
    for c, i in zip(chars, idx):
        if c.isspace():
            if prev_space:
                continue
            out.append(" ")
            out_idx.append(i)
            prev_space = True
        else:
            out.append(c)
            out_idx.append(i)
            prev_space = False
    while out and out[-1] == " ":
        out.pop()
        out_idx.pop()
    return "".join(out), out_idx


def locate(text: str, excerpt: str) -> list[tuple[int, int]] | None:
    """Return one (start, end) slice per excerpt fragment, or None.

    Per fragment, not one span across all of them. The first version spanned
    from the first fragment to the last, and on a long Stortinget report with
    a quote from §1 and another from §9 that produced a 358 KB window: the
    archive was quietly republishing the document it exists to avoid
    republishing. Each slice is padded on its own and marked with an elision
    between them, so the reader sees that the quotes come from different
    places instead of being handed a slab of unrelated text.

    norm() of each slice must still contain its own fragment. That check is
    what catches an offset map that drifted by a character: without it the
    archive would hold windows that look right and are clipped by a letter.
    """
    normed, idx = norm_with_map(text)
    frags = ve.fragments(excerpt)
    if not frags:
        return None

    spans: list[tuple[int, int]] = []
    cursor = 0
    for frag in frags:
        at = normed.find(frag, cursor)
        if at < 0:
            return None
        spans.append((at, at + len(frag)))
        cursor = at + len(frag)

    slices: list[tuple[int, int]] = []
    for fstart, fend in spans:
        start = max(0, fstart - PAD)
        end = min(len(normed), fend + PAD)
        frag = normed[fstart:fend]
        for _ in range(4):
            lo = idx[start] if start < len(idx) else len(text)
            hi = (idx[end - 1] + 1) if end > 0 else len(text)
            if frag in ve.norm(text[lo:hi]):
                slices.append((lo, hi))
                break
            # The map can land inside a composed character or a collapsed
            # space. Widen a little and retry, then give up rather than store
            # a window that does not hold its own quote.
            start = max(0, start - 40)
            end = min(len(normed), end + 40)
        else:
            return None
    return slices


# ELISION marks the join between two slices taken from different places in the
# same document. It is the same marker the excerpts themselves use, so a
# reader who has seen one claim card already knows what it means.
ELISION = " […] "


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", required=True, help="directory to write windows/ and manifest.json")
    ap.add_argument("--quiet", action="store_true")
    args = ap.parse_args()

    out = pathlib.Path(args.out).expanduser()
    (out / "windows").mkdir(parents=True, exist_ok=True)

    texts: dict[str, str | Exception] = {}
    entries: list[dict] = []
    no_excerpt = unlocatable = fetch_failed = 0
    problems: list[str] = []

    for f in sorted(EVIDENCE.glob("*.yaml")):
        d = yaml.safe_load(f.read_text(encoding="utf-8")) or {}
        excerpt = (d.get("verbatim_excerpt") or "").strip()
        if not excerpt:
            no_excerpt += 1
            continue
        # The key is the evidence `id` field, which is what the loader gives
        # the row and therefore what a page looks up. Keying on the filename
        # instead produced 316 windows that no page could find: the files are
        # 04-rodt-votering.yaml, the ids are ev-2026-04-21-rodt-votering.
        # Found before shipping, by the Go test that checks coverage against
        # the dataset rather than against the archive's own manifest.
        eid = (d.get("id") or "").strip()
        if not eid:
            problems.append(f"{f.stem}: has an excerpt but no id field in the YAML")
            unlocatable += 1
            continue
        url = d.get("source_url") or ""
        if not url:
            problems.append(f"{f.stem}: has an excerpt but no source_url")
            unlocatable += 1
            continue

        if url not in texts:
            try:
                texts[url] = _fetch_with_retry(url)
            except Exception as e:  # noqa: BLE001 - report, never hide
                texts[url] = e
        text = texts[url]
        if isinstance(text, Exception):
            fetch_failed += 1
            problems.append(f"{f.stem}: fetch failed: {text}")
            continue

        slices = locate(text, excerpt)
        if slices is None:
            unlocatable += 1
            problems.append(f"{f.stem}: excerpt not locatable in {url}")
            continue

        fetched_url = ve.full_document_url(url)
        cache_file = ve.CACHE / hashlib.sha256(fetched_url.encode()).hexdigest()
        body = cache_file.read_bytes()
        window = ELISION.join(text[lo:hi].strip() for lo, hi in slices)

        (out / "windows" / f"{eid}.txt").write_text(window + "\n", encoding="utf-8")
        entries.append(
            {
                "evidence_id": eid,
                "evidence_file": f.stem,
                "source_url": url,
                "fetched_url": fetched_url,
                "fetched_at": _date(cache_file),
                "document_sha256": hashlib.sha256(body).hexdigest(),
                "document_bytes": len(body),
                "window_file": f"windows/{eid}.txt",
                "window_chars": len(window),
                "excerpt_chars": len(excerpt),
                "fragments": len(slices),
            }
        )

    manifest = {
        "note": (
            "Bounded windows around each verified quote. The text is the source's own, "
            "kept so a claim stays checkable after the link dies. "
            "Built by tools/archive/build_archive.py."
        ),
        "pad_chars": PAD,
        "entries": len(entries),
    }
    (out / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    (out / "entries.jsonl").write_text(
        "".join(json.dumps(e, ensure_ascii=False) + "\n" for e in entries), encoding="utf-8"
    )

    total = len(entries)
    bytes_total = sum(e["window_chars"] for e in entries)
    sources = {e["source_url"] for e in entries}
    print(f"archived {total} of {total + unlocatable + fetch_failed + no_excerpt} evidence items")
    print(f"  without an excerpt (nothing to archive): {no_excerpt}")
    print(f"  unlocatable: {unlocatable}   fetch failed: {fetch_failed}")
    print(f"  windows: {total}  chars: {bytes_total}  median: {_median([e['window_chars'] for e in entries])}")
    print(f"  distinct sources: {len(sources)}")
    print(f"  written to: {out}")
    if problems:
        print("\nproblems:")
        for p in problems:
            print("  " + p)
        if fetch_failed:
            print(
                "\nA refused source is usually rate limiting, not a dead link: one host in\n"
                "this dataset answers 429 to a burst of requests. The cache keeps every\n"
                "document that succeeded, so re-running in a few minutes retries only the\n"
                "failures. This build is incomplete and exits non-zero either way, and\n"
                "check_archive.py refuses to call an archive with a missing window complete."
            )
    return 0 if not problems else 1


def _fetch_with_retry(url: str, attempts: int = 4) -> str:
    """Fetch one source, retrying only what is worth retrying.

    A first build fetches every distinct source, and one host in this dataset
    answers 429 when several of its documents are asked for in a row. That is
    transient back-pressure, not a broken source, but without a retry it ends
    a five-minute build with a partial archive and an exit code the newcomer
    has no way to interpret. Found on a clean-cache rebuild on 2026-10-24:
    four Rødt documents in a row, 312 of 316 windows, everything else fine.

    Retries 429 and 5xx, honoring Retry-After when the server sends one, and
    gives up after `attempts`. Any other error is raised at once: a 404 is a
    fact about the source and retrying it would only waste the reader's time.
    """
    delay = 2.0
    for attempt in range(1, attempts + 1):
        try:
            return ve.to_text(ve.fetch(url))
        except urllib.error.HTTPError as e:
            retryable = e.code == 429 or 500 <= e.code < 600
            if not retryable or attempt == attempts:
                raise
            wait = _retry_after(e) or delay
            host = urlparse(url).hostname or url
            print(
                f"  {e.code} from {host}, retrying in {wait:.0f}s ({attempt}/{attempts - 1})",
                file=sys.stderr,
            )
            time.sleep(wait)
            delay *= 3
    raise AssertionError("retry loop fell through")


def _retry_after(e: urllib.error.HTTPError) -> float | None:
    """The server's own wait, in seconds, when it sends one we can use."""
    raw = (e.headers or {}).get("Retry-After")
    if not raw:
        return None
    try:
        return max(1.0, float(raw))
    except ValueError:
        return None


def _date(cache_file: pathlib.Path) -> str:
    import datetime

    return datetime.datetime.fromtimestamp(cache_file.stat().st_mtime, datetime.UTC).date().isoformat()


def _median(vals: list[int]) -> int:
    if not vals:
        return 0
    s = sorted(vals)
    return s[len(s) // 2]


if __name__ == "__main__":
    raise SystemExit(main())
