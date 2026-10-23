# Design review, 2026-09-23 — triage

An external AI tool reviewed the live site's HTML/JS source (not just
appearance) and produced 11 points. Two were real, verifiable defects and
have already been fixed and deployed (see the `fix(ektevalgomat):` commit
same date). This doc triages the other 9, which are product/framing
decisions, not bugs — each needs a call from Jon-Mikkel, not a unilateral
rewrite, especially the statement-wording ones on a project whose whole
premise is neutral, evidence-first framing.

## Already fixed (not in this triage)

- No-JS fallback was a dead end (`<ul id="noscript-links">` was empty).
  Fixed: server now renders real `<a href="/utsagn/N">` links, locked in by
  a test.
- "Ingenting sendes til noen server" was true but unenforced. Fixed: a Go
  test now scans embedded static JS for `fetch`/`XHR`/`sendBeacon`/
  `WebSocket`/`EventSource` and fails the build if one appears unallowed.

## Needs a decision: statement wording (items 6, 7 in the review)

The reviewer flagged that utsagn-43 and utsagn-45's current text embeds a
contested framing rather than a neutral one:

- **utsagn-43** currently: "...begrense ACERs reguleringsmakt over norske
  kraftpriser." The reviewer's point, worth checking against ACER's own
  described mandate before accepting or rejecting: ACER's role is
  coordinating national regulators and ruling on certain cross-border
  methodologies when national regulators can't agree — not the same claim
  as "ACER sets Norwegian power prices." If that's accurate, the statement
  as written pre-loads an institutional claim into what should be a
  values question.
- **utsagn-45** currently: "...EU-mandater som kan begrense norsk
  operativt handlingsrom." "Som kan begrende" pre-loads a negative framing
  before the respondent answers.

**Before rewriting either:** independently verify the ACER characterization
against ACER's own published mandate description and NVE/RME material —
don't take the reviewing AI's characterization on faith any more than any
other unverified claim in this project's history. If it holds up, both
statements are candidates for a neutral rewrite that states the tradeoff
explicitly (national control vs. shared market rules) rather than
resolving it in the question's wording. This is the same class of mistake
as the fabricated dimension names — a subtly biased statement is worse
than an empty one.

**Decision needed:** rewrite now, defer to when addendum 43-45 get folded
into Dimension 2 proper (see roadmap.md), or leave as-is with a documented
rationale for why the current wording is fine.

## Needs a decision: first five statements' "friction" (item 5)

The reviewer's point: utsagn-01 through 05 (Demokratisk åpenhet) may be
phrased so that agreeing is obviously the sympathetic answer, with no
visible legitimate interest on the other side — a lobbyist register or
transparency mandate is easy to nod along to in the abstract. Suggested
fix pattern: reframe toward the actual tradeoff (e.g. "lobby contacts with
political leadership should be public, including when that reveals a
specific client relationship and the meeting's topic" — transparency vs.
confidentiality/business-sensitivity is now a real measured tension).

This is a genuine methodological point about what makes a valgomat useful
(a question everyone answers the same way measures nothing) but reworking
five already-partially-evidenced statements has a cost: any existing coded
party positions (Venstre/Rødt/MDG on utsagn-01) would need re-review
against the new wording. Decide whether this is worth doing before or
after Task 0 in `research-briefs/TASKS.md` (deepening evidence on the
current five statements) — doing it after means re-coding evidence twice.

## Needs a decision: intro length / progressive disclosure (items 3, 11)

Current order on `/`: snapshot badge → what-this-is-not → coding
methodology → data coverage → link to method page → privacy line → quiz
mount point. The reviewer's proposed alternative: title, one-line pitch,
"8 spørsmål · ca. 3 min", a Start button, three checkmark trust signals,
and a "hvordan fungerer dette?" disclosure — with everything else moved to
a second level.

This is a real, legitimate UX tradeoff (transparency-first vs.
attention-cost-first) but not a small change — it restructures the
landing page's whole information hierarchy. Compatible with the intro/
quiz split already used on statement pages (Scan/Forstå/Verifiser, see
`web/templates/statement.html.tmpl`), so the pattern exists; extending it
to the index page is the work.

## Needs a decision: one-question-at-a-time quiz flow (item 11 continued)

The reviewer's structure-render (a reconstruction, not a byte-identical
copy — it didn't have access to `style.css` or `quiz.js`) proposes:
theme → question → 5-point answer scale → "vet ikke / forklar meg først" →
optional depth tiers ("20 sek kort forklart" / "2 min kontekst" /
"dokumentasjon"). This is close to the existing statement-page pattern but
applied to the quiz-taking flow itself, which currently isn't described in
what I reviewed here — needs checking against `quiz.js`'s actual current
UI before deciding what changes.

## Needs a decision: separating evidensstyrke / count / konsistens (items
8, 9)

Real, specific UX confusion the reviewer identified: a cell can show
`evidensstyrke: "Sterk"` with `evidensstyrke_count: 1` and
`konsistens: "Utilstrekkelig grunnlag"` in the same view — which can read
as self-contradictory ("strong evidence" next to "insufficient basis") even
when it's correct (one very direct primary-source vote can genuinely be
strong evidence on its own, and "insufficient basis for a *consistency*
judgment" is a different, compatible claim about a different question:
does one datapoint support saying anything about pattern-over-time).

Proposed reframe: display three separately-labeled facts instead of one
overloaded word doing multiple jobs — source directness ("Direkte
primærkilde"), observation count ("1 kilde"), and consistency assessment
("Ikke vurderbar ennå") — so "sterk" only ever means one thing. This is a
presentation-layer change (`evidenceClassBadge`/`evidensstyrkeDots` in
`internal/server/server.go` and `statement.html.tmpl`), not a data-model
change — the fields already exist separately, they're just being narrated
together.

## Cosmetic, low-risk, cheap to accept whenever

- Rename "Snapshot" language toward something more legible ("Datagrunnlag
  per 23. september 2026") while keeping "snapshot" in technical metadata.
  Trivial copy change, no schema/logic impact.
- Sharpen the js-only messaging: not "ingen JavaScript" (the page does use
  JS) but "ingen konto, ingen server-innsending, quizlogikk kjører lokalt"
  — more accurate and, per the review, more interesting than a blanket
  "no JS" claim would be anyway. Also trivial.

## Not actionable yet

Item about the "struktur-render" HTML file itself: it's explicitly a
reconstruction from what was pasted into that tool's context, not a
byte-identical fetch of the live page (the reviewing tool's sandbox
couldn't load `style.css`/`quiz.js` from the real environment) — treat its
visual mockup as a sketch of an idea, not a spec to copy pixel-for-pixel.
