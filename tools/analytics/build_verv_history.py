#!/usr/bin/env python3
"""Diff two or more editions of the Storting verv register.

The register is a snapshot, so a role's `from` is either a date the register
states or the snapshot date. That leaves one question a single snapshot cannot
answer: which roles did a representative hold *before*, and which have ended?

Two dated editions answer it without inventing a single date. A role present in
an earlier edition and absent from the newest was held in between; a role in
both was held throughout. This script records exactly that, with the two dates
it is derived from, and nothing more.

Output: internal/data/content/verv/history.json (embedded, so the site can show
it). Not a role list: the page states "seen 2024-04-19, gone by 2026-10-30",
never a fabricated start or end.

Usage:
    python3 tools/analytics/build_verv_history.py \
        --snapshot 2024-04-19 <file.pdf> \
        --snapshot 2026-10-30 <file.pdf>
"""
from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import pathlib
import re
import sys

HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import build_verv_index as B  # noqa: E402 - reuse the one parser

OUT = HERE.parents[1] / "internal/data/content/verv/history.json"
KEEP = {2}  # section 2 is a verv/post; 3, 4 and 9 are prose about income and holdings


def sig(text: str) -> str:
    """Stable key for one register entry: same post, whatever the wrapping."""
    return re.sub(r"[^a-zæøå0-9]+", "", text.lower())[:64]


def parse_snapshot(date: str, pdf: pathlib.Path) -> dict:
    people, _ = B.parse(B.pdf_to_text(pdf))
    out = {}
    for p in people:
        key = (p["name"], p["party"])
        roles = {}
        for r in p["roles"]:
            if r["section"] not in KEEP:
                continue
            roles[sig(r["text"])] = {
                "section": r["section"],
                "kind": r["kind"],
                "text": re.sub(r"\s+", " ", r["text"].split("\n")[0])[:120],
            }
        out[key] = {"name": p["name"], "party": p["party"], "county": p["county"], "roles": roles}
    return out


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--snapshot", nargs=2, action="append", required=True,
                    metavar=("DATE", "PDF"))
    ap.add_argument("--out", type=pathlib.Path, default=OUT)
    args = ap.parse_args(argv)

    snaps = sorted((d, pathlib.Path(p)) for d, p in args.snapshot)
    for d, p in snaps:
        if not p.exists():
            print(f"missing {p}", file=sys.stderr)
            return 2
        dt.date.fromisoformat(d)

    parsed = [(d, parse_snapshot(d, p)) for d, p in snaps]
    first_date, first = parsed[0]
    last_date, last = parsed[-1]

    ended, carried, new = [], 0, []
    for key, rec in first.items():
        if key not in last:
            continue  # left the Storting between editions: not a role change
        now = last[key]["roles"]
        for s, role in rec["roles"].items():
            if s in now:
                carried += 1
            else:
                ended.append({
                    "name": rec["name"], "party": rec["party"],
                    "section": role["section"], "kind": role["kind"],
                    "text": role["text"], "seen": first_date, "gone_by": last_date,
                })
    for key, rec in last.items():
        if key not in first:
            continue
        was = first[key]["roles"]
        for s, role in rec["roles"].items():
            if s not in was:
                new.append({
                    "name": rec["name"], "party": rec["party"],
                    "section": role["section"], "kind": role["kind"],
                    "text": role["text"], "first_seen": last_date,
                })

    payload = {
        "generated_at": dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat(),
        "snapshots": [
            {"date": d, "pdf_sha256": hashlib.sha256(p.read_bytes()).hexdigest(),
             "people": len(recs)}
            for (d, p), (_, recs) in zip(snaps, parsed)
        ],
        "compared_between": [first_date, last_date],
        "note": (
            "A role present in an earlier edition and absent from the newest was held "
            "in between. The register states neither its start nor its end, so none is "
            "written here: only the two dates the comparison rests on."
        ),
        "counts": {
            "people_in_both": sum(1 for k in first if k in last),
            "roles_carried": carried,
            "roles_ended": len(ended),
            "roles_new": len(new),
        },
        "ended": sorted(ended, key=lambda r: (r["party"], r["name"], r["text"])),
        "new": sorted(new, key=lambda r: (r["party"], r["name"], r["text"])),
    }
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n")

    c = payload["counts"]
    print(f"wrote {args.out}")
    print(f"people in both {c['people_in_both']}  carried {c['roles_carried']}  "
          f"ended {c['roles_ended']}  new {c['roles_new']}")
    # A general election replaces a large share of the Storting, so two editions
    # two years apart legitimately share only ~120 people. Below 100 something is
    # wrong with the parse, not with the Storting.
    if c["people_in_both"] < 100:
        print(f"FAIL: only {c['people_in_both']} people in both editions; the parse is suspect")
        return 1
    print("PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
