#!/usr/bin/env python3
"""Build the Riksrevisjonen index: what the state's auditor published, and when.

Source: riksrevisjonen.no/rapporter/. Riksrevisjonen is the Storting's external
auditor, so its reports are the control on the power the rest of this project
describes: the body that tells the Storting whether the money was spent as decided.

What is collectable, and what is not, is recorded rather than glossed:

  COLLECTED   the listing page's ten most recent reports, with publication date,
              title, summary and the session taken from the URL.
  NOT         the archive behind them. /sitemap.xml answers 404, /robots.txt 403,
              and /rapporter/?page=2 returns the same ten reports as page one, so
              older reports cannot be enumerated by URL. Ten recent reports is a
              real index of a real source; it is not the whole archive, and the
              index says so in its own source note.

Dates are parsed from the page's own "Offentliggjort DD.MM.YYYY" text and stored as
ISO. A report whose date cannot be parsed is a failure, not a skip: an undated
audit finding is not much use for a timeline.
"""

from __future__ import annotations

import datetime
import html as htmllib
import json
import pathlib
import re
import sys
import urllib.request

PAGE = "https://www.riksrevisjonen.no/rapporter/"
OUT = pathlib.Path("internal/data/audits/index.json")

LIMITS = [
    "/sitemap.xml -> 404",
    "/sitemap_index.xml -> 404",
    "/robots.txt -> 403",
    "/rapporter/?page=2 -> same 10 reports as page 1",
]


def text(fragment: str) -> str:
    return re.sub(r"\s+", " ", htmllib.unescape(re.sub(r"<[^>]+>", " ", fragment))).strip()


def main() -> int:
    req = urllib.request.Request(PAGE, headers={"User-Agent": "ektevalgomat-audits/1"})
    with urllib.request.urlopen(req, timeout=60) as r:
        page = r.read().decode("utf-8", errors="replace")

    # The report link sits OUTSIDE the <article> (it wraps it), so an href search
    # inside each article finds nothing. Walk the document once, remember the last
    # report URL seen, and attach it to the article that follows.
    reports = []
    token = re.compile(r'href="(/rapporter-mappe/[^"]+)"|<article.*?</article>', re.S | re.I)
    last_url = None
    for m in token.finditer(page):
        if m.group(1):
            last_url = m.group(1)
            continue
        art = m.group(0)
        if not last_url:
            continue
        # Prefer the machine-readable time attribute; fall back to the visible text.
        d = re.search(r'<time[^>]+datetime="(\d{2})\.(\d{2})\.(\d{4})"', art)
        if not d:
            d = re.search(r"Offentliggjort\s+(\d{2})\.(\d{2})\.(\d{4})", text(art))
        if not d:
            print(f"no publication date for {last_url}", file=sys.stderr)
            return 1
        day, month, year = d.groups()
        published = f"{year}-{month}-{day}"
        # Each field comes from its own element. Deriving the title from the body
        # text was wrong: the heading carries no trailing period, so the split ran
        # the title and the summary together.
        h2 = re.search(r"<h2[^>]*>(.*?)</h2>", art, re.S | re.I)
        title = text(h2.group(1)) if h2 else ""
        if not title:
            print(f"no heading for {last_url}", file=sys.stderr)
            return 1
        para = re.search(r"<blockquote.*?<p[^>]*>(.*?)</p>", art, re.S | re.I)
        summary = text(para.group(1))[:400] if para else ""
        cats = re.search(r'rr-search-result-meta[^>]*>(.*?)</div>', art, re.S | re.I)
        categories = [c.strip() for c in text(cats.group(1)).split("/") if c.strip()] if cats else []
        session = re.search(r"/rapporter-mappe/(no-\d{4}-\d{4})/", last_url)
        reports.append(
            {
                "url": "https://www.riksrevisjonen.no" + last_url,
                "session": session.group(1) if session else "",
                "published": published,
                "title": title,
                "summary": summary,
                "categories": categories,
            }
        )
        last_url = None

    if len(reports) < 5:
        print(f"only {len(reports)} reports parsed: refusing to write", file=sys.stderr)
        return 2

    seen = set()
    for r in reports:
        if r["url"] in seen:
            print(f"duplicate report url {r['url']}", file=sys.stderr)
            return 1
        seen.add(r["url"])

    doc = {
        "source": {
            "name": "Riksrevisjonen, rapporter",
            "url": PAGE,
            "retrieved_at": datetime.date.today().isoformat(),
            "note": (
                "The listing page's most recent reports. This is not the whole archive: the "
                "site exposes no sitemap and no working pagination, so older reports cannot be "
                "enumerated. Ten recent reports are collected; the limit is recorded under "
                "not_collectable rather than left for a reader to discover."
            ),
        },
        "reports": sorted(reports, key=lambda r: r["published"], reverse=True),
        "not_collectable": {
            "archive": {
                "reason": (
                    "Older Riksrevisjonen reports cannot be enumerated by URL from the public "
                    "site: no sitemap, robots.txt refuses, and the page parameter is ignored."
                ),
                "attempts": LIMITS,
            }
        },
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n")

    print(f"wrote {OUT}: {len(reports)} reports, {reports[0]['published']} to {sorted(r['published'] for r in reports)[0]}")
    for r in sorted(reports, key=lambda r: r["published"], reverse=True)[:3]:
        print(f"  {r['published']}  {r['title'][:70]}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
