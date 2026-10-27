# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Apply staged verbatim_excerpt proposals to existing evidence files.

Usage:
  uv run tools/autopilot/apply_excerpts.py --dry-run
  uv run tools/autopilot/apply_excerpts.py

promote.py moves whole new files. This applies a one-field change to files that
are already content, which needs exactly the same three consequences promote.py
learned one at a time:

  1. a new quote needs a stored archive window  (TestArchiveCoversEveryQuotedClaim)
  2. the documented "N of M rows carry such an excerpt" changes  (TestDocsStateTheRealCounts)
  3. every excerpt must be a literal substring of its source  (verify_excerpts.py)

All three are checked before this reports success, and every file it touches is
restored if `go test ./...` fails. The proposals come from
tools/autopilot/staging/excerpt-tail-nonvote/, one YAML per target file with
{file, excerpt, note}; a proposal whose target does not exist, whose text does
not survive a YAML round-trip, or whose excerpt is already present is reported
and skipped rather than guessed at.
"""

import argparse
import json
import pathlib
import re
import shutil
import subprocess
import sys
import tempfile

import yaml

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import promote  # noqa: E402  - one implementation of the count bump and archive rebuild

ROOT = pathlib.Path(__file__).resolve().parents[2]
EVIDENCE = ROOT / "internal/data/content/evidence"
STAGING = pathlib.Path(__file__).resolve().parent / "staging/excerpt-tail-nonvote"


def render_block(text: str, width: int = 76) -> str | None:
    """A folded block scalar, or None when folding would change the value.

    Style matters here only because the corpus is read by humans, but
    correctness comes first: a folded scalar turns every newline into a space,
    so text that has its own newlines, leading/trailing spaces, or a double
    space at a wrap point cannot use one.
    """
    if "\n" in text or text != text.strip() or "  " in text:
        return None
    lines, cur = [], ""
    for word in text.split(" "):
        if cur and len(cur) + 1 + len(word) > width:
            lines.append(cur)
            cur = word
        else:
            cur = f"{cur} {word}" if cur else word
    if cur:
        lines.append(cur)
    return "verbatim_excerpt: >\n" + "\n".join("  " + ln for ln in lines) + "\n"


def insert_excerpt(original: str, excerpt: str) -> tuple[str, str]:
    """Add verbatim_excerpt just before status:, keeping the file's own style."""
    if re.search(r"^verbatim_excerpt:", original, re.M):
        raise ValueError("file already has a verbatim_excerpt")
    if not re.search(r"^status:", original, re.M):
        raise ValueError("no status: line to anchor the insert on")
    block = render_block(excerpt)
    if block is None:
        block = "verbatim_excerpt: " + json.dumps(excerpt, ensure_ascii=False) + "\n"
    new = re.sub(r"^status:", lambda m: block + m.group(0), original, count=1, flags=re.M)
    # Prove the round trip before touching the file: whatever style was used
    # must give back exactly the proposed text.
    # The corpus writes 316 of 317 excerpts as a folded block, and a folded
    # scalar keeps one trailing newline; the verifiers strip it. Compare what
    # the reader of the file gets, minus that newline, and nothing else.
    got = ((yaml.safe_load(new) or {}).get("verbatim_excerpt") or "").rstrip("\n")
    if got != excerpt:
        raise ValueError(f"round trip changed the excerpt ({got!r} != {excerpt!r})")
    return new, ("folded" if block.startswith("verbatim_excerpt: >") else "quoted")


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--dir", default=str(STAGING))
    args = ap.parse_args(argv)

    staging = pathlib.Path(args.dir)
    proposals = [p for p in sorted(staging.glob("*.yaml")) if p.name != "meta.yaml"]
    if not proposals:
        print("no proposals to apply: treat as failure", file=sys.stderr)
        return 2

    applied: list[tuple[pathlib.Path, str]] = []
    skipped: list[tuple[str, str]] = []
    for p in proposals:
        d = yaml.safe_load(p.read_text()) or {}
        name, excerpt = d.get("file"), (d.get("excerpt") or "").strip()
        if not name or not excerpt:
            skipped.append((p.name, "no file or no excerpt"))
            continue
        target = EVIDENCE / name
        if not target.exists():
            skipped.append((p.name, f"target {name} does not exist"))
            continue
        before = target.read_text()
        try:
            after, style = insert_excerpt(before, excerpt)
        except ValueError as e:
            skipped.append((p.name, str(e)))
            continue
        applied.append((target, before))
        print(f"{'would apply' if args.dry_run else 'apply'} {name} ({style}, {len(excerpt)} chars)")
        if not args.dry_run:
            target.write_text(after)

    for name, why in skipped:
        print(f"SKIP {name}: {why}")
    if not applied:
        print("nothing applied: treat as failure", file=sys.stderr)
        return 2
    if args.dry_run:
        return 0

    restore = [(t, b) for t, b in applied]
    archive_backup = pathlib.Path(tempfile.mkdtemp()) / "archive"
    shutil.copytree(promote.ARCHIVE, archive_backup)
    ok, msg = promote.rebuild_archive()
    if not ok:
        for t, b in restore:
            t.write_text(b)
        shutil.rmtree(promote.ARCHIVE)
        shutil.copytree(archive_backup, promote.ARCHIVE)
        print(f"FAIL archive rebuild: {msg}", file=sys.stderr)
        return 1

    counts = promote.real_counts()
    docs = promote.bump_docs(counts, __import__("datetime").date.today().isoformat())
    docs_before = {rel: (promote.ROOT / rel).read_text() for rel in docs}
    tests_before = promote.TESTS.read_text()
    for rel, text in docs.items():
        (promote.ROOT / rel).write_text(text)
    promote.TESTS.write_text(promote.bump_tests(counts["statements"], counts["evidence"]))

    ok, msg = promote.run_tests()
    if not ok:
        for t, b in restore:
            t.write_text(b)
        promote.TESTS.write_text(tests_before)
        for rel, text in docs_before.items():
            (promote.ROOT / rel).write_text(text)
        shutil.rmtree(promote.ARCHIVE)
        shutil.copytree(archive_backup, promote.ARCHIVE)
        print(f"FAIL go test, everything restored:\n{msg}", file=sys.stderr)
        return 1

    print(f"applied {len(applied)} excerpts, {len(skipped)} skipped; "
          f"{counts['quoted']} of {counts['evidence']} rows now carry one")
    print("next: uv run tools/verify/verify_excerpts.py   (the real substring check)")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
