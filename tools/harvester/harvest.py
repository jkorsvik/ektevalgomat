#!/usr/bin/env python3
"""Offline lead-finder against data.stortinget.no for ektevalgomat.

Implements docs/roadmap.md's "Next candidate work" item 1: query
Stortinget's own JSON export API directly (not the JS-rendered web
search UI, which the 2026-09-23 research pass hit real limits with)
across sessions, and surface candidate cases per open statement topic
(see topics.yaml) as leads for a human — or an agent doing the same
primary-source read utsagn-01/03/06 went through — to verify.

This script NEVER writes into internal/data/content/. It writes leads
(a Markdown report) and per-lead evidence-YAML SKELETONS with real
fields (id/henvisning/tittel/url) filled in and everything that
requires actually reading the case (party_id, position,
verbatim_excerpt, context_note) left as an explicit TODO — so a
skeleton can never be mistaken for real evidence and picked up by the
Go loader by accident. Skeletons land under
tools/harvester/reports/<date>/, which internal/data's readDir never
looks at.

Usage:
    uv run harvest.py                       # all sessions, all topics
    uv run harvest.py --start-year 2020      # narrower session range
    uv run harvest.py --topic utsagn-04      # one topic only
    uv run harvest.py --dry-run              # print session list, no network

Rate limiting: one request per session, in order, with a short delay —
this is a batch offline tool, not a live user-facing dependency, so
there is no reason to hammer stortinget.no's API.
"""

from __future__ import annotations

import argparse
import datetime
import json
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any

import yaml

API_BASE = "https://data.stortinget.no/eksport"
USER_AGENT = "ektevalgomat-harvester/0.1 (offline lead-finder, see docs/roadmap.md)"
CACHE_DIR = Path(__file__).parent / ".cache"
REPORTS_DIR = Path(__file__).parent / "reports"
DEFAULT_TOPICS_FILE = Path(__file__).parent / "topics.yaml"

# Stortinget's first session with this JSON export format online is
# 2009-2010 (the API predates that but coverage gets thin). "Back to
# 2011" in docs/roadmap.md meant "further back than the 2023-2026
# already checked by hand" — 2009 is the honest floor, not an arbitrary
# round number.
EARLIEST_SESSION_START_YEAR = 2009


def generate_sessions(start_year: int, end_year: int) -> list[str]:
    """Return Storting session identifiers like "2020-2021" for every
    session from start_year through end_year (inclusive), oldest first.
    Pure function, no network — the whole point is that this is
    trivially testable without hitting the real API.
    """
    if end_year < start_year:
        return []
    return [f"{y}-{y + 1}" for y in range(start_year, end_year + 1)]


def current_session_start_year() -> int:
    """Best-effort guess at the current session's start year from
    wall-clock time. A Storting session runs roughly October to
    September, so a date in Jan-Sep belongs to the session that started
    the previous calendar year — e.g. 2026-09-23 is still inside the
    2025-2026 session, not 2026-2027.
    """
    now = datetime.date.today()
    return now.year if now.month >= 10 else now.year - 1


def load_topics(path: Path) -> dict[str, dict[str, Any]]:
    """Load topics.yaml. Raises on missing/malformed file rather than
    returning an empty dict — a silently-empty topic set would make a
    harvest run report "found nothing" for a reason that has nothing to
    do with Stortinget's actual case record.
    """
    with path.open(encoding="utf-8") as f:
        data = yaml.safe_load(f)
    if not isinstance(data, dict) or not data:
        raise ValueError(f"{path}: expected a non-empty mapping of statement_id -> topic")
    for statement_id, topic in data.items():
        if "keywords" not in topic or not topic["keywords"]:
            raise ValueError(f"{path}: topic {statement_id!r} has no keywords")
    return data


def fetch_session_cases(session: str, *, use_cache: bool = True) -> list[dict[str, Any]]:
    """Fetch every case (sak) for one Storting session via the JSON
    export API. Caches the raw response on disk so re-running the
    harvester (e.g. after editing topics.yaml) doesn't re-hit the API
    for sessions already fetched — courteous to a public service, and
    makes iteration fast.
    """
    cache_path = CACHE_DIR / f"saker-{session}.json"
    if use_cache and cache_path.exists():
        with cache_path.open(encoding="utf-8") as f:
            payload = json.load(f)
    else:
        url = f"{API_BASE}/saker?sesjonid={session}&format=json"
        req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                raw = resp.read().decode("utf-8")
        except (urllib.error.URLError, urllib.error.HTTPError) as exc:
            print(f"  ! {session}: fetch failed ({exc}) — skipping", file=sys.stderr)
            return []
        payload = json.loads(raw)
        CACHE_DIR.mkdir(parents=True, exist_ok=True)
        with cache_path.open("w", encoding="utf-8") as f:
            json.dump(payload, f, ensure_ascii=False)
    return payload.get("saker_liste", [])


def case_search_text(case: dict[str, Any]) -> str:
    """Concatenate every field of a case worth keyword-matching against,
    lowercased once. Titles from this API sometimes contain a literal
    "?" where an æ/ø/å failed to encode (seen in the 2025-2026 session
    export) — normalized away so a keyword with the correct letter still
    matches.
    """
    parts = [case.get("tittel") or "", case.get("korttittel") or ""]
    for emne in case.get("emne_liste") or []:
        parts.append(emne.get("navn") or "")
    text = " ".join(parts).lower().strip()
    return re.sub(r"\?", "", text)


def match_case(case: dict[str, Any], keywords: list[str]) -> list[str]:
    """Return every keyword (lowercased) that appears in the case's
    search text. Empty list means no match. Pure function — no network,
    fully unit-testable.
    """
    text = case_search_text(case)
    return [kw for kw in keywords if kw.lower() in text]


def render_skeleton_yaml(statement_id: str, case: dict[str, Any], matched_keywords: list[str]) -> str:
    """Render an evidence-YAML-shaped skeleton for one lead. Every field
    the API can actually confirm (id, henvisning, tittel, url) is
    filled in for real; every field that requires reading the case is an
    explicit TODO — never a guess, never a placeholder that could pass
    for real evidence_class/position values if pasted into
    internal/data/content/evidence/ without editing.
    """
    case_id = case.get("id")
    henvisning = case.get("henvisning") or case.get("korttittel") or "UKJENT"
    tittel = case.get("tittel") or case.get("korttittel") or ""
    url = f"https://www.stortinget.no/no/Saker-og-publikasjoner/Saker/Sak/?p={case_id}"
    return (
        f"# LEAD — NOT EVIDENCE. Read the actual case before touching any TODO below.\n"
        f"# Matched keywords: {', '.join(matched_keywords)}\n"
        f"id: TODO-{statement_id}-{case_id}\n"
        f"statement_id: {statement_id}\n"
        f"party_id: \"TODO — one file per party, or \\\"\\\" for institutional context\"\n"
        f"actor: \"TODO\"\n"
        f"date: \"TODO — YYYY-MM-DD\"\n"
        f"evidence_type: \"TODO — see internal/data/loader.go for the field's known values\"\n"
        f"evidence_class: \"TODO — fakta | partiets_begrunnelse | analyse | uklart\"\n"
        f"scan_summary: \"TODO — one sentence\"\n"
        f"claim_action: >\n"
        f"  TODO — what actually happened, in the case's own terms\n"
        f"verbatim_excerpt: >\n"
        f"  TODO — quote from the primary source, not a paraphrase. See\n"
        f"  docs/roadmap.md's kjernekraft/partiloven entries for the bar.\n"
        f"status: \"TODO\"\n"
        f"position: null  # TODO — do not fill until the verbatim excerpt above justifies it\n"
        f"source_tier: \"TODO — A/B/C/D/E, see /metode kildehierarki\"\n"
        f"source_name: \"{henvisning}\"\n"
        f"source_url: \"{url}\"\n"
        f"claim_confidence: \"TODO\"\n"
        f"context_note: >\n"
        f"  TODO — why this position value and not a stronger/weaker one;\n"
        f"  note if this is committee-text-backed or floor-vote-only.\n"
        f"retrieved_at: \"{datetime.date.today().isoformat()}\"\n"
        f"# Case tittel for reference (not a YAML field): {tittel}\n"
    )


def run(topics: dict[str, dict[str, Any]], sessions: list[str], *, topic_filter: str | None, use_cache: bool) -> dict[str, list[dict[str, Any]]]:
    """Scan every session for every topic's keywords. Returns
    statement_id -> list of {case, matched_keywords, session} leads.
    """
    if topic_filter:
        if topic_filter not in topics:
            raise SystemExit(f"--topic {topic_filter!r} not found in topics.yaml (have: {', '.join(topics)})")
        topics = {topic_filter: topics[topic_filter]}

    leads: dict[str, list[dict[str, Any]]] = {sid: [] for sid in topics}
    for session in sessions:
        print(f"scanning {session}...", file=sys.stderr)
        cases = fetch_session_cases(session, use_cache=use_cache)
        for case in cases:
            for statement_id, topic in topics.items():
                matched = match_case(case, topic["keywords"])
                if matched:
                    leads[statement_id].append({"case": case, "matched_keywords": matched, "session": session})
        if not use_cache:
            time.sleep(0.3)  # be a courteous caller to a public, free API
    return leads


def write_report(leads: dict[str, list[dict[str, Any]]], topics: dict[str, dict[str, Any]]) -> Path:
    stamp = datetime.date.today().isoformat()
    out_dir = REPORTS_DIR / stamp
    out_dir.mkdir(parents=True, exist_ok=True)

    lines = [f"# Harvest report — {stamp}", ""]
    total = sum(len(v) for v in leads.values())
    lines.append(f"Total leads: {total} across {len(leads)} topic(s). Every lead below is a")
    lines.append("candidate to READ, not evidence — see the skeleton YAML files in this")
    lines.append("same directory, and docs/roadmap.md's harvester entry for the review bar.")
    lines.append("")

    for statement_id, items in leads.items():
        lines.append(f"## {statement_id}")
        lines.append("")
        lines.append(topics[statement_id]["text"].strip())
        lines.append("")
        if not items:
            lines.append("_No leads found in the scanned sessions._")
            lines.append("")
            continue
        for item in items:
            case = item["case"]
            henvisning = case.get("henvisning") or case.get("korttittel") or "?"
            tittel = case.get("tittel") or case.get("korttittel") or ""
            case_id = case.get("id")
            lines.append(f"- **{henvisning}** ({item['session']}) — {tittel}")
            lines.append(f"  matched: {', '.join(item['matched_keywords'])}")
            lines.append(f"  https://www.stortinget.no/no/Saker-og-publikasjoner/Saker/Sak/?p={case_id}")
            skeleton_path = out_dir / f"{statement_id}-{case_id}.yaml"
            skeleton_path.write_text(render_skeleton_yaml(statement_id, case, item["matched_keywords"]), encoding="utf-8")
            lines.append(f"  skeleton: {skeleton_path.relative_to(REPORTS_DIR.parent)}")
            lines.append("")

    report_path = out_dir / "REPORT.md"
    report_path.write_text("\n".join(lines), encoding="utf-8")
    return report_path


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--start-year", type=int, default=EARLIEST_SESSION_START_YEAR, help="First session's start year (default: earliest the API reliably serves)")
    parser.add_argument("--end-year", type=int, default=None, help="Last session's start year (default: current session)")
    parser.add_argument("--topic", default=None, help="Restrict to one statement_id from topics.yaml")
    parser.add_argument("--topics-file", type=Path, default=DEFAULT_TOPICS_FILE)
    parser.add_argument("--no-cache", action="store_true", help="Ignore any cached session JSON and re-fetch")
    parser.add_argument("--dry-run", action="store_true", help="Print the session list and exit, no network calls")
    args = parser.parse_args()

    end_year = args.end_year if args.end_year is not None else current_session_start_year()
    sessions = generate_sessions(args.start_year, end_year)

    if args.dry_run:
        print(f"Would scan {len(sessions)} sessions: {sessions[0]} .. {sessions[-1]}")
        return 0

    topics = load_topics(args.topics_file)
    leads = run(topics, sessions, topic_filter=args.topic, use_cache=not args.no_cache)
    report_path = write_report(leads, topics)

    total = sum(len(v) for v in leads.values())
    print(f"\n{total} lead(s) written. Report: {report_path}")
    print("Nothing here is evidence yet — read each lead's actual case text before filling in any skeleton TODO.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
