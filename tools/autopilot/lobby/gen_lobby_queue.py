# /// script
# requires-python = ">=3.11"
# dependencies = ["pyyaml"]
# ///
"""Generate the lane-L queue: one item per person in public office.

The queue carries only the Stortinget person id, party and constituency. It
carries no name, no birth date and no role. The record with names is written
later, by the tracker, and it is the thing the gate and the public page show.

    uv run tools/autopilot/lobby/gen_lobby_queue.py

Idempotent: a status already recorded for the same id is carried forward.
Offline: the roster fetch fails, the script says so and writes nothing. It does
not fall back to a remembered list.
"""

from __future__ import annotations

import json
import pathlib
import sys
import urllib.request

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[3]
OUT = pathlib.Path(__file__).resolve().parent / "queue.yaml"
ROSTER = "https://data.stortinget.no/eksport/representanter?format=json&sesjonid={}"

# The API speaks its own party codes (A, KrF, SV); the dataset speaks ids
# (ap, krf, sv). The queue must use the project's vocabulary, not the API's.
API_TO_ID = {"A": "ap", "FrP": "frp", "H": "h", "SV": "sv", "Sp": "sp",
             "R": "rodt", "MDG": "mdg", "KrF": "krf", "V": "v"}


def newest_session() -> str:
    import datetime

    today = datetime.date.today()
    start = today.year if today.month >= 10 else today.year - 1
    return f"{start}-{start + 1}"


def fetch(session: str) -> list[dict]:
    req = urllib.request.Request(ROSTER.format(session), headers={"User-Agent": "ektevalgomat-lobby/1"})
    with urllib.request.urlopen(req, timeout=30) as r:
        d = json.load(r)
    return d.get("representanter_liste") or []


def carried() -> dict[str, tuple[str, dict]]:
    if not OUT.exists():
        return {}
    try:
        doc = yaml.safe_load(OUT.read_text()) or {}
    except yaml.YAMLError:
        return {}
    return {it["id"]: (it.get("status") or "pending", it)
            for it in doc.get("items") or [] if isinstance(it, dict) and it.get("id")}


def main() -> int:
    session = newest_session()
    try:
        rows = fetch(session)
    except Exception as e:  # noqa: BLE001 - offline is a real result
        print(f"roster fetch failed for {session}: {type(e).__name__}: {e}", file=sys.stderr)
        print("wrote nothing; run again online", file=sys.stderr)
        return 2
    if not rows:
        print(f"roster for {session} was empty; wrote nothing", file=sys.stderr)
        return 2

    prev = carried()
    items = []
    for r in rows:
        pid = r.get("id")
        if not pid:
            continue
        status, old = prev.get(pid, ("pending", {}))
        party = API_TO_ID.get((r.get("parti") or {}).get("id") or "", "")
        fylke = (r.get("fylke") or {}).get("id") or ""
        item = {
            "id": pid,
            "lane": "lobby",
            "scope": "representant",
            "party_id": party,
            "fylke": fylke,
            "method": "method/person.md",
            "status": status,
        }
        for k in ("note", "result", "blocked_on"):
            if old.get(k):
                item[k] = old[k]
        items.append(item)
    items.sort(key=lambda i: i["id"])

    doc = {
        "generated_for_session": session,
        "rule": "one item per sitting representative; id only, no name",
        "items": items,
    }
    OUT.write_text(yaml.safe_dump(doc, allow_unicode=True, sort_keys=False))
    pending = sum(1 for i in items if i["status"] == "pending")
    print(f"session {session}: {len(items)} representatives -> {OUT.name}; "
          f"{pending} pending, {len(items) - pending} carried")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
