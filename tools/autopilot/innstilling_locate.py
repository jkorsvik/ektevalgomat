"""Locate the PDF of a Storting committee recommendation (innstilling) from a
`saker?sesjonid=` row.

Two facts make this small module necessary, both measured on 2026-11-11.

`innstilling_kode` is an enum, not a slug. It is "1" on a row whose sak has an
innstilling, and "0" or "2" on a row that has none. It answers WHETHER, never
WHICH. A recipe that treats it as a document identifier builds a URL that 404s
and looks like a missing document rather than a wrong field.

`henvisning` is the only field that names WHICH. It can name an innstilling
from a different session than the sak's own, because a document raised in one
session is handled in another: "Dokument 12:11 (2023-2024), Innst. 468 S
(2025-2026)". The session therefore comes from the parentheses and nowhere
else. Using the sak's session builds the wrong directory.

Nothing here reads document text. It maps metadata to a URL, so it is safe to
run anywhere, including on a model that must not see source bodies.
"""

import re

BASE = "https://www.stortinget.no/globalassets/pdf/innstillinger/stortinget"

# Any innstilling reference at all, whatever the chamber letter. Counted first
# so a reference that mixes S and L cannot pass by matching only its S half.
ANY_INNST = re.compile(r"Innst\.\s*(\d+)\s*([A-Z])\s*\((\d{4})-(\d{4})\)")


class Withheld(Exception):
    """No URL can be built. The caller must record the reason and emit nothing."""


def locate(henvisning, innstilling_kode):
    """Return the PDF URL for one `saker` row, or raise Withheld.

    Six conditions withhold, and each one is a shape that occurs in real
    payloads: the enum says no innstilling, the reference names none, the
    reference names more than one, the chamber is not the Storting, the two
    session years are not consecutive, or the number does not fit three digits.
    """
    kode = str(innstilling_kode if innstilling_kode is not None else "").strip()
    if kode != "1":
        raise Withheld("innstilling_kode %r: the sak has no innstilling" % kode)

    found = ANY_INNST.findall(henvisning or "")
    if not found:
        raise Withheld("henvisning names no Innst. N S (year-year)")
    if len(found) > 1:
        raise Withheld("henvisning names %d innstillinger" % len(found))

    number, chamber, first, second = found[0]
    if chamber != "S":
        raise Withheld("chamber %r is not the Storting" % chamber)
    if int(second) != int(first) + 1:
        raise Withheld("session %s-%s is not one year" % (first, second))

    nr = int(number)
    if not 1 <= nr <= 999:
        raise Withheld("innstilling number %d does not fit three digits" % nr)

    return "%s/%s-%s/inns-%s%s-%03ds.pdf" % (
        BASE, first, second, first, second[2:], nr,
    )
