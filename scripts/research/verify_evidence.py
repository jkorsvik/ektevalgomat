#!/usr/bin/env python3
"""Fast, parallel verification of claimed evidence URLs against live sources.

Built to answer one question fast: does this cited case/document actually
exist, and does its real title match the claim -- or was it fabricated /
misattributed by a delegated research session (see CONTEXT.md incidents)?

Design choice: for stortinget.no case pages (`?p=<id>`) this hits the
machine-readable XML export API (data.stortinget.no/eksport/sak?sakid=<id>)
instead of scraping the human HTML page. That API returns a compact,
structured payload (korttittel/kortvedtak/henvisning/komite) in one small
request, and a non-existent case reliably comes back as HTTP 500 -- so
existence + title checks are one fast request instead of a full page
render. Everything else gets a cheap HEAD/partial-GET status check.

No third-party dependencies (stdlib only) so it runs with a bare `python3`,
no venv setup needed for a quick check.

Usage:
    python3 verify_evidence.py --file urls.txt
    python3 verify_evidence.py https://www.stortinget.no/.../Sak/?p=105664 https://...
    echo "https://..." | python3 verify_evidence.py

    # Content check, not just existence: a real Innst./Dok8 ID can still be
    # cited for the wrong topic (misattribution, worse than a plain 404
    # because it looks verified at a glance). Pass a TSV file of
    # `url<TAB>expected substring` and each hit is checked for that
    # keyword in its korttittel/kortvedtak (Storting sak) or full body
    # (generic URL), case-insensitive:
    python3 verify_evidence.py --file evidence.tsv

Rate limiting: Stortinget's open data documents a 100 calls/minute cap
(https://data.stortinget.no/om-datatjenesten/). This script defaults to a
global ceiling of 80/min across all sak lookups, shared by all worker
threads, so it stays safely under the limit no matter the worker count.
"""
from __future__ import annotations

import argparse
import re
import sys
import threading
import time
import urllib.error
import urllib.request
import xml.etree.ElementTree as ET
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field

STORTING_SAK_RE = re.compile(r"stortinget\.no/.*?[?&]p=(\d+)", re.IGNORECASE)
STORTING_API_TMPL = "https://data.stortinget.no/eksport/sak?sakid={sakid}"
NS = {"s": "http://data.stortinget.no"}
USER_AGENT = "ektevalgomat-evidence-verifier/1.0 (+internal research QA)"


class RateLimiter:
    """Simple shared token bucket: at most `per_minute` calls/min, thread-safe."""

    def __init__(self, per_minute: int):
        self._interval = 60.0 / per_minute
        self._lock = threading.Lock()
        self._next_allowed = 0.0

    def wait(self) -> None:
        with self._lock:
            now = time.monotonic()
            start = max(now, self._next_allowed)
            self._next_allowed = start + self._interval
            delay = start - now
        if delay > 0:
            time.sleep(delay)


@dataclass
class Result:
    url: str
    kind: str  # "storting_sak" | "http"
    ok: bool
    status: str  # human-readable verdict
    detail: str = ""
    fetched: dict = field(default_factory=dict)


def fetch(url: str, timeout: float = 10.0, max_bytes: int = 8192) -> tuple[int, bytes]:
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read(max_bytes)
    except urllib.error.HTTPError as e:
        return e.code, b""
    except Exception as e:  # noqa: BLE001 - network errors of any shape
        return -1, str(e).encode()


def _keyword_hit(keyword: str, *texts: str) -> bool:
    needle = keyword.strip().lower()
    return any(needle in t.lower() for t in texts if t)


def check_storting_sak(url: str, sakid: str, limiter: RateLimiter, keyword: str = "") -> Result:
    limiter.wait()
    api_url = STORTING_API_TMPL.format(sakid=sakid)
    status, body = fetch(api_url, max_bytes=1 << 20)
    if status != 200 or not body:
        return Result(
            url=url,
            kind="storting_sak",
            ok=False,
            status=f"FABRICATED (sakid={sakid} -> HTTP {status})",
            detail="Case ID does not exist in Stortinget's export API.",
        )
    try:
        root = ET.fromstring(body)
    except ET.ParseError as e:
        return Result(url, "storting_sak", False, "UNPARSEABLE", str(e))

    def text(tag: str) -> str:
        el = root.find(f"s:{tag}", NS)
        return (el.text or "").strip() if el is not None else ""

    korttittel = text("korttittel")
    henvisning = text("henvisning")
    ferdigbehandlet = text("ferdigbehandlet")
    kortvedtak = text("kortvedtak")

    if keyword and not _keyword_hit(keyword, korttittel, kortvedtak, henvisning):
        return Result(
            url=url,
            kind="storting_sak",
            ok=False,
            status=f"MISMATCH: case exists but does not mention '{keyword}'",
            detail=f"sakid={sakid} -- likely misattribution, not fabrication",
            fetched={"korttittel": korttittel, "henvisning": henvisning, "kortvedtak": kortvedtak[:300]},
        )

    return Result(
        url=url,
        kind="storting_sak",
        ok=True,
        status="EXISTS",
        detail=f"sakid={sakid} ferdigbehandlet={ferdigbehandlet}",
        fetched={
            "korttittel": korttittel,
            "henvisning": henvisning,
            "kortvedtak": kortvedtak[:300],
        },
    )


def check_generic_url(url: str, keyword: str = "") -> Result:
    status, body = fetch(url, max_bytes=(1 << 20) if keyword else 8192)
    if status == 200:
        m = re.search(rb"<title[^>]*>(.*?)</title>", body, re.IGNORECASE | re.DOTALL)
        title = re.sub(rb"\s+", b" ", m.group(1)).decode("utf-8", "replace").strip() if m else ""
        text_body = body.decode("utf-8", "replace")
        if keyword and not _keyword_hit(keyword, text_body):
            return Result(
                url, "http", False,
                f"MISMATCH: page exists but does not mention '{keyword}'",
                fetched={"title": title},
            )
        return Result(url, "http", True, "200 OK", fetched={"title": title})
    if status == 404:
        return Result(url, "http", False, "404 NOT FOUND")
    if status == -1:
        return Result(url, "http", False, f"NETWORK ERROR: {body.decode('utf-8', 'replace')}")
    return Result(url, "http", False, f"HTTP {status}")


def classify_and_check(url: str, limiter: RateLimiter, keyword: str = "") -> Result:
    m = STORTING_SAK_RE.search(url)
    if m:
        return check_storting_sak(url, m.group(1), limiter, keyword)
    return check_generic_url(url, keyword)


def read_items(args: argparse.Namespace) -> list[tuple[str, str]]:
    """Return (url, expected_keyword) pairs. Keyword is '' when not given.

    Accepts plain URLs (one per line/arg) or TSV lines `url<TAB>keyword`.
    """
    lines: list[str] = [f"{u}" for u in args.urls]
    if args.file:
        with open(args.file, encoding="utf-8") as f:
            lines.extend(line.rstrip("\n") for line in f)
    if not args.urls and not args.file and not sys.stdin.isatty():
        lines.extend(line.rstrip("\n") for line in sys.stdin)

    seen: dict[str, str] = {}
    for line in lines:
        line = line.strip()
        if not line or not line.startswith("http"):
            continue
        if "\t" in line:
            url, keyword = line.split("\t", 1)
        else:
            url, keyword = line, ""
        url = url.strip().strip("\"'")
        if url not in seen:
            seen[url] = keyword.strip()
    return list(seen.items())



def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("urls", nargs="*", help="URLs to verify")
    ap.add_argument("--file", "-f", help="File with one URL per line")
    ap.add_argument("--workers", type=int, default=12, help="Thread pool size (default 12)")
    ap.add_argument("--storting-rate", type=int, default=80, help="Max Storting API calls/min (default 80, cap is 100)")
    ap.add_argument("--quiet-ok", action="store_true", help="Only print failures/fabrications")
    args = ap.parse_args()

    urls = read_items(args)
    if not urls:
        print("No URLs given. Use --file, positional args, or pipe via stdin.", file=sys.stderr)
        return 2

    limiter = RateLimiter(args.storting_rate)
    results: list[Result] = []
    start = time.monotonic()
    url_order = [u for u, _ in urls]
    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        futs = {pool.submit(classify_and_check, u, limiter, kw): u for u, kw in urls}
        for fut in as_completed(futs):
            results.append(fut.result())
    elapsed = time.monotonic() - start

    results.sort(key=lambda r: url_order.index(r.url))


    fabricated = [r for r in results if r.kind == "storting_sak" and not r.ok]
    dead = [r for r in results if r.kind == "http" and not r.ok]
    ok = [r for r in results if r.ok]

    for r in results:
        if args.quiet_ok and r.ok:
            continue
        marker = "OK " if r.ok else "FAIL"
        print(f"[{marker}] {r.status:<45} {r.url}")
        if r.fetched.get("korttittel"):
            print(f"        korttittel: {r.fetched['korttittel']}")
        if r.fetched.get("henvisning"):
            print(f"        henvisning: {r.fetched['henvisning']}")
        if r.fetched.get("title"):
            print(f"        <title>: {r.fetched['title']}")

    print()
    print(f"Checked {len(results)} URLs in {elapsed:.1f}s "
          f"({len(ok)} ok, {len(fabricated)} fabricated Storting cases, {len(dead)} dead/other failures)")
    return 1 if (fabricated or dead) else 0


if __name__ == "__main__":
    raise SystemExit(main())
