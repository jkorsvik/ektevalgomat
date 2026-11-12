#!/usr/bin/env python3
"""Build the committee-membership index from Stortinget's own committee pages.

The export API cannot answer this: `komite?id=` returns 400 naming
`HentKomiteMedlemmer`, and `komitemedlemmer`/`komitemedlem` are 404. The
representative payload carries no committee field. But each committee's page on
stortinget.no embeds the membership as JSON, with the same `perid` our person
records use, so membership joins without name matching.

What the pages give: the full member list, and roles where the page's own table
states them (Leder, Første nestleder, Nestleder, Medlem). What they do not give:
any start or end date. This is therefore a snapshot, exactly like the verv
register, and the artefact says so rather than inventing a start date.

Usage:
    uv run tools/analytics/build_committee_index.py
    uv run tools/analytics/build_committee_index.py --session 2026-2027 --out /tmp/komiteer.json
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import html
import json
import pathlib
import re
import sys
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / "internal/data/komiteer/index.json"
CACHE = pathlib.Path.home() / ".cache/ektevalgomat-sources"
BASE = "https://data.stortinget.no/eksport"
INDEX_PAGE = "https://www.stortinget.no/no/Representanter-og-komiteer/Komiteene/"
USER_AGENT = "ektevalgomat committee index"

MEMBER_RE = re.compile(r'perid=([A-Z0-9]{3,8})","memberImage":\{[^}]*\},"fullName":"([^"]+)"')
ROLE_RE = re.compile(
    r'"value":"(Leder|Første nestleder|Nestleder|Medlem)"[^{]*\{[^{]*\{[^}]*?"href":"[^"]*perid=([A-Z0-9]{3,8})"'
)
COMMITTEE_ID_RE = re.compile(r'"committeeId":"([A-ZÆØÅ0-9]+)"')
TITLE_RE = re.compile(r"<title>([^<]{3,120})</title>")
SLUG_RE = re.compile(r"/no/Representanter-og-komiteer/Komiteene/[A-Za-z0-9%\-]+/")


def fetch_text(url: str) -> str:
    """GET with the same URL-keyed cache the vote builder uses, so a rerun is offline."""
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / hashlib.sha256(url.encode()).hexdigest()
    if f.exists() and f.stat().st_size > 0:
        return f.read_text(encoding="utf-8", errors="replace")
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(req, timeout=60) as r:
        body = r.read()
    f.write_bytes(body)
    return body.decode("utf-8", errors="replace")


def unescape(raw: str) -> str:
    for a, b in (("\\u0022", '"'), ("\\u003c", "<"), ("\\u003e", ">"), ("\\/", "/"), ("\\\\", "")):
        raw = raw.replace(a, b)
    return raw


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--session", default="2026-2027")
    ap.add_argument("--out", default=str(OUT))
    args = ap.parse_args()

    names = {}
    payload = json.loads(fetch_text(f"{BASE}/komiteer?sesjonid={args.session}&format=json"))
    for row in payload.get("komiteer_liste") or []:
        names[row["id"]] = row["navn"]
    if not names:
        print("FAIL: the committee list came back empty", file=sys.stderr)
        return 2

    index_html = fetch_text(INDEX_PAGE)
    slugs = sorted({m.group(0) for m in SLUG_RE.finditer(index_html)})
    # The index page also links the group page "Andre faste komiteer" and the
    # historic committee list; those carry no committeeId and are dropped below.
    if len(slugs) < 10:
        print(f"FAIL: only {len(slugs)} committee links on the index page", file=sys.stderr)
        return 3

    committees = []
    parsed_urls: dict[str, str] = {}
    people: dict[str, list[dict]] = {}
    problems: list[str] = []
    for slug in slugs:
        url = "https://www.stortinget.no" + slug + "?tab=MembersCommittee"
        page = unescape(fetch_text(url))
        cid = COMMITTEE_ID_RE.search(page)
        if not cid:
            continue
        cid = cid.group(1)
        # The API list does not contain every body the pages publish: Europautvalget
        # appears there as committeeId EØS and is absent from the session's list. The
        # page title carries the name, so it is the fallback rather than the id.
        title = TITLE_RE.search(page)
        cname = names.get(cid) or (
            html.unescape(title.group(1)).replace(" - stortinget.no", "").strip() if title else cid
        )
        members = MEMBER_RE.findall(page)
        roles = dict((pid, role) for role, pid in ROLE_RE.findall(page))
        if len(members) < 5:
            problems.append(f"{cid}: parsed only {len(members)} members")
        seen = set()
        rows = []
        for pid, name in members:
            if pid in seen:
                continue
            seen.add(pid)
            rows.append({"person_id": pid, "name": name, "role": roles.get(pid) or "Medlem"})
            people.setdefault(pid, []).append(
                {"committee_id": cid, "committee_name": cname, "role": roles.get(pid) or "Medlem"}
            )
        parsed_urls[cid] = url
        committees.append(
            {
                "id": cid,
                "name": cname,
                "url": "https://www.stortinget.no" + slug,
                "members": sorted(rows, key=lambda r: (r["role"] != "Leder", r["name"])),
            }
        )

    if problems:
        print("FAIL: " + "; ".join(problems), file=sys.stderr)
        return 4
    if len(committees) < 10:
        print(f"FAIL: only {len(committees)} committees parsed", file=sys.stderr)
        return 5

    for c in committees:
        for m in c["members"]:
            if c["name"] == m["name"]:
                print(f"FAIL: {c['id']} is named after its own member {m['name']}", file=sys.stderr)
                return 7

    memberships = sum(len(c["members"]) for c in committees)
    if memberships < 100:
        print(f"FAIL: {memberships} memberships across {len(committees)} committees looks too low",
              file=sys.stderr)
        return 6

    # The committee pages under Komiteene cover the standing committees and
    # Europautvalget. Six bodies the API lists — Presidentskapet, Fullmaktskomiteen,
    # Valgkomiteen, Den utvidede utenriks- og forsvarskomité, the corona committee
    # and the EOS committee — have no page there; every URL tried returns 404.
    # They are reported rather than silently absent, so the site can say which
    # bodies it cannot describe.
    unpublished = [
        {"id": cid, "name": navn}
        for cid, navn in sorted(names.items())
        if cid not in {c["id"] for c in committees}
    ]

    out = {
        "generated_at": dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "source": "stortinget.no committee pages, ?tab=MembersCommittee, embedded membership JSON",
        "session": args.session,
        "snapshot": True,
        "snapshot_note": (
            "The pages state who sits on a committee now and no start or end date. "
            "This is the membership at fetch time, not when the seat began."
        ),
        "committees": sorted(committees, key=lambda c: c["id"]),
        "people": dict(sorted(people.items())),
        "unpublished": unpublished,
        "counts": {
            "committees": len(committees),
            "memberships": memberships,
            "people": len(people),
        },
    }
    pathlib.Path(args.out).parent.mkdir(parents=True, exist_ok=True)
    pathlib.Path(args.out).write_text(json.dumps(out, ensure_ascii=False, indent=1) + "\n")
    print(f"wrote {args.out}: {len(committees)} committees, {memberships} memberships, "
          f"{len(people)} people, {len(unpublished)} bodies without a published member page")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
