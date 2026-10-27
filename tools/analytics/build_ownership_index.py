#!/usr/bin/env python3
"""Build the state-ownership index: what the state owns, how much, and why.

Source: regjeringen.no, "Hva staten eier" (the annual state ownership report's
company list). It is the authoritative public statement of the state's holdings:
one row per company with the state's share and the ownership category that governs
it. Each row is enriched with Brønnøysund's register (organisation number, legal
form, employees, industry code) so a reader can go from "the state owns 67 % of
Equinor" to the company record itself.

Why this is a separate index from the party-finance one: it answers a different
question. Party finance is money going TO the parties; this is what the state owns
on behalf of everyone. Folding them together would be the same mistake as adding
the Partiloven register to the SSB accounts.

Two rules the code enforces rather than promises:

  - A company whose share cannot be parsed is a failure, not a skip. A silently
    dropped row is an ownership fact that disappears.
  - A company Brønnøysund does not match by exact name is stored WITHOUT an
    organisation number and listed in the report. Fuzzy matching would attach the
    wrong company's employees and legal form, which is worse than an empty field.

Ownership categories, as the report defines them:
  1  commercial objectives
  2  commercial objectives and sector policy
  3  sector policy objectives
  4  other (research, culture and similar)
  -  not categorised in the report (særlov companies, partial holdings)
"""

from __future__ import annotations

import datetime
import html as htmllib
import json
import pathlib
import re
import sys
import urllib.parse
import urllib.request

PAGE = "https://www.regjeringen.no/no/tema/naringsliv/statlig-eierskap/selskaper---ny/id2604524/"
BRREG = "https://data.brreg.no/enhetsregisteret/api/enheter"
OUT = pathlib.Path("internal/data/ownership/index.json")

# Two names in the report are not the names Brønnøysund holds, and both are
# resolvable without guessing. Each entry is an explicit statement about one
# company, with the organisation number it resolves to, so a reviewer can check it
# in one lookup. Anything not listed here stays unmatched rather than being fuzzy
# matched onto a different company.
ALIASES = {
    # The report uses the everyday short form; Brønnøysund holds the full company
    # name. Verified: 817209882.
    "AS Vinmonopolet": "Aktieselskapet Vinmonopolet",
}


def strip_parenthetical(name: str) -> str:
    """Drop a trailing "(...)" from a register name.

    Brønnøysund appends an abbreviation to some names, e.g.
    "NORDISK INSTITUTT FOR ODONTOLOGISKE MATERIALER AS (NIOM AS)". Removing the
    suffix is a rule about the register's format, not a similarity guess.
    """
    return re.sub(r"\s*\([^)]*\)\s*$", "", name).strip()


CATEGORIES = {
    "1": "kommersielle mål",
    "2": "kommersielle mål og sektorpolitiske mål",
    "3": "sektorpolitiske mål",
    "4": "andre mål",
    "-": "ikke kategorisert",
}


def get(url: str, timeout: int = 60) -> bytes:
    req = urllib.request.Request(url, headers={"User-Agent": "ektevalgomat-ownership/1"})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return r.read()


def text(cell: str) -> str:
    return re.sub(r"\s+", " ", htmllib.unescape(re.sub(r"<[^>]+>", " ", cell))).strip()


def parse_share(raw: str) -> float | None:
    m = re.search(r"(\d{1,3}(?:,\d+)?)\s*%", raw)
    if not m:
        return None
    return float(m.group(1).replace(",", "."))


def parse_companies(page: str) -> list[dict]:
    rows: list[dict] = []
    for table in re.findall(r"<table.*?</table>", page, re.S | re.I):
        for tr in re.findall(r"<tr.*?</tr>", table, re.S | re.I):
            cells = [text(c) for c in re.findall(r"<t[hd].*?</t[hd]>", tr, re.S | re.I)]
            if len(cells) < 2 or cells[0].lower().startswith("navn på selskap"):
                continue
            name = cells[0].replace("\xa0", " ").strip()
            share = parse_share(cells[1])
            if share is None:
                raise ValueError(f"unparseable share for {name!r}: {cells[1]!r}")
            category = cells[2].strip() if len(cells) > 2 else "-"
            special = "særlovselskap" in name
            clean = re.sub(r"\s*\(særlovselskap\)\s*", "", name).strip()
            rows.append(
                {
                    "company": clean,
                    "share": share,
                    "category": category,
                    "special_law_company": special,
                }
            )
    return rows


def brreg_lookup(name: str) -> dict | None:
    """Exact-name match only. A near match would attach another company's record."""
    url = f"{BRREG}?{urllib.parse.urlencode({'navn': name, 'size': 10})}"
    try:
        data = json.loads(get(url, timeout=30))
    except Exception as exc:  # noqa: BLE001 - a lookup failure must not lose the row
        print(f"  brreg lookup failed for {name}: {exc}", file=sys.stderr)
        return None
    target = re.sub(r"\s+", " ", ALIASES.get(name, name)).upper()
    for e in (data.get("_embedded") or {}).get("enheter") or []:
        registered = strip_parenthetical(re.sub(r"\s+", " ", e.get("navn", "")))
        if registered.upper() == target:
            return {
                "orgnr": e.get("organisasjonsnummer"),
                "legal_form": (e.get("organisasjonsform") or {}).get("kode"),
                "employees": e.get("antallAnsatte"),
                "industry_code": (e.get("naeringskode1") or {}).get("kode"),
                "industry": (e.get("naeringskode1") or {}).get("beskrivelse"),
                "municipality": (e.get("forretningsadresse") or {}).get("kommune"),
            }
    return None


def main() -> int:
    page = get(PAGE).decode("utf-8", errors="replace")
    rows = parse_companies(page)
    if len(rows) < 50:
        print(f"only {len(rows)} companies parsed: refusing to write", file=sys.stderr)
        return 2

    matched = 0
    for r in rows:
        hit = brreg_lookup(r["company"])
        if hit:
            r.update(hit)
            matched += 1
        else:
            r.update({"orgnr": "", "legal_form": "", "employees": None, "industry_code": "", "industry": "", "municipality": ""})

    unmatched = [r["company"] for r in rows if not r["orgnr"]]
    doc = {
        "source": {
            "name": "regjeringen.no, Hva staten eier",
            "url": PAGE,
            "retrieved_at": datetime.date.today().isoformat(),
            "brreg_api": BRREG,
            "note": (
                "The state's own list of its holdings, with the share and the ownership "
                "category that governs each company. Enriched from Brønnøysund by exact "
                "name match only; a company without an organisation number was not matched "
                "and is listed in unmatched_companies rather than guessed at."
            ),
        },
        "categories": CATEGORIES,
        "companies": sorted(rows, key=lambda r: (-r["share"], r["company"])),
        "unmatched_companies": unmatched,
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n")

    full = sum(1 for r in rows if r["share"] == 100)
    print(f"wrote {OUT}: {len(rows)} companies, {full} wholly owned, {matched} matched in Brønnøysund")
    print(f"unmatched ({len(unmatched)}): {', '.join(unmatched[:8])}{' …' if len(unmatched) > 8 else ''}")
    for cat in sorted({r["category"] for r in rows}):
        n = sum(1 for r in rows if r["category"] == cat)
        print(f"  category {cat:2} ({CATEGORIES.get(cat, 'ukjent')}): {n}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
