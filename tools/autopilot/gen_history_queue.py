#!/usr/bin/env python3
"""Generate the lane-H queue: one item per statement that needs a second date.

Why this exists: the konsistens score needs two dated observations of a cell,
and most cells hold one. `tools/autopilot/history.py` finds the earlier votes;
this script names the statements that the historian should work, in the order
that gives the most cells a second date first.

It is idempotent and does not invent content. It reads the real evidence files,
counts the coded cells and their distinct years, and writes
`tools/autopilot/queue-history.yaml`. A status already recorded for the same
item id in `queue.yaml` or in the existing `queue-history.yaml` is carried
forward, so a re-run never resets finished work.

    uv run tools/autopilot/gen_history_queue.py

Nothing here is evidence. It is a work list.
"""

from __future__ import annotations

import collections
import pathlib
import re
import sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
INNST = re.compile(r"Innst\.\s*(\d+)\s*([A-Za-z])\s*\((\d{4})\s*[–\-]\s*(\d{4})\)")
CONTENT = ROOT / "internal/data/content"
OUT = ROOT / "tools/autopilot/queue-history.yaml"
PREV = ROOT / "tools/autopilot/queue.yaml"


def year(date: str) -> int | None:
    try:
        y = int(str(date).strip()[:4])
    except (ValueError, TypeError):
        return None
    return y if 1900 <= y <= 2100 else None


def load_evidence() -> list[dict]:
    rows = []
    for p in sorted((CONTENT / "evidence").glob("*.yaml")):
        try:
            d = yaml.safe_load(p.read_text())
        except yaml.YAMLError as exc:
            print(f"skip malformed {p.name}: {exc}", file=sys.stderr)
            continue
        if isinstance(d, dict):
            rows.append(d)
    return rows


def load_statements() -> list[dict]:
    out = []
    for p in sorted((CONTENT / "statements").glob("*.yaml")):
        d = yaml.safe_load(p.read_text())
        if isinstance(d, dict):
            out.append(d)
    return out


def carried_statuses() -> dict[str, tuple[str, dict]]:
    """id -> (status, item) from both queues, so a rerun keeps finished work."""
    seen: dict[str, tuple[str, dict]] = {}
    for path in (PREV, OUT):
        if not path.exists():
            continue
        try:
            doc = yaml.safe_load(path.read_text()) or {}
        except yaml.YAMLError:
            continue
        for it in doc.get("items") or []:
            if isinstance(it, dict) and it.get("id"):
                seen[it["id"]] = (it.get("status") or "pending", it)
    return seen


def main() -> int:
    statements = {s["id"]: s for s in load_statements() if s.get("id")}
    evidence = load_evidence()

    # cell (statement, party) -> set of years its coded rows carry
    cell_years: dict[tuple[str, str], set[int]] = collections.defaultdict(set)
    cell_count: collections.Counter[tuple[str, str]] = collections.Counter()
    has_ref: dict[str, bool] = collections.defaultdict(bool)
    for e in evidence:
        sid, pid = e.get("statement_id"), e.get("party_id")
        if not sid or not pid or e.get("position") is None:
            continue
        cell_count[(sid, pid)] += 1
        if INNST.search(f"{e.get('source_name') or ''} {e.get('claim_action') or ''}"):
            has_ref[sid] = True
        y = year(e.get("date") or "")
        if y:
            cell_years[(sid, pid)].add(y)

    prev = carried_statuses()

    items = []
    per_statement = collections.defaultdict(lambda: {"cells": 0, "need": 0})
    for (sid, pid), yrs in cell_years.items():
        per_statement[sid]["cells"] += 1
        if len(yrs) < 2:
            per_statement[sid]["need"] += 1

    # Only statements that exist and have at least one cell needing a date.
    for sid, c in per_statement.items():
        if c["need"] == 0 or sid not in statements:
            continue
        item_id = f"voteringsspor-{sid}-second-dates"
        status, old = prev.get(item_id, ("pending", {}))
        item = {
            "id": item_id,
            "kind": "voteringsspor",
            "lane": "historikk",
            "statement": sid,
            "method": "method/historian.md",
            "cells_total": c["cells"],
            "cells_need_second_date": c["need"],
            "route": "innstilling" if has_ref.get(sid) else "votering-only",
            "status": status,
        }
        # Carry a note, result or blocked_on forward verbatim.
        for k in ("note", "result", "blocked_on"):
            if old.get(k):
                item[k] = old[k]
        items.append(item)

    # Most cells rescued first; then statement id for stability.
    items.sort(key=lambda i: (-i["cells_need_second_date"], i["statement"]))
    doc = {
        "generated_from": "internal/data/content",
        "rule": "one item per statement with at least one coded cell holding <2 distinct years",
        "items": items,
    }
    OUT.write_text(yaml.safe_dump(doc, allow_unicode=True, sort_keys=False))

    total_cells = sum(1 for _ in cell_years)
    need_cells = sum(1 for yrs in cell_years.values() if len(yrs) < 2)
    ready_cells = total_cells - need_cells
    print(f"coded cells: {total_cells}; need a second date: {need_cells}; "
          f"already 2+ years: {ready_cells}")
    print(f"statements in lane-H queue: {len(items)} -> {OUT.relative_to(ROOT)}")
    pending = sum(1 for i in items if i["status"] == "pending")
    print(f"pending: {pending}; carried as passed/blocked/failed: {len(items) - pending}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
