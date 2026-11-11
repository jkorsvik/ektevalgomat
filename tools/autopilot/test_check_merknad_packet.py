"""Tests for the merknad packet gate.

Run:  uv run --with pymupdf --with pyyaml python -m pytest tools/autopilot/test_check_merknad_packet.py -q

These are hermetic: no network, no document. They pin the two places where the
gate was wrong once and would be wrong again.

The party resolver rejected the canonical ids `sv` and `mdg` while its own
docstring promised to accept them, because it only consulted the label map. The
source feeder found it by running the gate. A resolver that rejects the language
the rest of the codebase speaks makes a correct packet look wrong.
"""

import pathlib
import sys

import pytest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import check_merknad_packet as g  # noqa: E402


@pytest.mark.parametrize("label,expected", [
    ("rodt", ["rodt"]),
    ("sv", ["sv"]),
    ("mdg", ["mdg"]),
    ("h", ["h"]),
    ("v", ["v"]),
    ("ap", ["ap"]),
])
def test_a_canonical_id_is_accepted_as_itself(label, expected):
    assert g.parties_of(label) == expected


@pytest.mark.parametrize("label,expected", [
    ("Rødt", ["rodt"]),
    ("Arbeiderpartiet", ["ap"]),
    ("Høyre", ["h"]),
    ("Sosialistisk Venstreparti", ["sv"]),
    ("Miljøpartiet De Grønne", ["mdg"]),
    ("Kristelig Folkeparti", ["krf"]),
    ("Senterpartiet", ["sp"]),
    ("Venstre", ["v"]),
])
def test_a_document_label_resolves(label, expected):
    assert g.parties_of(label) == expected


@pytest.mark.parametrize("label,expected", [
    ("Arbeiderpartiet og Senterpartiet", ["ap", "sp"]),
    ("ap og sp", ["ap", "sp"]),
    ("Arbeiderpartiet og sp", ["ap", "sp"]),
])
def test_a_coalition_resolves_every_part(label, expected):
    assert g.parties_of(label) == expected


def test_the_hyphen_wrap_resolves():
    # The printed heading breaks the second name across a line.
    assert g.parties_of("Arbeiderpartiet og Sen-\nterpartiet") == ["ap", "sp"]


@pytest.mark.parametrize("label", [
    "Norges Kommunistiske Parti",
    "Rødt og Norges Kommunistiske Parti",
    "",
    None,
])
def test_an_unresolvable_label_returns_none(label):
    assert g.parties_of(label) is None


def test_the_year_one_sentinel_is_not_a_date():
    payload = {"saksgang": {"saksgang_steg_liste": [
        {"navn": "Melding", "saksgang_hendelse_liste": [
            {"dato": "01.01.0001 00:00:00"},
            {"dato": "06.10.2023 00:00:00"}]},
        {"navn": "Komitebehandling", "saksgang_hendelse_liste": [
            {"dato": "27.11.2023 00:00:00"}]},
        {"navn": "Debatt og vedtak", "saksgang_hendelse_liste": [
            {"dato": "04.12.2023 00:00:00"}]},
    ]}}
    assert g.saksgang_dates(g.json.dumps(payload)) == ("2023-11-27", "2023-12-04")


def test_the_committee_date_is_not_the_vote_date():
    payload = {"saksgang": {"saksgang_steg_liste": [
        {"navn": "Komitebehandling", "saksgang_hendelse_liste": [
            {"dato": "27.11.2023 00:00:00"}]},
        {"navn": "Debatt og vedtak", "saksgang_hendelse_liste": [
            {"dato": "04.12.2023 00:00:00"}]},
    ]}}
    komite, vote = g.saksgang_dates(g.json.dumps(payload))
    assert komite != vote
