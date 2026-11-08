#!/usr/bin/env python3
"""Build a reproducible, stratified and blind party-position review sample."""

from __future__ import annotations

import argparse
import hashlib
import json
import random
from collections import Counter
from pathlib import Path

import yaml

SEED = "ektevalgomat-coding-audit-2026-11-01-v1"
ROOT = Path(__file__).resolve().parents[2]
EVIDENCE = ROOT / "internal/data/content/evidence"
STATEMENTS = ROOT / "internal/data/content/statements"
PARTIES = ROOT / "internal/data/content/parties"
SENSITIVE = {"position", "context_note", "scan_summary", "id", "evidence_id"}


def load_yaml_dir(path: Path) -> list[dict]:
    rows = []
    for file in sorted(path.glob("*.yaml")):
        value = yaml.safe_load(file.read_text(encoding="utf-8"))
        if isinstance(value, dict):
            value["_file"] = str(file.relative_to(ROOT))
            rows.append(value)
    if not rows:
        raise RuntimeError(f"no YAML records found in {path}")
    return rows


def evidence_for_cell(evidence: list[dict], statement_id: str, party_id: str) -> list[dict]:
    rows = []
    for row in evidence:
        if row.get("party_id") != party_id or row.get("position") is None:
            continue
        tags = [row.get("statement_id"), *row.get("secondary_statement_ids", [])]
        if statement_id in tags:
            rows.append(row)
    return sorted(rows, key=lambda row: (row.get("date", ""), row.get("id", "")))


def newest_position(rows: list[dict]) -> int:
    return rows[-1]["position"]


def build_cells() -> tuple[list[dict], list[dict], list[dict]]:
    evidence = load_yaml_dir(EVIDENCE)
    statements = load_yaml_dir(STATEMENTS)
    parties = load_yaml_dir(PARTIES)
    cells = []
    for statement in statements:
        for party in parties:
            rows = evidence_for_cell(evidence, statement["id"], party["id"])
            if not rows:
                continue
            cells.append({"statement": statement, "party": party, "rows": rows})
    return cells, statements, parties


def choose_sample(cells: list[dict], parties: list[dict]) -> list[dict]:
    rng = random.Random(SEED)
    chosen = []
    for party in parties:
        own = [cell for cell in cells if cell["party"]["id"] == party["id"]]
        buckets = {
            2: [cell for cell in own if abs(newest_position(cell["rows"])) == 2],
            1: [cell for cell in own if abs(newest_position(cell["rows"])) == 1],
            0: [cell for cell in own if newest_position(cell["rows"]) == 0],
        }
        picked = []
        for magnitude in (2, 1):
            bucket = buckets[magnitude] or own
            if not bucket:
                continue
            candidate = rng.choice(bucket)
            if candidate not in picked:
                picked.append(candidate)
        if len(picked) < 2:
            for candidate in buckets[0]:
                if candidate not in picked:
                    picked.append(candidate)
                    break
        if len(picked) != 2:
            raise RuntimeError(f"could not choose two distinct cells for {party['id']}")
        chosen.extend(picked)
    rng.shuffle(chosen)
    return chosen


def blind_row(row: dict) -> dict:
    allowed = (
        "actor", "date", "evidence_type", "evidence_class", "claim_action",
        "verbatim_excerpt", "status", "source_tier", "source_name", "source_url",
        "claim_confidence",
    )
    blind = {key: row.get(key) for key in allowed if row.get(key) not in (None, "")}
    if any(key in blind for key in SENSITIVE):
        raise RuntimeError("sensitive field reached blind row")
    return blind


def write_sample(output: Path) -> None:
    cells, _, parties = build_cells()
    selected = choose_sample(cells, parties)
    blind_packets = []
    key_packets = []
    for index, cell in enumerate(selected, 1):
        packet_id = f"packet-{index:02d}"
        statement = cell["statement"]
        party = cell["party"]
        blind_packets.append({
            "packet_id": packet_id,
            "statement": {
                "number": statement["number"],
                "text": statement["text"],
                "practical_stakes": statement.get("practical_stakes"),
            },
            "party": {"name": party["name"], "short": party["short"]},
            "evidence": [blind_row(row) for row in cell["rows"]],
        })
        key_packets.append({
            "packet_id": packet_id,
            "statement_id": statement["id"],
            "party_id": party["id"],
            "published_position": newest_position(cell["rows"]),
            "evidence_files": [row["_file"] for row in cell["rows"]],
        })

    payload = {
        "protocol": "docs/coding-review-protocol.md",
        "seed": SEED,
        "population_cells": len(cells),
        "sample_cells": len(blind_packets),
        "packets": blind_packets,
    }
    output.mkdir(parents=True, exist_ok=True)
    (output / "blind.json").write_text(
        json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    (output / "key.json").write_text(
        json.dumps({"packets": key_packets}, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    verify(output, cells, parties)
    digest = hashlib.sha256((output / "blind.json").read_bytes()).hexdigest()
    print(f"sample={len(blind_packets)} population={len(cells)} sha256={digest}")


def verify(output: Path, cells: list[dict] | None = None, parties: list[dict] | None = None) -> None:
    if cells is None or parties is None:
        cells, _, parties = build_cells()
    payload = json.loads((output / "blind.json").read_text(encoding="utf-8"))
    key = json.loads((output / "key.json").read_text(encoding="utf-8"))
    packets = payload["packets"]
    if len(packets) != 18 or len(key["packets"]) != 18:
        raise RuntimeError("audit must contain exactly 18 cells")
    counts = Counter(item["party_id"] for item in key["packets"])
    expected = {party["id"] for party in parties}
    if set(counts) != expected or set(counts.values()) != {2}:
        raise RuntimeError(f"party stratification failed: {dict(counts)}")
    serialized = json.dumps(payload, ensure_ascii=False)
    forbidden = ('"position"', '"context_note"', '"scan_summary"', '"evidence_files"')
    if any(term in serialized for term in forbidden):
        raise RuntimeError("blind payload leaks a published interpretation field")
    if payload["population_cells"] != len(cells) or payload["population_cells"] <= 0:
        raise RuntimeError("population count is empty or stale")
    print(f"verified packets=18 parties={len(counts)} per_party=2 population={len(cells)} blind_fields=ok")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path)
    parser.add_argument("--verify", action="store_true")
    parser.add_argument("--expected", type=Path, help="require byte-identical blind.json")
    args = parser.parse_args()
    if args.verify:
        verify(args.output)
    else:
        write_sample(args.output)
    expected = args.expected
    if args.verify and expected is None:
        expected = ROOT / "tools/audit/data/2026-11-01/blind.json"
    if expected:
        actual = (args.output / "blind.json").read_bytes()
        expected_bytes = expected.read_bytes()
        if actual != expected_bytes:
            raise RuntimeError(
                "blind.json differs from expected; use the recorded source revision "
                "or check the committed sample in tools/audit/data/2026-11-01"
            )
        print(f"verified byte-identical sample: {expected}")


if __name__ == "__main__":
    main()
