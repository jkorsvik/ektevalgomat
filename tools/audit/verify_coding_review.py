#!/usr/bin/env python3
"""Verify committed inputs, immutable evidence hashes and generated audit output."""

from __future__ import annotations

import hashlib
import json
import subprocess
from pathlib import Path

import yaml

from build_coding_review import DEFAULT_INPUT, DEFAULT_OUTPUT, INPUT_NAMES, build


def main() -> None:
    output = json.loads(DEFAULT_OUTPUT.read_text(encoding="utf-8"))
    revision = output["meta"]["source_revision"]
    for name in INPUT_NAMES:
        actual = hashlib.sha256((DEFAULT_INPUT / name).read_bytes()).hexdigest()
        if output["meta"]["input_sha256"].get(name) != actual:
            raise RuntimeError(f"input hash mismatch: {name}")

    checked = 0
    for packet in output["packets"]:
        for row in packet["reviewed_evidence"]:
            raw = subprocess.check_output(["git", "show", f"{revision}:{row['file']}"])
            if hashlib.sha256(raw).hexdigest() != row["sha256"]:
                raise RuntimeError(f"reviewed evidence hash mismatch: {packet['packet_id']} {row['id']}")
            doc = yaml.safe_load(raw)
            if doc["id"] != row["id"] or doc.get("date", "") != row["date"]:
                raise RuntimeError(f"reviewed evidence identity mismatch: {packet['packet_id']}")
            checked += 1
    if checked != 20:
        raise RuntimeError(f"checked {checked} evidence hashes, want 20")

    expected = build(DEFAULT_INPUT)
    if expected != output:
        raise RuntimeError("generated audit output differs from internal/data/coding_review/index.json")
    print(f"verified inputs=6 packets={len(output['packets'])} evidence_hashes={checked} output=identical")


if __name__ == "__main__":
    main()
