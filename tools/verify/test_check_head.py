"""The pre-push head check must exist, be wired, and be able to fail.

On 2026-11-12 a commit shipped nine templates calling {{explore}} without the
line that registers the function. The working tree built and its tests passed;
the commit panicked at template parse, and the deploy host kept the previous
container serving for 19 minutes. tools/verify/check_head.sh closes that hole by
building the committed tree from an archive.

A gate that is not installed does not gate. These tests keep the script, the
hook and the wiring from being deleted or emptied without anyone noticing, and
they check that the script fails on a broken ref rather than only passing on a
good one.
"""

import os
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "tools" / "verify" / "check_head.sh"
HOOK = ROOT / ".githooks" / "pre-push"


def test_the_check_script_exists_and_is_executable():
    assert SCRIPT.is_file(), "tools/verify/check_head.sh is missing"
    assert os.access(SCRIPT, os.X_OK), "tools/verify/check_head.sh is not executable"


def test_the_script_builds_the_committed_tree_not_the_working_tree():
    body = SCRIPT.read_text()
    assert "git -C \"$ROOT\" archive" in body, (
        "the check must build HEAD from an archive; building the working tree is "
        "what failed to catch the missing registration"
    )
    assert "go test ./..." in body, "the check must run the tests"


def test_the_script_treats_an_empty_test_run_as_failure():
    body = SCRIPT.read_text()
    assert 'PKGS" -eq 0' in body, (
        "a run that reports no package ok is a broken invocation, not a green tree"
    )


def test_the_pre_push_hook_runs_the_check():
    assert HOOK.is_file(), ".githooks/pre-push is missing"
    assert os.access(HOOK, os.X_OK), ".githooks/pre-push is not executable"
    body = HOOK.read_text()
    assert "tools/verify/check_head.sh" in body, "the hook does not call the check"
    assert "EKV_SKIP_HEAD_CHECK" in body, (
        "there must be one named escape hatch, so a skip is visible in the log"
    )


def test_the_hook_is_wired_through_a_committed_hooks_path():
    """core.hooksPath is machine-local, so the repo documents the one command."""
    contrib = (ROOT / "CONTRIBUTING.md").read_text()
    assert "core.hooksPath .githooks" in contrib, (
        "CONTRIBUTING.md must tell a fresh clone how to install the hook"
    )


def test_the_check_fails_on_a_ref_that_does_not_build():
    """The gate must be able to fail. Build a ref with the registration removed."""
    head = subprocess.run(
        ["git", "-C", str(ROOT), "rev-parse", "HEAD"],
        capture_output=True, text=True, check=True,
    ).stdout.strip()

    # A tree that is known bad: 5caf73d, the commit that shipped without the
    # explore registration. It stays reachable in this repository's history.
    bad = "5caf73d"
    exists = subprocess.run(
        ["git", "-C", str(ROOT), "cat-file", "-e", f"{bad}^{{commit}}"],
        capture_output=True,
    )
    if exists.returncode != 0:
        import pytest

        pytest.skip(f"{bad} is not in this clone; cannot prove the gate can fail")

    result = subprocess.run(
        [str(SCRIPT), bad], capture_output=True, text=True, timeout=600,
    )
    assert result.returncode != 0, (
        f"check_head.sh passed on {bad}, which panics at template parse"
    )
    assert "FAIL" in result.stdout + result.stderr

    # And it passes on the commit we are standing on.
    good = subprocess.run(
        [str(SCRIPT), head], capture_output=True, text=True, timeout=600,
    )
    assert good.returncode == 0, good.stdout + good.stderr
