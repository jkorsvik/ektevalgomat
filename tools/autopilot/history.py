# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Find historical Stortinget votes that bear on an already-coded cell.

Usage:
  uv run tools/autopilot/history.py                    # every statement with coded cells
  uv run tools/autopilot/history.py --statement utsagn-12
  uv run tools/autopilot/history.py --min-score 3 --top 8

Why this exists: 510 of the 518 coded cells hold exactly ONE evidence item,
and only 7 cells hold two different dates. The konsistens score needs two
dated observations to say anything, so it reports "utilstrekkelig grunnlag"
almost everywhere. That is not a bug in the score, it is the shape of the
data. This tool finds the second date.

The chain, every step a real API field (no title guessing):

  evidence file  --source_name-->  "Innst. 535 L (2024-2025)"
                 --henvisning-->   sak id in that session's list
                 --emne_liste-->   the case's subject ids
                 --same emne-->    saker in EARLIER sessions
                 --voteringer-->   the votes taken, with forslag text and tally

It writes a report only. Nothing here is evidence, and this tool never touches
internal/data/content/: a candidate is a lead a human must open and read, the
same rule tools/harvester/README.md states for its own skeletons. Output goes
to tools/autopilot/reports/history/<today>/ (gitignored).

A zero-candidate run is reported as zero candidates, never as "nothing exists":
matching is on emne ids plus title tokens, and a real vote can sit under a case
whose title says none of the statement's words.
"""

import argparse
import collections
import datetime
import json
import pathlib
import re
import sys
import unicodedata

import yaml

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import votes as V  # noqa: E402  - reuse its cached, backoff-aware fetch

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONTENT = ROOT / "internal/data/content"
REPORTS = ROOT / "tools/autopilot/reports/history"

# Sessions back to 2009-2010: the Stortinget export's own horizon, and far
# enough back that every party's position has a record to compare against.
FIRST_YEAR = 2009

# An innstilling reference as evidence files actually write it, with either
# dash: "Innst. 535 L (2024–2025)", "Innst. 2 S (2025-2026)".
INNST = re.compile(r"Innst\.\s*(\d+)\s*([A-Za-z])\s*\((\d{4})\s*[–\-]\s*(\d{4})\)")

MAX_VOTES = 12  # per candidate case; a 58-vote case is not worth the fetch budget

STOP = set("""
og i til av for med på som er det en et den de har skal bør kan må ikke om fra
seg sin sitt sine blir ble vært være mot ved når hvor hva hvem hvilken denne
dette disse slik også bare ennå allerede fortsatt kun både eller men samtlige
alle noen noe hver andre annet egen eget egne seg
""".split())


def norm(s: str) -> str:
    return unicodedata.normalize("NFKC", s or "").lower().strip()


def tokens(s: str) -> set[str]:
    """Words that carry meaning, Norwegian letters kept literal."""
    return {w for w in re.findall(r"[a-zæøå0-9]+", norm(s)) if len(w) > 3 and w not in STOP}


def sessions(first_year: int = FIRST_YEAR) -> list[str]:
    today = datetime.date.today()
    last = today.year if today.month >= 10 else today.year - 1
    return [f"{y}-{y + 1}" for y in range(first_year, last + 1)]


def session_saker(ses: str) -> list[dict]:
    return V.fetch(V.BASE + f"/saker?format=json&sesjonid={ses}").get("saker_liste") or []


def build_index(ses_list: list[str]) -> tuple[dict, dict]:
    """Two indexes over every sak in every session.

    by_ref:  "innst. 535 l|2024-2025" -> [sak, ...]   what evidence cites
    by_emne: emne id -> [(session, sak), ...]          how we walk back in time
    """
    by_ref: dict[str, list[dict]] = collections.defaultdict(list)
    by_emne: dict[int, list[tuple[str, dict]]] = collections.defaultdict(list)
    total = 0
    for ses in ses_list:
        saker = session_saker(ses)
        total += len(saker)
        for s in saker:
            for m in INNST.finditer(s.get("henvisning") or ""):
                by_ref[f"innst. {m.group(1)} {m.group(2).lower()}|{m.group(3)}-{m.group(4)}"].append(
                    {"session": ses, "sak": s}
                )
            for e in s.get("emne_liste") or []:
                eid = e.get("hovedemne_id") if not e.get("er_hovedemne") else e.get("id")
                if eid:
                    by_emne[eid].append((ses, s))
    print(f"indexed {total} saker across {len(ses_list)} sessions "
          f"({len(by_ref)} innstilling refs, {len(by_emne)} emner)", file=sys.stderr)
    return by_ref, by_emne


def load_content() -> tuple[list[dict], list[dict]]:
    statements = [yaml.safe_load(p.read_text()) for p in sorted((CONTENT / "statements").glob("*.yaml"))]
    evidence = [yaml.safe_load(p.read_text()) for p in sorted((CONTENT / "evidence").glob("*.yaml"))]
    return statements, evidence


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--statement", action="append", default=[], help="limit to one statement id (repeatable)")
    ap.add_argument("--min-score", type=int, default=2, help="minimum shared title tokens for a candidate")
    ap.add_argument("--top", type=int, default=6, help="candidates kept per statement")
    ap.add_argument("--first-year", type=int, default=FIRST_YEAR)
    ap.add_argument("--resume", action="store_true",
                    help="keep the candidates already in today's candidates.json and skip those statements")
    args = ap.parse_args(argv)

    statements, evidence = load_content()
    statements = [s for s in statements if not args.statement or s["id"] in args.statement]
    keep = {s["id"] for s in statements}

    # cell -> the innstilling refs it already rests on, and the session of each.
    cells: dict[tuple[str, str], list[str]] = collections.defaultdict(list)
    for e in evidence:
        if e.get("statement_id") not in keep or e.get("position") is None:
            continue
        for m in INNST.finditer(f"{e.get('source_name') or ''} {e.get('claim_action') or ''}"):
            ref = f"innst. {m.group(1)} {m.group(2).lower()}|{m.group(3)}-{m.group(4)}"
            if ref not in cells[(e["statement_id"], e["party_id"])]:
                cells[(e["statement_id"], e["party_id"])].append(ref)

    by_ref, by_emne = build_index(sessions(args.first_year))

    # statement -> emner its own coded evidence deals with, and the earliest
    # session that evidence covers (we walk strictly before it).
    topics: dict[str, set[int]] = collections.defaultdict(set)
    earliest: dict[str, str] = {}
    resolved = unresolved = 0
    for (sid, _party), refs in cells.items():
        for ref in refs:
            hits = by_ref.get(ref) or []
            if not hits:
                unresolved += 1
                continue
            resolved += 1
            for h in hits:
                for e in h["sak"].get("emne_liste") or []:
                    eid = e.get("hovedemne_id") if not e.get("er_hovedemne") else e.get("id")
                    if eid:
                        topics[sid].add(eid)
                ses = ref.split("|")[1]
                if sid not in earliest or ses < earliest[sid]:
                    earliest[sid] = ses
    print(f"resolved {resolved} innstilling refs to saker, {unresolved} did not resolve", file=sys.stderr)
    if not topics:
        print("no statement had a resolvable innstilling reference: nothing to search with", file=sys.stderr)
        return 2

    today = datetime.date.today().isoformat()
    dest = REPORTS / today
    dest.mkdir(parents=True, exist_ok=True)

    out: list[dict] = []
    if args.resume and (dest / "candidates.json").exists():
        out = json.loads((dest / "candidates.json").read_text())
    done = {e["statement"] for e in out}

    def flush() -> None:
        (dest / "candidates.json").write_text(json.dumps(out, ensure_ascii=False, indent=2))

    for st in statements:
        sid = st["id"]
        if sid not in topics or sid in done:
            continue
        want = tokens(st.get("text", "")) | tokens(st.get("short_name", "")) | tokens(st.get("primary_question", ""))
        cutoff = earliest.get(sid, "9999")
        seen, cands = set(), []
        for eid in topics[sid]:
            for ses, sak in by_emne.get(eid, []):
                if ses >= cutoff:
                    continue  # we want a strictly earlier observation
                key = sak["id"]
                if key in seen:
                    continue
                seen.add(key)
                title = " ".join(filter(None, [sak.get("korttittel"), sak.get("tittel")]))
                score = len(want & tokens(title))
                if score >= args.min_score:
                    cands.append((score, ses, sak))
        cands.sort(key=lambda t: (-t[0], t[1]))
        cands = cands[: args.top]
        # A statement with no candidate is a RESULT, not an omission: dropping
        # it lets a later reader believe it was never searched. It is appended
        # with an empty list, and the report says so in words.

        rows = []
        for score, ses, sak in cands:
            try:
                vo = V.fetch(V.VOTERINGER.format(sak["id"])).get("sak_votering_liste") or []
            except Exception as exc:  # noqa: BLE001 - one bad case must not stop the run
                rows.append({"score": score, "session": ses, "sak_id": sak["id"],
                             "title": sak.get("korttittel"), "error": str(exc)[:120], "votes": []})
                continue
            vrows = []
            for v in vo[:MAX_VOTES]:
                vid = v.get("votering_id")
                # The forslag text and the tally live behind their own
                # endpoints, exactly as votes.py reads them. voteringer?sakid
                # carries neither.
                try:
                    fl = V.fetch(V.FORSLAG.format(vid)).get("voteringsforslag_liste") or []
                    forslag = " | ".join(
                        x for x in (V.strip_html(y.get("forslag_tekst")) for y in fl) if x
                    ) or V.strip_html(v.get("votering_tema"))
                    resultat = V.fetch(V.RESULTAT.format(vid)).get("voteringsresultat_liste") or []
                except Exception as exc:  # noqa: BLE001 - one bad vote must not stop the run
                    vrows.append({"votering_id": vid, "error": str(exc)[:120]})
                    continue
                vrows.append({"votering_id": vid, "vedtatt": v.get("vedtatt"),
                              "resultat_type": v.get("votering_resultat_type_tekst"),
                              "forslag": forslag[:400], "tally": V.tally(resultat)})
            rows.append({"score": score, "session": ses, "sak_id": sak["id"],
                         "title": sak.get("korttittel"), "henvisning": sak.get("henvisning"),
                         "votes": vrows})
        out.append({"statement": sid, "number": st.get("number"), "text": st.get("text"),
                    "cutoff_session": cutoff, "candidates": rows})
        flush()
        print(f"[{len(out)}/{len(topics)}] {sid}: {len(rows)} candidate cases", file=sys.stderr, flush=True)

    flush()

    lines = [f"# Historical vote candidates, {today}", "",
             f"Statements searched: {len(out)}. Rules: same emne as the statement's own coded",
             f"evidence, in a session before {earliest and min(earliest.values())}, title sharing at least",
             f"{args.min_score} meaningful words. Nothing here is evidence until a human opens the case.", ""]
    for entry in out:
        lines.append(f"## {entry['statement']} — {entry['text'].strip()[:120]}")
        lines.append("")
        lines.append(f"Own evidence earliest session: {entry['cutoff_session']}")
        lines.append("")
        if not entry["candidates"]:
            lines.append("No candidate case in an earlier session matched. That is not proof that no "
                         "earlier vote exists: matching is on emne id plus title words, and a real "
                         "case can be titled with none of the statement's words.")
            lines.append("")
        for c in entry["candidates"]:
            lines.append(f"- **sak {c['sak_id']}** ({c['session']}, score {c['score']}) — {c['title']}")
            if c.get("henvisning"):
                lines.append(f"  - {c['henvisning']}")
            for v in c.get("votes", []):
                if v.get("error"):
                    lines.append(f"  - votering {v['votering_id']} — fetch error: {v['error']}")
                    continue
                t = v.get("tally") or {}
                summ = ", ".join(
                    f"{p}:{d['for']}f/{d['mot']}m" for p, d in sorted(t.items()) if d["for"] or d["mot"]
                ) or "no per-party rows"
                lines.append(f"  - votering {v['votering_id']} vedtatt={v.get('vedtatt')} — {v['forslag'][:200]}")
                lines.append(f"    - {summ}")
        lines.append("")
    (dest / "REPORT.md").write_text("\n".join(lines))
    print(f"wrote {dest}/REPORT.md ({len(out)} statements with candidates)")
    return 0 if out else 2


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
