#!/usr/bin/env python3
"""Derived trends over the vote index and the evidence cells.

This does not create evidence and does not touch content. It reads
`internal/data/votes/index.json` (25 773 divisions, 2011-2026) and the
evidence YAML, and writes two files under `tools/autopilot/reports/trends/`:

  summary.json  the numbers, for a page or a chart later
  summary.md    the same numbers, for a human

Two things it answers honestly:

1. **Party cohesion over time.** Per session and party: how many divisions it
   took part in, how often it voted for the proposal, how often it split, and
   how often it was absent. A split is not a side; `absent` is not a no.

2. **Evidence readiness.** How many coded cells hold one date, two or more
   dates, and which cells changed direction between their first and last
   observation. This is the number the konsistens score depends on.

A number with no data behind it is omitted, never printed as zero.

    uv run tools/analytics/trends.py
"""

from __future__ import annotations

import collections
import datetime
import json
import pathlib
import sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
VOTES = ROOT / "internal/data/votes/index.json"
CONTENT = ROOT / "internal/data/content"
OUT = ROOT / "tools/autopilot/reports/trends"

PARTIES = ["ap", "frp", "h", "sv", "sp", "rodt", "mdg", "krf", "v"]
SIDES = ("for", "mot", "split", "absent")


def year(date: str) -> int | None:
    try:
        y = int(str(date).strip()[:4])
    except (ValueError, TypeError):
        return None
    return y if 1900 <= y <= 2100 else None


def load_cells() -> dict[tuple[str, str], list[dict]]:
    cells: dict[tuple[str, str], list[dict]] = collections.defaultdict(list)
    for p in sorted((CONTENT / "evidence").glob("*.yaml")):
        d = yaml.safe_load(p.read_text())
        if not isinstance(d, dict):
            continue
        sid, pid = d.get("statement_id"), d.get("party_id")
        if sid and pid and d.get("position") is not None:
            cells[(sid, pid)].append(d)
    return cells


def dimensions() -> dict[str, str]:
    out = {}
    for p in sorted((CONTENT / "statements").glob("*.yaml")):
        d = yaml.safe_load(p.read_text())
        if isinstance(d, dict) and d.get("id"):
            out[d["id"]] = d.get("dimension") or "ukjent"
    return out


def evidence_readiness() -> dict:
    cells = load_cells()
    dims = dimensions()
    by_year_count = collections.Counter()
    changed = []
    per_party_ready = collections.Counter()
    per_dim_ready = collections.Counter()
    per_dim_total = collections.Counter()
    for (sid, pid), rows in cells.items():
        yrs = sorted({year(r.get("date", "")) for r in rows} - {None})
        by_year_count[min(len(yrs), 3)] += 1  # 0,1,2,3+ buckets
        per_dim_total[dims.get(sid, "ukjent")] += 1
        if len(yrs) >= 2:
            per_party_ready[pid] += 1
            per_dim_ready[dims.get(sid, "ukjent")] += 1
            # earliest vs newest by date
            ordered = sorted(rows, key=lambda r: (str(r.get("date") or ""), str(r.get("id") or "")))
            first, last = ordered[0].get("position"), ordered[-1].get("position")
            if first is not None and last is not None and sign(first) != sign(last):
                # A sign change is not always an opposing reversal: a cell can
                # move to or from the neutral 0, which is a weaker claim. Keep
                # the distinction so a report cannot call all of them reversals.
                if {first, last} & {0}:
                    kind = "to_zero" if last == 0 else "from_zero"
                else:
                    kind = "opposing"
                changed.append({"statement": sid, "party": pid,
                                "from_year": year(ordered[0].get("date", "")),
                                "to_year": year(ordered[-1].get("date", "")),
                                "from": first, "to": last, "kind": kind})
    return {
        "coded_cells": len(cells),
        "by_distinct_years": {"0": by_year_count[0], "1": by_year_count[1],
                              "2": by_year_count[2], "3+": by_year_count[3]},
        "cells_with_two_or_more_years": sum(1 for r in cells.values()
                                            if len({year(x.get("date", "")) for x in r} - {None}) >= 2),
        "ready_per_party": dict(per_party_ready),
        "ready_per_dimension": {k: {"ready": per_dim_ready[k], "total": per_dim_total[k]}
                                for k in sorted(per_dim_total)},
        "direction_changes": changed,
        "direction_change_kinds": dict(collections.Counter(c["kind"] for c in changed)),
    }


def sign(p) -> int:
    return (p > 0) - (p < 0)


def vote_trends() -> dict:
    doc = json.loads(VOTES.read_text())
    votes = doc.get("votes") or []
    sessions = collections.defaultdict(list)
    for v in votes:
        if v.get("session"):
            sessions[v["session"]].append(v)

    per_session = []
    for ses in sorted(sessions):
        rows = sessions[ses]
        divisions = [v for v in rows if not v.get("unanimous")]
        tally = collections.Counter()
        for v in divisions:
            for s in (v.get("parties") or {}).values():
                tally[s] += 1
        per_session.append({
            "session": ses,
            "votes": len(rows),
            "divided": len(divisions),
            "unanimous": len(rows) - len(divisions),
            "sides": {k: tally[k] for k in SIDES if tally[k]},
        })

    # Per party: participation, for-share, split-rate, by session.
    per_party = {}
    for pid in PARTIES:
        series = []
        for ses in sorted(sessions):
            rows = [v for v in sessions[ses] if not v.get("unanimous")]
            n = f = m = sp = ab = 0
            for v in rows:
                s = (v.get("parties") or {}).get(pid)
                if s is None:
                    continue
                n += 1
                f += s == "for"
                m += s == "mot"
                sp += s == "split"
                ab += s == "absent"
            if n:
                series.append({"session": ses, "n": n, "for_share": round(f / max(f + m, 1), 3),
                               "split_rate": round(sp / n, 3), "absent_rate": round(ab / n, 3)})
        if series:
            per_party[pid] = series

    # Pairwise agreement per session, over divided votes where both voted.
    pairs = []
    for i, a in enumerate(PARTIES):
        for b in PARTIES[i + 1:]:
            total = agree = 0
            for v in votes:
                if v.get("unanimous"):
                    continue
                pa = (v.get("parties") or {}).get(a)
                pb = (v.get("parties") or {}).get(b)
                if pa in ("for", "mot") and pb in ("for", "mot"):
                    total += 1
                    agree += pa == pb
            if total:
                pairs.append({"a": a, "b": b, "n": total, "agreement": round(agree / total, 3)})
    pairs.sort(key=lambda p: -p["agreement"])
    return {"per_session": per_session, "per_party": per_party,
            "pairwise_agreement": pairs, "generated_from_votes": len(votes)}


def render(ev: dict, vt: dict) -> str:
    L = [f"# Trender, {datetime.date.today().isoformat()}", ""]
    L.append("Derived from the vote index and the evidence cells. Not evidence.")
    L.append("")
    L.append("## Konsistens: er det nok data?")
    L.append("")
    b = ev["by_distinct_years"]
    L.append(f"- Kodede celler: **{ev['coded_cells']}**")
    L.append(f"- Med 1 dato: **{b['1']}** — kan ikke si noe om utvikling")
    L.append(f"- Med 2 datoer: **{b['2']}**")
    L.append(f"- Med 3+ datoer: **{b['3+']}**")
    L.append(f"- Klare til å si noe: **{ev['cells_with_two_or_more_years']}** "
             f"({round(100 * ev['cells_with_two_or_more_years'] / max(ev['coded_cells'], 1))}%)")
    L.append("")
    if ev["direction_changes"]:
        L.append("### Dokumenterte retningsendringer")
        L.append("")
        kinds = ev.get("direction_change_kinds") or {}
        L.append(f"Motsatte omslag: **{kinds.get('opposing', 0)}**. "
                 f"Til eller fra nøytral 0: **{kinds.get('to_zero', 0) + kinds.get('from_zero', 0)}**.")
        L.append("")
        for c in ev["direction_changes"][:20]:
            label = {"opposing": "omslag", "to_zero": "til 0",
                     "from_zero": "fra 0"}.get(c.get("kind"), "endring")
            L.append(f"- {c['statement']} / {c['party']}: {c['from']:+d} ({c['from_year']}) "
                     f"→ {c['to']:+d} ({c['to_year']}) — {label}")
        L.append("")
    L.append("## Per dimensjon")
    L.append("")
    L.append("| dimensjon | klare | celler |")
    L.append("|---|---|---|")
    for d, v in ev["ready_per_dimension"].items():
        L.append(f"| {d} | {v['ready']} | {v['total']} |")
    L.append("")
    L.append("## Partisamhold over tid")
    L.append("")
    L.append("Per sesjon: andel forslag partiet stemte for, andel splitt, andel fravær.")
    L.append("")
    for pid, series in vt["per_party"].items():
        L.append(f"### {pid}")
        L.append("")
        L.append("| sesjon | n | for | splitt | fravær |")
        L.append("|---|---|---|---|---|")
        for s in series:
            L.append(f"| {s['session']} | {s['n']} | {s['for_share']:.0%} | "
                     f"{s['split_rate']:.0%} | {s['absent_rate']:.0%} |")
        L.append("")
    L.append("## Hvem stemmer likt")
    L.append("")
    L.append("Andel av delte voteringer der to partier stemte samme vei, hele perioden.")
    L.append("")
    L.append("| par | andel | n |")
    L.append("|---|---|---|")
    for p in vt["pairwise_agreement"][:15]:
        L.append(f"| {p['a']}–{p['b']} | {p['agreement']:.0%} | {p['n']} |")
    L.append("")
    return "\n".join(L)


def main() -> int:
    ev = evidence_readiness()
    vt = vote_trends()
    today = datetime.date.today().isoformat()
    dest = OUT / today
    dest.mkdir(parents=True, exist_ok=True)
    (dest / "summary.json").write_text(json.dumps({"evidence": ev, "votes": vt},
                                                  ensure_ascii=False, indent=2))
    (dest / "summary.md").write_text(render(ev, vt))
    b = ev["by_distinct_years"]
    k = ev.get("direction_change_kinds") or {}
    print(f"coded cells {ev['coded_cells']}: 1 date {b['1']}, 2 dates {b['2']}, 3+ {b['3+']}; "
          f"direction changes {len(ev['direction_changes'])} "
          f"(opposing {k.get('opposing', 0)}, to/from zero "
          f"{k.get('to_zero', 0) + k.get('from_zero', 0)})")
    print(f"votes {vt['generated_from_votes']} over {len(vt['per_session'])} sessions")
    print(f"wrote {dest.relative_to(ROOT)}/summary.{{json,md}}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
