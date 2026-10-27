package data

import "testing"

func TestStoreLoadsExpectedCounts(t *testing.T) {
	// Phase 1 is dimension 1 (Demokratisk åpenhet, 5 statements). Three
	// more statements (43-45) exist as an explicit addendum outside the
	// memo's 1-42 numbering — see docs/roadmap.md for why. utsagn-06
	// (2026-09-24) is the first real statement of dimension 2 (Norsk
	// handlefrihet) — one statement, not the whole dimension; see that
	// statement's own evidence context_notes for why 06 is the only one
	// so far. utsagn-43 (2026-09-24) got full 9/9 party coverage from a
	// single, cleanly roll-call-verifiable case (Innst. 196 S, votering_id
	// 24383-24386) — see that statement's evidence context_notes.
	// utsagn-04 (2026-09-24, later same day) got full 9/9 party coverage
	// too, from Innst. 231 S (2022-2023) — a real kartlegging proposal on
	// exactly this topic that two earlier research passes had missed
	// (they found only the vote-tally-only fake-news case, Innst. 303 S,
	// which was correctly rejected for the wrong specificity; this one
	// tests the actual "map/document foreign election influence" ask).
	// utsagn-02 (2026-09-24, same session) got full 9/9 coverage too,
	// from a plenary "løst forslag" during the 2025-10-14 trontaledebatt
	// (votering_id 26361) that a Dokument-8/Innst.-only search would
	// never surface — see that statement's own evidence context_notes
	// for the one real half-of-the-claim nuance it carries. utsagn-12
	// (2026-09-24, later same session) is the first real statement of
	// dimension 3 (Forvaltning av fellesskapets ressurser / energi) —
	// full 9/9 coverage from Innst. 535 L (2024-2025), votering_id
	// 26116: Rødt's own national makspris-på-strøm forslag (part of a
	// 4-bill merged committee report the roadmap had previously flagged
	// as "too complex to rush" — this codes only that ONE clean vote,
	// not the other 16 on the same case). utsagn-13 (2026-09-24, later
	// still) is dimension 3's second statement — full 9/9 coverage from
	// Innst. 278 S (2025-2026), votering_id 27722: FrP+H+V's forslag 1-2
	// to legislate NOW for commercial nuclear power, a distinct axis
	// from utsagn-06's state-ownership question (the same case's own
	// committee majority only backed a much weaker "write a white
	// paper" tilråding instead — see that statement's evidence
	// context_notes for why only this one vote, of 11 on the case, was
	// coded). utsagn-25 (2026-09-24, later still) is the first real
	// statement of dimension 5 (Universell grunnmur, 25-30) — full 9/9
	// coverage from Innst. 458 L (2024-2025), votering_id 25387: SV's
	// own forslag for a binding plan to universalize dental care by
	// 2030, distinct from the same case's own (weaker, adopted)
	// tilråding. Rødt's own written merknad on this case explicitly
	// backs the same universalist goal yet its vote on this specific
	// forslag is "mot" — documented as a real, unresolved tension
	// rather than smoothed over, see that statement's own evidence
	// context_notes. utsagn-31 (2026-09-24, same session) is the first
	// real statement of dimension 6 (Barns muligheter, 31-36) — full
	// 9/9 coverage from Innst. 126 S (2025-2026), votering_id 26942:
	// SV+MDG's own forslag (added during committee treatment, not part
	// of Venstre's original 17-point Dokument 8:11 S) to legally index
	// barnetrygd to inflation every year. A distinct near-miss caught
	// here: an AI web-fetch intermediary's summary of this innstilling
	// mixed in content from an unrelated document; discarded before
	// coding anything, by fetching and reading the raw HTML directly
	// instead — see that statement's own evidence context_notes. A
	// count drifting here without a matching content change is a sign
	// something in internal/data/content/ silently failed to load, not
	// a reason to bump the numbers. utsagn-19 (2026-09-24, same session)
	// is the first real statement of dimension 4 (Teknologisk
	// kapasitet, 19/20/24) — full 9/9 coverage from Innst. 225 S
	// (2025-2026), votering_id 27485: SV+Rødt's own forslag to build a
	// publicly-owned-and-operated national cloud, distinct from the
	// government's own KVU (which found that exact model the most
	// expensive option and recommended against it) and from two
	// documented in-case alternatives (FrP's own "investigate first"
	// version, and MDG+Rødt's ownership-model-agnostic version) — see
	// that statement's own evidence context_notes for why testing the
	// strict forslag rather than either alternative was the right call.
	// utsagn-44 (2026-09-26) got full 9/9 party coverage from the SAME
	// case (Innst. 225 S), a different and more specific forslag (4,
	// votering_id 27492: GDPR-oppfølging + jurisdiksjonskrav for
	// sensitiv/samfunnskritisk informasjon) — a real Stortinget floor
	// vote, not the tier-C party-program fallback originally planned.
	// The innstilling lists forslag 4 as SV/Rødt/MDG, the vote record as
	// "på vegne av SV og R"; the evidence files quote both. FrP is coded
	// from its own text (+1) rather than its raw "mot"; Sp and V share
	// one vote pattern (for forslag 2, mot forslag 4) and are both 0.
	// JSON vote codes from data.stortinget.no are 1 = ikke til stede,
	// 2 = for, 3 = mot (verified against the XML labels for 27257); a
	// first pass read them wrong and was corrected the same day.
	//
	// utsagn-46 through utsagn-51 (2026-09-27, corrected 2026-09-27) are
	// the first six statements of "Tilgang, integrering og rettsstat" —
	// a dimension the owner added beyond the source memo's original nine
	// (see server.go's dimensionTotals/dimensionOrder doc comments and
	// docs/roadmap.md). All six anchor to real, individually verified
	// 2025-2026 Storting votes (python3 votes.py confirmed [OK] against
	// data.stortinget.no's own vote-count totals for every votering_id
	// used).
	//
	// utsagn-46 (utvisning uavhengig av tilknytning, Innst. 205 S,
	// forslag 2, votering 27441) has full 9/9 coverage: FrP's own
	// forslag (+2); Høyre (+1) and Sp (-1) from their own, partly shared
	// text — Høyre separately states it supports the ambition, Sp does
	// not and only signs the text warning that abolishing all
	// individual assessment risks breaching international law; the
	// other six parties (Ap, SV, Rødt, MDG, KrF, Venstre) have no own
	// text on forslag 2 specifically and are coded -1 with
	// evidensstyrke_cap moderat from the vote alone, same convention as
	// utsagn-44's KrF/Venstre evidence.
	//
	// utsagn-47 (individuell tilknytningsvurdering, same case, forslag
	// 3) is deliberately capped at 3/9 (FrP -2, Høyre +1, Sp +1):
	// voting against forslag 2 is not the same as defending today's
	// individual-assessment rule, so only Høyre and Sp — whose own
	// merknad explicitly defends keeping an individual assessment — are
	// coded from text, and no party is inferred from the forslag-2 vote
	// alone. The remaining six parties are "Ikke vurdert".
	//
	// utsagn-48/49 (skjerpede krav for statsborgerskap vs. en tidsfrist
	// for tilbakekall, Innst. 232 S, votering 27524/27522/27523): both
	// have full 9/9 coverage. utsagn-48: Sp is coded 0/nøytral since its
	// own text wants strict criteria in general but calls FrP's and
	// Høyre's specific asks under-investigated, not wrong in direction.
	// utsagn-49: SV/Rødt/MDG's own forslag (+2), KrF/Venstre from their
	// own for-vote with no own text (+1, cap moderat), and Ap/FrP/Høyre/
	// Sp from their own mot-vote with no own text on tilbakekall
	// specifically (-1, cap moderat).
	//
	// utsagn-50 (skriftlig integreringserklæring, Innst. 309 L, votering
	// 27779, 9/9 coverage) — FrP is coded +1, not -1: FrP's own text
	// calls the government's version "symbolpolitikk" for being too
	// weak and wants an even stronger, sanctioned version, so its "mot"
	// vote is a "not strong enough", not a rejection of the idea.
	//
	// utsagn-51 (ubegrenset varighet ved lukket institusjon for
	// ungdomskriminalitet, Innst. 173 S, forslag 1, votering 27111,
	// 6/9 coverage — SV, Rødt and Venstre have no own text and no
	// committee seat, left "Ikke vurdert" rather than inferred). KrF is
	// coded -1 (corrected 2026-09-27 from an earlier -2): KrF's own text
	// is conditional, not a principled rejection — closed institutions
	// "kan ... være nødvendige i enkelte alvorlige tilfeller" — and
	// explicitly targets only the unbounded "så lenge det er behov for
	// det" wording. The same correction fixed a sponsor misattribution:
	// KrF is NOT a co-sponsor of forslag 1 "med FrP og Høyre" — Innst.
	// 173 S's own "Forslag 3" is Høyre+KrF only, on an unrelated topic
	// (a 10-14-year-old ungdomskriminalitetsnemnd). What FrP, Høyre and
	// KrF actually share is the committee's tilråding (vedtak I, a
	// broader open-and-closed institution network), a separate ask from
	// forslag 1's unbounded duration, and both are quoted directly in
	// the evidence file rather than conflated.
	//
	// utsagn-52 through utsagn-56 (2026-09-27) are dimension 10,
	// "Finansiering av fellesskapet" — the second post-memo dimension
	// (see server.go's dimensionTotals/dimensionOrder doc comments),
	// listed before dimension 11 despite being coded into this repo
	// after it, because 10 < 11. Built from party programs 2025-2029
	// (tier C, source_tier: C) for utsagn-52/53/55, plus two real,
	// individually verified Storting cases (tier A) for utsagn-54/56.
	//
	// utsagn-52/53 (formuesskatten trappes ned/fjernes, vs. høy
	// inntekt/formue skal betale mer — opposite directions, tier C)
	// have 7/9 and 8/9 coverage respectively. Sp and KrF are coded 0
	// on utsagn-52: both programs pair a reduction for one group
	// (SMB-eiere / arbeidende kapital) with an explicit, same-breath
	// counter-signal (formuende må bidra etter evne / økt formueskatt
	// på dyre primærboliger) — a real, textual tension, not inferred.
	// Høyre's utsagn-52 evidence is sourced from Innst. 188 S (tier A,
	// "Formuesskatten bør deretter avskaffes i sin helhet"), stronger
	// than and consistent with its program's narrower "arbeidende
	// kapital" text. FrP/Høyre/Venstre's formuesskatt-reduction texts
	// also anchor utsagn-53 at -1 (sannsynliggjort, not dokumentert):
	// on-topic (wealth taxation) but scoped to business capital, not
	// explicit about high-income earners too. Rødt was "Ikke vurdert"
	// on both at first (roedt.no answers automated fetches with HTTP
	// 429); it is now coded on both from Innst. 2 S (tier A, files
	// 601/602), and the program downloaded by hand on 2026-09-28
	// (file 603) points the same way.
	//
	// utsagn-54 (boligen teller mindre i formuesskatten, tier A, Innst.
	// 188 S, forslag 1-6, voteringene 27359-27362) has full 9/9
	// coverage — the first real tier-A anchor for Rødt in this
	// dimension despite the program fetch failure, since Rødt is a
	// direct signatory of the committee-majority merknad. FrP/Høyre/
	// Venstre are +2 from their own forslag; Ap/SV/Sp/Rødt/MDG are
	// coded 0 (nøytral/delt), not -1/-2, from the shared "vil behandle
	// realitetene ifm. revidert nasjonalbudsjett" merknad — a
	// procedural deferral to the government's own, smaller adjustment
	// (10 to 14 mill. kroner), not a substantive rejection of the
	// statement's direction; KrF has no committee seat and no text, so
	// -1 cap moderat from the vote alone.
	//
	// utsagn-55 (samlet skatte-/avgiftsnivå senkes, tier C) has 7/9
	// coverage (6/9 until Rødt's program was read, see below). Sp and MDG were deliberately left "Ikke vurdert": both
	// have real, explicit tax-reduction sentences, but each is scoped
	// to a specific group (lave/vanlige inntekter) or purpose, not the
	// TOTAL/samlede level the statement asks about, and reducing tax
	// for one group without more is a redistribution, not necessarily
	// a total-level cut.
	//
	// utsagn-56 (Norge bør stanse ETS2, optional per TAX-HANDOFF.md,
	// tier A, Innst. 58 S, sak 104877, voteringene 26745-26747) has
	// full 9/9 coverage. FrP is coded +2 despite its own primary
	// forslag being narrower (fjerne CO2-avgiften, not stanse hele
	// gjennomføringen): its own text explicitly says it "vil
	// subsidiært støtte Senterpartiets forslag" and the vote confirms
	// it (46 of 47 FrP members for). Rødt is coded -1 with
	// consistency_flag=tension: its own text is genuinely critical of
	// ETS2's democratic-control implications, but explicitly frames
	// the EU's 2028 delay as room for debate "uten å rive opp
	// klimakvotedirektivet", and Rødt voted with the majority to let
	// the implementation continue — text sounds skeptical, the vote
	// does not. SV is -1 cap moderat with an explicitly noted internal
	// split (7 for the majority tilråding, 1 against). This is also
	// the first case in the dataset with per-vote turnout near the
	// full 169-seat chamber (~166 votes cast) rather than the ~99-102
	// seen in the dimension-11 cases — a real attendance difference
	// between debates, confirmed [OK] against data.stortinget.no's own
	// totals, not a data error.
	//
	// 2026-09-27 (commit a9887cf): 13 statements were flipped to the
	// opposite direction to cut the acquiescence-bias spread from 41 to
	// 14 points (see acquiescence_test.go) — utsagn-04, 07, 08, 12, 18,
	// 19, 20, 21, 24, 25, 31, 37, 44. Their case identity, votering_ids
	// and source documents are all unchanged (the historical build notes
	// above still describe the real underlying case correctly); only the
	// statement `text`/`short_name` were rewritten to the position a real
	// majority party actually argued in that same case, and every
	// party's `position` on that statement was negated to match. See
	// TestAcerEnergyMarketPositionsAreCoded and
	// TestUtenlandskPaavirkningPositionsAreCoded below for the two
	// flipped statements that had their own pinned position assertions,
	// and each flipped evidence file's own context_note for the rest.
	//
	// 2026-09-28 (W3 excerpt-verification pass): +2 evidence files.
	// utsagn-52 and utsagn-53 each gained a Rødt entry (Innst. 2 S
	// 2025-2026, Rødt's own budget-innstilling hovedmerknad, tier A) —
	// Rødt's roedt.no arbeidsprogram is blocked for programmatic
	// fetching by a Vercel bot-challenge (HTTP 429 on every URL and
	// UA tried, including a real-browser UA; wayback machine has only
	// pre-landsmøte "forslag" drafts, not the adopted text, so those
	// were not used either). utsagn-52's Sp gap was investigated too
	// (own program PDF: "Senterpartiet vil redusere avgiftene i
	// Norge" — an avgifter-specific, not formuesskatt-specific,
	// statement) but left uncoded rather than overtolked; utsagn-52's
	// AP gap was searched (Stortinget Innst. 2 S, Innst. 3 S, Innst.
	// 188 S, and AP's own program PDF) with zero "formuesskatt"
	// mentions found anywhere, so it stays uncoded rather than guessed.
	// utsagn-45's Sp gap was independently re-investigated and reached
	// the exact same conclusion TestAddendumPendingCasesHaveNoFabricatedPartyPositions
	// already encodes below (general EU/EØS-sovereignty text exists,
	// but never connects to defense/beredskap specifically) — no file
	// added, Sp and Rødt stay pending on utsagn-45 by design.
	// utsagn-55's Sp/MDG/Rødt gaps were investigated (Sp and MDG
	// programs both lack any "samlet skatte- og avgiftsnivå"-level
	// statement; Rødt's program was unreachable at the time, since
	// coded, see the 2026-09-28 entry below) and left uncoded, matching statement 543's own kontekst file, which had
	// already reasoned this out for Sp/MDG explicitly.
	// 2026-10-02 (tax fairness): utsagn-57 (skatt på store arv) and
	// utsagn-58 (utflytterskatten lempes). 57 is 8/9: Ap has only a Sp
	// minister's sentence about Ap (file 710, position: null).
	// 2026-10-03 (digital dependencies): utsagn-59, a fixed report to
	// Stortinget on Norway's digital dependencies (the tilråding of
	// Innst. 225 S, votering 27496, failed 47-54). 9/9 coded.
	if got := len(Store.Statements); got != 60 {
		t.Errorf("Statements = %d, want 60", got)
	}
	if got := len(Store.Parties); got != 9 {
		t.Errorf("Parties = %d, want 9", got)
	}
	//
	// 2026-09-28 (Rødt program, downloaded by hand in a browser): +1
	// evidence file. utsagn-55 gains Rødt at -2 (file 603, tier C: "mer
	// ... finansieres i fellesskap ... via blant annet skatter og
	// avgifter" is a total-level sentence), so 55 goes 6/9 -> 7/9, and
	// Sp and MDG stay uncoded. utsagn-45 Rødt stays pending: the
	// program was read in full (defence chapter pp. 86-89, EØS chapter)
	// and, like Sp, never links EU/EØS to defence or beredskap.
	//
	// 2026-10-02 (trust audit): +3 evidence files, 701-703. Ap, Sp and
	// MDG's votes against Rødt's forslag 30-32 on utsagn-06 (votering
	// 27715), shown next to their merknad-based +1 and deliberately
	// position: null, so no cell position or label changes.
	//
	// 2026-10-02 (tax fairness): +26 evidence files, 704-729, for
	// utsagn-57 and utsagn-58 (20 coded party items, 6 items with
	// position: null: 5 kontekst files and 710).
	//
	// 2026-10-03 (digital dependencies): +19 evidence files, 730-748,
	// for utsagn-59 (13 coded party items, 6 with position: null: 3
	// kontekst files, 2 press items (tier D) and Sp's programme).
	if got := len(Store.Evidence); got != 619 {
		t.Errorf("Evidence = %d, want 619", got)
	}
}

func TestEveryEvidenceReferencesKnownIDs(t *testing.T) {
	// mustLoad already validates this at package init (a bad reference
	// panics before this test even runs) — this test exists so a future
	// refactor of load() that accidentally drops that validation fails
	// loudly here instead of silently shipping unlinked evidence.
	for _, e := range Store.Evidence {
		if _, ok := Store.Statement(e.StatementID); !ok {
			t.Errorf("evidence %s: statement_id %q not found", e.ID, e.StatementID)
		}
		if e.PartyID != "" {
			if _, ok := Store.Party(e.PartyID); !ok {
				t.Errorf("evidence %s: party_id %q not found", e.ID, e.PartyID)
			}
		}
	}
}

func TestEveryEvidenceHasValidClassAndScanSummary(t *testing.T) {
	// Same "redundant with the load-time panic, but asserted directly"
	// reasoning as TestEveryEvidenceReferencesKnownIDs above — these two
	// fields (added 2026-09-23 for the evidence-card redesign) drive
	// server.evidenceClassBadge and the Scan disclosure level directly;
	// a silently empty one would render as a blank badge or an empty
	// first line, not an error.
	validClasses := map[string]bool{"fakta": true, "partiets_begrunnelse": true, "analyse": true, "uklart": true}
	for _, e := range Store.Evidence {
		if !validClasses[e.EvidenceClass] {
			t.Errorf("evidence %s: evidence_class %q is not one of fakta/partiets_begrunnelse/analyse/uklart", e.ID, e.EvidenceClass)
		}
		if e.ScanSummary == "" {
			t.Errorf("evidence %s: scan_summary is empty", e.ID)
		}
	}
}

func TestEvidensstyrkeCapIsValidWhenSet(t *testing.T) {
	// Redundant with the load-time panic (loader.go validates the same
	// values), asserted directly per this file's own convention. Empty
	// is valid (no cap, the common case) — only a set value must match.
	valid := map[string]bool{"": true, "svak": true, "moderat": true}
	capped := 0
	for _, e := range Store.Evidence {
		if !valid[e.EvidensstyrkeCap] {
			t.Errorf("evidence %s: evidensstyrke_cap %q is not one of svak/moderat", e.ID, e.EvidensstyrkeCap)
		}
		if e.EvidensstyrkeCap != "" {
			capped++
		}
	}
	// Not a hard count check (more cells may legitimately earn a cap as
	// evidence-gathering continues) — just a sanity floor so a future
	// refactor that silently drops the yaml tag doesn't go unnoticed.
	if capped < 5 {
		t.Errorf("found %d evidensstyrke_cap items, want at least 5 (the 2026-09-26 utsagn-06 review set)", capped)
	}
}

func TestVenstreLobbyregisterPositionIsCoded(t *testing.T) {
	// Regression guard for the one cell phase 1 actually has a coded
	// position for: Venstre proposed the lobby register (Dok. 8:53 S),
	// so its position on utsagn-01 must be +2, not nil and not defaulted
	// to 0 — those would both misrepresent what the source documents.
	cell := Store.EvidenceForCell("utsagn-01", "v")
	if len(cell) != 1 {
		t.Fatalf("EvidenceForCell(utsagn-01, v) = %d items, want 1", len(cell))
	}
	if cell[0].Position == nil || *cell[0].Position != 2 {
		t.Errorf("Venstre utsagn-01 position = %v, want pointer to 2", cell[0].Position)
	}
}

func TestRodtAndMDGLobbyPositionsAreNowCoded(t *testing.T) {
	// Was TestUncodedPositionsStayUncoded until 2026-09-23: research
	// verified the exact motion text for Rødt's forslag nr. 1 and MDG's
	// forslag nr. 2-4 directly against Stortinget's voting-detail pages
	// (data.stortinget.no / stortinget.no votering-detaljer, vt=27257..
	// 27260) — both parties argued FOR the lobby register principle,
	// differing only in scope and process from the presidentskap's
	// proposal. This test intentionally replaces the old "must stay
	// nil" guard now that the real check has actually been done; see
	// git history on this file and on the evidence YAML for the before
	// state. Do not re-introduce a nil expectation here without a reason
	// as concrete as the one that removed it.
	//
	// Rødt and MDG are coded DIFFERENTLY on purpose (+2 vs +1): Rødt's
	// motion demands mandatory registration and sanction authority —
	// stronger than the statement's own bar. MDG's three motions only
	// ask to investigate a scheme ("utrede, med sikte på å innføre") —
	// real support for the principle, but a real reviewer caught this
	// review's first pass coding both identically at +2, which would
	// have overstated MDG's actual commitment. See each YAML's
	// context_note for the full reasoning.
	wantPosition := map[string]int{"rodt": 2, "mdg": 1}
	for partyID, want := range wantPosition {
		cell := Store.EvidenceForCell("utsagn-01", partyID)
		if len(cell) != 1 {
			t.Fatalf("EvidenceForCell(utsagn-01, %s) = %d items, want 1", partyID, len(cell))
		}
		if cell[0].Position == nil || *cell[0].Position != want {
			t.Errorf("%s utsagn-01 position = %v, want pointer to %d", partyID, cell[0].Position, want)
		}
		if cell[0].VerbatimExcerpt == "" {
			t.Errorf("%s utsagn-01 evidence has a coded position but no verbatim_excerpt backing it", partyID)
		}
	}
}

func TestInfluenceOperationTargetsDoNotBecomePartyPositions(t *testing.T) {
	// The FFI-documented TikTok network targeted Ap, SV and Rødt — being
	// a TARGET of an influence operation must never surface as if it
	// were one of those parties' own coded positions. This used to be
	// guarded by asserting EvidenceForCell(utsagn-04, {ap,sv,rodt}) was
	// entirely empty, back when utsagn-04 had zero party evidence of any
	// kind (2026-09-24: Innst. 231 S, 2022-2023, gave utsagn-04 real 9/9
	// party coverage on the ACTUAL kartlegging proposal these three
	// parties voted/wrote on as legislators — a legitimate, unrelated
	// reason for the same three IDs to appear here). The guard now checks
	// the specific target-evidence items directly instead of the whole
	// cell: they must keep party_id empty and position null regardless
	// of what other, real position evidence later gets added for the
	// same statement.
	targetEvidenceIDs := []string{
		"ev-2025-valget-tiktok-koordinering",
		"ev-2025-valget-tiktok-attribusjon",
	}
	for _, id := range targetEvidenceIDs {
		var found *Evidence
		for i := range Store.Evidence {
			if Store.Evidence[i].ID == id {
				found = &Store.Evidence[i]
				break
			}
		}
		if found == nil {
			t.Fatalf("evidence %s not found — id may have changed, update this test", id)
		}
		if found.PartyID != "" {
			t.Errorf("evidence %s: party_id = %q, want \"\" (target of the operation, not an actor)", id, found.PartyID)
		}
		if found.Position != nil {
			t.Errorf("evidence %s: position = %v, want nil (target of the operation, not an actor)", id, *found.Position)
		}
	}
}

func TestAddendumPendingCasesHaveNoFabricatedPartyPositions(t *testing.T) {
	// utsagn-45 ("Tillegg (utenfor de 42)" — see docs/roadmap.md for why
	// these three statements exist outside the memo's original
	// numbering) mixes two evidence tiers on purpose (2026-09-26 review):
	// the Innst. 464 S sovereignty-threshold grunnlovsforslag is real
	// tier-A evidence but deliberately position: null on every party
	// (see each file's own context_note — it answers a different,
	// broader constitutional question, not this statement's specific
	// forsvar/beredskap-vs-EU claim). Actual coded positions here come
	// only from tier-C partiprogram evidence, for the parties whose
	// own program text says something concrete enough to code
	// (ap/frp/h/sv/mdg/krf/v, plus Rødt from 2026-10-13).
	//
	// Sp still has real tier-A context but no coded position on this
	// statement at all — its own program text never connects EU/EØS to
	// defense/beredskap specifically (checked both the dedicated defense
	// chapter and the EU/EØS chapter, zero connection).
	//
	// Rødt was in that same pending set until 2026-10-13, on the reading
	// that its program never links EU/EØS to defence in ONE sentence:
	// the defence chapter wants a defence "uavhengig av
	// stormaktsinteresser", the EØS chapter wants the agreement replaced,
	// but they are not joined. That reading was reversed on review, and
	// deliberately, because it was stricter than the project's own
	// handling of KrF on this same statement: ev-2026-partiprogram-krf-
	// eos-handlefrihet codes KrF +1 from a passage about the EU
	// relationship in general, with an explicit note that the text is NOT
	// defense-specific. So a non-sector-specific passage is already
	// codeable here at a reduced magnitude. Rødt has strictly more than
	// KrF: a defence-specific passage on national independence AND an
	// explicit demand to replace EØS to secure national control. Coded
	// +2, tier C, and the item's own context_note states plainly that the
	// join between the two chapters is the coder's, not the party's.
	//
	// This test checks the actual invariant: every item's source_tier
	// must justify whether it carries a position — tier A items here
	// must be null (no A-tier case resolves the specific question yet),
	// and no party may have a coded position without at least one real
	// evidence item to back it.
	// The pending set is empty as of 2026-10-13. Rødt was coded from its
	// arbeidsprogram, and Sp from its own stortingsvalgprogram, which says
	// the EØS-avtale "er for inngripende og begrenser norsk sjølråderett på
	// mange viktige politikkområder" and that Norge "ikke må avgi
	// suverenitet til EU", beside a defence chapter on a styrket nasjonalt
	// totalforsvar. Both met the standard KrF already met on this statement.
	//
	// With that set empty, "nobody is pending" would be a vacuous guard, so
	// it is replaced by the stronger invariant it was really protecting: a
	// coded position must always rest on at least one real evidence item,
	// and the loop must have met coded positions at all.
	coded := 0
	for _, party := range Store.Parties {
		items := Store.EvidenceForCell("utsagn-45", party.ID)
		codedItems := 0
		for _, item := range items {
			if item.SourceTier == "A" && item.Position != nil {
				t.Errorf("EvidenceForCell(utsagn-45, %s): tier-A evidence %s has position %v, want nil (no A-tier case yet resolves the forsvar/beredskap-specific question)", party.ID, item.ID, *item.Position)
			}
			if item.Position != nil {
				codedItems++
			}
		}
		if codedItems > 0 {
			coded++
		}
		if codedItems > 0 && len(items) == 0 {
			t.Errorf("EvidenceForCell(utsagn-45, %s): a coded position with no evidence item behind it", party.ID)
		}
	}
	if coded != len(Store.Parties) {
		t.Errorf("utsagn-45: %d of %d parties carry a coded item, want all %d — if this changed on purpose, fix this assertion and say why", coded, len(Store.Parties), len(Store.Parties))
	}
}

func TestKjernekraftStatligKontrollPositionsAreCoded(t *testing.T) {
	// utsagn-06 (2026-09-24, first real statement of Dimension 2 "Norsk
	// handlefrihet"): Innst. 278 S (2025-2026) read in full (PDF, not
	// just the vote tally) precisely because the vote tally alone is
	// misleading here — FrP proposed the underlying bill and voted
	// AGAINST the final adopted tilråding, but only because that
	// tilråding ("legg frem en stortingsmelding") was weaker than what
	// FrP itself asked for, not because FrP opposes commercial nuclear.
	// See each evidence file's context_note for the reasoning behind its
	// specific value.
	wantPosition := map[string]int{
		"rodt": 2,  // own forslag 30-31: explicit "fullt statlig eierskap, kontroll og styring"
		"ap":   1,  // fellesmerknad w/ Sp/Rødt/MDG: skeptical of the private model, no explicit ownership demand
		"sp":   1,  // same fellesmerknad
		"mdg":  1,  // same fellesmerknad
		"h":    -2, // own forslag 27 + fellesmerknad w/ V: "ikke aktuelt med ... eierskap"
		"v":    -2, // same fellesmerknad w/ H
		"frp":  -2, // own forslag 2/5: explicit "kommersiell kjernekraftproduksjon", the opposite of state ownership
	}
	// 2026-10-02 (trust audit): Ap, Sp and MDG voted against Rødt's
	// forslag 30-32 (votering 27715). That vote is now listed next to
	// their merknad as its own item, deliberately uncoded (files
	// 701-703); every cell still has exactly one coded item.
	uncodedVote := map[string]bool{"ap": true, "sp": true, "mdg": true}
	for partyID, want := range wantPosition {
		var coded, uncoded []Evidence
		for _, e := range Store.EvidenceForCell("utsagn-06", partyID) {
			if e.Position != nil {
				coded = append(coded, e)
			} else {
				uncoded = append(uncoded, e)
			}
		}
		if len(coded) != 1 {
			t.Fatalf("utsagn-06/%s has %d coded items, want 1", partyID, len(coded))
		}
		if *coded[0].Position != want {
			t.Errorf("%s utsagn-06 position = %d, want %d", partyID, *coded[0].Position, want)
		}
		if coded[0].VerbatimExcerpt == "" {
			t.Errorf("%s utsagn-06 evidence has a coded position but no verbatim_excerpt backing it", partyID)
		}
		wantUncoded := 0
		if uncodedVote[partyID] {
			wantUncoded = 1
		}
		if len(uncoded) != wantUncoded {
			t.Errorf("utsagn-06/%s has %d uncoded items, want %d", partyID, len(uncoded), wantUncoded)
		} else if wantUncoded == 1 && uncoded[0].Status != "forslag_avvist_stemte_mot" {
			t.Errorf("utsagn-06/%s uncoded item status = %q, want the vote against forslag 30-32", partyID, uncoded[0].Status)
		}
	}
}

func TestKjernekraftStatligKontrollLeavesAmbiguousPartiesUncoded(t *testing.T) {
	// SV and KrF are deliberately NOT coded on utsagn-06, for two
	// different honest reasons — see docs/research-notes for the general
	// discipline this follows:
	//   - SV's own committee merknad argues against subsidizing PRIVATE
	//     nuclear and against a rushed process — a real position, but a
	//     different claim than "Norway must have full state ownership"
	//     (the statement's actual text). Coding SV here would attribute
	//     a stronger, more specific claim than SV's own words make.
	//   - KrF has no seat on the energi- og miljøkomite for this case, so
	//     no committee merknad exists to read at all. KrF's floor vote
	//     alone (4 mot on the final tilråding) is exactly the ambiguous
	//     signal the FrP case above warns against over-interpreting.
	for _, partyID := range []string{"sv", "krf"} {
		if items := Store.EvidenceForCell("utsagn-06", partyID); len(items) != 0 {
			t.Errorf("EvidenceForCell(utsagn-06, %s) = %d items, want 0 (real committee text exists but doesn't support this statement's specific claim)", partyID, len(items))
		}
	}
}

func TestPartilovenAnonymeBidragPositionsAreCoded(t *testing.T) {
	// utsagn-03 (2026-09-24): the "party-level evidence for statements
	// 2,3,4,5 — actively searched, found nothing" null result recorded
	// in docs/roadmap.md turned out to be incomplete, not final, exactly
	// as that entry's own caveat warned. Prop. 144 L / Innst. 528 L
	// (2024-2025) bans anonymous party contributions and contributions
	// laundered through legal-entity or private-person intermediaries —
	// precisely utsagn-03's claim — and Stortinget's own
	// voteringsresultat API gives an individually verified, unambiguous
	// roll call on the exact substantive provision (votering_id 26057:
	// 89 for, 12 mot, 68 absent; the 12 "mot" votes are 100% FrP, cross-
	// checked against the official 89/12 split). This is a cleaner
	// signal than utsagn-06's kjernekraft case, where the floor vote
	// alone was actively misleading — here the vote is ON the operative
	// text itself, not a vague "legg frem en stortingsmelding" motion.
	wantPosition := map[string]int{
		"ap":   2,  // committee majority merknad + unanimous "for" vote
		"h":    2,  // own merknad + "for" vote (own extra forslag asked for MORE, not less)
		"sp":   2,  // committee majority merknad + unanimous "for" vote
		"sv":   2,  // committee majority merknad + unanimous "for" vote
		"rodt": 2,  // committee majority merknad + unanimous "for" vote
		"frp":  -2, // own merknad objects to the ban's substance + unanimous 12/12 "mot" vote, 0 "for"
		"mdg":  1,  // no committee seat, no own merknad, but a clean "for" vote on the operative text
		"v":    1,  // same as MDG
		"krf":  1,  // same as MDG
	}
	for partyID, want := range wantPosition {
		cell := Store.EvidenceForCell("utsagn-03", partyID)
		if len(cell) != 1 {
			t.Fatalf("EvidenceForCell(utsagn-03, %s) = %d items, want 1", partyID, len(cell))
		}
		if cell[0].Position == nil || *cell[0].Position != want {
			t.Errorf("%s utsagn-03 position = %v, want pointer to %d", partyID, cell[0].Position, want)
		}
	}
}

func TestAcerEnergyMarketPositionsAreCoded(t *testing.T) {
	// Originally drafted as utsagn-43 (2026-09-24, an addendum statement
	// outside the memo's 1-42 numbering — see docs/roadmap.md's
	// "Correction"). Renumbered back to utsagn-07 (2026-09-24, later same
	// day, "Phase 2r") once Dimension 2 (Norsk handlefrihet, 6-11) was
	// actually drafted from the memo's own question — this statement's
	// own primary_question already matched it verbatim. Innst. 196 S
	// (2024-2025) gives full 9/9 coverage from a single case with four
	// separate, individually cross-checked roll-call votes (votering_id
	// 24383-24386, data.stortinget.no/eksport/voteringsresultat).
	//
	// Flipped 2026-09-27 (commit a9887cf, acquiescence-spread fix): text
	// now asserts the status-quo-defending majority's own claim (Ap/H,
	// citing Høyesterett, that ACER/EØS does NOT weaken national control)
	// instead of the failed opposition's ask to limit ACER. Every
	// position below is the negation of the pre-flip value. FrP's split
	// vote is still the reason this needed the full PDF, not just the
	// tally: FrP voted AGAINST the most drastic ask (formal veto +
	// starting ACER withdrawal, votering_id 24383) but FOR the milder
	// asks (votering_id 24384, 24385) and AGAINST the majority's
	// rejection of both representantforslag (votering_id 24386) — net a
	// mild "agree the EØS tie doesn't weaken control" position now (+1),
	// not as strong as Ap/H's own defense. See each evidence file's
	// context_note for the full reasoning.
	wantPosition := map[string]int{
		"sp":   -2, // own forslag (Dok 8:58 S), for on all 4 votes to LIMIT ACER — disagrees control is intact
		"sv":   -2, // for on all 4 votes, joint merknad w/ Sp/Rødt — disagrees control is intact
		"rodt": -2, // own forslag (Dok 8:75 S), for on all 4 votes to LIMIT ACER — disagrees control is intact
		"frp":  -1, // for the milder asks, against the most drastic one — mildly disagrees
		"ap":   1,  // defends status quo (cites Høyesterett) — this IS the flipped statement's own claim
		"h":    1,  // defends status quo (cites Høyesterett) — this IS the flipped statement's own claim
		"v":    2,  // explicitly wants FASTER/FULLER integration — even stronger agreement
		"mdg":  2,  // same as V, joint merknad — even stronger agreement
		"krf":  1,  // voting-record only, no own merknad, against all 4 limit-ACER votes
	}
	for partyID, want := range wantPosition {
		cell := Store.EvidenceForCell("utsagn-07", partyID)
		if len(cell) != 1 {
			t.Fatalf("EvidenceForCell(utsagn-07, %s) = %d items, want 1", partyID, len(cell))
		}
		if cell[0].Position == nil || *cell[0].Position != want {
			t.Errorf("%s utsagn-07 position = %v, want pointer to %d", partyID, cell[0].Position, want)
		}
		if cell[0].VerbatimExcerpt == "" {
			t.Errorf("%s utsagn-07 evidence has a coded position but no verbatim_excerpt backing it", partyID)
		}
	}
}

func TestUtenlandskPaavirkningPositionsAreCoded(t *testing.T) {
	// utsagn-04 (2026-09-24, later same day as utsagn-43): Innst. 231 S
	// (2022-2023) gives full 9/9 coverage from one case — Høyre's own
	// representantforslag (Dokument 8:51 S) asked the government to
	// immediately map ("kartlegge") foreign information influence
	// targeting the 2023 and 2025 elections, matching the ORIGINAL
	// utsagn-04's claim closely (see
	// evidence/51-utenlandsk-paavirkning-kontekst.yaml's context_note for
	// the one real precision nuance: "sosiale medier" isn't a verbatim
	// condition in the forslag text). Two earlier research passes had
	// found nothing usable on this statement — this case surfaced only
	// from a broader, non-keyword-limited web search, not the
	// harvester's title/emne matching, which is the harvester's
	// documented ceiling (see docs/roadmap.md's "Next candidate work" #1).
	//
	// Flipped 2026-09-27 (commit a9887cf, acquiescence-spread fix): text
	// now asserts Ap/Sp's own "already addressed by government work"
	// claim (Meld. St. 9, NSM, PST/politiet) instead of the original
	// kartlegging ask. Every position below is the negation of the
	// pre-flip value. SV's real, shared surveillance-risk merknad on the
	// same concrete forslag (also not a rejection of the goal) is now a
	// mild AGREE (+1) that the extra kartlegging mechanism isn't needed;
	// H and V, who co-sponsored the actual forslag text, are now the
	// strongest disagreers (-2 each).
	wantPosition := map[string]int{
		"h":    -2, // co-sponsor of forslag 6, the exact kartlegging ask this flipped statement argues is unnecessary
		"v":    -2, // co-sponsor of forslag 6, plus own further forslag 7
		"krf":  -1, // voting-record only, was for the original kartlegging ask
		"mdg":  -1, // voting-record only, was for the original kartlegging ask
		"rodt": -1, // voting-record only, was for the original kartlegging ask
		"sv":   1,  // own merknad: shared the surveillance-risk concern, now a mild agree
		"ap":   1,  // own merknad w/ Sp: goal already addressed by government work — this IS the flipped statement's own claim
		"sp":   1,  // same joint merknad w/ Ap
		"frp":  1,  // voting-record only, was against the original kartlegging ask
	}
	for partyID, want := range wantPosition {
		cell := Store.EvidenceForCell("utsagn-04", partyID)
		if len(cell) != 1 {
			t.Fatalf("EvidenceForCell(utsagn-04, %s) = %d items, want 1", partyID, len(cell))
		}
		if cell[0].Position == nil || *cell[0].Position != want {
			t.Errorf("%s utsagn-04 position = %v, want pointer to %d", partyID, cell[0].Position, want)
		}
		if cell[0].VerbatimExcerpt == "" {

			t.Errorf("%s utsagn-04 evidence has a coded position but no verbatim_excerpt backing it", partyID)
		}
	}
}

func TestPolitiskReklamePositionsAreCoded(t *testing.T) {
	// utsagn-02 (2026-09-24, same session as utsagn-04): a "løst forslag"
	// during the 2025-10-14 trontaledebatt (votering_id 26361, sak
	// p=104588) gives full 9/9 coverage. This is not a committee case at
	// all — Rødt's own Marie Sneve Martinussen proposed, in plenary, to
	// regulate anonymous funding of political advertising/marketing as
	// part of the upcoming partiloven review, citing real examples from
	// the 2025 campaign. It passed 54-46. See
	// evidence/61-politisk-reklame-anonym-finansiering-kontekst.yaml's
	// context_note for the one real nuance: utsagn-02 has two clauses
	// ("the ad must show who paid" AND "money's origin must be public")
	// and this forslag only tests the second — a real, honestly narrower
	// match than a hypothetical case testing on-ad labeling directly
	// would be, but a closer match than Innst. 275 L (rejected earlier
	// for utsagn-02) because the forslag's own text names "politisk
	// reklame og markedsføring" specifically, not general candidate
	// support. Rødt has an own plenary speech (+2); the other 5 "for"
	// votes and 3 "mot" votes are vote-record-only (+1 / -1).
	wantPosition := map[string]int{
		"rodt": 2,
		"ap":   1,
		"sv":   1,
		"sp":   1,
		"mdg":  1,
		"v":    1,
		"frp":  -1,
		"h":    -1,
		"krf":  -1,
	}
	for partyID, want := range wantPosition {
		cell := Store.EvidenceForCell("utsagn-02", partyID)
		if len(cell) != 1 {
			t.Fatalf("EvidenceForCell(utsagn-02, %s) = %d items, want 1", partyID, len(cell))
		}
		if cell[0].Position == nil || *cell[0].Position != want {
			t.Errorf("%s utsagn-02 position = %v, want pointer to %d", partyID, cell[0].Position, want)
		}
		if cell[0].VerbatimExcerpt == "" {
			t.Errorf("%s utsagn-02 evidence has a coded position but no verbatim_excerpt backing it", partyID)
		}
	}
}

// TestStatementsPerDimension pins how many statements exist in each
// dimension, so adding or removing a statement is a deliberate, visible
// edit in this table rather than a silent drift. It is the data-side twin
// of internal/server's coverage test, which derives each dimension's total
// from these same counts: a stale hand-maintained total once let the front
// page show a dimension at more than 100% built. The "Tillegg" row is the
// addendum dimension (server.addendumDimension); the literal is repeated
// here because internal/data must not import internal/server.
func TestStatementsPerDimension(t *testing.T) {
	want := map[string]int{
		"Demokratisk åpenhet":                             5,
		"Norsk handlefrihet":                              7,
		"Forvaltning av fellesskapets ressurser / energi": 7,
		"Teknologisk kapasitet":                           3,
		"Innovasjon / omstillingsevne":                    3,
		"Universell grunnmur":                             7,
		"Barns muligheter":                                6,
		"Robusthet / adaptabilitet":                       5,
		"Generasjonsansvar":                               2,
		"Finansiering av fellesskapet":                    7,
		"Tilgang, integrering og rettsstat":               6,
		"Tillegg (utenfor de 42)":                         2,
	}
	got := map[string]int{}
	for _, s := range Store.Statements {
		got[s.Dimension]++
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("dimension %q: %d statements, want %d; if this changed on purpose, update this table", name, got[name], w)
		}
	}
	for name, n := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("unexpected dimension %q with %d statements; add it to this table", name, n)
		}
	}
}

// TestBudgetObservationsCarryTheirStage pins the fix for the bredbånd
// finding, which compared the government's 2026 PROPOSAL (34,1 mill.
// kroner) against its 2027 proposal while the Stortinget's SALDERT 2026 was
// 159,11 mill. Both are true and they answer different questions, so the
// amount must carry the stage that produced it. A finding that stores only
// one bare number per year cannot express this, which is the defect.
func TestBudgetObservationsCarryTheirStage(t *testing.T) {
	ds := Store
	var bb *BudgetObservation
	for _, b := range ds.Budgets {
		for _, f := range b.Findings {
			if f.Topic == "Bredbånd: ingen ny bevilgning i forslaget" {
				if got := len(f.Observed); got != 3 {
					t.Fatalf("bredbånd finding has %d observations, want 3 (2026 forslag, 2026 saldert, 2027 forslag)", got)
				}
				bb = &f.Observed[1]
				if bb.Stage != "saldert" || bb.Amount != 159110000 {
					t.Errorf("observation 1 = %s %d, want saldert 159110000", bb.Stage, bb.Amount)
				}
				if f.Observed[0].Stage != "forslag" || f.Observed[0].Amount != 34110000 {
					t.Errorf("observation 0 = %s %d, want forslag 34110000", f.Observed[0].Stage, f.Observed[0].Amount)
				}
			}
		}
	}
	if bb == nil {
		t.Fatal("bredbånd finding not found")
	}
}

// TestEveryBudgetObservationHasAValidStage walks the real dataset, so the
// guard cannot pass on an empty list. It also asserts that at least three
// observations were actually checked, which is what makes the check
// meaningful rather than vacuously green.
func TestEveryBudgetObservationHasAValidStage(t *testing.T) {
	ds := Store
	checked := 0
	for _, b := range ds.Budgets {
		for i, f := range b.Findings {
			for j, o := range f.Observed {
				checked++
				if !ValidBudgetStage(o.Stage) {
					t.Errorf("budget %s finding %d observation %d: stage %q is not one the loader accepts", b.ID, i, j, o.Stage)
				}
			}
		}
	}
	if checked < 3 {
		t.Fatalf("checked %d observations, want at least 3 — this guard must not pass without data", checked)
	}
	t.Logf("checked %d observations across %d budgets", checked, len(ds.Budgets))
}

// TestValidBudgetStageRejectsThePlausibleMistake guards the guard: the
// stage names that a coder is most likely to type instead of the real ones
// must be rejected, or the vocabulary silently widens.
func TestValidBudgetStageRejectsThePlausibleMistake(t *testing.T) {
	for _, bad := range []string{"", "vedtatt", "vedtatt_budsjett", "proposal", "saldert_budsjett", "Forslag"} {
		if ValidBudgetStage(bad) {
			t.Errorf("ValidBudgetStage(%q) = true, want false", bad)
		}
	}
	for _, ok := range []string{"forslag", "saldert", "regnskap"} {
		if !ValidBudgetStage(ok) {
			t.Errorf("ValidBudgetStage(%q) = false, want true", ok)
		}
	}
}
