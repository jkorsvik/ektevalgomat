# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Promote a passed lane-L person record out of staging.

Default is a dry run. It re-runs the gate with the fetch ON, then shows the
destination. `--apply` writes the file.

    uv run tools/autopilot/lobby/promote_people.py tools/autopilot/lobby/staging/<id>
    uv run tools/autopilot/lobby/promote_people.py tools/autopilot/lobby/staging/<id> --apply

Destination is `internal/data/content/people/<id>.yaml`, inside the
embedded content root so the Go loader can read it. The Go loader and the `/personer` page are the next
loop item; until they exist, a promoted record is verified but not served. That
is honest: a record the site cannot show is not yet a claim on the site.

A destination that already exists is not overwritten silently. A changed record
is a review event: pass `--force` only after reading the diff.
"""

from __future__ import annotations

import argparse
import pathlib
import shutil
import sys

import yaml

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import gate_people as G  # noqa: E402 - reuse the one validator

ROOT = pathlib.Path(__file__).resolve().parents[3]
DEST = ROOT / "internal/data/content/people"
STAGING = pathlib.Path(__file__).resolve().parent / "staging"


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("path", nargs="?", default="--all",
                    help="staging directory or person.yaml (default: --all)")
    ap.add_argument("--all", action="store_true", help="promote every staged record")
    ap.add_argument("--apply", action="store_true", help="write the file (default: dry run)")
    ap.add_argument("--force", action="store_true", help="overwrite an existing record")
    ap.add_argument("--no-fetch", action="store_true", help="skip the source fetch (offline)")
    args = ap.parse_args(argv)

    if args.all or args.path == "--all":
        return promote_all(args)

    target = pathlib.Path(args.path)
    if not target.exists():
        print(f"no such path: {target}", file=sys.stderr)
        return 2
    files = G.collect(target)
    if len(files) != 1:
        print(f"expected exactly one person record, found {len(files)}", file=sys.stderr)
        return 2
    src = files[0]

    scope = yaml.safe_load(G.SCOPE.read_text())
    snap = G.snapshot_date()
    errs = G.validate(src, scope, snap, not args.no_fetch)
    if errs:
        print(f"FAIL {src}")
        for e in errs:
            print(f"   - {e}")
        print("not promoted")
        return 1

    doc = yaml.safe_load(src.read_text())
    pid = doc["id"]
    out = DEST / f"{pid}.yaml"
    if out.exists() and not args.force:
        print(f"destination exists: {out.relative_to(ROOT)} (use --force after reading the diff)")
        return 1

    print(f"PASS {src}")
    print(f"-> {out.relative_to(ROOT)}")
    if not args.apply:
        print("dry run; pass --apply to write")
        return 0

    DEST.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(src, out)
    print(f"wrote {out.relative_to(ROOT)}")
    return 0


def promote_all(args) -> int:
    """Promote every staged record in one process. Existing files need --force."""
    scope = yaml.safe_load(G.SCOPE.read_text())
    snap = G.snapshot_date()
    DEST.mkdir(parents=True, exist_ok=True)
    ok = skipped = failed = 0
    for src in sorted(G.collect(STAGING)):
        errs = G.validate(src, scope, snap, False)
        if errs:
            failed += 1
            print(f"FAIL {src}: {errs[0]}")
            continue
        doc = yaml.safe_load(src.read_text())
        out = DEST / f"{doc['id']}.yaml"
        if out.exists() and not args.force:
            skipped += 1
            continue
        if args.apply:
            shutil.copyfile(src, out)
        ok += 1
    print(f"promoted {ok}, unchanged {skipped}, failed {failed}"
          f"{'' if args.apply else ' (dry run)'}")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
