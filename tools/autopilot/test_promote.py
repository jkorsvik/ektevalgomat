"""Checks on promote.py, offline.

Run:  python3 tools/autopilot/test_promote.py     (also works under pytest)

Why the first test exists: promote.py was written before
TestDocsStateTheRealCounts landed. The Go test pins the README and the roadmap
to the dataset's real totals; promote.py bumped only loader_test.go. So every
evidence promotion failed go test and rolled back: the tool that moves content
could not move any, and nothing said so because a rollback looks like a clean
tree. The fix is only durable if the two lists of claims cannot drift apart, so
this test reads the Go test's own claim table and refuses a promote.py that
does not carry every entry.
"""

import pathlib
import re
import sys
import tempfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import promote  # noqa: E402

GO_TEST = pathlib.Path(__file__).resolve().parents[2] / "internal/scoring/docs_claims_test.go"


def go_claim_patterns() -> list[str]:
    """Every (file, pattern) pair the Go test asserts on, as written there."""
    src = GO_TEST.read_text()
    block = src[src.index("claims := []claim{"):src.index("}\n\n\troot := filepath.Join")]
    return re.findall(r'"([^"]+)",\s*`([^`]+)`', block)


def test_every_go_doc_claim_is_known_to_promote():
    pairs = go_claim_patterns()
    assert pairs, "parsed no claims out of docs_claims_test.go: the parse is broken, not the list"
    known = {(rel, pat) for rel, pat, _ in promote.DOC_CLAIMS}
    missing = [p for p in pairs if p not in known]
    assert not missing, (
        "docs_claims_test.go asserts claims promote.py cannot rewrite, so promotion will roll back: "
        + ", ".join(f"{rel}:{pat}" for rel, pat in missing)
    )


def test_bump_docs_sets_every_number():
    with tempfile.TemporaryDirectory() as d:
        root = pathlib.Path(d)
        (root / "docs").mkdir()
        (root / "README.md").write_text(
            "Current state (2026-10-24): 60 statements, 615 evidence\n"
            "files. 54 of 60 statements have a coded position. 316 of\n"
            "the 615 evidence rows carry such an excerpt.\n"
        )
        (root / "docs/roadmap.md").write_text(
            "## Current state (2026-10-03)\n\n- **60 statements, 615 evidence files.**\n"
            "- **54 of 60 statements have a coded position.**\n"
        )
        old = promote.ROOT
        promote.ROOT = root
        try:
            out = promote.bump_docs({"statements": 61, "evidence": 700, "quoted": 320, "full": 55},
                                    "2026-11-01")
        finally:
            promote.ROOT = old
        readme, roadmap = out["README.md"], out["docs/roadmap.md"]
        assert "61 statements, 700 evidence" in readme, readme
        assert "55 of 61 statements have a coded position" in readme, readme
        assert re.search(r"320 of\s+the 700 evidence rows carry such an excerpt", readme), readme
        assert "Current state (2026-11-01)" in readme, readme
        assert "**61 statements, 700 evidence files.**" in roadmap, roadmap
        assert "**55 of 61 statements have a coded position.**" in roadmap, roadmap
        assert "## Current state (2026-11-01)" in roadmap, roadmap


def test_bump_docs_is_idempotent():
    """Promoting twice must not shift a number twice: a second run sees the
    already-correct text and leaves it alone."""
    with tempfile.TemporaryDirectory() as d:
        root = pathlib.Path(d)
        (root / "docs").mkdir()
        (root / "README.md").write_text("Current state (2026-10-24): 60 statements, 615 evidence\nfiles.\n")
        (root / "docs/roadmap.md").write_text("## Current state (2026-10-03)\n\n- **60 statements, 615 evidence files.**\n")
        counts = {"statements": 60, "evidence": 615, "quoted": 316, "full": 54}
        old = promote.ROOT
        promote.ROOT = root
        try:
            first = promote.bump_docs(counts, "2026-11-01")
            for rel, text in first.items():
                (root / rel).write_text(text)
            second = promote.bump_docs(counts, "2026-11-01")
        finally:
            promote.ROOT = old
        for rel in first:
            assert (root / rel).read_text() == second[rel], f"{rel} changed on a second identical run"


def test_real_counts_agrees_with_the_dataset():
    counts = promote.real_counts()
    files = list((promote.EVIDENCE).glob("*.yaml"))
    assert counts["evidence"] == len(files), (counts, len(files))
    assert counts["statements"] == len(list((promote.CONTENT / "statements").glob("*.yaml")))
    assert 0 < counts["quoted"] <= counts["evidence"], counts
    assert 0 < counts["full"] <= counts["statements"], counts


if __name__ == "__main__":
    fails = 0
    for name, fn in sorted(globals().items()):
        if name.startswith("test_") and callable(fn):
            try:
                fn()
                print(f"ok   {name}")
            except AssertionError as e:
                fails += 1
                print(f"FAIL {name}: {e}")
    print(f"\n{len([n for n in globals() if n.startswith('test_')]) - fails}/"
          f"{len([n for n in globals() if n.startswith('test_')])} passed")
    sys.exit(1 if fails else 0)
