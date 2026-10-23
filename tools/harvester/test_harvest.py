"""Unit tests for harvest.py's pure functions — no network involved.

Run with: cd tools/harvester && uv run pytest
"""

from __future__ import annotations

import textwrap
from pathlib import Path

import pytest
import yaml

from harvest import (
    case_search_text,
    current_session_start_year,
    generate_sessions,
    load_topics,
    match_case,
    render_skeleton_yaml,
)


def test_generate_sessions_basic_range():
    assert generate_sessions(2020, 2022) == ["2020-2021", "2021-2022", "2022-2023"]


def test_generate_sessions_single_year():
    assert generate_sessions(2020, 2020) == ["2020-2021"]


def test_generate_sessions_empty_when_end_before_start():
    assert generate_sessions(2022, 2020) == []


def test_current_session_start_year_mid_session_stays_in_that_session(monkeypatch):
    # 2026-09-23 (this repo's own "today") is still inside the
    # 2025-2026 session, which started October 2025 — a date in
    # Jan-Sep must resolve to the PREVIOUS calendar year's session
    # start, not the current calendar year. An earlier version of this
    # function had a no-op ternary that always returned now.year,
    # silently reporting one session too far into the future on every
    # run between January and September.
    import datetime as dt

    import harvest

    class FixedDate(dt.date):
        @classmethod
        def today(cls):
            return cls(2026, 9, 23)

    monkeypatch.setattr(harvest.datetime, "date", FixedDate)
    assert current_session_start_year() == 2025


def test_current_session_start_year_after_october_is_current_calendar_year(monkeypatch):
    import datetime as dt

    import harvest

    class FixedDate(dt.date):
        @classmethod
        def today(cls):
            return cls(2026, 11, 5)

    monkeypatch.setattr(harvest.datetime, "date", FixedDate)
    assert current_session_start_year() == 2026


def test_case_search_text_joins_title_and_emner():
    case = {
        "tittel": "Om kjernekraft",
        "korttittel": "Kjernekraft-saken",
        "emne_liste": [{"navn": "Energi"}, {"navn": "Sikkerhet"}],
    }
    text = case_search_text(case)
    assert "kjernekraft" in text
    assert "energi" in text
    assert "sikkerhet" in text


def test_case_search_text_strips_encoding_artifact_question_marks():
    # Seen for real in the 2025-2026 session export: an æ/ø/å that
    # failed to encode renders as a literal "?" in tittel/korttittel.
    # Stripping it (not replacing with a letter — we don't know which
    # one) means a keyword like "gjøre" still won't false-positive
    # match "gj?re", but a keyword typed as the ASCII-safe substring
    # still finds it.
    case = {"tittel": "Om utenlandsk p?virkning i valg", "korttittel": ""}
    text = case_search_text(case)
    assert "p?virkning" not in text
    assert "pvirkning" in text  # the "?" is dropped, not replaced


def test_case_search_text_handles_missing_fields():
    assert case_search_text({}) == ""


def test_match_case_finds_keyword_in_title():
    case = {"tittel": "Representantforslag om politisk reklame på nett", "korttittel": ""}
    matched = match_case(case, ["politisk reklame", "noe annet"])
    assert matched == ["politisk reklame"]


def test_match_case_is_case_insensitive():
    case = {"tittel": "POLITISK REKLAME", "korttittel": ""}
    matched = match_case(case, ["politisk reklame"])
    assert matched == ["politisk reklame"]


def test_match_case_returns_empty_list_when_no_keyword_matches():
    case = {"tittel": "Om jordbruksoppgjøret", "korttittel": ""}
    matched = match_case(case, ["politisk reklame", "kjernekraft"])
    assert matched == []


def test_match_case_can_return_multiple_matches():
    case = {"tittel": "Politisk reklame og desinformasjon valg", "korttittel": ""}
    matched = match_case(case, ["politisk reklame", "desinformasjon valg", "uansett noe"])
    assert set(matched) == {"politisk reklame", "desinformasjon valg"}


def test_load_topics_rejects_empty_file(tmp_path: Path):
    path = tmp_path / "empty.yaml"
    path.write_text("{}", encoding="utf-8")
    with pytest.raises(ValueError):
        load_topics(path)


def test_load_topics_rejects_topic_without_keywords(tmp_path: Path):
    path = tmp_path / "bad.yaml"
    path.write_text(
        textwrap.dedent(
            """
            utsagn-02:
              text: "Some statement"
              note: "no keywords field at all"
            """
        ),
        encoding="utf-8",
    )
    with pytest.raises(ValueError):
        load_topics(path)


def test_load_topics_parses_real_topics_file():
    real_path = Path(__file__).parent / "topics.yaml"
    topics = load_topics(real_path)
    assert "utsagn-02" in topics
    assert "utsagn-04" in topics
    assert len(topics["utsagn-02"]["keywords"]) > 0
    # utsagn-05 must NOT have a topic entry — see topics.yaml's own
    # comment for why (meta-statement, no case to search for).
    assert "utsagn-05" not in topics


def test_render_skeleton_yaml_fills_known_fields_and_todos_the_rest():
    case = {
        "id": 12345,
        "henvisning": "Dokument 8:999 S (2025–2026)",
        "tittel": "Representantforslag om noe",
        "korttittel": "Noe-saken",
    }
    rendered = render_skeleton_yaml("utsagn-04", case, ["desinformasjon valg"])
    parsed = yaml.safe_load(rendered)

    # Real, confirmed fields must be filled in for real.
    assert parsed["statement_id"] == "utsagn-04"
    assert parsed["source_name"] == "Dokument 8:999 S (2025–2026)"
    assert "p=12345" in parsed["source_url"]

    # Anything requiring an actual read must stay an explicit TODO
    # marker, and position must stay null — never a guessed value that
    # could slip past review looking like real evidence.
    assert parsed["position"] is None
    assert "TODO" in parsed["party_id"]
    assert "TODO" in parsed["context_note"]
    assert "TODO" in parsed["verbatim_excerpt"]


def test_render_skeleton_yaml_output_is_valid_yaml():
    # A skeleton that fails to parse would be useless as a starting
    # point for the human/agent filling it in.
    case = {"id": 1, "henvisning": "X", "tittel": "Y"}
    rendered = render_skeleton_yaml("utsagn-02", case, ["politisk reklame"])
    parsed = yaml.safe_load(rendered)
    assert isinstance(parsed, dict)
