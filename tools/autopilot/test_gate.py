"""Checks on the autopilot gate itself, offline and dependency-light.

Run:  python3 tools/autopilot/test_gate.py     (also works under pytest)

The queue check is here because of a real miss: an edit to queue.yaml left
an orphaned closing quote, nothing parsed the file, and the breakage went
unnoticed until the gate started reading it. A file that names every unit of
work has to be validated by something.
"""

import pathlib
import sys
import tempfile

import yaml

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import gate  # noqa: E402


def test_queue_parses_and_every_item_is_whole():
    q = yaml.safe_load(gate.QUEUE.read_text(encoding="utf-8"))
    assert q and q.get("items"), "queue.yaml has no items"
    for it in q["items"]:
        assert it.get("id"), f"queue item without an id: {it}"
        # "running" belongs here: the ektevalgomat-autopilot skill's loop step 2
        # sets an item running while its worker writes, so a set without it
        # rejects the state the documented workflow creates. Same shape as the
        # gate's CONFIDENCE set rejecting a value the corpus already used: the
        # check was calibrated against one author's idea, not the artefacts.
        assert it.get("status") in {"pending", "running", "passed", "failed", "blocked"}, \
            f"{it['id']}: unexpected status {it.get('status')!r}"


def test_queue_statuses_lookup():
    st = gate.queue_statuses()
    assert st.get("program-rodt") == "passed", st.get("program-rodt")


def test_empty_item_is_not_a_failure():
    """A promoted or not-yet-started item has nothing to check. Counting it
    as failed is how a gate trains its readers to ignore red."""
    with tempfile.TemporaryDirectory() as d:
        item = pathlib.Path(d) / "sp-45"
        item.mkdir()
        r = gate.check_item(item, set(), set(), set())
    assert r["files"] == 0
    checked, failed, empty = gate.summarise([r], {"sp-45": "passed"})
    assert checked == 0
    assert failed == 0, "an empty item must not count as a failure"
    assert empty == [("sp-45", "already promoted")], empty


def test_pending_empty_item_says_so():
    with tempfile.TemporaryDirectory() as d:
        item = pathlib.Path(d) / "budget-bredband"
        item.mkdir()
        r = gate.check_item(item, set(), set(), set())
    _, failed, empty = gate.summarise([r], {"budget-bredband": "pending"})
    assert failed == 0
    assert empty == [("budget-bredband", "pending, nothing staged yet")], empty


def test_a_real_failure_is_still_counted():
    """The guard's teeth: a file that fails must count, and be counted once."""
    good = {"item": "a", "files": 1, "ok": True, "reports": []}
    bad = {"item": "b", "files": 1, "ok": False, "reports": []}
    checked, failed, empty = gate.summarise([good, bad], {})
    assert checked == 2
    assert failed == 1, f"expected exactly 1 failure, got {failed}"
    assert empty == []


def test_summary_counts_files_not_items():
    one = {"item": "a", "files": 3, "ok": True, "reports": []}
    checked, failed, empty = gate.summarise([one], {})
    assert checked == 3
    assert failed == 0 and empty == []


def test_empty_item_with_ok_false_is_still_not_a_failure():
    """The precise fix, and the mutation target. The old gate both returned
    ok=False for a zero-file item AND counted it again in a separate
    files==0 sum. This pins the condition that actually changed."""
    r = {"item": "leftover", "files": 0, "ok": False, "reports": []}
    checked, failed, empty = gate.summarise([r], {"leftover": "passed"})
    assert checked == 0
    assert failed == 0, "a zero-file item must not count as failed even when ok is False"
    assert empty == [("leftover", "already promoted")], empty

def main() -> int:
    tests = [v for k, v in sorted(globals().items()) if k.startswith("test_") and callable(v)]
    bad = 0
    for t in tests:
        try:
            t()
            print(f"ok   {t.__name__}")
        except AssertionError as e:  # noqa: PERF203
            bad += 1
            print(f"FAIL {t.__name__}: {e}")
    print(f"\n{len(tests) - bad}/{len(tests)} passed")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
