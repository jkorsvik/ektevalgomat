# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Build the vote index: every cited votering, with each party's side on it.

Usage:
  uv run tools/analytics/build_vote_index.py            # write the index
  uv run tools/analytics/build_vote_index.py --dry-run

Why this file exists: the site can say what one party did on one statement, but
it cannot answer "who votes with whom", because nothing in the dataset holds a
vote as a party-by-party record. The evidence rows cite 400+ voterings and name
each party's side in prose; the tallies themselves are never stored. A
vote-alignment map needs them as data.

So this walks every evidence file that cites a vote, fetches the official tally
for each cited votering, and writes one generated artefact under
internal/data/votes/. It is generated, never hand-written: re-running it against
the same sources must reproduce the same file, and a test checks that every
cit votering in the content appears in the index.

What it is not: an opinion about what a vote "means". It records, per vote, the
proposal text as the Stortinget API states it, the date, whether it carried, and
each party's side computed from the official per-representative rows. Reading
those as a position stays a human act with a citation attached.

Exhaustive mode: by default this walks only the voterings the content cites, so the
index is exactly as wide as the evidence. `--sessions` widens it: for each session it
lists every sak, asks for that sak's voterings, and adds them all. Two facts cost
real time to find and are recorded here so nobody rediscovers them: the API has no
session-level votering listing (voteringer?sesjonid= answers 400, "missing SakId"),
and the per-sak payload names its list `sak_votering_liste` with the id under
`votering_id`. Reading `voteringer_liste` returns an empty list on every sak and
looks like absence of data, which is how this was first misdiagnosed.

Which of those claims are re-checked, and which are inherited (2026-10-27):

  verified against cached payloads
    saker?sesjonid=2025-2026 -> `saker_liste`, 663 saker; `sak_votering_liste`
    present in 543 cached per-sak payloads, `voteringer_liste` in none.
  inherited, not re-checked
    the 400 from voteringer?sesjonid=. No cached error body exists to check it
    against. Treat it as the prior session's finding until one does.

A walk resumes from the fetch cache, so a run killed by a network loss costs only
the saker it had not reached. Set `session` on a vote only when the walk found it:
an empty `session` means "reached through a citation, session unknown", never "this
vote belongs to no session". Nothing reads `session` yet -- the Go `Vote` type does
not carry it, so it is stored, not shown. Add the field there only when a page
needs it; an empty column is worse than a missing one.

Codes, as verified in tools/verify/verify_votes.py against vote 26387's official
totals: 2 = for, 3 = mot, 1 = ikke tilstede.
"""

import argparse
import datetime
import hashlib
import json
import os
import pathlib
import re
import sys
import time
import urllib.error
import urllib.request

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "tools/verify"))
sys.path.insert(0, str(ROOT / "tools/autopilot"))
import verify_votes as vv  # noqa: E402  - one implementation of id extraction and party mapping

EVIDENCE = ROOT / "internal/data/content/evidence"
OUT = ROOT / "internal/data/votes"
CACHE = pathlib.Path.home() / ".cache/ektevalgomat-sources"
BASE = "https://data.stortinget.no/eksport"
FORSLAG = BASE + "/voteringsforslag?voteringid={}&format=json"
RESULTAT = BASE + "/voteringsresultat?voteringid={}&format=json"
VOTERINGER = BASE + "/voteringer?sakid={}&format=json"
SAKER = BASE + "/saker?sesjonid={}&format=json"

CODE = {2: "for", 3: "mot", 1: "ikke_tilstede"}
API_TO_ID = {v: k for k, v in vv.PARTY.items()}  # "A" -> "ap"


def fetch(url: str, tries: int = 8) -> dict:
    """GET + JSON, cached by URL under CACHE; 429-backed-off so a burst is safe."""
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / hashlib.sha256(url.encode()).hexdigest()
    if f.exists() and f.stat().st_size > 0:
        return json.loads(f.read_bytes())
    backoff = 2.0
    for _ in range(tries):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": "ektevalgomat vote index"})
            with urllib.request.urlopen(req, timeout=60) as r:
                body = r.read()
            tmp = f.with_suffix(f".{os.getpid()}.tmp")
            tmp.write_bytes(body)
            os.replace(tmp, f)
            return json.loads(body)
        except urllib.error.HTTPError as e:
            if e.code == 429:
                time.sleep(backoff)
                backoff = min(backoff * 1.7, 30)
                continue
            raise
    raise RuntimeError(f"gave up after {tries} tries: {url}")


def strip_html(s: str | None) -> str:
    if not s:
        return ""
    s = re.sub(r"<[^>]+>", " ", s)
    s = s.replace("&nbsp;", " ").replace("&amp;", "&").replace("&lt;", "<").replace("&gt;", ">")
    return re.sub(r"\s+", " ", s).strip()


def side(rows: list[dict]) -> dict[str, str]:
    """Per party id -> for | mot | split | absent.

    A split is not a side: 3 for and 3 against is a different fact from either,
    and collapsing it to a majority would silently invent a position.
    """
    tally: dict[str, dict[str, int]] = {}
    for r in rows:
        api = (r.get("representant") or {}).get("parti", {}).get("id")
        pid = API_TO_ID.get(api or "")
        if not pid:
            continue
        bucket = tally.setdefault(pid, {"for": 0, "mot": 0})
        label = CODE.get(r.get("votering"))
        if label in ("for", "mot"):
            bucket[label] += 1
    out = {}
    for pid, b in tally.items():
        if b["for"] and b["mot"]:
            out[pid] = "split"
        elif b["for"]:
            out[pid] = "for"
        elif b["mot"]:
            out[pid] = "mot"
        else:
            out[pid] = "absent"
    return out


def session_votes(sessions: list[str]) -> dict[int, str]:
    """votering id -> session, from every sak in every named session."""
    out: dict[int, str] = {}
    for ses in sessions:
        sak_list = fetch(SAKER.format(ses)).get("saker_liste") or []
        got = 0
        for sak in sak_list:
            sid = sak.get("id")
            if not sid:
                continue
            try:
                body = fetch(VOTERINGER.format(sid))
            except Exception as e:  # noqa: BLE001 - report, never hide
                print(f"  sak {sid} ({ses}): fetch failed: {e}", file=sys.stderr)
                continue
            for v in body.get("sak_votering_liste") or []:
                vid = v.get("votering_id")
                if not vid or vid < 0:
                    continue
                out[int(vid)] = ses
                got += 1
            time.sleep(0.34)  # the API answers 429 to a fast scan; stay under it
        print(f"  {ses}: {len(sak_list)} saker, {got} voterings", file=sys.stderr)
    return out


def cited_votes() -> dict[str, dict]:
    """votering id -> {evidence_ids, party_ids}, from what the content cites."""
    cites: dict[str, dict] = {}
    for p in sorted(EVIDENCE.glob("*.yaml")):
        d = yaml.safe_load(p.read_text()) or {}
        ids = vv.vote_ids(f"{d.get('source_name') or ''} {d.get('source_url') or ''} "
                          f"{d.get('context_note') or ''} {d.get('claim_action') or ''}")
        for vid in ids:
            e = cites.setdefault(vid, {"evidence_ids": [], "parties": []})
            e["evidence_ids"].append(d.get("id") or p.stem)
            pid = (d.get("party_id") or "").strip()
            if pid and pid not in e["parties"]:
                e["parties"].append(pid)
    return cites


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--sessions", default="",
                    help="comma-separated sessions to walk exhaustively, e.g. 2025-2026,2024-2025")
    args = ap.parse_args(argv)

    cites = cited_votes()
    if not cites:
        print("no cited votes found in the content: treat as failure", file=sys.stderr)
        return 2
    print(f"content cites {len(cites)} distinct voterings", file=sys.stderr)

    wide: dict[int, str] = {}
    sessions = [x.strip() for x in args.sessions.split(",") if x.strip()]
    if sessions:
        wide = session_votes(sessions)
        print(f"{len(wide)} voterings found by walking {len(sessions)} session(s)", file=sys.stderr)

    allvids = sorted({int(v) for v in cites} | set(wide))
    votes, failed = [], 0
    for vid in allvids:
        vid = str(vid)
        try:
            rows = fetch(RESULTAT.format(vid)).get("voteringsresultat_liste") or []
        except Exception as e:  # noqa: BLE001 - report, never hide
            print(f"  votering {vid}: fetch failed: {type(e).__name__}: {e}", file=sys.stderr)
            failed += 1
            continue
        if not rows:
            print(f"  votering {vid}: no per-representative rows (unanimous or not a vote)",
                  file=sys.stderr)
        fl = fetch(FORSLAG.format(vid)).get("voteringsforslag_liste") or []
        forslag = " | ".join(x for x in (strip_html(y.get("forslag_tekst")) for y in fl) if x)
        # No date. voteringsresultat carries no sak_id and there is no
        # votering->sak endpoint (eksport/votering returns 404), so a date could
        # only be had by scanning every case's voterings across 18 sessions. A
        # mostly-empty date column reads as missing data; the year axis belongs to
        # the position view, which has real evidence dates, not to this one.
        votes.append({
            "votering_id": int(vid),
            # A vote with no per-representative rows was taken unanimously (the
            # API returns resultat_type 5 and no party tally). It compares nothing,
            # so it is marked rather than dropped: the coverage test requires the
            # index to hold every cited votering, and silently omitting one would
            # make that test pass while the map lost a vote it cites.
            "unanimous": not rows,
            "forslag": forslag[:600],
            "parties": side(rows),
            "cited_by": (cites.get(vid) or {}).get("evidence_ids") or [],
            "coded_parties": (cites.get(vid) or {}).get("parties") or [],
            "session": wide.get(int(vid), ""),
        })

    if not votes:
        print("fetched no votes at all: treat as failure", file=sys.stderr)
        return 2

    doc = {"generated_at": datetime.date.today().isoformat(), "votes": votes}
    body = json.dumps(doc, ensure_ascii=False, indent=1) + "\n"
    if args.dry_run:
        print(f"would write {len(votes)} votes to {OUT}/index.json ({len(body)} bytes)")
        return 0
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / "index.json").write_text(body)
    print(f"wrote {OUT}/index.json: {len(votes)} votes, {failed} fetch failures")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
