# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml", "pymupdf", "rapidfuzz"]
# ///
"""Check that every evidence file's verbatim_excerpt really appears in its source.

Usage:
  uv run tools/verify/verify_excerpts.py                 # all evidence files
  uv run tools/verify/verify_excerpts.py FILE [FILE...]  # only these files

Fetches each distinct source_url once (cached under ~/.cache/ektevalgomat-sources),
extracts text (HTML or PDF), normalises whitespace, quotes, dashes and soft
hyphens, and requires every excerpt fragment to occur in order. Fragments are
split on "[...]", "[…]", "..." and "…" so elided quotes still verify.

On a miss it reports the closest window in the source and a similarity score:
  EDITED  score >= 85: the quote is there but was changed (paraphrased label,
          swapped words). Fix: copy the source text exactly.
  MISS    score <  85: not found in this source. Wrong URL, or not a quote.

Exit 0 only if every checked excerpt is found. Prints one line per failure and a
summary with the checked count, so a run that checked nothing is visible.
"""

import hashlib
import html
import json
import os
import pathlib
import re
import sys
import unicodedata
import urllib.request

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
EVIDENCE = ROOT / "internal/data/content/evidence"
BUDGET = ROOT / "internal/data/content/budget"
CACHE = pathlib.Path.home() / ".cache/ektevalgomat-sources"
UA = "Mozilla/5.0 (ektevalgomat evidence verifier; +https://ektevalgomat.no)"


def full_document_url(url: str) -> str:
    """Stortinget serves long innstillinger in pages; only ?all=true is the whole text.
    source_url in the data is left as written, only the fetch is widened."""
    if "stortinget.no" in url and "/Publikasjoner/" in url and "all=true" not in url:
        return url + ("&" if "?" in url else "?") + "all=true"
    return url


def fetch(url: str) -> bytes:
    url = full_document_url(url)
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / hashlib.sha256(url.encode()).hexdigest()
    if f.exists() and f.stat().st_size > 0:
        return f.read_bytes()
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=60) as r:
        body = r.read()
    tmp = f.with_suffix(f".{os.getpid()}.tmp")  # atomic: parallel runs share the cache
    tmp.write_bytes(body)
    os.replace(tmp, f)
    return body


def to_text(body: bytes) -> str:
    if body[:5] == b"%PDF-":
        import pymupdf

        doc = pymupdf.open(stream=body, filetype="pdf")
        # PDF line-break hyphenation ("utenlands-\nkabler") is layout, not text.
        # Only here: in HTML a hyphen before a newline is real ("bort-\neller", W2).
        return re.sub(r"-\s*\n\s*", "", " ".join(p.get_text() for p in doc))
    s = body.decode("utf-8", "replace")
    if s.lstrip().startswith(("{", "[")):
        try:
            return json.dumps(json.loads(s), ensure_ascii=False)
        except ValueError:
            pass
    s = re.sub(r"(?is)<(script|style)\b.*?</\1>", " ", s)
    s = re.sub(r"(?s)<[^>]+>", " ", s)
    return html.unescape(s)


def norm(s: str) -> str:
    s = unicodedata.normalize("NFKC", s).replace("\xad", "")
    for a in "«»“”„\"‘’‚'":
        s = s.replace(a, "'")
    for a in "–—‐‑−":
        s = s.replace(a, "-")
    return re.sub(r"\s+", " ", s).strip().lower()


def fragments(excerpt: str) -> list[str]:
    parts = re.split(r"\[\s*(?:\.\.\.|…)\s*\]|\.\.\.|…", excerpt)
    return [norm(p).strip(" '.,;:") for p in parts if len(norm(p).strip(" '.,;:")) >= 12]


def check_budget(texts: dict) -> tuple[int, int]:
    """The budget page makes quote claims too. Check each finding's quote
    against its source_url the same way, so a page-level claim is covered by
    a verifier and not only by the evidence files behind it."""
    checked = failed = 0
    for f in sorted(BUDGET.glob("*.yaml")):
        d = yaml.safe_load(f.read_text(encoding="utf-8")) or {}
        for i, fi in enumerate(d.get("findings") or []):
            q = (fi.get("quote") or "").strip()
            if not q:
                continue
            url = fi.get("source_url") or ""
            checked += 1
            if url not in texts:
                try:
                    texts[url] = norm(to_text(fetch(url)))
                except Exception as e:  # noqa: BLE001 - report, never hide
                    texts[url] = e
            t = texts[url]
            if isinstance(t, Exception):
                failed += 1
                print(f"FETCH  {f.name} finding {i}  {url}  {t}")
                continue
            if norm(q) not in t:
                failed += 1
                from rapidfuzz import fuzz

                al = fuzz.partial_ratio_alignment(norm(q), t)
                kind = "EDITED" if al.score >= 85 else "MISS  "
                print(f"{kind} {f.name} finding {i}  score={al.score:.0f}  {url}")
                print(f"       quote : {q[:200]!r}")
    return checked, failed


def main(argv: list[str]) -> int:
    files = [pathlib.Path(a) for a in argv] or sorted(EVIDENCE.glob("*.yaml"))
    texts: dict[str, str | Exception] = {}
    checked = failed = skipped = 0
    for f in files:
        d = yaml.safe_load(f.read_text(encoding="utf-8"))
        ex = (d.get("verbatim_excerpt") or "").strip()
        url = d.get("source_url") or ""
        if not ex:
            skipped += 1
            continue
        if url not in texts:
            try:
                texts[url] = norm(to_text(fetch(url)))
            except Exception as e:  # noqa: BLE001 - report, never hide
                texts[url] = e
        t = texts[url]
        checked += 1
        if isinstance(t, Exception):
            failed += 1
            print(f"FETCH  {f.name}  {url}  {t}")
            continue
        frags = fragments(ex)
        if not frags:
            failed += 1
            print(f"SHORT  {f.name}  excerpt has no fragment >= 12 chars")
            continue
        pos = 0
        for fr in frags:
            i = t.find(fr, pos)
            if i < 0:
                failed += 1
                from rapidfuzz import fuzz

                al = fuzz.partial_ratio_alignment(fr, t)
                kind = "EDITED" if al.score >= 85 else "MISS  "
                order = " (out of order)" if t.find(fr) >= 0 else ""
                print(f"{kind} {f.name}  score={al.score:.0f}{order}  {url}")
                print(f"       excerpt: {fr[:200]!r}")
                print(f"       source : {t[al.dest_start:al.dest_end][:200]!r}")
                break
            pos = i + len(fr)
    bchecked = bfailed = 0
    if not argv:
        bchecked, bfailed = check_budget(texts)
    total_checked = checked + bchecked
    total_failed = failed + bfailed
    print(f"\nchecked {checked} evidence excerpts and {bchecked} budget quotes "
          f"from {len(texts)} sources, {total_failed} failed, {skipped} files without excerpt")
    if total_checked == 0:
        print("nothing checked: treat as failure")
        return 2
    return 1 if total_failed else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
