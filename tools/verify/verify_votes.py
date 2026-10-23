# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Check every "stemte for/mot" evidence file against Stortinget's official vote record.

Usage:
  uv run tools/verify/verify_votes.py                 # all evidence files
  uv run tools/verify/verify_votes.py FILE [FILE...]

For each file whose status contains stemte_for / stemte_mot and that has a
party_id, it extracts vote ids (voteringid=N, vt=N, "votering_id N",
"Voteringene N og M") from source_name and source_url, then context_note,
fetches https://data.stortinget.no/eksport/voteringsresultat for each, and
compares the party's majority of present representatives with the status.

Codes in the API, checked against the official totals of vote 26387
(7 for, 86 mot, 76 ikke tilstede): 2 = for, 3 = mot, 1 = ikke tilstede.

Result per file:
  OK        every cited vote matches
  MIXED     several votes cited, at least one matches (typical: the party's
            own proposal plus the vote the status refers to). Not a failure.
  MISMATCH  no cited vote shows the party on the status's side
  SPLIT     the party split (no clear majority) on a cited vote
  ABSENT    no representative of the party voted
  NOID      status says the party voted, but no vote id is cited
A vote on "Komiteens tilråding: <Dokument 8> – vedtas ikke" is read inverted:
"for" the tilråding means against the proposal (vote 27597, statement 10).

Exit 1 if any MISMATCH. NOID is a coverage gap, reported but not a failure.
"""

import hashlib
import json
import pathlib
import re
import sys
import os
import urllib.request

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
EVIDENCE = ROOT / "internal/data/content/evidence"
CACHE = pathlib.Path.home() / ".cache/ektevalgomat-sources"
API = "https://data.stortinget.no/eksport/voteringsresultat?voteringid={}&format=json"
FORSLAG = "https://data.stortinget.no/eksport/voteringsforslag?voteringid={}&format=json"
PARTY = {"ap": "A", "frp": "FrP", "h": "H", "sv": "SV", "sp": "Sp",
         "rodt": "R", "mdg": "MDG", "krf": "KrF", "v": "V"}
CODE = {2: "for", 3: "mot"}

ID_PATTERNS = [
    r"voteringid=(\d{4,6})",
    r"[?&]vt=(\d{4,6})",
    r"votering[_ -]?id[\s:=]*((?:\d{4,6})(?:\s*(?:,|og)\s*\d{4,6})*)",
    r"[Vv]otering(?:en|ene)?\s+((?:\d{4,6})(?:\s*(?:,|og)\s*\d{4,6})*)",  # "votering 27523" (W6 found it)
]


def vote_ids(text: str) -> list[str]:
    out: list[str] = []
    for p in ID_PATTERNS:
        for m in re.findall(p, text):
            for n in re.findall(r"\d{4,6}", m):
                if n not in out:
                    out.append(n)
    return out


def rejects_proposal(vid: str) -> bool:
    """True if the only thing voted on is a tilråding that the proposal "vedtas ikke"."""
    fl = fetch_json(FORSLAG.format(vid)).get("voteringsforslag_liste") or []
    return len(fl) == 1 and "tilråding" in (fl[0].get("forslag_betegnelse") or "").lower() \
        and "vedtas ikke" in (fl[0].get("forslag_tekst") or "")


def fetch_votes(vid: str) -> list[dict]:
    return fetch_json(API.format(vid)).get("voteringsresultat_liste") or []


def fetch_json(url: str) -> dict:
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / hashlib.sha256(url.encode()).hexdigest()
    if not (f.exists() and f.stat().st_size > 0):
        with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": "ektevalgomat verifier"}), timeout=60) as r:
            body = r.read()
        tmp = f.with_suffix(f".{os.getpid()}.tmp")
        tmp.write_bytes(body)
        os.replace(tmp, f)
    return json.loads(f.read_bytes())


def party_side(rows: list[dict], party: str) -> tuple[str, dict]:
    n = {"for": 0, "mot": 0}
    for r in rows:
        if r["representant"]["parti"]["id"] == party and r.get("votering") in CODE:
            n[CODE[r["votering"]]] += 1
    if n["for"] == n["mot"] == 0:
        return "absent", n
    if n["for"] and n["mot"] and min(n.values()) * 4 > max(n.values()):
        return "split", n  # minority above 20 %: no clear party line
    return ("for" if n["for"] > n["mot"] else "mot"), n


def main(argv: list[str]) -> int:
    files = [pathlib.Path(a) for a in argv] or sorted(EVIDENCE.glob("*.yaml"))
    stats = {k: 0 for k in ("OK", "MIXED", "MISMATCH", "SPLIT", "ABSENT", "NOID", "ERROR")}
    for f in files:
        d = yaml.safe_load(f.read_text(encoding="utf-8"))
        st, pid = str(d.get("status") or ""), d.get("party_id") or ""
        m = re.search(r"stemte_(for|mot)$", st)
        if not m or pid not in PARTY:
            continue
        want = m.group(1)
        ids = vote_ids(f"{d.get('source_name', '')} {d.get('source_url', '')}") or \
            vote_ids(str(d.get("context_note", "")))
        if not ids:
            stats["NOID"] += 1
            print(f"NOID      {f.name}  status={st}")
            continue
        sides, detail = [], []
        for vid in ids:
            try:
                side, n = party_side(fetch_votes(vid), PARTY[pid])
                if side in ("for", "mot") and rejects_proposal(vid):
                    side = {"for": "mot", "mot": "for"}[side]
                    vid += "(tilråding vedtas ikke, inverted)"
            except Exception as e:  # noqa: BLE001 - report, never hide
                side, n = "error", {"err": str(e)}
            sides.append(side)
            detail.append(f"{vid}:{side} {n.get('for', '')}/{n.get('mot', '')}")
        if "error" in sides:
            worst = "ERROR"
        elif all(x == want for x in sides):
            worst = "OK"
        elif want in sides:
            worst = "MIXED"
        elif any(x in ("for", "mot") for x in sides):
            worst = "MISMATCH"
        else:
            worst = "SPLIT" if "split" in sides else "ABSENT"
        stats[worst] += 1
        if worst != "OK":
            print(f"{worst:<9} {f.name}  {pid} status={st}  votes {' '.join(detail)} (for/mot)")
    checked = sum(v for k, v in stats.items() if k != "NOID")
    print("\n" + "  ".join(f"{k} {v}" for k, v in stats.items()) + f"  (checked {checked} files against the vote record)")
    if checked == 0:
        print("nothing checked: treat as failure")
        return 2
    return 1 if stats["MISMATCH"] or stats["ERROR"] else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
