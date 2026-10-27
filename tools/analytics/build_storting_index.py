#!/usr/bin/env python3
"""Build the Storting index: who holds the seats, and which committees exist.

Source: data.stortinget.no, the Storting's own open data.

Two endpoints answer, one does not, and the difference is the point of this file:

  representanter?sesjonid=...  works. 169 representatives with party and county.
  komiteer?sesjonid=...        works. 16 committees, id and name.
  committee MEMBERS            does not. Every plausible endpoint name answers 404
                               and /eksport/komite answers 400 with "unknown
                               parameter", so the per-party distribution of
                               committee seats and chair posts cannot be collected
                               from here. That refusal is recorded in the index
                               with the attempts, so the next run does not repeat
                               the search from scratch.

What the index therefore answers: how many seats each party holds, where those
seats are, and which committees exist. What it cannot answer: who sits on them.

Seat counts are derived, never written down: the index stores the representatives
and the loader counts them, so a seat total cannot drift from the list it describes.
"""

from __future__ import annotations

import datetime
import json
import pathlib
import sys
import urllib.request

API = "https://data.stortinget.no/eksport"
SESSION = "2025-2026"
OUT = pathlib.Path("internal/data/storting/index.json")

# The API names parties in Norwegian; the dataset names them by id. An unmapped
# party is a failure: a representative who cannot be attributed is a seat that
# disappears from every count.
PARTIES = {
    "Arbeiderpartiet": "ap",
    "Fremskrittspartiet": "frp",
    "Høyre": "h",
    "Sosialistisk Venstreparti": "sv",
    "Senterpartiet": "sp",
    "Rødt": "rodt",
    "Miljøpartiet De Grønne": "mdg",
    "Kristelig Folkeparti": "krf",
    "Venstre": "v",
}

COMMITTEE_ATTEMPTS = [
    "komiteamedlemmer?komiteid=ENERGI&sesjonid=2025-2026 -> 404",
    "komitemedlemmer?komiteid=ENERGI&sesjonid=2025-2026 -> 404",
    "komitemedlem?komiteid=ENERGI&sesjonid=2025-2026 -> 404",
    "komite_medlemmer?komiteid=ENERGI&sesjonid=2025-2026 -> 404",
    "komiteamedlem?komiteid=ENERGI&sesjonid=2025-2026 -> 404",
    "komiteer_medlemmer?komiteid=ENERGI&sesjonid=2025-2026 -> 404",
    "komite?komiteid=ENERGI&sesjonid=2025-2026 -> 400 (unknown parameter)",
    "komite?id=ENERGI / komite_id=ENERGI / sesjonid first -> 400",
]


def get(path: str) -> dict:
    url = f"{API}/{path}"
    req = urllib.request.Request(url, headers={"User-Agent": "ektevalgomat-storting/1"})
    with urllib.request.urlopen(req, timeout=60) as r:
        return json.loads(r.read().decode())


# A vote where the whole Storting was recorded, used to cross-check the seat counts.
# The per-party row totals on a votering (for + mot + ikke til stede) are the number
# of representatives that party had at that moment, so they must equal the counts
# from the representative list. Two artefacts of the same authority, independently
# produced, agreeing is a real check; one endpoint agreeing with itself is not.
CROSS_CHECK_VOTERING = 27721
VOTE_CODES = {2: "for", 3: "mot", 1: "ikke_tilstede"}


def party_totals_from_vote(votering_id: int) -> dict[str, int]:
    d = get(f"voteringsresultat?voteringid={votering_id}&format=json")
    totals: dict[str, int] = {}
    for row in d.get("voteringsresultat_liste") or []:
        code = ((row.get("representant") or {}).get("parti") or {}).get("id")
        if not code:
            continue
        totals[code] = totals.get(code, 0) + 1
    return totals


def main() -> int:
    reps = get(f"representanter?sesjonid={SESSION}&format=json").get("representanter_liste") or []
    committees = get(f"komiteer?sesjonid={SESSION}&format=json").get("komiteer_liste") or []
    if not reps:
        print("no representatives returned: refusing to write", file=sys.stderr)
        return 2
    if not committees:
        print("no committees returned: refusing to write", file=sys.stderr)
        return 2

    rows, unmapped = [], set()
    for r in reps:
        # parti and fylke are objects, not strings. Reading them as strings is how
        # the first version of this file crashed on a dict.
        party = r.get("parti") or {}
        county = r.get("fylke") or {}
        name = party.get("navn") or ""
        pid = PARTIES.get(name)
        if pid is None:
            unmapped.add(name or party.get("id") or "")
            pid = ""
        rows.append(
            {
                "id": r.get("id"),
                "first_name": r.get("fornavn"),
                "last_name": r.get("etternavn"),
                "party_name": name,
                "party_code": party.get("id") or "",
                "party_id": pid,
                "county": county.get("navn") or "",
                "county_code": county.get("id") or "",
                "deputy": bool(r.get("vara_representant")),
            }
        )
    if unmapped:
        print(f"unmapped party names: {sorted(unmapped)}", file=sys.stderr)
        return 1

    counts: dict[str, int] = {}
    codes: dict[str, str] = {}
    for row in rows:
        counts[row["party_id"]] = counts.get(row["party_id"], 0) + 1
        codes[row["party_id"]] = row["party_code"]

    vote_totals = party_totals_from_vote(CROSS_CHECK_VOTERING)
    disagreements = [
        {"party": pid, "seats": counts[pid], "vote_total": vote_totals.get(code)}
        for pid, code in sorted(codes.items())
        if vote_totals.get(code) != counts[pid]
    ]
    if disagreements:
        print(f"seat counts disagree with votering {CROSS_CHECK_VOTERING}: {disagreements}", file=sys.stderr)
        return 1

    doc = {
        "source": {
            "name": "Stortinget, åpne data",
            "api": API,
            "human_url": "https://data.stortinget.no/",
            "session": SESSION,
            "retrieved_at": datetime.date.today().isoformat(),
            "note": (
                "Representatives and committees from the Storting's open data. Seat counts are "
                "derived from the representative list, never stored. Committee membership is NOT "
                "included: no endpoint answers for it, and the attempts are recorded under refused."
            ),
        },
        "parties": PARTIES,
        "representatives": sorted(rows, key=lambda r: (r["party_id"], r["last_name"] or "", r["first_name"] or "")),
        "committees": [{"id": c.get("id"), "name": c.get("navn")} for c in committees],
        "cross_check": {
            "votering_id": CROSS_CHECK_VOTERING,
            "method": (
                "Per-party row totals on the votering (for + mot + ikke til stede) must equal the "
                "seat counts derived from the representative list. The vote record and the "
                "representative list are produced separately, so agreement is evidence."
            ),
            "per_party_vote_totals": vote_totals,
        },
        "refused": {
            "committee_membership": {
                "reason": (
                    "No open-data endpoint returns which representatives sit on which committee. "
                    "Every candidate name answers 404 and the singular form answers 400 with an "
                    "unknown-parameter error, so the per-party distribution of committee seats and "
                    "chair posts cannot be collected from this source."
                ),
                "attempts": COMMITTEE_ATTEMPTS,
            }
        },
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n")

    print(f"wrote {OUT}: {len(rows)} representatives, {len(committees)} committees, session {SESSION}")
    print("seats per party: " + ", ".join(f"{p} {counts[p]}" for p in sorted(counts, key=lambda x: -counts[x])))
    print(f"committee membership refused; {len(COMMITTEE_ATTEMPTS)} attempts recorded")
    return 0


if __name__ == "__main__":
    sys.exit(main())
