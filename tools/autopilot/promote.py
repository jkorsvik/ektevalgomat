# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Promote a gate-passing staging item into internal/data/content/.

Run only after tools/autopilot/gate.py says PASS for the item. This step is
deterministic and reversible: it assigns the next free evidence number, moves
the files, updates the count assertions in internal/data/loader_test.go to the
real new totals, runs `go test ./...`, and rolls everything back if the build
or the tests fail. Nothing an agent wrote reaches the repo without go test
having passed on it.

Usage:
  uv run tools/autopilot/promote.py <item-dir> [<item-dir> ...]
  uv run tools/autopilot/promote.py --dry-run <item-dir>

Writes the new status back into tools/autopilot/queue.yaml (passed / failed)
so the loop knows what is done.
"""

import datetime
import pathlib
import re
import shutil
import subprocess
import sys
import tempfile

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONTENT = ROOT / "internal/data/content"
EVIDENCE = CONTENT / "evidence"
PROGRAMS = CONTENT / "programs"
TESTS = ROOT / "internal/data/loader_test.go"
ARCHIVE = ROOT / "internal/data/archive"
ARCHIVE_TOOL = ROOT / "tools/archive/build_archive.py"
VOTES = ROOT / "internal/data/votes"
VOTE_TOOL = ROOT / "tools/analytics/build_vote_index.py"
QUEUE = pathlib.Path(__file__).resolve().parent / "queue.yaml"


def next_number() -> int:
    nums = [int(m.group(1)) for f in EVIDENCE.glob("*.yaml")
            if (m := re.match(r"(\d+)-", f.name))]
    return max(nums, default=0) + 1


def real_counts() -> dict[str, int]:
    """Everything the documents state, derived from the dataset.

    TestDocsStateTheRealCounts (internal/scoring/docs_claims_test.go) pins the
    README and the roadmap to these numbers. Promotion used to bump only
    internal/data/loader_test.go, so every evidence promotion since that test
    landed failed go test and rolled back: the tool that moves content could
    not move any. The counts are computed here the same way the Go test
    computes them, from the dataset, never from a stored copy.
    """
    statements = [p for p in CONTENT.glob("statements/*.yaml")]
    evidence = [p for p in EVIDENCE.glob("*.yaml")]
    parties = {yaml.safe_load(p.read_text())["id"] for p in (CONTENT / "parties").glob("*.yaml")}
    coded: dict[str, set[str]] = {}
    quoted = 0
    for f in evidence:
        d = yaml.safe_load(f.read_text()) or {}
        if (d.get("verbatim_excerpt") or "").strip():
            quoted += 1
        if d.get("position") is not None and d.get("party_id") and d.get("statement_id"):
            coded.setdefault(d["statement_id"], set()).add(d["party_id"])
    full = sum(1 for s in statements
               if len(coded.get(yaml.safe_load(s.read_text())["id"], set()) & parties) == len(parties))
    return {"statements": len(statements), "evidence": len(evidence),
            "quoted": quoted, "full": full}


# The claims the documents make, with the same patterns docs_claims_test.go
# uses, so the two cannot drift silently. test_promote.py fails if the Go test
# grows a claim this table does not carry.
DOC_CLAIMS: list[tuple[str, str, tuple[str, ...]]] = [
    ("README.md", r"Current state \(\d{4}-\d{2}-\d{2}\): (\d+) statements, (\d+) evidence\s+files",
     ("statements", "evidence")),
    ("README.md", r"(\d+) of (\d+) statements have a coded position", ("full", "statements")),
    ("README.md", r"(\d+) of\s+the (\d+) evidence rows carry such an excerpt", ("quoted", "evidence")),
    ("docs/roadmap.md", r"\*\*(\d+) statements, (\d+) evidence files\.\*\*", ("statements", "evidence")),
    ("docs/roadmap.md", r"\*\*(\d+) of (\d+) statements have a coded position", ("full", "statements")),
]

# "Current state (<date>)" is a date as well as a number: the state changed, so
# the date it is stated as is no longer true.
DOC_DATES = [
    ("README.md", r"Current state \(\d{4}-\d{2}-\d{2}\)", "Current state ({today})"),
    ("docs/roadmap.md", r"## Current state \(\d{4}-\d{2}-\d{2}\)", "## Current state ({today})"),
]


def bump_docs(counts: dict[str, int], today: str) -> dict[str, str]:
    """Return path -> new text, with every documented number set to the data."""
    out: dict[str, str] = {}
    for rel, pattern, keys in DOC_CLAIMS:
        path = ROOT / rel
        text = out.get(rel) or path.read_text()
        m = re.search(pattern, text)
        if not m:
            continue  # the Go test already fails loudly on a vanished claim
        # Replace by position, never by string search: two captured numbers can
        # share digits (60 and 615), and replace() would hit the wrong one.
        pieces, prev = [], 0
        for i, key in enumerate(keys, start=1):
            a, b = m.span(i)
            pieces.append(m.group(0)[prev:a - m.start()])
            pieces.append(str(counts[key]))
            prev = b - m.start()
        pieces.append(m.group(0)[prev:])
        out[rel] = text[:m.start()] + "".join(pieces) + text[m.end():]
    for rel, pattern, replacement in DOC_DATES:
        path = ROOT / rel
        text = out.get(rel) or path.read_text()
        out[rel] = re.sub(pattern, replacement.format(today=today), text, count=1)
    return out


def bump_tests(stmts: int, ev: int) -> str:
    s = TESTS.read_text()
    s = re.sub(r"(if got := len\(Store\.Statements\); got != )\d+", rf"\g<1>{stmts}", s)
    s = re.sub(r'("Statements = %d, want )\d+', rf"\g<1>{stmts}", s)
    s = re.sub(r"(if got := len\(Store\.Evidence\); got != )\d+", rf"\g<1>{ev}", s)
    s = re.sub(r'("Evidence = %d, want )\d+', rf"\g<1>{ev}", s)
    return s


def quoted_claim_added(paths: list[pathlib.Path]) -> bool:
    """True when a promoted file carries an excerpt, so the archive must grow.

    TestArchiveCoversEveryQuotedClaim (internal/data/archive_test.go) requires
    one stored window per quoted claim, keyed by the evidence `id`. Promotion
    that skips the rebuild therefore fails go test on any new quote; this is
    the third reader of the promotion contract that promote.py did not know
    about, after loader_test.go and the documented counts.
    """
    for p in paths:
        try:
            d = yaml.safe_load(p.read_text()) or {}
        except Exception:  # noqa: BLE001 - a broken file fails the gate, not here
            return True
        if (d.get("verbatim_excerpt") or "").strip():
            return True
    return False


def cites_votes(paths: list[pathlib.Path]) -> bool:
    """True when a promoted file cites a votering, so the vote index must grow.

    TestVoteIndexCoversEveryCitedVotering (internal/data/votes_test.go) requires
    the index to hold every votering the content cites. Promotion that skips the
    rebuild therefore fails go test on any new vote citation. This is the same
    contract as the archive: an index generated from the content must be
    regenerated by the step that changes the content, not by a later memory.
    """
    for p in paths:
        try:
            if "votering" in p.read_text().lower():
                return True
        except Exception:  # noqa: BLE001 - a broken file fails the gate, not here
            return True
    return False


def rebuild_votes() -> tuple[bool, str]:
    p = subprocess.run(["uv", "run", str(VOTE_TOOL)],
                       cwd=ROOT, capture_output=True, text=True,
                       env={**__import__("os").environ,
                            "PATH": "/opt/homebrew/bin:" + __import__("os").environ["PATH"]})
    if p.returncode != 0:
        return False, (p.stdout + p.stderr).strip()[-500:]
    return True, ""


def rebuild_archive() -> tuple[bool, str]:
    p = subprocess.run(["uv", "run", str(ARCHIVE_TOOL), "--out", str(ARCHIVE), "--quiet"],
                       cwd=ROOT, capture_output=True, text=True,
                       env={**__import__("os").environ,
                            "PATH": "/opt/homebrew/bin:" + __import__("os").environ["PATH"]})
    if p.returncode != 0:
        return False, (p.stdout + p.stderr).strip()[-500:]
    return True, ""


def set_status(item_id: str, status: str, error: str = "") -> None:
    q = yaml.safe_load(QUEUE.read_text())
    for it in q["items"]:
        if it["id"] == item_id:
            it["status"] = status
            if error:
                it["last_error"] = error
            elif "last_error" in it:
                del it["last_error"]
    QUEUE.write_text(yaml.safe_dump(q, allow_unicode=True, sort_keys=False, width=100))


def promote(item: pathlib.Path, dry: bool) -> tuple[bool, str]:
    meta = yaml.safe_load((item / "meta.yaml").read_text()) if (item / "meta.yaml").exists() else {}
    kind = meta.get("kind", "evidence")
    item_id = meta.get("id", item.name)
    yamls = [f for f in item.glob("*.yaml") if f.name != "meta.yaml"]
    if not yamls:
        return False, "no yaml to promote"

    if kind == "program":
        party = meta["party"]
        dest = PROGRAMS / f"{party}.yaml"
        if dry:
            return True, f"would write {dest.relative_to(ROOT)}"
        PROGRAMS.mkdir(parents=True, exist_ok=True)
        backup = dest.read_text() if dest.exists() else None
        shutil.copy(yamls[0], dest)
        ok, msg = run_tests()
        if not ok:
            if backup is None:
                dest.unlink(missing_ok=True)
            else:
                dest.write_text(backup)
            return False, msg
        return True, f"wrote {dest.relative_to(ROOT)}"

    # evidence: assign the next free numbers, keep each file's slug
    moved: list[tuple[pathlib.Path, pathlib.Path]] = []
    n = next_number()
    for f in sorted(yamls):
        slug = re.sub(r"^\d+-", "", f.name)
        dest = EVIDENCE / f"{n:03d}-{slug}"
        n += 1
        if dry:
            moved.append((f, dest))
            continue
        shutil.move(str(f), dest)
        moved.append((f, dest))
    if dry:
        return True, "would move " + ", ".join(d.name for _, d in moved)

    # The archive is generated from the dataset, so it is restored, not merged:
    # a half-rebuilt archive would look complete to the coverage test's count.
    archive_backup = None
    if quoted_claim_added([d for _, d in moved]):
        archive_backup = pathlib.Path(tempfile.mkdtemp()) / "archive"
        shutil.copytree(ARCHIVE, archive_backup)
        ok, msg = rebuild_archive()
        if not ok:
            for src, dest in moved:
                shutil.move(str(dest), str(src))
            return False, f"archive rebuild failed: {msg}"

    # The vote index is generated from the content too: if the new rows cite a
    # votering the index does not hold, every later test run fails on it.
    votes_backup = None
    if cites_votes([d for _, d in moved]):
        votes_backup = pathlib.Path(tempfile.mkdtemp()) / "votes"
        if VOTES.exists():
            shutil.copytree(VOTES, votes_backup)
        ok, msg = rebuild_votes()
        if not ok:
            for src, dest in moved:
                shutil.move(str(dest), str(src))
            if archive_backup is not None:
                shutil.rmtree(ARCHIVE)
                shutil.copytree(archive_backup, ARCHIVE)
            return False, f"vote index rebuild failed: {msg}"

    tests_before = TESTS.read_text()
    counts = real_counts()
    docs = bump_docs(counts, datetime.date.today().isoformat())
    docs_before = {rel: (ROOT / rel).read_text() for rel in docs}
    TESTS.write_text(bump_tests(counts["statements"], counts["evidence"]))
    for rel, text in docs.items():
        (ROOT / rel).write_text(text)
    ok, msg = run_tests()
    if not ok:
        for src, dest in moved:
            shutil.move(str(dest), str(src))
        TESTS.write_text(tests_before)
        for rel, text in docs_before.items():
            (ROOT / rel).write_text(text)
        if archive_backup is not None:
            shutil.rmtree(ARCHIVE)
            shutil.copytree(archive_backup, ARCHIVE)
        if votes_backup is not None and votes_backup.exists():
            shutil.rmtree(VOTES)
            shutil.copytree(votes_backup, VOTES)
        return False, msg
    return True, f"promoted {len(moved)} files ({counts['evidence']} evidence now)"


def run_tests() -> tuple[bool, str]:
    p = subprocess.run(["go", "test", "./..."], cwd=ROOT, capture_output=True, text=True,
                       env={**__import__("os").environ, "PATH": "/opt/homebrew/bin:" + __import__("os").environ["PATH"]})
    lines = p.stdout.strip().splitlines() + p.stderr.strip().splitlines()
    if p.returncode == 0:
        return True, "\n".join(lines[-4:])
    # A failing promotion used to return the last four lines, which for a package
    # failure is "ok ... / ok ... / FAIL": the name of the test that failed was
    # thrown away, and the driver had to rebuild the promotion by hand to read it.
    # Keep the lines that name a failure, bounded, and fall back to the tail.
    interesting = [ln for ln in lines
                   if ln.startswith(("--- FAIL", "FAIL", "panic:", "# "))
                   or ".go:" in ln or "Error" in ln or "error:" in ln]
    if interesting:
        return False, "\n".join(interesting[:40])
    return False, "\n".join(lines[-10:])


def main(argv: list[str]) -> int:
    dry = "--dry-run" in argv
    items = [pathlib.Path(a) for a in argv if not a.startswith("--")]
    if not items:
        print(__doc__)
        return 2
    fails = 0
    for item in items:
        ok, msg = promote(item, dry)
        print(f"{'PASS' if ok else 'FAIL'} {item.name}: {msg}")
        if not dry:
            set_status(item.name, "passed" if ok else "failed", "" if ok else msg)
        fails += not ok
    return 1 if fails else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
