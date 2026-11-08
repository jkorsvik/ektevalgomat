#!/usr/bin/env python3
"""Build the immutable public coding audit from committed review inputs."""

from __future__ import annotations

import argparse
import hashlib
import json
from collections import Counter
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_INPUT = ROOT / "tools/audit/data/2026-11-01"
DEFAULT_OUTPUT = ROOT / "internal/data/coding_review/index.json"
INPUT_NAMES = (
    "blind.json", "key.json", "reviewer-a.json", "reviewer-b.json",
    "adjudication.json", "reviewed-evidence.json",
)

AMBIGUITY_NB = {
    "figure_requirement": "Hovedregelen krever et eget instrument «med et tall», mens KrF-eksempelet gir −2 for et programpunkt uten siffer. Et siffer kan derfor ikke være et absolutt krav. Vår lesning: eget instrument avgjør, og et tall er et typisk tegn, ikke et vilkår. Derfor står Venstre på +2 i packet-03. Med et absolutt tallkrav ville cellen vært +1.",
    "joint_forslag": "Skalaen regner et eget lovforslag som instrument, men sier også at medforslagsstillelse ikke er instrument. Et mindretallsforslag fra flere partier er begge deler. Vår lesning: felles forslag er medforslag, fordi eksemplene i skalaen bruker dokumenter fra ett parti. Hvis felles forslag telles som eget instrument, går flere celler tilbake til ±2.",
    "merknad_operativ": "Skalaen unntar bare komitemerknader som «states only a direction». Den sier ikke eksplisitt hva som skjer når en merknad beskriver en full operativ regel. Vår lesning: en merknad i en annen innstilling er ikke et av de tre instrumenttypene, og den har ingen tallfestet forpliktelse. Derfor blir disse cellene magnitude 1.",
    "vote_answers_exact_question": "Protokollen sier at en stemme mot ett design ikke automatisk beviser motstand mot prinsippet. Når avstemningen gjelder utsagnets egen formulering eller dens eksakte motsetning, vurderer vi den som et svar med magnitude 1. En strengere lesning kan gi null. Dette gjelder packet-15 og packet-16.",
}


def read(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def by_packet(doc: dict) -> dict:
    return {row["packet_id"]: row for row in doc["packets"]}


def sign(position: int | None) -> int:
    if position in (None, 0):
        return 0
    return 1 if position > 0 else -1


def current_evidence_by_id() -> dict:
    rows = {}
    for file in sorted((ROOT / "internal/data/content/evidence").glob("*.yaml")):
        row = yaml.safe_load(file.read_text(encoding="utf-8"))
        rows[row["id"]] = row
    if not rows:
        raise RuntimeError("no current evidence found")
    return rows


def build(input_dir: Path) -> dict:
    for name in INPUT_NAMES:
        if not (input_dir / name).is_file():
            raise RuntimeError(f"missing committed audit input: {input_dir / name}")
    input_hashes = {
        name: hashlib.sha256((input_dir / name).read_bytes()).hexdigest()
        for name in INPUT_NAMES
    }
    blind = read(input_dir / "blind.json")
    key = by_packet(read(input_dir / "key.json"))
    reviewer_a = by_packet(read(input_dir / "reviewer-a.json"))
    reviewer_b = by_packet(read(input_dir / "reviewer-b.json"))
    adjudication_doc = read(input_dir / "adjudication.json")
    adjudicated = by_packet(adjudication_doc)
    reviewed_doc = read(input_dir / "reviewed-evidence.json")
    reviewed = by_packet(reviewed_doc)
    current = current_evidence_by_id()
    agreement = Counter()
    resolutions = Counter()
    party_counts = Counter()
    packets = []

    for packet_id in sorted(key):
        k = key[packet_id]
        a_row = reviewer_a[packet_id]
        b_row = reviewer_b[packet_id]
        a, b = a_row["cell_position"], b_row["cell_position"]
        if a == b:
            kind = "exact"
        elif sign(a) == sign(b):
            kind = "same_direction"
        else:
            kind = "substantive"
        agreement[kind] += 1

        adj = adjudicated.get(packet_id)
        final = adj["final_position"] if adj else k["published_position"]
        if adj and adj["resolution"] == "unresolved":
            resolution = "unresolved"
        else:
            resolution = "changed" if final != k["published_position"] else "kept"
        resolutions[resolution] += 1
        party_counts[k["party_id"]] += 1

        reviewed_rows = reviewed[packet_id]["evidence"]
        if not reviewed_rows:
            raise RuntimeError(f"{packet_id}: no reviewed evidence identity")
        newest = max(reviewed_rows, key=lambda row: (row["date"], row["id"]))
        if resolution == "changed":
            current_row = current.get(newest["id"])
            if current_row is None or current_row.get("position") != final:
                actual = None if current_row is None else current_row.get("position")
                raise RuntimeError(
                    f"{packet_id}: correction {newest['id']} has {actual}, want {final}"
                )

        packets.append({
            "packet_id": packet_id,
            "statement_id": k["statement_id"],
            "party_id": k["party_id"],
            "reviewer_a": a,
            "reviewer_b": b,
            "published_position": k["published_position"],
            "final_position": final,
            "agreement": kind,
            "adjudicated": adj is not None,
            "resolution": resolution,
            "dispute_type": adj["dispute_type"] if adj else "agreement",
            "reviewer_a_reason": a_row["reason"],
            "reviewer_b_reason": b_row["reason"],
            "adjudication_rationale": adj["rationale"] if adj else "",
            "source_urls": (
                adj["source_urls"] if adj else
                sorted(set(a_row["sources_checked"] + b_row["sources_checked"]))
            ),
            "reviewed_evidence": reviewed_rows,
        })

    if agreement != Counter({"exact": 15, "same_direction": 1, "substantive": 2}):
        raise RuntimeError(f"unexpected reviewer agreement: {dict(agreement)}")
    expected_counts = adjudication_doc["meta"]["counts"]
    if resolutions["changed"] != expected_counts["change"] or resolutions["unresolved"] != expected_counts["unresolved"]:
        raise RuntimeError(f"resolution counts disagree with adjudication: {dict(resolutions)}")
    if len(packets) != 18 or len(party_counts) != 9 or set(party_counts.values()) != {2}:
        raise RuntimeError(f"bad sample shape: {len(packets)} packets, {dict(party_counts)}")

    return {
        "meta": {
            "reviewed_at": "2026-11-01",
            "source_revision": reviewed_doc["source_revision"],
            "input_sha256": input_hashes,
            "protocol": "docs/coding-review-protocol.md",
            "scale": "tools/autopilot/method/position-scale.md",
            "seed": blind["seed"],
            "population_cells": blind["population_cells"],
            "sample_cells": len(packets),
            "reviewer_a": "source-checker, DeepSeek V4.1 Flash 1017",
            "reviewer_b": "RAIWork, moonshotai/kimi-k3.1-pro",
            "adjudicator": "trust-auditor, DeepSeek V4.1 Flash 1017, separate context",
            "exact_agreement": agreement["exact"],
            "same_direction": agreement["same_direction"],
            "substantive_disagreement": agreement["substantive"],
            "changed_positions": resolutions["changed"],
            "kept_positions": resolutions["kept"],
            "unresolved_positions": resolutions["unresolved"],
            "limitations": "Agentgjennomgang, ikke uavhengig menneskelig kontroll. Dette er en stratifisert pilot med 18 celler, ikke et pålitelighetsmål for alle 517. To lesere så ingen publiserte posisjoner. En tredje kontekst vurderte de 11 cellene der svarene eller den publiserte kodingen var omstridt, på nytt mot primærkildene.",
        },
        "ambiguities": [
            {"id": row["id"], "text": AMBIGUITY_NB[row["id"]], "affects": row["affects"]}
            for row in adjudication_doc["ambiguities"]
        ],
        "packets": packets,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input-dir", type=Path, default=DEFAULT_INPUT)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    args = parser.parse_args()
    doc = build(args.input_dir)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    loaded = read(args.output)
    print(
        f"wrote {args.output} packets={len(loaded['packets'])} "
        f"exact={loaded['meta']['exact_agreement']} changed={loaded['meta']['changed_positions']} "
        f"revision={loaded['meta']['source_revision']}"
    )


if __name__ == "__main__":
    main()
