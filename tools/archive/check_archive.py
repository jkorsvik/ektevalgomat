# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Re-prove the source archive against the data, with no network.

Usage:
  uv run tools/archive/check_archive.py --archive DIR

What it proves, and why each part is needed:

1. Coverage both ways. Every evidence item with a verbatim_excerpt has an
   entry, and every entry belongs to an item with an excerpt. A one-way check
   would pass an archive that silently dropped the claims whose links broke,
   which is the exact set the archive exists for.

2. Each fragment is still in its window. The archive's promise is not "we
   stored something", it is "the words this claim rests on are in here". Both
   are checked with the verifier's own norm(), so the archive and the
   verifier cannot disagree about what "the same text" means.

3. Provenance is complete. source_url matches the data, document_sha256 is a
   real digest, fetched_at is a date, document_bytes is positive. A window
   with no provenance is an anonymous quotation and cannot be checked by
   anyone.

4. It says how much it checked and fails when that is zero. A gate that
   quietly verifies nothing is worse than no gate: it manufactures assurance.

Exit 0 only if every check passes.
"""

import argparse
import hashlib
import json
import pathlib
import re
import sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "tools/verify"))
import verify_excerpts as ve  # noqa: E402

EVIDENCE = ROOT / "internal/data/content/evidence"
HEX64 = re.compile(r"^[0-9a-f]{64}$")
ISO_DATE = re.compile(r"^\d{4}-\d{2}-\d{2}$")
ELISION = " […] "


def load_data() -> dict[str, dict]:
    out = {}
    for f in sorted(EVIDENCE.glob("*.yaml")):
        d = yaml.safe_load(f.read_text(encoding="utf-8")) or {}
        out[f.stem] = {
            "excerpt": (d.get("verbatim_excerpt") or "").strip(),
            "source_url": d.get("source_url") or "",
        }
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--archive", required=True, help="directory holding entries.jsonl and windows/")
    args = ap.parse_args()

    root = pathlib.Path(args.archive).expanduser()
    entries_file = root / "entries.jsonl"
    if not entries_file.exists():
        print(f"FAIL: {entries_file} does not exist. Build the archive first.")
        return 1

    entries = [json.loads(line) for line in entries_file.read_text(encoding="utf-8").splitlines() if line.strip()]
    data = load_data()
    quoted = {k: v for k, v in data.items() if v["excerpt"]}

    fails: list[str] = []
    checked_fragments = 0

    if not entries:
        print("FAIL: the archive has no entries. This gate checked nothing.")
        return 1
    if not quoted:
        print("FAIL: the data has no quoted claims. This gate checked nothing.")
        return 1

    seen: dict[str, dict] = {}
    for e in entries:
        eid = e.get("evidence_id") or ""
        if eid in seen:
            fails.append(f"{eid}: archived twice")
        seen[eid] = e
        if eid not in data:
            fails.append(f"{eid}: entry has no evidence file")
            continue
        if not data[eid]["excerpt"]:
            fails.append(f"{eid}: entry for a claim with no excerpt")
            continue
        if e.get("source_url") != data[eid]["source_url"]:
            fails.append(f"{eid}: source_url drifted from the data")
        if not HEX64.match(str(e.get("document_sha256") or "")):
            fails.append(f"{eid}: document_sha256 is not a sha256")
        if not ISO_DATE.match(str(e.get("fetched_at") or "")):
            fails.append(f"{eid}: fetched_at is not a date: {e.get('fetched_at')!r}")
        if not isinstance(e.get("document_bytes"), int) or e["document_bytes"] <= 0:
            fails.append(f"{eid}: document_bytes is not positive")

        path = root / str(e.get("window_file") or "")
        if not path.exists():
            fails.append(f"{eid}: window file missing: {e.get('window_file')}")
            continue
        window = path.read_text(encoding="utf-8")
        normed = ve.norm(window)
        frags = ve.fragments(data[eid]["excerpt"])
        if not frags:
            fails.append(f"{eid}: excerpt has no usable fragment")
            continue
        for frag in frags:
            checked_fragments += 1
            if frag not in normed:
                fails.append(f"{eid}: a fragment of the excerpt is not in the window")
        # The elision is how a reader is told the quotes came from different
        # places. A multi-fragment window with no elision is either one
        # contiguous quote (fine) or a join that hides a jump (not fine).
        if len(frags) > 1 and ELISION.strip() not in window:
            fails.append(f"{eid}: joins {len(frags)} fragments without an elision marker")

    missing = sorted(set(quoted) - set(seen))
    for eid in missing:
        fails.append(f"{eid}: quoted claim has no archived window")

    print(f"entries: {len(entries)}   quoted claims: {len(quoted)}   fragments checked: {checked_fragments}")
    print(f"coverage: {len(quoted) - len(missing)}/{len(quoted)} quoted claims archived")
    if checked_fragments == 0:
        print("FAIL: checked no fragments.")
        return 1
    if fails:
        print(f"\n{len(fails)} problem(s):")
        for f in fails[:40]:
            print("  " + f)
        if len(fails) > 40:
            print(f"  ... and {len(fails) - 40} more")
        return 1
    print("OK: every quoted claim has a window, and every window holds its quote.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
