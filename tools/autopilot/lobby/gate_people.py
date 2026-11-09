# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Mechanical gate for lane L person records.

The evidence gate (`gate.py`) re-fetches every quote and every vote. This gate
does the same job for person data, with the added job that the data is about a
named individual. It enforces the scope rule, the forbidden fields, and a
public source with a date for every role.

    uv run tools/autopilot/lobby/gate_people.py tools/autopilot/lobby/staging/<id>
    uv run tools/autopilot/lobby/gate_people.py <dir> --no-fetch   # offline

PASS means the record is in scope, has no forbidden field, and every role has a
source and a date. It does not mean the role is TRUE; `ekv-person-verifier`
re-reads the source for that. Gate and verifier fail differently.

Exit: 0 all files pass, 1 any file fails, 2 no record found.
"""

from __future__ import annotations

import argparse
import datetime
import pathlib
import re
import sys
import urllib.error
import urllib.request

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[3]
SCOPE = pathlib.Path(__file__).resolve().parent / "scope.yaml"
VIEWMODEL = ROOT / "internal/server/viewmodel.go"
COMPANY_ROLES = {"styrerverv", "daglig_leder"}
YEAR = re.compile(r"^(\d{4})(?:-\d{2}-\d{2})?$")
ORGRNR = re.compile(r"^\d{9}$")


def snapshot_date() -> str:
    try:
        text = VIEWMODEL.read_text()
    except OSError:
        return ""
    m = re.search(r'SnapshotDate\s*=\s*"([^"]+)"', text)
    return m.group(1) if m else ""


def collect(path: pathlib.Path) -> list[pathlib.Path]:
    if path.is_file():
        return [path]
    out = []
    for p in sorted(path.rglob("*.yaml")):
        try:
            d = yaml.safe_load(p.read_text())
        except yaml.YAMLError:
            out.append(p)  # let validation report the parse error
            continue
        if isinstance(d, dict) and d.get("roles") is not None:
            out.append(p)
    return out


def walk_keys(node, found: set[str]) -> None:
    if isinstance(node, dict):
        for k, v in node.items():
            found.add(str(k).lower())
            walk_keys(v, found)
    elif isinstance(node, list):
        for v in node:
            walk_keys(v, found)


def fetchable(url: str) -> tuple[bool, str]:
    if not url.startswith(("http://", "https://")):
        return False, "not an http(s) url"
    req = urllib.request.Request(url, headers={"User-Agent": "ektevalgomat-gate-people/1"})
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return 200 <= r.status < 400, f"http {r.status}"
    except urllib.error.HTTPError as e:
        return False, f"http {e.code}"
    except Exception as e:  # noqa: BLE001 - any fetch failure is a FAIL, named
        return False, f"{type(e).__name__}: {e}"[:80]


def validate(path: pathlib.Path, scope: dict, snap: str, do_fetch: bool) -> list[str]:
    errs: list[str] = []
    try:
        d = yaml.safe_load(path.read_text())
    except yaml.YAMLError as e:
        return [f"yaml parse: {e}"]
    if not isinstance(d, dict):
        return ["not a mapping"]

    allowed_scope = set(scope.get("in_scope") or [])
    forbidden = {str(k).lower() for k in scope.get("forbidden_fields") or []}
    required = list(scope.get("required_for_each_role") or [])
    tiers = set(scope.get("source_tiers") or {})

    pid = d.get("id")
    if not pid or not isinstance(pid, str):
        errs.append("missing string id")
    if d.get("scope") not in allowed_scope:
        errs.append(f"scope {d.get('scope')!r} not in {sorted(allowed_scope)}")
    if d.get("consent_basis") != "public_office":
        errs.append("consent_basis must be public_office")

    keys: set[str] = set()
    walk_keys(d, keys)
    bad = sorted(keys & forbidden)
    if bad:
        errs.append(f"forbidden field(s): {', '.join(bad)}")

    roles = d.get("roles")
    if not isinstance(roles, list) or not roles:
        errs.append("roles must be a non-empty list")
        return errs

    for i, r in enumerate(roles):
        where = f"roles[{i}]"
        if not isinstance(r, dict):
            errs.append(f"{where}: not a mapping")
            continue
        for k in required:
            if r.get(k) in (None, ""):
                errs.append(f"{where}: missing {k}")
        tier = r.get("source_tier")
        if tier and tier not in tiers:
            errs.append(f"{where}: source_tier {tier!r} not in {sorted(tiers)}")
        frm = str(r.get("from") or "")
        if frm and not YEAR.match(frm):
            errs.append(f"{where}: from {frm!r} is not YYYY or YYYY-MM-DD")
        to = str(r.get("to") or "")
        if to:
            if not YEAR.match(to):
                errs.append(f"{where}: to {to!r} is not YYYY or YYYY-MM-DD")
            elif frm and YEAR.match(frm) and to[:4] < frm[:4]:
                errs.append(f"{where}: to {to} is before from {frm}")
        if r.get("role") in COMPANY_ROLES:
            if not ORGRNR.match(str(r.get("orgrnr") or "")):
                errs.append(f"{where}: company role needs a 9-digit orgnr")
        if snap:
            for field in ("from", "to"):
                v = str(r.get(field) or "")
                if v[:4].isdigit() and v[:4] > snap[:4]:
                    errs.append(f"{where}: {field} {v} is after SnapshotDate {snap}")
        if do_fetch:
            ok, why = fetchable(str(r.get("source_url") or ""))
            if not ok:
                errs.append(f"{where}: source_url not fetchable ({why})")
    return errs


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("path", help="a person.yaml, or a directory to scan")
    ap.add_argument("--no-fetch", action="store_true", help="skip the source_url fetch (offline)")
    args = ap.parse_args(argv)

    target = pathlib.Path(args.path)
    if not target.exists():
        print(f"no such path: {target}", file=sys.stderr)
        return 2
    scope = yaml.safe_load(SCOPE.read_text())
    snap = snapshot_date()

    files = collect(target)
    if not files:
        print(f"no person record found under {target}", file=sys.stderr)
        return 2

    failed = 0
    for p in files:
        errs = validate(p, scope, snap, not args.no_fetch)
        mark = "PASS" if not errs else "FAIL"
        print(f"{mark} {p}")
        for e in errs:
            print(f"   - {e}")
        failed += 1 if errs else 0
    print(f"\n{len(files)} file(s), {len(files) - failed} PASS, {failed} FAIL "
          f"(snapshot {snap or 'unknown'}, fetch {'off' if args.no_fetch else 'on'})")
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
