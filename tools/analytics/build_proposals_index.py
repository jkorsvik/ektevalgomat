#!/usr/bin/env python3
"""Build the representative-proposals index: what each party put on the table.

Source: data.stortinget.no, /eksport/saker for a session, filtered to the records
that carry proposers. A representative proposal (representantforslag, "Dokument 8")
is the cheapest way a party can state a demand in the Storting, so counting them per
party per session is a direct measure of initiative: not what a party promised in a
programme, and not how it voted, but what it chose to spend political capital on.

Three decisions the code makes explicit rather than silent:

  - A proposal is kept with every proposer, not attributed to one party. Most are
    co-signed, and picking the first signer would undercount the others.
  - A party the dataset does not score (Pasientfokus appears in this session) is kept
    with an empty party_id and reported. Dropping the proposal would delete a real
    initiative; inventing a party id would corrupt the counts.
  - The status code is stored as the source's own code. Its meaning is not guessed:
    the codes seen are 1, 2 and 5, and this file does not claim to know which is
    "rejected" or "in progress". A reader who needs that must consult the source.

Subjects (emne_liste) are often empty for these records, so the index does not build
a topic breakdown on them; it stores what is there and no more.
"""

from __future__ import annotations

import collections
import datetime
import json
import pathlib
import sys
import urllib.request

API = "https://data.stortinget.no/eksport"
SESSION = "2025-2026"
OUT = pathlib.Path("internal/data/proposals/index.json")

# The nine parties the dataset scores. Any other party is stored with an empty id.
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


def get(path: str) -> dict:
    req = urllib.request.Request(f"{API}/{path}", headers={"User-Agent": "ektevalgomat-proposals/1"})
    with urllib.request.urlopen(req, timeout=90) as r:
        return json.loads(r.read().decode())


def main() -> int:
    d = get(f"saker?sesjonid={SESSION}&dokumenttypeid=8&format=json")
    saker = d.get("saker_liste") or []
    if not saker:
        print("no saker returned: refusing to write", file=sys.stderr)
        return 2

    proposals, unscored = [], collections.Counter()
    for s in saker:
        proposers = s.get("forslagstiller_liste") or []
        if not proposers:
            # 301 of the 663 records are committee reports, delegation reports and
            # other business with no proposer. They are not party initiatives.
            continue
        people = []
        for p in proposers:
            party = p.get("parti") or {}
            name = party.get("navn") or ""
            pid = PARTIES.get(name, "")
            if name and not pid:
                unscored[name] += 1
            people.append(
                {
                    "id": p.get("id"),
                    "name": f"{p.get('fornavn') or ''} {p.get('etternavn') or ''}".strip(),
                    "party_name": name,
                    "party_id": pid,
                }
            )
        komite = s.get("komite") or {}
        proposals.append(
            {
                "id": str(s.get("id")),
                "reference": s.get("henvisning") or "",
                "title": s.get("korttittel") or s.get("tittel") or "",
                "status_code": s.get("status"),
                "committee_id": komite.get("id") or "",
                "committee": komite.get("navn") or "",
                "subjects": [e.get("navn") for e in (s.get("emne_liste") or []) if e.get("navn")],
                "proposers": people,
            }
        )

    if not proposals:
        print("no proposal with a proposer: refusing to write", file=sys.stderr)
        return 2

    per_party = collections.Counter()
    for p in proposals:
        for party in {x["party_id"] for x in p["proposers"] if x["party_id"]}:
            per_party[party] += 1
    missing = sorted(set(PARTIES.values()) - set(per_party))

    doc = {
        "source": {
            "name": "Stortinget, åpne data: saker med forslagstillere",
            "api": API,
            "human_url": "https://data.stortinget.no/",
            "session": SESSION,
            "retrieved_at": datetime.date.today().isoformat(),
            "note": (
                "Representative proposals and other business that carries named proposers. "
                "A proposal is attributed to every proposer's party, not to the first signer. "
                "status_code is the source's own code and is not interpreted here."
            ),
        },
        "parties": PARTIES,
        "proposals": sorted(proposals, key=lambda p: p["reference"]),
        "proposals_per_party": dict(sorted(per_party.items(), key=lambda kv: -kv[1])),
        "unscored_parties": dict(unscored),
        "scored_parties_without_proposals": missing,
        "records_without_proposers": len(saker) - len(proposals),
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n")

    print(f"wrote {OUT}: {len(proposals)} proposals with proposers, {len(saker)} records in the session")
    print("per party: " + ", ".join(f"{k} {v}" for k, v in doc["proposals_per_party"].items()))
    if unscored:
        print(f"unscored parties kept with empty id: {dict(unscored)}")
    if missing:
        print(f"scored parties with no proposal this session: {missing}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
