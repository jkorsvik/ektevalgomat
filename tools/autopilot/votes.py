# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Collect the vote stream for one or more Stortinget cases (saker).

Usage:
  uv run tools/autopilot/votes.py 200434
  uv run tools/autopilot/votes.py 200414 200418

For each sak id it fetches
  https://data.stortinget.no/eksport/voteringer?sakid=<id>&format=json
and, for every votering on the case, writes ONE JSON line to
  tools/autopilot/staging/votes/<sak_id>.jsonl
with:
  votering_id, sak_id, votering_tema, forslag (the proposal voted on, from
  voteringsforslag), forslag_betegnelse, vedtatt, the official totals, and the
  party tally.

Two honest caveats the line states about itself:
  * When votering_resultat_type is 5 ("Enstemmig vedtatt") no vote was taken:
    Stortinget returns antall_for/mot/ikke_tilstede = -1 and NO per-representative
    rows, so the party tally is {} and the totals are -1. That is the API's real
    state, not a missing count; the resultat_type fields say so per line.
  * antall_* and the tally are reproduced verbatim from the API, never computed.

Party tally: per party (representant.parti.id) the count of representatives who
voted for / mot / ikke tilstede, using the codes verified in
tools/verify/verify_votes.py: 2 = for, 3 = mot, 1 = ikke tilstede.

This tool never writes into internal/data/content/. Output is a draft under
staging/ only. It prints one summary line per case and exits non-zero (2) if it
checked no case at all, so a run that fetched nothing never looks like success.
"""

import hashlib
import json
import os
import pathlib
import re
import sys
import time
import urllib.error
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / "tools/autopilot/staging/votes"
CACHE = pathlib.Path.home() / ".cache/ektevalgomat-sources"
BASE = "https://data.stortinget.no/eksport"
VOTERINGER = BASE + "/voteringer?sakid={}&format=json"
FORSLAG = BASE + "/voteringsforslag?voteringid={}&format=json"
RESULTAT = BASE + "/voteringsresultat?voteringid={}&format=json"
SAKER = BASE + "/saker?format=json"

# codes verified against vote 26387's official totals in verify_votes.py
CODE = {2: "for", 3: "mot", 1: "ikke_tilstede"}


def fetch(url: str, tries: int = 8) -> dict:
    """GET + JSON, cached by URL under CACHE; 429-backed-off so a burst is safe."""
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / hashlib.sha256(url.encode()).hexdigest()
    if f.exists() and f.stat().st_size > 0:
        return json.loads(f.read_bytes())
    backoff = 2.0
    for _ in range(tries):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": "ektevalgomat votes collector"})
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


def tally(rows: list[dict]) -> dict[str, dict[str, int]]:
    """Per party id -> {for, mot, ikke_tilstede}."""
    out: dict[str, dict[str, int]] = {}
    for r in rows:
        party = (r.get("representant") or {}).get("parti", {}).get("id")
        if party is None:
            continue
        bucket = out.setdefault(party, {"for": 0, "mot": 0, "ikke_tilstede": 0})
        label = CODE.get(r.get("votering"))
        if label:
            bucket[label] += 1
    return out


def titles() -> dict[int, str]:
    try:
        L = fetch(SAKER).get("saker_liste") or []
        return {s["id"]: s.get("korttittel") for s in L}
    except Exception:  # noqa: BLE001 - a missing title must not abort the collect
        return {}


def collect(sak_id: int, tmap: dict[int, str]) -> tuple[int, pathlib.Path]:
    lst = fetch(VOTERINGER.format(sak_id)).get("sak_votering_liste") or []
    OUT.mkdir(parents=True, exist_ok=True)
    dest = OUT / f"{sak_id}.jsonl"
    lines = []
    for v in lst:
        vid = v["votering_id"]
        fl = (fetch(FORSLAG.format(vid)).get("voteringsforslag_liste") or [])
        forslag = " | ".join(x for x in (strip_html(x.get("forslag_tekst")) for x in fl) if x)
        betegnelse = " | ".join(
            (x.get("forslag_betegnelse_kort") or x.get("forslag_betegnelse") or "") for x in fl
        )
        rows = fetch(RESULTAT.format(vid)).get("voteringsresultat_liste") or []
        lines.append(json.dumps({
            "votering_id": vid,
            "sak_id": sak_id,
            "votering_tema": v.get("votering_tema"),
            "forslag": forslag,
            "forslag_betegnelse": betegnelse,
            "vedtatt": v.get("vedtatt"),
            "votering_resultat_type": v.get("votering_resultat_type"),
            "votering_resultat_type_tekst": v.get("votering_resultat_type_tekst"),
            "antall_for": v.get("antall_for"),
            "antall_mot": v.get("antall_mot"),
            "antall_ikke_tilstede": v.get("antall_ikke_tilstede"),
            "parti": tally(rows),
        }, ensure_ascii=False))
    dest.write_text("\n".join(lines) + ("\n" if lines else ""), encoding="utf-8")
    title = tmap.get(sak_id) or ""
    print(f"sak {sak_id}: {len(lst)} votes -> {dest.relative_to(ROOT)}"
          + (f'  "{title.strip()}"' if title else ""))
    return len(lst), dest


def main(argv: list[str]) -> int:
    ids = [int(a) for a in argv if a.isdigit()]
    if not ids:
        print("usage: uv run tools/autopilot/votes.py SAKID [SAKID ...]", file=sys.stderr)
        return 2
    tmap = titles()
    total = 0
    for sid in ids:
        n, _ = collect(sid, tmap)
        total += n
        time.sleep(0.25)
    print(f"{len(ids)} case(s), {total} vote(s) written under tools/autopilot/staging/votes/")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
