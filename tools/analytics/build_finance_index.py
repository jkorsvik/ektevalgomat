#!/usr/bin/env python3
"""Build the party-finance index from SSB table 11233.

This is reference data, not evidence: it carries no position and says nothing
about what a party wants. It answers the money half of the god's-eye view, "who
paid for the parties, and how much came from the state", with the state's own
statistics rather than with documents about them.

Source: SSB 11233 "Partifinansiering, etter regnskapspost, politisk parti,
statistikkvariabel og år". The table is annual, per party, per accounting post,
in kroner, and covers 2005 onwards. It is fetched over the open API, so the index
can be rebuilt and the numbers re-derived at any time; the retrieved_at date is
recorded so a stale index is visible as stale.

What this index deliberately does NOT contain: anonymous donations. The table has
no such field. Its nearest is A.3.5 "Bidrag fra andre, ikke kategoriserte givere",
which is a residual, not an anonymity count, and reading it as one would be a
fabrication. Anonymous and intermediary-routed contributions are the Partiloven
register's business, and that is a different source with a different shape.

Reconciliation is the point of the fetch being more than a copy: A.1 + A.2 + A.3
must equal A for every party-year, and the five contribution subtypes must sum to
A.3. A row that does not reconcile is reported, not silently stored, because an
income composition that does not add up is not usable as a denominator.
"""

from __future__ import annotations

import datetime
import json
import pathlib
import sys
import urllib.request

API = "https://data.ssb.no/api/v0/no/table/11233"

# The nine parties the dataset scores. SSB's codes are not the dataset's ids, and
# the mapping is explicit so a missing party fails loudly instead of vanishing.
PARTIES = {
    "01": "ap",
    "02": "frp",
    "03": "h",
    "04": "krf",
    "05": "sp",
    "06": "sv",
    "07": "v",
    "08": "mdg",
    "55": "rodt",
}

# Accounting posts, with the meaning the table gives them. A.3.5 is kept but
# named for what it is.
POSTS = {
    "A": "Inntekter i alt",
    "A.1": "Offentlig støtte",
    "A.1.1": "Statlig støtte",
    "A.1.2": "Kommunal/fylkeskommunal støtte",
    "A.2": "Egen virksomhet i alt",
    "A.3": "Bidrag i alt",
    "A.3.1": "Bidrag fra privatpersoner",
    "A.3.2": "Bidrag fra kommersielle foretak",
    "A.3.3": "Bidrag fra organisasjoner i arbeidslivet",
    "A.3.4": "Bidrag fra andre organisasjoner",
    "A.3.5": "Bidrag fra andre, ikke kategoriserte givere",
}

OUT = pathlib.Path("internal/data/finance/index.json")

# The second source: Statsforvalteren's register of contributions reported under
# Partiloven, served behind partifinansiering.no. It answers a different question
# from SSB 11233 and must not be added to it: SSB reports the parties' accounts,
# the register reports individual contributions above the statutory threshold.
# Same kroner, two definitions, so they are stored as two blocks and never summed.
REGISTER = "https://partistatistikkclient.statsforvalteren.no/financing"
REGISTER_HUMAN = "https://www.partifinansiering.no/nb/arlige-bidrag/"

# The register names parties; the dataset names ids. An unmapped scored party is a
# failure, not a silent omission.
REGISTER_PARTIES = {
    "Arbeiderpartiet": "ap",
    "Høyre": "h",
    "Fremskrittspartiet": "frp",
    "Sosialistisk Venstreparti": "sv",
    "Rødt": "rodt",
    "Venstre": "v",
    "Senterpartiet": "sp",
    "Miljøpartiet De Grønne": "mdg",
    "Kristelig Folkeparti": "krf",
}


def fetch(query: dict) -> dict:
    req = urllib.request.Request(
        API,
        data=json.dumps(query).encode(),
        headers={"Content-Type": "application/json", "User-Agent": "ektevalgomat-finance/1"},
    )
    with urllib.request.urlopen(req, timeout=60) as r:
        return json.loads(r.read().decode())


def flatten(js: dict) -> list[dict]:
    """json-stat2 -> one dict per cell, in the order the payload declares."""
    dim_ids = js["id"]
    cats = []
    for dim in dim_ids:
        idx = js["dimension"][dim]["category"]["index"]
        cats.append([code for code, _ in sorted(idx.items(), key=lambda kv: kv[1])])
    rows = []
    import itertools

    for i, combo in enumerate(itertools.product(*cats)):
        rows.append({"cells": dict(zip(dim_ids, combo)), "value": js["value"][i]})
    return rows


def get_json(url: str) -> object:
    with urllib.request.urlopen(url, timeout=60) as r:
        return json.loads(r.read().decode())


def fetch_register() -> dict:
    """Per-party and total contributions per year, cross-checked against each other.

    The register publishes both a per-party breakdown and a yearly total. They must
    agree; a difference means one of the two endpoints is filtered differently, and
    storing the disagreement would make every derived share wrong without saying so.
    """
    totals = get_json(f"{REGISTER}/AnnualContributionsTotalPerYear")
    per_party = get_json(f"{REGISTER}/AnnualContributionsTotalPerYearForParties")
    if not totals or not per_party:
        raise RuntimeError("register returned no rows")

    by_year = {int(r["Year"]): int(r["Amount"]) for r in totals}
    parties, sums = [], {}
    for p in per_party:
        rows = [
            {"year": str(r["Year"]), "amount": int(r["Amount"])}
            for r in p.get("AnnualPartyContributions") or []
            if r.get("Amount") is not None
        ]
        for r in rows:
            sums[int(r["year"])] = sums.get(int(r["year"]), 0) + r["amount"]
        parties.append(
            {
                "name": p["PartyName"],
                "list_code": p.get("ListCode"),
                "party_id": REGISTER_PARTIES.get(p["PartyName"], ""),
                "by_year": sorted(rows, key=lambda r: r["year"]),
            }
        )

    mismatched = [
        {"year": str(y), "published": by_year[y], "sum_of_parties": sums.get(y, 0)}
        for y in sorted(by_year)
        if sums.get(y, 0) != by_year[y]
    ]
    if mismatched:
        print(f"register does not reconcile: {mismatched[:3]}", file=sys.stderr)

    missing = sorted(set(REGISTER_PARTIES.values()) - {p["party_id"] for p in parties})
    if missing:
        print(f"scored parties absent from the register: {missing}", file=sys.stderr)

    return {
        "source": {
            "name": "Statsforvalteren, register over bidrag etter partiloven",
            "api": REGISTER,
            "human_url": REGISTER_HUMAN,
            "retrieved_at": datetime.date.today().isoformat(),
            "unit": "kroner",
            "note": (
                "Contributions above the statutory threshold, as reported by the parties. "
                "Not the same measure as SSB 11233 A.3 and never added to it. "
                "Donor-level rows are not included: the register's paginated endpoint "
                "answers 500 for every parameter shape tried, so the site's own table "
                "could not be reproduced; this block holds the published aggregates."
            ),
        },
        "totals_by_year": [{"year": str(y), "amount": by_year[y]} for y in sorted(by_year)],
        "parties": sorted(parties, key=lambda p: p["name"]),
        "mismatched_years": mismatched,
        "missing_scored_parties": missing,
    }


def main() -> int:
    meta = json.loads(urllib.request.urlopen(API, timeout=60).read().decode())
    title = meta["title"]

    query = {
        "query": [
            {"code": "PolitParti", "selection": {"filter": "item", "values": list(PARTIES)}},
            {"code": "Regnskapsp", "selection": {"filter": "item", "values": list(POSTS)}},
            {"code": "ContentsCode", "selection": {"filter": "item", "values": ["Partifinans"]}},
            {"code": "Tid", "selection": {"filter": "all", "values": ["*"]}},
        ],
        "response": {"format": "json-stat2"},
    }
    js = fetch(query)
    rows = flatten(js)
    if not rows:
        print("no rows returned: refusing to write an empty index", file=sys.stderr)
        return 2

    years = sorted({r["cells"]["Tid"] for r in rows})
    # party-year -> post -> kroner
    table: dict[tuple[str, str], dict[str, int | None]] = {}
    for r in rows:
        c = r["cells"]
        table.setdefault((c["PolitParti"], c["Tid"]), {})[c["Regnskapsp"]] = r["value"]

    out_rows, missing_cells, unreconciled = [], 0, []
    for code, pid in PARTIES.items():
        for year in years:
            post = table.get((code, year))
            if post is None:
                missing_cells += 1
                continue
            if post.get("A") is None:
                # A party with no published accounts that year. Kept as absent,
                # never as zero: zero income is a claim, no accounts is a gap.
                missing_cells += 1
                continue
            parts = [post.get(k) for k in ("A.1", "A.2", "A.3")]
            if any(p is not None for p in parts):
                s = sum(p or 0 for p in parts)
                # SSB rounds to whole kroner; a residual of a few kroner is
                # rounding, a larger one is a post we are not reading.
                if abs(s - post["A"]) > 10:
                    unreconciled.append((pid, year, post["A"], s))
            subs = [post.get(k) for k in ("A.3.1", "A.3.2", "A.3.3", "A.3.4", "A.3.5")]
            if post.get("A.3") is not None and any(p is not None for p in subs):
                if abs(sum(p or 0 for p in subs) - post["A.3"]) > 10:
                    unreconciled.append((pid, year, f"A.3={post['A.3']}", sum(p or 0 for p in subs)))
            out_rows.append(
                {
                    "party": pid,
                    "year": year,
                    "income": post.get("A"),
                    "public_support": post.get("A.1"),
                    "state_support": post.get("A.1.1"),
                    "own_activity": post.get("A.2"),
                    "contributions": post.get("A.3"),
                    "from_private_persons": post.get("A.3.1"),
                    "from_companies": post.get("A.3.2"),
                    "from_worker_orgs": post.get("A.3.3"),
                    "from_other_orgs": post.get("A.3.4"),
                    "from_uncategorised": post.get("A.3.5"),
                }
            )

    if not out_rows:
        print("every party-year was empty: refusing to write", file=sys.stderr)
        return 2

    doc = {
        "source": {
            "table": "11233",
            "title": title,
            "api": API,
            "human_url": "https://www.ssb.no/statbank/table/11233",
            "retrieved_at": datetime.date.today().isoformat(),
            "unit": "kroner",
            "note": (
                "Annual party accounts as reported to SSB. No field measures anonymous "
                "donations; A.3.5 is a residual for uncategorised givers and must not be "
                "read as an anonymity count."
            ),
        },
        "posts": POSTS,
        "parties": {pid: {"ssb_code": code} for code, pid in PARTIES.items()},
        "years": years,
        "rows": out_rows,
    }
    doc["register"] = fetch_register()

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n")

    print(f"{title}")
    print(f"wrote {OUT}: {len(out_rows)} party-years, {len(years)} years ({years[0]}-{years[-1]})")
    print(f"absent party-years (no published accounts): {missing_cells}")
    if unreconciled:
        print(f"UNRECONCILED rows: {len(unreconciled)}", file=sys.stderr)
        for r in unreconciled[:10]:
            print("   ", r, file=sys.stderr)
        return 1
    # A sanity read on the thing the index exists for: state money, per party.
    latest = years[-1]
    for pid in PARTIES.values():
        r = next((x for x in out_rows if x["party"] == pid and x["year"] == latest), None)
        if r and r["state_support"] is not None:
            print(f"  {pid:5s} {latest}: income {r['income']:>12,}  state {r['state_support']:>11,}")
    reg = doc["register"]
    print(
        f"register: {len(reg['parties'])} parties, "
        f"{len(reg['totals_by_year'])} years, "
        f"{len(reg['mismatched_years'])} unreconciled years, "
        f"{len(reg['missing_scored_parties'])} scored parties missing"
    )
    return 1 if reg["mismatched_years"] or reg["missing_scored_parties"] else 0


if __name__ == "__main__":
    sys.exit(main())
