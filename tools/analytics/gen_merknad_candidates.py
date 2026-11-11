"""Rank earlier innstillinger as candidates for a statement that lacks an
earlier year.

The site can only say a party stayed consistent when a cell holds two distinct
years. 447 of 517 cells hold one. The committee recommendation (innstilling) for
an earlier session on the same question is the strongest second source,
because the party states its own position in its own words.

This script does the part that needs no reading of document text: it pairs each
statement with earlier innstillinger whose official subject tags overlap the
statement's own wording. It emits ids, sessions, scores and PDF URLs. It never
emits a title, a summary or any prose, because the driving model must not see
people and because a ranking is not evidence.

A candidate is a lead. It becomes evidence only after the excerpt is taken from
the document and admitted on its own words.

MEASURED YIELD, 2026-11-11, and it is a negative result. Over 15 sessions there
are 7,124 innstillinger with a resolvable PDF URL, and 165 distinct subject tags.
Across the thirteen statements this was built for, the best overlap score is 4,
one statement reaches it, and nothing reaches 5. Scores of 2 to 4 are two to
four shared words drawn from a 165-word vocabulary, so they do not identify a
document that asks the same question. Six of the thirteen statements score zero.

So this is a breadth tool, not a locator. Use it to enumerate what exists in a
session, not to decide what to read. Finding the earlier document that asks the
same question needs a reader: the subject tags cannot carry meaning that fine.
The API offers no help either: sak_relasjon_liste links siblings inside one
session (four rows share one innstilling), never a previous session's version of
the same case.
"""

import argparse
import collections
import json
import pathlib
import re
import sys
import urllib.request

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "autopilot"))
from innstilling_locate import Withheld, locate  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONTENT = ROOT / "internal" / "data" / "content"
CACHE = pathlib.Path.home() / ".cache" / "ektevalgomat-sources"
API = "https://data.stortinget.no/eksport/saker?sesjonid=%s&format=json"
UA = "ektevalgomat/1.0 (+https://ektevalgomat.no)"

SESSIONS = [f"{y}-{y + 1}" for y in range(2011, 2026)]

STOP = set("""
alle alltid annen annet også bare både blant bli blir blitt bør dalle denne deres
disse dette eller etter fleste flere fra før fås godt gjennom gjøre gjør gjelde
gjelder gå ha har hatt her hos hvor hvordan hver hvilken hvis ikke ingen inn imot
kan komme kommer kunne langt lett man mange med meg men mer mest mot må måte ned
noe noen nokså når ny nye nær og om opp oss over på samme seg selv sett skal skje
skulle slik som stor store står svært synes så sånn til to trenger under uten ved
vei være vært vil vite være åpen åpent år
""".split())

WORD = re.compile(r"[a-zæøå]{4,}")


def keywords(*texts):
    out = set()
    for t in texts:
        if not t:
            continue
        for w in WORD.findall(str(t).lower()):
            if w not in STOP:
                out.add(w)
    return out


def fetch(session):
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / ("saker-%s.json" % session)
    if f.exists() and f.stat().st_size > 0:
        return json.loads(f.read_text(encoding="utf-8"))
    req = urllib.request.Request(API % session, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=60) as r:
        body = r.read()
    f.write_bytes(body)
    return json.loads(body.decode("utf-8"))


def tags_of(row):
    """Subject tags plus the short title, tokenised.

    The short title is included as matching input only: it is never printed and
    never stored, because a short title can name a person on a
    representantforslag. The output carries ids, sessions, scores and URLs.
    """
    out = set()
    for e in list(row.get("emne_liste") or []) + [row.get("korttittel") or ""]:
        s = e if isinstance(e, str) else (e.get("navn") or "")
        for w in WORD.findall(str(s).lower()):
            if w not in STOP:
                out.add(w)
        # keep the whole tag too, so a two-word tag can match on the pair
    return out


def innstilling_rows(session):
    """Every sak in the session that carries an innstilling, with its PDF URL."""
    out = []
    for row in fetch(session).get("saker_liste") or []:
        if str(row.get("innstilling_kode")) != "1":
            continue
        try:
            url = locate(row.get("henvisning"), row.get("innstilling_kode"))
        except Withheld:
            continue
        out.append({"session": session, "sak_id": row.get("id"), "url": url,
                    "tags": tags_of(row)})
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("statements", nargs="+")
    ap.add_argument("--min-score", type=int, default=2)
    ap.add_argument("--top", type=int, default=6)
    ap.add_argument("--out", default="tools/autopilot/staging/merknad-candidates")
    a = ap.parse_args()

    index = collections.defaultdict(list)
    for s in SESSIONS:
        for r in innstilling_rows(s):
            index[s].append(r)
    total = sum(len(v) for v in index.values())
    print("sessions %d, innstillinger with a PDF URL %d" % (len(index), total))

    dest = ROOT / a.out
    dest.mkdir(parents=True, exist_ok=True)
    for sid in a.statements:
        import yaml
        st = None
        for cand in sorted(CONTENT.joinpath("statements").glob("*.yaml")):
            d = yaml.safe_load(cand.read_text(encoding="utf-8"))
            if str(d.get("id")) == sid:
                st = d
                break
        if st is None:
            print("  %s: no statement file" % sid)
            continue
        kw = keywords(st.get("text"), st.get("short_name"), st.get("case_tag"),
                      st.get("dimension"))
        scored = []
        for session, rows in index.items():
            for r in rows:
                hit = kw & r["tags"]
                if len(hit) >= a.min_score:
                    scored.append({"session": session, "sak_id": r["sak_id"],
                                   "url": r["url"], "score": len(hit),
                                   "matched_tags": len(hit)})
        scored.sort(key=lambda x: (-x["score"], x["session"]))
        keep = scored[: a.top]
        (dest / ("%s.yaml" % sid)).write_text(
            yaml.safe_dump({"statement": sid, "keywords": sorted(kw),
                            "candidates": keep}, allow_unicode=True, sort_keys=False),
            encoding="utf-8")
        print("  %s: %d candidates above %d, kept %d (best %s)"
              % (sid, len(scored), a.min_score, len(keep),
                 keep[0]["score"] if keep else "-"))
    print("wrote %s" % a.out)


if __name__ == "__main__":
    main()
