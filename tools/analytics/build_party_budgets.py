# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Read the party-budget tables out of the budget innstillinger.

Every committee's budget innstilling carries a table headed

    Kap. | Post | Formål | Prop. 1 S (2025–2026) | A, SV, Sp, R og MDG | FrP | H | SV | R | KrF | V

where the fourth column is the government's own proposal, the fifth is the
budsjettforlik that several parties stand behind jointly, and the rest are single
parties' alternative budgets. Each cell is an amount with that party's deviation
from the government in parentheses: "415 000(+415 000)", "81 500(0)". The unit is
stated in the section header, "(i tusen kroner)".

The column map is read from the header text, never assumed, because the columns
are not in a fixed order and the forlik's member list differs per document: a
guess by position silently attributes one party's money to another.

Every record is checked against the document's own arithmetic before it is
written: the per-row deviations in a column must add up to the deviation printed
on that column's "Sum utgifter" row, and netto must equal utgifter minus
inntekter. A table that fails is reported and not written.

Usage:
  uv run tools/analytics/build_party_budgets.py                 # 2025-2026, numbers 1-25
  uv run tools/analytics/build_party_budgets.py --numbers 2,8,12
"""

import argparse
import hashlib
import html
from html.parser import HTMLParser
import json
import pathlib
import re
import sys
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / "internal/data/budget"
CACHE = pathlib.Path.home() / ".cache/ektevalgomat-sources"
UA = "Mozilla/5.0 (ektevalgomat party-budget reader; +https://ektevalgomat.no)"

# The nine parties the dataset compares, keyed by how the tables spell them.
PARTY_NAME = {
    "a": "ap", "ap": "ap", "arbeiderpartiet": "ap",
    "sv": "sv", "sosialistisk venstreparti": "sv",
    "sp": "sp", "senterpartiet": "sp",
    "r": "r", "rødt": "r", "rodt": "r",
    "mdg": "mdg", "miljøpartiet de grønne": "mdg",
    "frp": "frp", "fremskrittspartiet": "frp",
    "h": "h", "høyre": "h", "hoyre": "h",
    "krf": "krf", "kristelig folkeparti": "krf",
    "v": "v", "venstre": "v",
}
# Some innstillinger head the section "Utgifter rammeområde 5 (i tusen kroner)"
# and others only "Utgifter (i tusen kroner)", with the area printed in the
# paragraph above the table. Requiring the area here dropped every row of the
# eight tables that use the short form, which is how rammeområde 1, 4, 5, 6, 8
# and 14 first looked empty.
SECTION = re.compile(r"^(Utgifter|Inntekter)\b", re.I)
# The total row is printed "Sum utgifter rammeområde 12" in most documents and only
# "Sum utgifter" in the ones that head their section without an area. Requiring the
# area left Sums empty for seven tables, and an empty Sums made every party cell on
# /budsjett render as a dash, which claims the party proposed nothing there. That is
# a false statement about a real proposal, so the area is optional here.
TOTAL = re.compile(r"^Sum\s+(utgifter|inntekter|netto)\b", re.I)
AMOUNT = re.compile(r"^(-?[\d][\d\s\u00a0]*)(?:\(([+-]?[\d\s\u00a0]*)\))?$")
# A third form: a party column that prints only its deviation, signed and without a
# parenthesised amount ("+209 000"). Innst. 3 S rammeområde 21 uses it in its total
# rows. Reading that as an absolute amount would report every party as unchanged
# there, so a leading sign with no parentheses is the deviation and no base.
SIGNED_ONLY = re.compile(r"^([+-][\d\s\u00a0]*)$")


def session_dir(session: str) -> tuple[str, str]:
    """("2025-2026") -> the URL's year pairs: ("2025-2026", "202526")."""
    first, second = session.split("-")
    return session, f"{first}{second[2:]}"


def innstilling_url(session: str, number: int) -> str:
    d, compact = session_dir(session)
    return (f"https://www.stortinget.no/no/Saker-og-publikasjoner/Publikasjoner/"
            f"Innstillinger/Stortinget/{d}/inns-{compact}-{number:03d}s/?all=true")


def full_document_url(url: str) -> str:
    """Only ?all=true is the whole document; the paged form truncates the tables."""
    if "all=true" in url:
        return url
    return url + ("&" if "?" in url else "?") + "all=true"


def fetch(url: str) -> str:
    url = full_document_url(url)
    CACHE.mkdir(parents=True, exist_ok=True)
    f = CACHE / hashlib.sha256(url.encode()).hexdigest()
    if f.exists() and f.stat().st_size > 0:
        return f.read_bytes().decode("utf-8", "replace")
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=60) as r:
        body = r.read()
    f.write_bytes(body)
    return body.decode("utf-8", "replace")


def norm(s: str) -> str:
    return re.sub(r"\s+", " ", html.unescape(re.sub(r"<[^>]+>", " ", s))).strip()


class _TableParser(HTMLParser):
    """Collect every table's own rows, keeping nested tables separate.

    A regex cannot do this. The innstillinger nest tables inside layout tables, and
    a non-greedy `<table>.*?</table>` ends at the inner table's close tag, mixing two
    tables' rows into one. That produced a party table whose "sum" row was the
    government-only table's four-cell total, and it was the row check that caught it:
    the SV column added to 529 000 000 against a printed +52 000 000.

    Colspan is expanded into empty cells so a row's indices line up with the header's,
    which is what makes the column map from the header usable on every row.
    """

    def __init__(self, line_starts: list[int]) -> None:
        super().__init__(convert_charrefs=True)
        self.line_starts = line_starts
        self.stack: list[dict] = []
        self.done: list[tuple[int, list[list[str]]]] = []
        self.cell: list[str] | None = None
        self.row: list[str] | None = None
        self.span = 1

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str]]) -> None:
        if tag == "table":
            line, col = self.getpos()
            offset = self.line_starts[line - 1] + col if line - 1 < len(self.line_starts) else 0
            self.stack.append({"rows": [], "offset": offset})
        elif tag == "tr" and self.stack:
            self.row = []
        elif tag in ("td", "th") and self.stack and self.row is not None:
            self.cell = []
            try:
                self.span = max(1, int(dict(attrs).get("colspan") or 1))
            except ValueError:
                self.span = 1
        elif tag == "br" and self.cell is not None:
            self.cell.append(" ")

    def handle_data(self, data: str) -> None:
        if self.cell is not None:
            self.cell.append(data)

    def handle_endtag(self, tag: str) -> None:
        if tag in ("td", "th") and self.cell is not None and self.row is not None:
            text = re.sub(r"\s+", " ", "".join(self.cell)).strip()
            self.row.append(text)
            self.row.extend([""] * (self.span - 1))
            self.cell = None
            self.span = 1
        elif tag == "tr" and self.row is not None and self.stack:
            if any(c for c in self.row):
                self.stack[-1]["rows"].append(self.row)
            self.row = None
        elif tag == "table" and self.stack:
            top = self.stack.pop()
            self.done.append((top["offset"], top["rows"]))


def tables_with_offsets(doc: str) -> list[tuple[int, list[list[str]]]]:
    """Every <table> with the byte offset where it starts, so a label printed above
    a table can be attributed to it."""
    starts, pos = [], 0
    for line in doc.splitlines(keepends=True):
        starts.append(pos)
        pos += len(line)
    p = _TableParser(starts)
    p.feed(doc)
    p.close()
    return [t for t in p.done if t[1]]


def tables(doc: str) -> list[list[list[str]]]:
    """Every <table> in the document as a list of rows of cell text."""
    return [rows for _, rows in tables_with_offsets(doc)]


def area_before(doc: str, offset: int) -> int | None:
    """The rammeområde a table belongs to, from the nearest mention above it.

    Innst. 6, 7, 10 and 16 S print the label in a paragraph before the table
    instead of in a cell, so a table-only search took those four documents' party
    tables for area-less and dropped them.
    """
    best = None
    for m in re.finditer(r"[Rr]ammeområde\s+(\d+)", doc[:offset]):
        best = int(m.group(1))
    return best


def party_of(label: str) -> list[str]:
    """The parties a column header names, in the header's own order."""
    parts = re.split(r",|\s+og\s+", label)
    found = []
    for p in parts:
        p = p.strip()
        if not p:
            continue
        pid = PARTY_NAME.get(p.lower())
        if pid and pid not in found:
            found.append(pid)
    return found


def column_map(header: list[str]) -> dict | None:
    """col index -> role, or None when this is not a party-budget table."""
    if len(header) < 6:
        return None
    if not any("Prop" in c for c in header):
        return None
    # A bevilgning table has a Kap. and a Post column. Innst. 2 S carries a
    # per-committee summary headed "Komité | Nr." whose columns also name parties;
    # reading it as a post table attributed a committee summary to rammeområde 21.
    if not any(c.strip().lower().startswith("kap") for c in header):
        return None
    roles = {}
    parties = 0
    for i, cell in enumerate(header):
        if "Prop" in cell:
            roles[i] = {"kind": "government"}
            continue
        members = party_of(cell)
        if not members:
            continue
        if len(members) > 1:
            roles[i] = {"kind": "forlik", "parties": members}
        else:
            roles[i] = {"kind": "party", "party": members[0]}
            parties += 1
    return roles if parties >= 3 else None


def amount(cell: str, deviation_ok: bool = False) -> tuple[int, int] | None:
    """'415 000(+415 000)' -> (415_000_000, 415_000_000) in kroner; '81 500' -> (81_500_000, 0).

    A base amount can be negative: an inntekt post carried as a cost, and any
    "Sum netto" row, read like "-23 176(-477 000)". Requiring a digit first dropped
    every such cell and with it the whole row, which is what made the SV column of
    Innst. 9 S rammeområde 12 add to 529 000 000 against a printed +52 000 000.
    """
    cell = cell.replace("\u00a0", " ").replace("\u2212", "-").strip()
    # Only a party or forlik column may print a bare deviation. The government
    # column always prints an absolute amount, and reading its "-397 642" as a
    # change made Innst. 16 S rammeområde 1 add to -397 642 000 against a sum of 0.
    so = SIGNED_ONLY.match(cell) if deviation_ok else None
    if so:
        raw = so.group(1).replace(" ", "")
        return 0, int(raw) * 1000
    m = AMOUNT.match(cell)
    if not m:
        return None
    digits = re.sub(r"\D", "", m.group(1)) or "0"
    base = int(digits) * (-1 if m.group(1).strip().startswith("-") else 1)
    change = 0
    if m.group(2) is not None:
        raw = m.group(2).replace(" ", "")
        change = int(raw) if raw not in ("", "+", "-") else 0
    return base * 1000, change * 1000


def extract(rows: list[list[str]], roles: dict) -> dict:
    """One table -> a record with sections, party totals and the sum rows."""
    unit = None
    sections: dict[str, list] = {}
    sums: dict[str, dict] = {}
    current = None
    for cells in rows:
        joined = " ".join(cells)
        sm = SECTION.search(cells[0] if cells else "")
        if sm:
            current = sm.group(1).lower()
            sections.setdefault(current, [])
            u = re.search(r"\(i\s*([^)]+)\)", joined)
            if u:
                unit = u.group(1)
            continue
        label = cells[2] if len(cells) > 2 else ""
        tm = TOTAL.search(label) if len(label) < 60 else None
        if tm:
            kind = tm.group(1).lower()
            line = {}
            for i, role in roles.items():
                if i < len(cells):
                    a = amount(cells[i], role["kind"] != "government")
                    if a:
                        line[i] = a
            sums[kind] = line
            continue
        if current is None or len(cells) < 4:
            continue
        # A bevilgning always carries a post number. A row without one is a heading
        # or a breakdown of the row above it, and its amount restates that row's:
        # Innst. 9 S rammeområde 12 prints "post 24 Driftsresultat:" and then
        # "3 Lete- og feltutviklingsutgifter" with the same figure, and
        # Innst. 8 S rammeområde 9 prints a kapittel's total and then its parts.
        # Adding both doubled the column, which the sum check caught twice.
        if not (cells[1].strip().isdigit() if len(cells) > 1 else False):
            continue
        posted = {}
        for i, role in roles.items():
            if i < len(cells):
                a = amount(cells[i], role["kind"] != "government")
                if a:
                    posted[i] = a
        if not posted:
            continue
        sections[current].append({
            "kap": cells[0] if len(cells) > 0 else "",
            "post": cells[1] if len(cells) > 1 else "",
            "formal": cells[2] if len(cells) > 2 else "",
            "amounts": {str(i): {"kr": a[0], "change": a[1]} for i, a in posted.items()},
        })
    return {"unit": unit, "sections": sections, "sums": sums}


def check(rec: dict, roles: dict) -> list[str]:
    """The document's own arithmetic, per column: rows must add to the printed sum."""
    bad = []
    for kind in ("utgifter", "inntekter"):
        rows = rec["sections"].get(kind) or []
        total = rec["sums"].get(kind) or {}
        if not rows or not total:
            continue
        for i, role in roles.items():
            want = total.get(i)
            if want is None:
                continue
            got = sum(r["amounts"].get(str(i), {}).get("change", 0) for r in rows)
            if got != want[1]:
                name = role.get("party") or role.get("kind")
                bad.append(f"{kind} column {i} ({name}): rows add to {got}, sum row says {want[1]}")
    return bad


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--session", default="2025-2026")
    ap.add_argument("--numbers", default=",".join(str(n) for n in range(1, 26)))
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args(argv)

    numbers = [int(x) for x in args.numbers.split(",") if x.strip()]
    records, failures, absent, no_area = [], [], [], []
    for n in numbers:
        url = innstilling_url(args.session, n)
        try:
            doc = fetch(url)
        except Exception as e:  # noqa: BLE001 - a missing number is a fact to report
            absent.append((n, f"{type(e).__name__}"))
            continue
        title = re.search(r"<title>([^<]*)</title>", doc)
        title = norm(title.group(1)) if title else f"inns-{n:03d}s"
        found = 0
        for offset, tbl in tables_with_offsets(doc):
            roles = column_map(tbl[0])
            if not roles:
                continue
            rec = extract(tbl[1:], roles)
            # The rammeområde label can sit anywhere in the table, not only in its
            # first rows: scanning just the top skipped four whole documents
            # (Innst. 6, 7, 10 and 16 S) without counting them as either accepted or
            # rejected, so rammeområde 1, 4, 5, 6, 8 and 14 looked absent from the
            # source when they were simply never read.
            rm = None
            for cells in tbl:
                mm = re.search(r"rammeområde\s+(\d+)", " ".join(cells), re.I)
                if mm:
                    rm = int(mm.group(1))
                    break
            if rm is None:
                rm = area_before(doc, offset)
            if rm is None:
                no_area.append((title, " | ".join(tbl[0])[:90]))
                continue
            bad = check(rec, roles)
            entry = {
                "document": title,
                "number": n,
                "url": url,
                "rammeomrade": rm,
                "unit": rec["unit"],
                "columns": {str(i): roles[i] for i in roles},
                "sums": {k: {str(i): {"kr": v[0], "change": v[1]} for i, v in line.items()}
                         for k, line in rec["sums"].items()},
                "posts": {k: v for k, v in rec["sections"].items()},
                "check_failures": bad,
            }
            if bad:
                failures.append((title, rm, bad))
            else:
                records.append(entry)
            found += 1
        print(f"  {n:3d} {title[:44]:44s} party tables: {found}", file=sys.stderr)

    print(f"\n  documents read: {len(numbers) - len(absent)}  absent: {len(absent)}  "
          f"tables accepted: {len(records)}  rejected: {len(failures)}", file=sys.stderr)
    for a in absent:
        print(f"    absent {a[0]:3d}: {a[1]}", file=sys.stderr)
    for title, hdr in no_area:
        print(f"    NO RAMMEOMRÅDE in {title[:40]}: table header {hdr}", file=sys.stderr)
    for title, rm, bad in failures:
        print(f"    REJECTED {title} rammeområde {rm}: {bad[0]}", file=sys.stderr)
    if not records:
        print("  no table passed its own arithmetic: treat as failure", file=sys.stderr)
        return 2
    doc = {"generated_at": "2025-2026", "session": args.session, "tables": records}
    body = json.dumps(doc, ensure_ascii=False, indent=1) + "\n"
    if args.dry_run:
        print(f"  would write {len(records)} tables ({len(body)} bytes)")
        return 0
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / "partier.json").write_text(body)
    print(f"  wrote {OUT}/partier.json: {len(records)} tables, {len(body)} bytes")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
