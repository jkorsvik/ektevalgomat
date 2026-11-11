"""Tests for the innstilling locator.

Run:  uv run --with pytest python -m pytest tools/autopilot/test_innstilling_locate.py -q

Every URL asserted here was fetched and answered 200 with the byte length in
the comment. They are a fixture, so a change to the transform fails here
instead of quietly producing 404s in a collector log.

The live test at the bottom is a separate thing and stays opt-in: a unit
fixture proves the mapping, a live test proves the site still serves it. A
past 200 is not a promise about tomorrow, and a fixture is not a promise about
every document.
"""

import os
import sys
import pathlib
import urllib.request

import pytest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from innstilling_locate import BASE, Withheld, locate  # noqa: E402


def test_the_tarbudget_innstilling():
    # The document the extraction recipe was measured on.
    assert locate("Dokument 8:34 S (2024-2025), Innst. 2 S (2025-2026)", "1") == (
        BASE + "/2025-2026/inns-202526-002s.pdf"
    )


def test_session_comes_from_the_parentheses():
    # A 2023-2024 document handled by a 2025-2026 innstilling. Using the sak's
    # own session here builds /2023-2024/ and 404s.
    url = locate("Dokument 12:11 (2023-2024), Innst. 468 S (2025-2026)", "1")
    assert url == BASE + "/2025-2026/inns-202526-468s.pdf"


def test_a_dokument_8_reference_with_two_sessions():
    url = locate("Dokument 8:253 S (2021-2022), Innst. 257 S (2022-2023)", "1")
    assert url == BASE + "/2022-2023/inns-202223-257s.pdf"


def test_three_digit_zero_padding():
    # 005, 007 and 010 all answered 200 in three different sessions, so the
    # padding is measured and not a guess about the site's convention.
    assert locate("Innst. 5 S (2024-2025)", "1").endswith("/2024-2025/inns-202425-005s.pdf")
    assert locate("Innst. 7 S (2023-2024)", "1").endswith("/2023-2024/inns-202324-007s.pdf")
    assert locate("Innst. 10 S (2021-2022)", "1").endswith("/2021-2022/inns-202122-010s.pdf")


def test_a_number_above_999_is_withheld():
    with pytest.raises(Withheld, match="three digits"):
        locate("Innst. 1000 S (2025-2026)", "1")


@pytest.mark.parametrize("kode", ["0", "2", "", None, "-1"])
def test_no_innstilling_enum_value_withholds(kode):
    # Measured distribution in 2025-2026: "1" on 429 rows, "0" on 139, "2" on
    # 95. Only "1" can carry a document.
    with pytest.raises(Withheld, match="no innstilling"):
        locate("Dokument 11 (2025-2026)", kode)


def test_enum_one_without_a_reference_withholds():
    # One such row exists in the 2025-2026 payload: the enum says an
    # innstilling exists, the reference does not name it. Withhold, never
    # fall back to the sak's session.
    with pytest.raises(Withheld, match="names no Innst"):
        locate("Dokument 11 (2025-2026)", "1")


def test_two_references_withhold():
    with pytest.raises(Withheld, match="names 2"):
        locate("Innst. 10 S (2025-2026), Innst. 11 S (2025-2026)", "1")


def test_a_lagting_reference_withholds():
    # The letter is the chamber. Only S is covered; L is pre-2009 and outside
    # a corpus that starts in 2011.
    with pytest.raises(Withheld, match="not the Storting"):
        locate("Innst. 5 L (2005-2006)", "1")


def test_a_mixed_chamber_reference_withholds():
    # Counting only S references would let the S half of a mixed pair through
    # and silently ignore the other document.
    with pytest.raises(Withheld, match="names 2"):
        locate("Innst. 5 S (2025-2026), Innst. 6 L (2005-2006)", "1")


@pytest.mark.parametrize("years", ["(2022-2025)", "(2025-2025)", "(2026-2025)"])
def test_non_consecutive_session_years_withhold(years):
    # A Storting session spans exactly two consecutive years. Anything else is
    # a misparse, and a misparse here builds a directory that does not exist.
    with pytest.raises(Withheld, match="not one year"):
        locate("Innst. 12 S %s" % years, "1")


VERIFIED_200 = [
    ("/2025-2026/inns-202526-002s.pdf", 3563108),
    ("/2025-2026/inns-202526-121s.pdf", 1633087),
    ("/2025-2026/inns-202526-468s.pdf", 629634),
    ("/2024-2025/inns-202425-143s.pdf", 517371),
    ("/2024-2025/inns-202425-005s.pdf", 2164105),
    ("/2023-2024/inns-202324-007s.pdf", 1530550),
    ("/2021-2022/inns-202122-010s.pdf", 495104),
]


@pytest.mark.skipif(os.environ.get("EKV_LIVE") != "1", reason="set EKV_LIVE=1 to hit the network")
@pytest.mark.parametrize("path,size", VERIFIED_200)
def test_live_documents_still_serve(path, size):
    req = urllib.request.Request(BASE + path, method="HEAD",
                                 headers={"User-Agent": "Mozilla/5.0"})
    with urllib.request.urlopen(req, timeout=25) as r:
        assert r.status == 200
        assert r.headers.get("Content-Length") == str(size)
