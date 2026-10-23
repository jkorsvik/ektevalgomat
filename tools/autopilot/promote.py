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

import pathlib
import re
import shutil
import subprocess
import sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONTENT = ROOT / "internal/data/content"
EVIDENCE = CONTENT / "evidence"
PROGRAMS = CONTENT / "programs"
TESTS = ROOT / "internal/data/loader_test.go"
QUEUE = pathlib.Path(__file__).resolve().parent / "queue.yaml"


def next_number() -> int:
    nums = [int(m.group(1)) for f in EVIDENCE.glob("*.yaml")
            if (m := re.match(r"(\d+)-", f.name))]
    return max(nums, default=0) + 1


def real_counts() -> tuple[int, int]:
    return len(list(CONTENT.glob("statements/*.yaml"))), len(list(EVIDENCE.glob("*.yaml")))


def bump_tests(stmts: int, ev: int) -> str:
    s = TESTS.read_text()
    s = re.sub(r"(if got := len\(Store\.Statements\); got != )\d+", rf"\g<1>{stmts}", s)
    s = re.sub(r'("Statements = %d, want )\d+', rf"\g<1>{stmts}", s)
    s = re.sub(r"(if got := len\(Store\.Evidence\); got != )\d+", rf"\g<1>{ev}", s)
    s = re.sub(r'("Evidence = %d, want )\d+', rf"\g<1>{ev}", s)
    return s


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

    tests_before = TESTS.read_text()
    stmts, ev = real_counts()
    TESTS.write_text(bump_tests(stmts, ev))
    ok, msg = run_tests()
    if not ok:
        for src, dest in moved:
            shutil.move(str(dest), str(src))
        TESTS.write_text(tests_before)
        return False, msg
    return True, f"promoted {len(moved)} files ({ev} evidence now)"


def run_tests() -> tuple[bool, str]:
    p = subprocess.run(["go", "test", "./..."], cwd=ROOT, capture_output=True, text=True,
                       env={**__import__("os").environ, "PATH": "/opt/homebrew/bin:" + __import__("os").environ["PATH"]})
    tail = "\n".join(p.stdout.strip().splitlines()[-4:] + p.stderr.strip().splitlines()[-4:])
    return p.returncode == 0, tail


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
