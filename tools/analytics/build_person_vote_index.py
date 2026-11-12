#!/usr/bin/env python3
"""Build a per-representative vote index from the same payloads the party index uses.

The division payload (`voteringsresultat`) carries one row per representative, each
with `representant.id`, `representant.parti.id` and a `votering` code. The party
index throws those rows away and keeps only a per-party line, so nothing on the
site has ever said how an individual voted. This builds that, and it cross-checks
itself against the party index: if the party line derived from the per-person rows
disagrees with the stored line for any division, the extraction is wrong and the
build fails rather than shipping a plausible-looking file.

The interesting number is `deviation`: how often a representative votes against
the line their own party took in the same division. That is a fact about the
person, not the party, and it is the reason to build this at all.

Usage:
    uv run tools/analytics/build_person_vote_index.py            # full pass
    uv run tools/analytics/build_person_vote_index.py --limit 50 # smoke test
"""

from __future__ import annotations

import argparse
import collections
import datetime as dt
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import build_vote_index as bvi  # noqa: E402  reuse CODE, fetch, RESULTAT, role()

ROOT = pathlib.Path(__file__).resolve().parents[2]
PARTY_INDEX = ROOT / "internal/data/votes/index.json"
OUT = ROOT / "internal/data/votes/persons.json"


def rows_for(votering_id: int) -> list[dict]:
    payload = bvi.fetch(bvi.RESULTAT.format(votering_id))
    return payload.get("voteringsresultat_liste") or []


def person_of(row: dict) -> tuple[str, str, str]:
    """(person id, party id from the API, display name) for one row."""
    rep = row.get("representant") or {}
    parti = (rep.get("parti") or {}).get("id") or ""
    name = " ".join(x for x in (rep.get("fornavn"), rep.get("etternavn")) if x).strip()
    return rep.get("id") or "", parti, name


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--limit", type=int, default=0, help="only the first N divisions")
    ap.add_argument("--out", default=str(OUT))
    args = ap.parse_args()

    index = json.loads(PARTY_INDEX.read_text())
    divisions = index["votes"]
    if args.limit:
        divisions = divisions[: args.limit]

    people: dict[str, dict] = {}
    mismatches: list[dict] = []
    checked = 0
    with_rows = 0
    row_total = 0
    sessions_seen: dict[str, set[str]] = collections.defaultdict(set)

    for n, division in enumerate(divisions, 1):
        vid = division["votering_id"]
        rows = rows_for(vid)
        checked += 1
        if not rows:
            continue
        with_rows += 1
        row_total += len(rows)

        # The party line, derived by the same function the party index uses, so a
        # disagreement here is a real disagreement and not two implementations
        # of "majority" drifting apart.
        derived = bvi.role(rows)
        stored = division.get("parties") or {}
        for pid, line in stored.items():
            if pid in derived and derived[pid] != line:
                mismatches.append(
                    {"votering_id": vid, "party": pid, "stored": line, "derived": derived[pid]}
                )

        session = division.get("session") or ""
        for row in rows:
            pid, party_api, name = person_of(row)
            if not pid:
                continue
            party = bvi.API_TO_ID.get(party_api, party_api.lower())
            p = people.setdefault(
                pid,
                {
                    "name": name,
                    "party": party,
                    "for": 0,
                    "mot": 0,
                    "absent": 0,
                    "deviation": 0,
                    "divisions": 0,
                },
            )
            if name and not p["name"]:
                p["name"] = name
            code = row.get("votering")
            label = bvi.CODE.get(code)
            if label is None:
                continue
            p["divisions"] += 1
            if label == "for":
                p["for"] += 1
            elif label == "mot":
                p["mot"] += 1
            else:
                p["absent"] += 1
            # Deviation is counted only where both sides actually voted and the
            # party had a line: absent is not dissent, and "split" is not a line.
            line = derived.get(party) or stored.get(party)
            if label in ("for", "mot") and line in ("for", "mot") and label != line:
                p["deviation"] += 1
            if session:
                sessions_seen[pid].add(session)

        if n % 2000 == 0:
            print(f"  {n}/{len(divisions)} divisions, {len(people)} people", flush=True)

    # A gate that cannot fail is not a gate: the party line is re-derived here
    # from rows this script read itself, and any disagreement with the stored
    # index stops the build.
    if mismatches:
        print(f"FAIL: {len(mismatches)} party-line mismatches between the per-person "
              f"rows and the stored party index; first: {mismatches[:3]}", file=sys.stderr)
        return 2
    if with_rows == 0:
        print("FAIL: no division yielded representative rows; nothing was measured", file=sys.stderr)
        return 3

    for pid, p in people.items():
        p["sessions"] = sorted(sessions_seen.get(pid, ()))
        p["voted"] = p["for"] + p["mot"]

    out = {
        "generated_at": dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "source": "data.stortinget.no voteringsresultat, one row per representative",
        "codes": {str(k): v for k, v in bvi.CODE.items()},
        "divisions_in_party_index": len(index["votes"]),
        "divisions_checked": checked,
        "divisions_with_rows": with_rows,
        "representative_rows": row_total,
        "party_line_mismatches": 0,
        "persons": dict(sorted(people.items())),
    }
    pathlib.Path(args.out).write_text(json.dumps(out, ensure_ascii=False, indent=1) + "\n")
    devs = sum(1 for p in people.values() if p["deviation"])
    print(f"wrote {args.out}: {len(people)} people, {row_total} rows over {with_rows} divisions, "
          f"{devs} with at least one deviation, 0 party-line mismatches")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
