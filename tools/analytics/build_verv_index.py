#!/usr/bin/env python3
"""Build a per-person index from the Storting verv register PDF.

The register is one dated PDF for the whole Storting, refreshed a few times a
year.  This script parses it once into a machine-readable index so the person
trackers do not each re-parse 3 900 lines.

What is stored: name, party, county, and the text of §2 (board and committee
posts), §3 (self-employed business type), §4 (paid position) plus company names
from §9 (company interests, names only, no share counts or values).

What is never stored: birth dates, addresses, phone, e-mail, health, family,
gifts, travel, debt, property, or any amount in kroner.  Those sections are
counted and dropped.

Usage:
    python3 tools/analytics/build_verv_index.py --pdf /tmp/ekv-register.pdf
"""
from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parents[2]
OUT = REPO / "internal/data/verv/index.json"
SOURCE_URL = (
    "https://www.stortinget.no/globalassets/pdf/verv-og-okonomiske-interesser/"
    "arkiv_2026-2027/pr-30-oktober-2026.pdf"
)
SNAPSHOT_DATE = "2026-10-30"

PARTIES = {"A", "FrP", "H", "KrF", "MDG", "R", "Sp", "SV", "V", "PF", "Uav"}

# § -> short label. Only these are kept.
KEEP = {2: "verv", 3: "naering", 4: "lonnet_stilling", 9: "selskapsinteresse"}
DROP = {
    5: "tidligere_arbeidsgiver",
    6: "framtidig_arbeidsgiver",
    7: "okonomisk_stotte",
    8: "eiendom",
    10: "utenlandsreiser",
    11: "gaver",
    15: "andre_forhold",
}

HEADER = re.compile(r"^(?P<name>[^§()]{3,80}?) \((?P<party>[A-Za-zÆØÅæøå]{1,6}), (?P<county>[^)]+)\)\s*$")
SECTION = re.compile(r"^§(?P<num>\d{1,2})\s{2,}(?P<label>[A-Za-zÆØÅæøå. ]{3,40}?)(?:\s{2,}(?P<text>.*))?$")


def pdf_to_text(pdf: Path) -> str:
    with tempfile.NamedTemporaryFile(suffix=".txt", delete=False) as fh:
        tmp = Path(fh.name)
    subprocess.run(["pdftotext", "-layout", str(pdf), str(tmp)], check=True)
    text = tmp.read_text(encoding="utf-8", errors="replace")
    tmp.unlink(missing_ok=True)
    return text


def clean(chunks: list[str]) -> str:
    """Join continuation lines but keep the line breaks: one line is often one post."""
    out = []
    for c in chunks:
        s = re.sub(r"\s{2,}", " ", c.strip())
        if s:
            out.append(s)
    return "\n".join(out)[:1200]


def parse(text: str):
    people = []
    cur = None
    sec = None
    buf: list[str] = []
    dropped: dict[str, int] = {}

    def flush():
        nonlocal sec, buf
        if cur is None or sec is None:
            buf, sec = [], None
            return
        body = clean(buf)
        if not body or body.replace("\n", " ") in {"-", "Ingen", "Har ingen registreringspliktige interesser"}:
            buf, sec = [], None
            return
        if sec in KEEP:
            cur["roles"].append({"section": sec, "kind": KEEP[sec], "text": body})
        elif sec in DROP:
            dropped[DROP[sec]] = dropped.get(DROP[sec], 0) + 1
        buf, sec = [], None

    for raw in text.splitlines():
        line = raw.rstrip()
        if not line.strip():
            continue
        m = HEADER.match(line.strip())
        if m and m.group("party") in PARTIES:
            flush()
            cur = {
                "name": m.group("name").strip(),
                "party": m.group("party"),
                "county": m.group("county").strip(),
                "roles": [],
            }
            people.append(cur)
            continue
        s = SECTION.match(line)
        if s:
            flush()
            n = int(s.group("num"))
            sec = n
            if s.group("text"):
                buf.append(s.group("text"))
            continue
        # continuation line of the current section
        if sec is not None and line.startswith("   "):
            buf.append(line)
    flush()
    return people, dropped


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--pdf", type=Path, required=True)
    ap.add_argument("--out", type=Path, default=OUT)
    ap.add_argument("--check", action="store_true", help="verify the existing index only")
    args = ap.parse_args()

    if args.check:
        if not args.out.exists():
            print(f"FAIL: {args.out} does not exist")
            return 1
        data = json.loads(args.out.read_text())
        n = len(data["people"])
        with_roles = sum(1 for p in data["people"] if p["roles"])
        print(f"index: {n} people, {with_roles} with roles, snapshot {data['snapshot_date']}")
        ok = n >= 150 and with_roles >= 80
        print("PASS" if ok else "FAIL: index is thinner than expected")
        return 0 if ok else 1

    text = pdf_to_text(args.pdf)
    people, dropped = parse(text)

    for p in people:
        # §9 company names only: strip any digits/amounts that slipped through
        for r in p["roles"]:
            if r["section"] == 9:
                r["text"] = re.sub(r"[\d.,]+\s*(%|NOK|kr)", "", r["text"]).strip()

    people.sort(key=lambda p: p["name"])
    names = [p["name"] for p in people]
    dupes = sorted({n for n in names if names.count(n) > 1})

    payload = {
        "generated_at": dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat(),
        "snapshot_date": SNAPSHOT_DATE,
        "source_url": SOURCE_URL,
        "source_sha256": hashlib.sha256(args.pdf.read_bytes()).hexdigest(),
        "not_stored": sorted(DROP.values()),
        "retained_note": (
            "Section text is the register's own wording. Section 9 keeps company "
            "names only; share counts and kroner amounts are stripped. Compensation "
            "percentages in section 2 describe the office, not the person's wealth."
        ),
        "counts": {
            "people": len(people),
            "with_roles": sum(1 for p in people if p["roles"]),
            "roles": sum(len(p["roles"]) for p in people),
            "dropped_sections": dropped,
        },
        "people": people,
    }
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n")

    c = payload["counts"]
    print(f"wrote {args.out}")
    print(f"people {c['people']}  with roles {c['with_roles']}  roles {c['roles']}")
    print(f"dropped sections: {c['dropped_sections'] or 'none'}")
    if dupes:
        print(f"WARN duplicate names ({len(dupes)}): {', '.join(dupes[:5])}")

    if c["people"] < 150:
        print(f"FAIL: only {c['people']} people parsed; expected 169 representatives")
        return 1
    if c["with_roles"] < 80:
        print(f"FAIL: only {c['with_roles']} with roles; the parser is probably misaligned")
        return 1
    print("PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
