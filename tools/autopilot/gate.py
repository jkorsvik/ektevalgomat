# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml", "pymupdf", "rapidfuzz"]
# ///
"""The autopilot gate: the one thing that decides whether staged work ships.

An agent writes evidence into tools/autopilot/staging/<item>/. This gate reads
it and decides PASS or FAIL, mechanically. Only PASS items are promoted into
internal/data/content/ by promote.py. Nothing an agent says about its own work
is read here — only the files it wrote, checked against the live sources.

Three checks, same discipline as tools/verify/ (imported, not re-implemented):

  1. SCHEMA   every field the loader requires, every enum, and the two rules
              the loader itself enforces (unknown statement_id / party_id /
              evidence_class / source_tier / cap fail the build). Plus two
              rules the loader does not: a claim whose status says the party
              voted must name a vote id, and a verbatim_excerpt must have a
              source_url.
  2. EXCERPT  every verbatim_excerpt must be a substring of its source_url,
              fetched now (verify_excerpts.fetch/to_text/norm/fragments).
  3. VOTE     every status of the form stemte_for/stemte_mot must match the
              official record on data.stortinget.no (verify_votes.*).

Usage:
  uv run tools/autopilot/gate.py                 # every item in staging/
  uv run tools/autopilot/gate.py <item> [<item>] # one or more item dirs
  uv run tools/autopilot/gate.py --json          # machine summary to stdout

Exit 0 if every checked item passed, 1 if any failed, 2 if nothing was checked
(a run that checks nothing must never look like success).
"""

import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "verify"))

import yaml  # noqa: E402

from verify_excerpts import fetch, fragments, norm, to_text  # noqa: E402
from verify_votes import PARTY, fetch_votes, party_side, rejects_proposal, vote_ids  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parents[2]
STAGING = pathlib.Path(__file__).resolve().parent / "staging"
QUEUE = pathlib.Path(__file__).resolve().parent / "queue.yaml"
CONTENT = ROOT / "internal/data/content"

REQUIRED = ("id", "statement_id", "party_id", "actor", "date", "evidence_type",
            "evidence_class", "scan_summary", "claim_action", "status",
            "source_tier", "source_name", "source_url", "claim_confidence",
            "context_note", "retrieved_at")
# party_id must be present but MAY be empty: institutional evidence ("what the
# state did", a vedtak the government made) is a real category, 70 files use
# party_id: "". Requiring it non-empty would reject valid work the build accepts.
PRESENT_ONLY = {"party_id"}
CLASSES = {"fakta", "partiets_begrunnelse", "analyse", "uklart"}
TIERS = {"A", "B", "C", "D", "E"}
CAPS = {"", "svak", "moderat"}
# The vocabulary must cover what the build already accepts: internal/data does
# not validate claim_confidence at all, so this set is the only enforcement
# point, and it has to admit every value the corpus uses. It did not:
# "sannsynliggjort" is used by 7 committed evidence files (including 464 and
# 465, the two siblings on utsagn-45) while "sannsynlig" was used by none, so
# the gate rejected files the loader accepts. Same failure shape as a gate
# calibrated against its author's idea of the format rather than against the
# artefacts the consumer has already taken.
CONFIDENCE = {"dokumentert", "sannsynlig", "sannsynliggjort", "usikkert", "ukjent"}


def known_ids() -> tuple[set[str], set[str]]:
    stmts = {yaml.safe_load(p.read_text())["id"] for p in (CONTENT / "statements").glob("*.yaml")}
    parties = {yaml.safe_load(p.read_text())["id"] for p in (CONTENT / "parties").glob("*.yaml")}
    return stmts, parties


def known_dimensions() -> set[str]:
    return {yaml.safe_load(p.read_text())["dimension"] for p in (CONTENT / "statements").glob("*.yaml")}


def program_errors(d: dict, parties: set[str], dims: set[str], stmts: set[str]) -> list[str]:
    """Programme items have their own schema (a Program, not an Evidence): a
    source, and a list of verbatim programme points. Checked the same way —
    every quote must be a substring of the source, fetched now."""
    errs = []
    if not str(d.get("id") or "").strip():
        errs.append("missing id (party_id)")
    elif d["id"] not in parties:
        errs.append(f"unknown party_id {d['id']!r}")
    if not str(d.get("source_url") or "").strip():
        errs.append("missing source_url")
    if not str(d.get("source_name") or "").strip():
        errs.append("missing source_name")
    pts = d.get("points") or []
    if not isinstance(pts, list) or not pts:
        errs.append("points is empty or not a list")
        return errs
    for i, pt in enumerate(pts):
        q = str((pt or {}).get("quote") or "").strip()
        if len(q) < 12:
            errs.append(f"point {i}: quote missing or shorter than 12 chars")
        dim = str((pt or {}).get("dimension") or "").strip()
        if dim and dim not in dims:
            errs.append(f"point {i}: unknown dimension {dim!r}")
        st = str((pt or {}).get("statement") or "").strip()
        if st and st not in stmts:
            errs.append(f"point {i}: unknown statement {st!r}")
    return errs


def program_excerpt_errors(d: dict) -> list[str]:
    url = d.get("source_url") or ""
    try:
        text = norm(to_text(fetch(url)))
    except Exception as e:  # noqa: BLE001 - report, never hide
        return [f"FETCH {url}: {type(e).__name__}: {e}"]
    errs = []
    for i, pt in enumerate(d.get("points") or []):
        q = str((pt or {}).get("quote") or "").strip()
        if not q:
            continue
        frags = fragments(q)
        if not frags:
            errs.append(f"point {i}: quote has no fragment >= 12 chars")
            continue
        pos = 0
        for fr in frags:
            j = text.find(fr, pos)
            if j < 0:
                from rapidfuzz import fuzz
                al = fuzz.partial_ratio_alignment(fr, text)
                kind = "EDITED" if al.score >= 85 else "MISS"
                errs.append(f"point {i}: {kind} score={al.score:.0f} quote={fr[:110]!r}")
                break
            pos = j + len(fr)
    return errs


def schema_errors(d: dict, stmts: set[str], parties: set[str]) -> list[str]:
    errs = [f"missing field {f}" for f in REQUIRED if f not in d]
    errs += [f"empty field {f}" for f in REQUIRED
             if f not in PRESENT_ONLY and not str(d.get(f, "")).strip()]
    if d.get("statement_id") and d["statement_id"] not in stmts:
        errs.append(f"unknown statement_id {d['statement_id']!r}")
    if d.get("party_id") and d["party_id"] not in parties:
        errs.append(f"unknown party_id {d['party_id']!r}")
    if d.get("evidence_class") and d["evidence_class"] not in CLASSES:
        errs.append(f"bad evidence_class {d['evidence_class']!r}")
    if d.get("source_tier") and d["source_tier"] not in TIERS:
        errs.append(f"bad source_tier {d['source_tier']!r}")
    if d.get("evidensstyrke_cap", "") not in CAPS:
        errs.append(f"bad evidensstyrke_cap {d['evidensstyrke_cap']!r}")
    if d.get("claim_confidence") and d["claim_confidence"] not in CONFIDENCE:
        errs.append(f"bad claim_confidence {d['claim_confidence']!r}")
    pos = d.get("position")
    if pos is not None and (not isinstance(pos, int) or pos not in (-2, -1, 0, 1, 2)):
        errs.append(f"position must be -2..2 or absent, got {pos!r}")
    st = str(d.get("status") or "")
    if ("stemte_for" in st or "stemte_mot" in st):
        if not d.get("party_id"):
            errs.append("status says a party voted, but party_id is empty")
        if not (vote_ids(f"{d.get('source_name','')} {d.get('source_url','')}")
                or vote_ids(str(d.get("context_note", "")))):
            errs.append("status says a party voted, but no vote id is cited")
    if d.get("verbatim_excerpt") and not d.get("source_url"):
        errs.append("verbatim_excerpt without source_url")
    return errs


def excerpt_error(d: dict) -> str | None:
    ex = (d.get("verbatim_excerpt") or "").strip()
    url = d.get("source_url") or ""
    if not ex:
        return None
    try:
        text = norm(to_text(fetch(url)))
    except Exception as e:  # noqa: BLE001 - report, never hide
        return f"FETCH {url}: {type(e).__name__}: {e}"
    frags = fragments(ex)
    if not frags:
        return "excerpt has no fragment >= 12 chars"
    pos = 0
    for fr in frags:
        i = text.find(fr, pos)
        if i < 0:
            from rapidfuzz import fuzz
            al = fuzz.partial_ratio_alignment(fr, text)
            kind = "EDITED" if al.score >= 85 else "MISS"
            return f"{kind} score={al.score:.0f} excerpt={fr[:120]!r}"
        pos = i + len(fr)
    return None


def vote_error(d: dict) -> str | None:
    st, pid = str(d.get("status") or ""), d.get("party_id") or ""
    import re
    m = re.search(r"stemte_(for|mot)$", st)
    if not m or pid not in PARTY:
        return None
    want = m.group(1)
    ids = vote_ids(f"{d.get('source_name','')} {d.get('source_url','')}") or \
        vote_ids(str(d.get("context_note", "")))
    if not ids:
        return "no vote id cited"
    for vid in ids:
        try:
            side, n = party_side(fetch_votes(vid), PARTY[pid])
            if side in ("for", "mot") and rejects_proposal(vid):
                side = {"for": "mot", "mot": "for"}[side]
        except Exception as e:  # noqa: BLE001
            return f"vote {vid}: FETCH {type(e).__name__}: {e}"
        if side == want:
            return None
    return f"party {pid} is not {want} on any cited vote {ids}"


def check_item(item: pathlib.Path, stmts: set[str], parties: set[str], dims: set[str]) -> dict:
    meta = {}
    if (item / "meta.yaml").exists():
        meta = yaml.safe_load((item / "meta.yaml").read_text()) or {}
    kind = meta.get("kind", "evidence")
    reports = []
    for f in sorted(item.glob("*.yaml")):
        if f.name in ("meta.yaml", "item.yaml"):
            continue
        try:
            d = yaml.safe_load(f.read_text(encoding="utf-8")) or {}
        except yaml.YAMLError as e:
            reports.append({"file": f.name, "ok": False, "errors": [f"YAML: {e}"]})
            continue
        if kind == "program":
            errs = program_errors(d, parties, dims, stmts)
            if not errs:
                errs = program_excerpt_errors(d)
        else:
            errs = schema_errors(d, stmts, parties)
            for label, fn in (("excerpt", excerpt_error), ("vote", vote_error)):
                if not errs:
                    e = fn(d)
                    if e:
                        errs.append(f"{label}: {e}")
        reports.append({"file": f.name, "ok": not errs, "errors": errs,
                        "id": d.get("id"), "statement_id": d.get("statement_id"),
                        "party_id": d.get("party_id")})
    return {"item": item.name, "kind": kind, "ok": all(r["ok"] for r in reports),
            "files": len(reports), "reports": reports}


def queue_statuses() -> dict[str, str]:
    """id -> status from queue.yaml, so an empty staging dir can say WHY.

    A queue.yaml that does not parse is a hard error, not a shrug: it is
    repo state that names every unit of work, and a malformed one silently
    stopped being checked by anything until this function started reading
    it (2026-10-13). Fails loudly rather than returning an empty map, which
    would look like "no items are queued".
    """
    try:
        text = QUEUE.read_text(encoding="utf-8")
    except OSError as e:
        raise SystemExit(f"queue.yaml unreadable: {e}") from e
    try:
        q = yaml.safe_load(text) or {}
    except yaml.YAMLError as e:
        raise SystemExit(f"queue.yaml is not valid YAML: {e}") from e
    return {str(it.get("id")): str(it.get("status", "")) for it in q.get("items", [])}


def summarise(results: list[dict], statuses: dict[str, str]) -> tuple[int, int, list[tuple[str, str]]]:
    """Count what the run actually checked.

    An item with no staged files is EMPTY, not failed. promote.py moves a
    passed item's files out of staging, and an item nobody has started yet
    has nothing to check either. Counting those as failures is how a gate
    teaches its readers to ignore red, which is worse than a gate that is
    merely wrong once: the next real failure gets skimmed past too.

    The empties are named instead, with the queue's own status, and the
    summary always states how many there were. A run where nothing at all
    was checked still exits 2, so "green" is never silent about an empty
    estate.
    """
    checked = sum(r["files"] for r in results)
    failed = sum(1 for r in results if r["files"] and not r["ok"])
    empty: list[tuple[str, str]] = []
    for r in results:
        if r["files"] != 0:
            continue
        st = statuses.get(r["item"], "")
        why = {
            "passed": "already promoted",
            "pending": "pending, nothing staged yet",
        }.get(st, st or "not a queue item (leftover or ad-hoc staging)")
        empty.append((r["item"], why))
    return checked, failed, empty


def main(argv: list[str]) -> int:
    as_json = "--json" in argv
    args = [a for a in argv if not a.startswith("--")]
    items = [pathlib.Path(a) for a in args] or sorted(p for p in STAGING.iterdir() if p.is_dir()) \
        if STAGING.exists() else []
    if not items:
        print("no staged items to gate: treat as failure")
        return 2
    stmts, parties = known_ids()
    dims = known_dimensions()
    results = [check_item(i, stmts, parties, dims) for i in items]
    checked, failed, empty = summarise(results, queue_statuses())
    if as_json:
        print(json.dumps({"checked": checked, "failed": failed,
                          "empty": [{"item": i, "why": w} for i, w in empty],
                          "results": results}, ensure_ascii=False, indent=2))
    else:
        for r in results:
            if r["files"] == 0:
                continue
            mark = "PASS" if r["ok"] else "FAIL"
            print(f"{mark} {r['item']}  ({r['files']} files)")
            for fr in r["reports"]:
                if not fr["ok"]:
                    for e in fr["errors"]:
                        print(f"       {fr['file']}: {e}")
        for name, why in empty:
            print(f"TOM  {name}  (0 files: {why})")
    print(f"\nchecked {checked} files in {len(results)} items, {failed} failed, {len(empty)} empty")
    if checked == 0:
        print("nothing checked: treat as failure")
        return 2
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
