#!/usr/bin/env python3
"""Build the hearing-submissions index: which organisations are heard, and on what.

This is lane L's access dataset. The question it answers is not "what does the
organisation think" but "who gets a seat at the table, on which case, and when".

Sources, both proven on 2026-11-09:
  horinger?sesjonid=YYYY-YYYY        the hearings, committee and case
  horingsinnspill?horingid=<id>      the written submissions, one per organisation

Only `organisasjon`, `dato` and the hearing/case keys are stored. The `tekst`
field is free text that can name individuals and is never stored.

    uv run tools/analytics/build_hearings_index.py [--session 2026-2027]

Writes internal/data/hearings/index.json. Derived reference data, no position.
"""

from __future__ import annotations

import argparse
import collections
import datetime
import json
import pathlib
import re
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[2] / "tools/autopilot"))
import votes as V  # noqa: E402 - reuse the cached, backoff-aware fetch

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / "internal/data/hearings/index.json"
DATE = re.compile(r"/Date\((\d+)")


def iso(raw: str) -> str:
    m = DATE.search(raw or "")
    if not m:
        return ""
    return datetime.datetime.fromtimestamp(
        int(m.group(1)) / 1000, datetime.timezone(datetime.timedelta(hours=1))
    ).date().isoformat()


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--session", default="2026-2027")
    args = ap.parse_args(argv)

    hearings = V.fetch(V.BASE + f"/horinger?format=json&sesjonid={args.session}").get("horinger_liste") or []
    if not hearings:
        print(f"no hearings for {args.session}; wrote nothing", file=sys.stderr)
        return 2

    rows, org_count = [], collections.Counter()
    for i, h in enumerate(hearings, 1):
        hid = h.get("id")
        case = (h.get("horing_sak_info_liste") or [{}])[0]
        subs, err = [], ""
        try:
            d = V.fetch(V.BASE + f"/horingsinnspill?horingid={hid}&format=json")
            for s in d.get("horingsinnspill_liste") or []:
                org = (s.get("organisasjon") or "").strip()
                subs.append({"organisasjon": org, "dato": iso(s.get("dato") or "")})
                if org:
                    org_count[org] += 1
        except Exception as e:  # noqa: BLE001 - one hearing must not stop the run
            err = f"{type(e).__name__}: {e}"[:120]
        rows.append({
            "hearing_id": hid,
            "komite": (h.get("komite") or {}).get("navn") or "",
            "komite_id": (h.get("komite") or {}).get("id") or "",
            "status": h.get("horing_status") or "",
            "start_dato": iso(h.get("start_dato") or ""),
            "sak_id": case.get("sak_id"),
            "sak_tittel": case.get("sak_korttittel") or case.get("sak_tittel") or "",
            "n_submissions": len(subs),
            "organisasjoner": subs,
            **({"error": err} if err else {}),
        })
        if i % 50 == 0:
            print(f"  {i}/{len(hearings)} hearings", file=sys.stderr, flush=True)

    doc = {
        "generated_at": datetime.date.today().isoformat(),
        "session": args.session,
        "hearings": rows,
        "top_organisasjoner": [
            {"organisasjon": o, "hearings": n} for o, n in org_count.most_common(50)
        ],
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(doc, ensure_ascii=False, indent=1))

    total_subs = sum(r["n_submissions"] for r in rows)
    with_subs = sum(1 for r in rows if r["n_submissions"])
    print(f"hearings {len(rows)}, with submissions {with_subs}, submissions {total_subs}, "
          f"distinct organisations {len(org_count)}")
    print(f"wrote {OUT.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
