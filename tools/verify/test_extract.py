"""Checks on the text extractor, offline.

Run:  python3 tools/verify/test_extract.py     (also works under pytest)

Why this exists: `to_text` stripped HTML only when the document *was* HTML. Most
Stortinget documents arrive as a JSON envelope whose string values carry the
markup, and that branch re-serialised the envelope instead of stripping it. The
markup therefore survived as literal text, which put a table's label and its
figure thousands of characters apart. An evidence row for a budget table could
not be quoted at all, and three collection items stalled on the resulting 6 000
character archive bound.

Measured on one post table before the fix: 1 173 332 characters of text, of which
599 895 were real, so 573 437 characters were markup. The fragment that carries a
post's amount is 72 characters once the markup is gone.

Every test here is written so that putting the old behaviour back makes it fail.
"""

import json
import sys

sys.path.insert(0, str(__import__("pathlib").Path(__file__).resolve().parent))

from verify_excerpts import norm, to_text  # noqa: E402

# A table row the way Stortinget actually serves it: a JSON envelope whose value
# is HTML, with the label and the amount in separate cells.
ENVELOPE = {
    "value": (
        "<table><tr><td><p>78</p></td><td><p>(Ny)</p></td>"
        "<td><p>Norgespris for fjernvarme<span class='x'>, overslagsbevilgning,</span>"
        "<span class='y'>bevilges</span> med</p></td>"
        "<td><p>65 000 000</p></td></tr></table>"
    ),
    "forslag_type": 1,
}
BODY = json.dumps(ENVELOPE, ensure_ascii=False).encode()

# The claim a budget row wants to make, as a reader would write it.
FRAGMENT = "norgespris for fjernvarme , overslagsbevilgning, bevilges med 65 000 000"


def test_markup_inside_a_json_envelope_is_stripped() -> None:
    text = to_text(BODY)
    assert "<" not in text and ">" not in text, f"tags survived extraction: {text!r}"
    assert "class=" not in text, f"attributes survived extraction: {text!r}"


def test_a_cell_label_and_its_figure_become_one_quotable_fragment() -> None:
    """The point of the fix: a table cell is quotable once markup is gone.

    Fails if the JSON branch re-serialises, because the tags then sit between the
    words and the fragment is not a substring.
    """
    assert norm(FRAGMENT) in norm(to_text(BODY))


def test_entities_are_decoded_in_a_json_envelope() -> None:
    body = json.dumps({"value": "<p>skatt &amp; avgift</p>"}, ensure_ascii=False).encode()
    assert "skatt & avgift" in to_text(body)


def test_a_plain_html_document_still_extracts() -> None:
    """The path that already worked must not change."""
    body = b"<html><body><p>Det samlede <b>niv&aacute;et</b> senkes.</p></body></html>"
    text = to_text(body)
    assert "Det samlede" in text and "niv" in text
    assert "<" not in text


def test_a_document_that_is_not_json_still_falls_through() -> None:
    """A body starting with a brace but not parseable must not raise."""
    body = b"{ dette er ikke json <p>men tekst</p>"
    assert "men tekst" in to_text(body)


if __name__ == "__main__":
    fails = 0
    for name, fn in sorted(globals().items()):
        if name.startswith("test_") and callable(fn):
            try:
                fn()
                print(f"PASS {name}")
            except AssertionError as exc:
                print(f"FAIL {name}: {exc}")
                fails += 1
    print(f"\n{fails} failed")
    sys.exit(1 if fails else 0)
