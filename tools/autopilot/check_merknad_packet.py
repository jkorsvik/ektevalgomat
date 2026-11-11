"""Check a merknad packet against the four conditions that make it admissible.

A packet is one candidate row: a PDF URL, a party (or a named coalition), an
exact excerpt, and the sak it came from. This script answers four questions and
nothing else. It never prints the excerpt, the document text or a title, so it
is safe to run on a model that must not see people or source prose.

  1. Is the excerpt an exact substring of the text the gate reads?
     The gate reads tools/verify/verify_excerpts.py:to_text, so a quote taken
     with pdftotext is a different string and cannot be archived.
  2. Is the padded archive window under 6,000 characters?
  3. Does the date come from saksgang, and is it the committee's date rather
     than the vote's?
  4. Is the party one canonical id, or an explicitly named coalition of them?

A pass here is not an admission. Admission is a semantic judgement about
whether the excerpt states the cell's position in its own words.
"""

import argparse
import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "tools" / "verify"))
sys.path.insert(0, str(ROOT / "tools" / "archive"))
sys.path.insert(0, str(ROOT / "tools" / "autopilot"))

import verify_excerpts as ve  # noqa: E402
import build_archive as ba  # noqa: E402

CANONICAL = {"ap", "frp", "h", "krf", "mdg", "rodt", "sp", "sv", "v"}
WINDOW_MAX = 6000

# Party label as the document writes it -> dataset id. A coalition heading is
# split on " og " and each half must resolve, which is why this is a lookup and
# not a fuzzy match.
LABEL_TO_ID = {
    "arbeiderpartiet": "ap",
    "fremskrittspartiet": "frp",
    "hoyre": "h",
    "høyre": "h",
    "kristelig folkeparti": "krf",
    "miljopartiet de gronne": "mdg",
    "miljøpartiet de grønne": "mdg",
    "rodt": "rodt",
    "rødt": "rodt",
    "senterpartiet": "sp",
    "sosialistisk venstreparti": "sv",
    "venstre": "v",
}


def norm_label(s):
    s = re.sub(r"[-\u00ad]\s*", "", str(s or ""))       # hyphen wrap
    s = re.sub(r"\s+", " ", s).strip().lower()
    return s


def resolve_part(part):
    """One heading part as a canonical id, or None.

    A part may already BE a canonical id: `sv` and `mdg` are how the rest of the
    codebase names those parties, so a packet that speaks the dataset's language
    must not be rejected for not speaking the document's. Checked before the
    label map because `h` is both an id and the start of a longer label.
    """
    if part in CANONICAL:
        return part
    return LABEL_TO_ID.get(part)


def parties_of(label):
    """Return the canonical ids a heading names, or None if one does not resolve."""
    whole = norm_label(label)
    if whole in CANONICAL or whole in LABEL_TO_ID:
        return [resolve_part(whole)]
    parts = re.split(r"\s+og\s+", whole)
    ids = []
    for part in parts:
        pid = resolve_part(part)
        if pid is None:
            return None
        ids.append(pid)
    return ids if ids else None


def saksgang_dates(sak_body):
    """(komitebehandling date, vote date) as ISO strings, or (None, None)."""
    d = json.loads(sak_body)
    komite = vote = None
    for step in d.get("saksgang", {}).get("saksgang_steg_liste") or []:
        name = str(step.get("navn") or "").lower()
        for h in step.get("saksgang_hendelse_liste") or []:
            raw = str(h.get("dato") or "")
            m = re.match(r"(\d{2})\.(\d{2})\.(\d{4})", raw)
            if not m or m.group(3) == "0001":       # null sentinel, not a date
                continue
            iso = f"{m.group(3)}-{m.group(2)}-{m.group(1)}"
            if "komite" in name:
                komite = iso
            elif "debatt" in name or "vedtak" in name:
                vote = iso
    return komite, vote


def check(packet, sak_body=None):
    out = {"packet": packet.get("id") or packet.get("statement"), "checks": {}, "ok": True}

    def put(name, ok, detail=""):
        out["checks"][name] = {"ok": bool(ok), "detail": detail}
        if not ok:
            out["ok"] = False

    url = packet.get("source_url") or ""
    excerpt = packet.get("excerpt") or ""

    ids = parties_of(packet.get("party_label") or packet.get("party") or "")
    put("party", ids is not None,
        "canonical %s" % ",".join(ids) if ids else "heading does not resolve to canonical ids")

    if not url.lower().endswith(".pdf"):
        put("excerpt_is_substring", False, "source is not a PDF; this check is PDF only")
        put("window_under_limit", False, "not measured")
    else:
        text = ve.to_text(ve.fetch(url))
        spans = ba.locate(text, excerpt)
        put("excerpt_is_substring", spans is not None,
            "%d span(s)" % len(spans) if spans else "not found in to_text")
        if spans:
            # Two numbers, because they differ by the whitespace each edge
            # carries: build_archive strips the slice before writing it, so the
            # raw span is the wider of the two. Report both rather than pick the
            # flattering one.
            raw_window = sum(e - s for s, e in spans)
            stripped = sum(len(text[s:e].strip()) for s, e in spans)
            put("window_under_limit", raw_window <= WINDOW_MAX,
                "%d chars raw, %d stripped (limit %d)"
                % (raw_window, stripped, WINDOW_MAX))
        else:
            put("window_under_limit", False, "cannot measure a quote that does not locate")

    if sak_body is None:
        put("date_from_saksgang", False, "no sak payload supplied")
    else:
        komite, vote = saksgang_dates(sak_body)
        claimed = str(packet.get("date") or "")
        put("date_from_saksgang", bool(komite) and claimed == komite,
            "saksgang %s, vote %s, packet says %s" % (komite, vote, claimed or "-"))

    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("packets", nargs="+", help="YAML files with url, excerpt, party, date")
    ap.add_argument("--sakid", type=int, help="sak id whose saksgang carries the date")
    ap.add_argument("--json", action="store_true")
    a = ap.parse_args()

    import yaml
    body = None
    if a.sakid:
        body = ve.fetch("https://data.stortinget.no/eksport/sak?sakid=%d&format=json" % a.sakid)

    results = []
    for p in a.packets:
        packet = yaml.safe_load(pathlib.Path(p).read_text(encoding="utf-8"))
        results.append(check(packet, body))

    if a.json:
        print(json.dumps(results, ensure_ascii=False, indent=2))
    else:
        for r in results:
            print("%s %s" % ("PASS" if r["ok"] else "HOLD", r["packet"]))
            for name, c in r["checks"].items():
                print("   %-22s %s  %s" % (name, "ok" if c["ok"] else "NO", c["detail"]))
        print("\n%d packets, %d pass, %d hold"
              % (len(results), sum(1 for r in results if r["ok"]), sum(1 for r in results if not r["ok"])))
    return 0 if all(r["ok"] for r in results) else 1


if __name__ == "__main__":
    sys.exit(main())
