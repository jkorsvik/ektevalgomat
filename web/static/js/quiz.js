// Ekte valgomat — quiz client. Vanilla JS, no build step, no framework.
// Everything the quiz needs (statements, parties, precomputed cells) is
// injected server-side as a JSON script tag — this file never fetches
// anything and never sends anything: the user's answers live only in
// this tab's memory plus localStorage, never a server request.
(function () {
  "use strict";

  var dataEl = document.getElementById("ektevalgomat-data");
  if (!dataEl) return;
  var DATA = JSON.parse(dataEl.textContent);

  var STORAGE_KEY = "ektevalgomat-svar-v1";
  // The five positions on the -2..+2 scale. 0 ("Nøytral") is a genuine
  // middle position and still counts in the match — it is NOT the same as
  // "vet ikke".
  var SCALE = [
    { value: -2, label: "Sterkt uenig" },
    { value: -1, label: "Uenig" },
    { value: 0, label: "Nøytral" },
    { value: 1, label: "Enig" },
    { value: 2, label: "Sterkt enig" },
  ];
  // "Vet ikke / hopp over" is a separate sixth choice, deliberately OFF
  // the scale. SKIP is a string, not a number, so pct() returns null for
  // it and the answer is excluded from every party-match computation —
  // neither agreement nor disagreement. This is the same distinction the
  // resultat text already makes for parties: "Ikke vurdert" does not mean
  // "nøytralt" (see renderResults below). The pure scoring lives in
  // scoring.js, loaded before this file.
  var SKIP = EkteScoring.SKIP;
  var SKIP_LABEL = "Vet ikke / hopp over";
  var isPosition = EkteScoring.isPosition;
  var pct = EkteScoring.pct;
  var weightOf = EkteScoring.weightOf;
  var MAX_WEIGHT = EkteScoring.MAX_WEIGHT;
  var leaders = EkteScoring.leaders;

  // 2026-09-27: these statements were rewritten to the opposite direction
  // (and every party code negated) so that always-agree no longer lands on
  // one side. An answer saved before that date refers to the OLD text, so
  // it is negated once on load; without this, a returning user's "enig"
  // would silently count as agreement with the opposite position. The
  // snapshot date is locked on purpose, so the version field marks this.
  var DIRECTION_VERSION = 2;
  var FLIPPED_2026_09_27 = [
    "utsagn-04", "utsagn-07", "utsagn-08", "utsagn-12", "utsagn-18",
    "utsagn-19", "utsagn-20", "utsagn-21", "utsagn-24", "utsagn-25",
    "utsagn-31", "utsagn-37", "utsagn-44",
  ];

  // Ephemeral mode ("midlertidig økt"): the visitor asked to take the quiz
  // without anything being stored on this device, e.g. on a shared phone.
  // When true, saveAnswers() and clearAllAnswers() never touch localStorage
  // at all — not a write, not a removal. It is a per-session choice and is
  // deliberately NOT persisted itself: persisting it would be the very write
  // it promises not to make. Default off.
  var ephemeral = false;

  // View mode: "enkel" (one question at a time, the default) or "utforsker"
  // (the full list, for scanning everything). Persisted in its own key so
  // the choice survives a reload. It is a display preference, not an answer;
  // a write for it is skipped while ephemeral mode is on.
  var VIEW_KEY = "ektevalgomat-visning-v1";
  function loadViewMode() {
    try {
      return localStorage.getItem(VIEW_KEY) === "utforsker" ? "utforsker" : "enkel";
    } catch (e) {
      return "enkel";
    }
  }
  function persistViewMode(mode) {
    if (ephemeral) return;
    try {
      localStorage.setItem(VIEW_KEY, mode === "utforsker" ? "utforsker" : "enkel");
    } catch (e) {
      /* localStorage unavailable (private mode etc.) — the choice just
         won't survive a reload; the quiz still works */
    }
  }
  var viewMode = loadViewMode();

  // Importance weighting (2026-10-13). A weight says how much a statement
  // should pull the result; it is NOT a position and never enters an
  // answer. It lives in its own key, keyed by the same snapshot date, so a
  // changed dataset drops it exactly like the answers. One weight per
  // statement: absent means "vanlig" (1).
  var WEIGHT_KEY = "ektevalgomat-vekt-v1";
  var WEIGHT_LEVELS = [
    { value: 1, label: "Vanlig" },
    { value: 2, label: "Viktig" },
    { value: 3, label: "Sv\u00e6rt viktig" },
  ];
  // --- persistence (the rules live in store.js) ------------------------
  // A visitor's choices are the one thing on this site that cannot be
  // rebuilt, so every read and write goes through EkteStore. Measured on
  // 2026-10-16 before this change: two open tabs erased each other's answers,
  // a failed write still said "kept", a data refresh dropped every answer, and
  // a reload with 30 answered started again at question 1.
  var Store = EkteStore;
  var storage = (function () {
    try { return window.localStorage; } catch (e) { return null; }
  })();
  // What THIS tab last loaded or saved. write() sends only the difference, so
  // a tab that loaded earlier cannot erase what another tab saved since.
  var savedAnswers = {};
  var savedWeights = {};
  // Trade-off records (docs/design/2026-10-13-avveining.md): own key, same
  // rules as the answers and weights, each stored under the PAIR of
  // statements it decides between. A record only EXPLAINS the weights, so it
  // is believed only while they still match it (AV.consistent). That makes a
  // stale write from another tab self-correcting instead of leaving a log
  // that names a winner the score is not using.
  var AV = EkteAvveining;
  var AVV_KEY = "ektevalgomat-avveining-v1";
  var savedRecords = {};
  var records = [];
  var avvLater = false; // "Ikke nå": this page view only, never stored
  var loadReport = { kept: 0, dropped: 0 };
  var noteEl = document.getElementById("storage-note");
  var noteState = { nosave: false, dropped: null, resume: null };

  // paintNote shows at most one note: a save that cannot happen outranks
  // everything, then the report of answers a refresh could not carry over,
  // then the "you left off here" line. In ephemeral mode nothing is saved
  // on purpose, so the cannot-save warning would be noise.
  //
  // #storage-note is a role="status" container that is ALWAYS in the DOM and
  // empty until there is something to say; the note is a child inserted into
  // it. A live region must exist before its content changes, so a container
  // toggled with `hidden` (removed from the accessibility tree) announces
  // nothing. And it is its own region on purpose: the same sentence sent
  // through announce() was overwritten in the same tick by "Utsagn 1: Sterkt
  // uenig.", so a screen-reader user heard the answer was kept at the very
  // moment it was not.
  function paintNote() {
    if (!noteEl) return;
    var warn = noteState.nosave && !ephemeral;
    var text = warn ? Store.NOSAVE_MESSAGE : (noteState.dropped || noteState.resume || "");
    var cls = "storage-note" + (warn ? " storage-note-warn" : "");
    var cur = noteEl.firstChild;
    if (cur && cur.textContent === text && cur.className === cls) return; // do not re-announce
    noteEl.textContent = "";
    if (!text) return;
    var p = document.createElement("p");
    p.className = cls;
    p.textContent = text;
    noteEl.appendChild(p);
  }
  function setNoSave(flag) {
    if (flag === noteState.nosave) return;
    noteState.nosave = flag;
    paintNote();
  }
  // settleSave is the only place a write result is read. A failed write is
  // never reported as saved, and the changes stay unsaved in memory so the
  // next successful write carries them.
  function settleSave(res, target, which) {
    if (!res.ok) { setNoSave(true); return; }
    if (which === "records") {
      savedRecords = Object.assign({}, res.map);
      records = AV.consistent(weights, AV.fromMap(res.map));
    } else {
      Store.replaceInPlace(target, res.map);
      if (which === "answers") {
        savedAnswers = Object.assign({}, res.map);
      } else {
        savedWeights = Object.assign({}, res.map);
        // The merge may have taken in another tab's weights, so a record that
        // explained the old ones may no longer be true.
        records = AV.consistent(weights, records);
      }
    }
    setNoSave(false);
  }

  function loadWeights() {
    var r = Store.read(storage, WEIGHT_KEY, "weights", DATA);
    savedWeights = Object.assign({}, r.map);
    if (r.legacy) saveWeights(r.map);
    return r.map;
  }
  function saveWeights(w) {
    if (ephemeral) return;
    settleSave(Store.write(storage, WEIGHT_KEY, "weights", DATA, w, savedWeights), w, "weights");
  }
  function loadRecords() {
    var r = Store.read(storage, AVV_KEY, "avveininger", DATA, AV.storeOpts);
    savedRecords = Object.assign({}, r.map);
    return AV.consistent(weights, AV.fromMap(r.map));
  }
  function saveRecords() {
    if (ephemeral) return;
    settleSave(
      Store.write(storage, AVV_KEY, "avveininger", DATA, AV.toMap(records), savedRecords, null, AV.storeOpts),
      null,
      "records"
    );
  }
  function countWeighted() {
    var n = 0;
    DATA.statements.forEach(function (st) {
      if (weightOf(weights, st.id) > 1) n++;
    });
    return n;
  }

  function loadAnswers() {
    var r = Store.read(storage, STORAGE_KEY, "answers", DATA);
    loadReport = { kept: r.kept, dropped: r.dropped };
    var saved = r.map;
    savedAnswers = Object.assign({}, saved);
    var version = (r.blob && r.blob.direction_version) || 1;
    // Only migrate what was actually kept: saving an empty set here would
    // overwrite the stored answers a refresh just could not carry over,
    // before the visitor was even told.
    if (r.blob && r.kept > 0 && (version < DIRECTION_VERSION || r.legacy)) {
      if (version < DIRECTION_VERSION) {
        FLIPPED_2026_09_27.forEach(function (id) {
          if (typeof saved[id] === "number") saved[id] = -saved[id];
        });
      }
      // A blob from before fingerprints is rewritten the first time it is
      // seen, so a returning visitor who only looks still keeps their answers
      // through the next data refresh.
      saveAnswers(saved);
    }
    return saved;
  }

  function saveAnswers(a) {
    // Ephemeral mode must genuinely not write: not the answers, not a
    // migration, nothing. This early return is the whole guarantee.
    if (ephemeral) return;
    settleSave(
      Store.write(storage, STORAGE_KEY, "answers", DATA, a, savedAnswers, { direction_version: DIRECTION_VERSION }),
      a,
      "answers"
    );
  }

  var answers = loadAnswers();
  var weights = loadWeights();
  records = loadRecords();
  var root = document.getElementById("quiz-root");
  // Which page this script is running on: "quiz" (the 44 statement
  // cards, on /) or "resultat" (the running match, on /resultat — see
  // resultat.html.tmpl). Read from the body itself, not guessed from
  // which elements exist, since both pages share this one file
  // (2026-09-26: Resultat used to be an in-page tab here, but that made
  // it un-linkable and missing from the nav next to Om/Metode — it's
  // now a real page with its own URL, still rendered purely from
  // localStorage, never a server round trip).
  var PAGE = document.body.getAttribute("data-page") || "quiz";

  // A reply counts as "besvart" (answered) only when it is a position on
  // the scale. A stored SKIP is a real reply with no position, so it is
  // counted and shown separately (countSkipped) and never enters the
  // match. Both count over DATA.statements, not the raw answers map, so a
  // stale id can never inflate a number.
  function countAnswered() {
    var n = 0;
    DATA.statements.forEach(function (st) {
      if (isPosition(answers[st.id])) n++;
    });
    return n;
  }

  function countSkipped() {
    var n = 0;
    DATA.statements.forEach(function (st) {
      if (answers[st.id] === SKIP) n++;
    });
    return n;
  }

  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }

  // A polite live region so a screen-reader user hears that an answer
  // registered (or that all answers were deleted) even though focus does
  // not move — the same information the visual pressed state and the
  // running progress count give a sighted user. Created once, appended
  // outside #quiz-root, so the re-render that clears the quiz never
  // removes it. Visually hidden via CSS, never style="" (style-src has no
  // 'unsafe-inline' here).
  var statusRegion = el("p", "visually-hidden");
  statusRegion.setAttribute("role", "status");
  statusRegion.setAttribute("aria-live", "polite");
  document.body.appendChild(statusRegion);

  function announce(message) {
    statusRegion.textContent = message;
  }

  // appendArrow adds the decorative link arrow as a child element that
  // assistive tech skips. Set as part of textContent (until 2026-10-13) it
  // ended up in the link's accessible name, so a screen reader announced
  // "Se kilden bak dette utsagnet, pil hoyre". The arrow carries no meaning
  // the words do not already carry.
  function appendArrow(parent) {
    var arrow = el("span", null, "\u2192");
    arrow.setAttribute("aria-hidden", "true");
    parent.appendChild(arrow);
  }

  // slugify must produce the exact same anchor id as the Go side's
  // dimSlug (internal/server/server.go) for a given dimension name, or
  // the front page's "Kjerneverdiene" links land nowhere. Same rule,
  // including the æ/ø/å transliteration: html/template percent-encodes
  // non-ASCII bytes in the href="#..." it renders, so the slug must be
  // pure ASCII to avoid depending on browsers percent-decoding the
  // fragment correctly before matching it against an element id.
  function slugify(name) {
    return name
      .toLowerCase()
      .replace(/æ/g, "ae")
      .replace(/ø/g, "o")
      .replace(/å/g, "a")
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "");
  }

  // Groups DATA.statements by dimension, preserving each dimension's
  // first-appearance order — NOT simply "did the dimension change
  // since the last statement", because statement numbering is not
  // always contiguous per dimension: utsagn 19/20/24 are all
  // "Teknologisk kapasitet" with utsagn 21-23 ("Innovasjon /
  // omstillingsevne") built in between them (see docs/roadmap.md's
  // Phase 2 history — dimensions were filled in the order evidence was
  // found, not strict number order). A naive "heading on change"
  // approach would render a second "Teknologisk kapasitet" heading
  // later with a duplicate id, breaking the front page's dimension
  // links. Computed once, not per render call, since DATA never
  // changes after page load.
  function groupStatementsByDimension(statements) {
    var order = [];
    var byDim = {};
    statements.forEach(function (st) {
      if (!byDim[st.dimension]) {
        byDim[st.dimension] = [];
        order.push(st.dimension);
      }
      byDim[st.dimension].push(st);
    });
    return order.map(function (name) {
      return { name: name, slug: slugify(name), statements: byDim[name] };
    });
  }

  var DIMENSION_GROUPS = groupStatementsByDimension(DATA.statements);

  // Keeps the front page's "Kjerneverdiene" dimension rows (rendered
  // server-side in index.html.tmpl, defaulting to "0 av Y besvart" /
  // an empty bar) in sync with the visitor's own answers. Those rows
  // live outside #quiz-root, in the static part of the page, so this
  // runs as its own DOM pass rather than through renderQuiz() — a
  // no-op wherever the rows don't exist (the /resultat page, or a
  // future dimension with zero statements built yet, which keeps its
  // honest "0 av Y" default). This is deliberately NOT the "Built/
  // Total utsagn bygget" dev-completeness number that used to live
  // here (2026-09-26 review: "ikke vis hvor mange som er bygd ut, men
  // hvor mange du har svart på") — it's the visitor's own progress per
  // dimension, nothing about the project's authoring state.
  function updateDimensionOverview() {
    DIMENSION_GROUPS.forEach(function (group) {
      var row = document.getElementById("dim-row-" + group.slug);
      if (!row) return;
      var total = group.statements.length;
      var answered = group.statements.filter(function (st) {
        return isPosition(answers[st.id]);
      }).length;
      var complete = total > 0 && answered === total;
      row.classList.toggle("dim-row-complete", complete);
      var countEl = row.querySelector(".dim-count");
      if (countEl) {
        countEl.textContent = answered + " av " + total + " besvart" + (complete ? " ✓" : "");
      }
      var fill = row.querySelector(".bar-fill");
      if (fill) {
        var pct = total > 0 ? Math.round((answered / total) * 100) : 0;
        fill.className = "bar-fill " + pctStepClass(pct);
      }
    });
  }

  // renderQuiz is the single re-render entry point: every answer click
  // calls it again. It just dispatches on PAGE — the quiz page (44
  // cards) and the resultat page (running match) are two different
  // routes now, never two panes of the same page (2026-09-26).
  function renderQuiz() {
    root.innerHTML = "";
    if (PAGE === "resultat") {
      renderResultatPage();
    } else {
      renderQuizPage();
    }
    updateDimensionOverview();
  }

  // markAnswered updates ONE card's buttons in place after an answer,
  // instead of re-rendering the whole quiz. This is the focus fix
  // (2026-10-12): the old code called renderQuiz() on every click, which
  // does root.innerHTML = "" and rebuilds every card — removing the very
  // button that had focus, so a keyboard user lost their place after each
  // question. Updating in place keeps focus on the activated button, its
  // visible pressed state and its <details> state intact (WCAG 2.4.7
  // focus visible; 2.4.3 focus order unchanged). Buttons are created in
  // SCALE order, so index i maps to SCALE[i].value; the off-scale skip
  // button is handled by its own class, since it is not a SCALE point.
  // statementCardId is the ONE place a statement card's DOM id is built.
  // It was built in two places until 2026-10-13: buildStatementCard set
  // "utsagn-" + st.number while markWeight looked for "statement-" + st.id.
  // getElementById returned null, markWeight returned early, and the
  // importance control kept showing "Vanlig" after the visitor pressed
  // "Viktig" — a control that misreported the visitor's own input, on the
  // one axis the trade-off step later reads. Single-sourced so the two
  // cannot drift again; web/dom_id_guard_test.go fails if a second scheme
  // is ever reintroduced.
  function statementCardId(number) {
    return "utsagn-" + number;
  }

  function markAnswered(st, answeredValue) {
    var card = document.getElementById(statementCardId(st.number));
    if (!card) return;
    var scale = card.querySelector(".scale");
    if (!scale) return;
    var buttons = scale.querySelectorAll("button");
    SCALE.forEach(function (opt, i) {
      var btn = buttons[i];
      if (!btn) return;
      var isSelected = answeredValue === opt.value;
      btn.classList.toggle("selected", isSelected);
      btn.setAttribute("aria-pressed", isSelected ? "true" : "false");
    });
    var skipBtn = scale.querySelector("button.scale-skip");
    if (skipBtn) {
      var skipSelected = answeredValue === SKIP;
      skipBtn.classList.toggle("selected", skipSelected);
      skipBtn.setAttribute("aria-pressed", skipSelected ? "true" : "false");
    }
  }

  // progressStepClass maps a 0..1 fraction to one of style.css's discrete
  // 5%-step width classes (.pct-0 .. .pct-100) — the same ladder the bar
  // chart uses. Never an inline style="" or el.style write (CSP style-src
  // has no 'unsafe-inline').
  function progressStepClass(fraction) {
    var step = Math.round((fraction * 100) / 5) * 5;
    if (step < 0) step = 0;
    if (step > 100) step = 100;
    return "pct-" + step;
  }

  // updateProgress rewrites the small position indicator in place after an
  // answer, so it stays live without a re-render. In the one-question view
  // it shows the position ("Spørsmål 6 av 60") and a discreet line; in
  // utforskermodus it shows how many statements are answered, since there is
  // no single question in focus there. It never claims a completion time.
  function updateProgress() {
    var text = document.getElementById("quiz-progress-text");
    var fill = document.getElementById("quiz-progress-fill");
    var totalN = DATA.statements.length;
    if (viewMode === "utforsker") {
      var answered = countAnswered();
      if (text) text.textContent = "Besvart " + answered + " av " + totalN + " utsagn";
      if (fill) fill.className = "quiz-progress-fill " + progressStepClass(totalN ? answered / totalN : 0);
      return;
    }
    if (reviewing) {
      if (text) text.textContent = "Se over svarene dine";
      if (fill) fill.className = "quiz-progress-fill pct-100";
      return;
    }
    if (showingResult) {
      if (text) text.textContent = "Resultat";
      if (fill) fill.className = "quiz-progress-fill pct-100";
      return;
    }
    var pos = currentIndex + 1;
    if (text) text.textContent = "Spørsmål " + pos + " av " + totalN;
    if (fill) {
      var cls = progressStepClass(totalN ? pos / totalN : 0);
      // A sliver at the first question, so the line is never invisible while
      // the position is real.
      if (cls === "pct-0" && pos > 0) cls = "pct-5";
      fill.className = "quiz-progress-fill " + cls;
    }
  }

  // clearAllAnswers wipes the stored answers from this browser: the
  // in-memory object AND the localStorage entry. Answers are political
  // opinions — sensitive personal data — and this site's stated stance is
  // data minimisation, so deletion must be a real, complete action, not
  // just a visual reset.
  function clearAllAnswers() {
    answers = {};
    weights = {};
    savedAnswers = {};
    savedWeights = {};
    records = [];
    savedRecords = {};
    avvLater = false;
    noteState.dropped = null;
    noteState.resume = null;
    paintNote();
    // In ephemeral mode nothing was ever stored here, so there is nothing to
    // remove — and touching localStorage at all would break the promise that
    // ephemeral mode writes nothing.
    if (ephemeral) return;
    Store.clear(storage, [STORAGE_KEY, WEIGHT_KEY, AVV_KEY]);
  }

  // buildDeleteControl is the visible, accessible "delete all answers"
  // control. A single click never destroys anything: it opens an inline
  // confirmation, moves focus into it, and returns focus to the trigger on
  // cancel, so a keyboard user is never stranded and the action cannot be
  // hit by accident (WCAG 2.4.3 focus order, 2.4.7 focus visible). After a
  // confirmed delete the quiz is re-rendered empty and focus lands back on
  // the (freshly rendered) delete button.
  function buildDeleteControl() {
    var wrap = el("div", "privacy-controls");

    var heading = el("h2", "privacy-controls-heading", "Dine svar");
    wrap.appendChild(heading);

    var help = el("p", "lede privacy-help");
    help.id = "slett-svar-hjelp";
    help.textContent =
      "Svarene dine lagres bare i denne nettleseren og sendes aldri til en server. Vil du slette dem, forsvinner de helt herfra.";
    wrap.appendChild(help);

    var idle = el("div", "delete-idle");
    var deleteBtn = el("button", "delete-btn", "Slett alle svar fra denne nettleseren");
    deleteBtn.type = "button";
    deleteBtn.id = "slett-svar-knapp";
    deleteBtn.setAttribute("aria-describedby", "slett-svar-hjelp");
    idle.appendChild(deleteBtn);
    wrap.appendChild(idle);

    var confirm = el("div", "delete-confirm");
    confirm.hidden = true;
    var q = el("p", "lede delete-confirm-text", "Slette alle svarene dine permanent? Dette kan ikke angres.");
    q.id = "slett-svar-bekreft";
    confirm.appendChild(q);
    var yes = el("button", "delete-btn delete-confirm-yes", "Ja, slett alle svar");
    yes.type = "button";
    yes.setAttribute("aria-describedby", "slett-svar-bekreft");
    var no = el("button", "delete-cancel", "Avbryt");
    no.type = "button";
    confirm.appendChild(yes);
    confirm.appendChild(no);
    wrap.appendChild(confirm);

    deleteBtn.addEventListener("click", function () {
      idle.hidden = true;
      confirm.hidden = false;
      yes.focus();
      announce("Bekreft sletting av alle svar.");
    });

    no.addEventListener("click", function () {
      confirm.hidden = true;
      idle.hidden = false;
      deleteBtn.focus();
      announce("Sletting avbrutt.");
    });

    yes.addEventListener("click", function () {
      clearAllAnswers();
      renderQuiz();
      announce("Alle svar er slettet.");
      var freshBtn = document.getElementById("slett-svar-knapp");
      if (freshBtn) freshBtn.focus();
    });

    return wrap;
  }

  // The Resultat page shows real match data as soon as you've answered
  // anything at all, not only once every one of the 44 statements is
  // done — "totaliteten" (the full documented dataset) is already
  // fixed regardless of how far along you are; what grows as you
  // answer is just how much of it you've compared yourself against.
  // Every bar carries its own "N utsagn" sample-size caveat (see
  // renderBarChart), so 3 answers is never presented as equivalent to
  // 44 — it's just visible immediately, not gated behind finishing.
  function renderResultatPage() {
    var answeredCount = countAnswered();
    var skippedCount = countSkipped();
    var total = DATA.statements.length;
    // A skip is a real reply: show the results page as soon as there is
    // anything to show, including a run of "vet ikke" answers (which are
    // then reported as skipped, not silently dropped).
    root.appendChild(
      answeredCount + skippedCount > 0 ? renderResults(answeredCount, total) : renderResultsEmptyState()
    );
    root.appendChild(buildDeleteControl());
  }

  // ---- Quiz page (the / route) ----
  // The thematic heading a statement belongs to is deliberately NOT shown
  // above the question. This is a precaution against category priming, not
  // a proven requirement for the labels this site uses. What was actually
  // tested: Kamoen, van de Pol, Krouwel, de Vreese and Holleman, "Issue
  // framing in online voting advice applications" (PLOS ONE 14(2):
  // e0212555, 2019, N=27,404). The headers they manipulated were NOT neutral
  // topic labels — they were short labels chosen for their party
  // association (left-wing frames "nature and environment", "social policy";
  // right-wing frames "finances", "economy", "transportation"). The finding
  // was that ANY header shifted answers in the same, more left-wing
  // direction whether the header leaned left or right, and only for
  // respondents with low political sophistication. So the study does not
  // show that a neutral label such as "Skatt og fordeling" biases an answer.
  // Removing the heading is still the safe move, and the theme is not lost:
  // it lives in the neutral progress overview on the front page
  // (updateDimensionOverview, which fills the server-rendered dimension
  // rows), a different place on the screen, never a label on the question —
  // in any mode, including utforskermodus.

  var currentIndex = 0;      // which statement is in focus (one-question view)
  var reviewing = false;     // showing the "se over svarene dine" step
  var showingResult = false; // ephemeral inline result

  function clampIndex(i) {
    var max = DATA.statements.length - 1;
    if (max < 0) return 0;
    if (i < 0) return 0;
    if (i > max) return max;
    return i;
  }

  function answerLabel(value) {
    if (value === undefined) return "Ikke besvart";
    if (value === SKIP) return SKIP_LABEL;
    for (var i = 0; i < SCALE.length; i++) {
      if (SCALE[i].value === value) return SCALE[i].label;
    }
    return "Ikke besvart";
  }

  // buildStatementCard is the one rendering of a single statement, shared by
  // the one-question view, the utforskermodus list and a review jump. It
  // carries no dimension heading (see the note above). The statement text is
  // focusable (tabindex=-1) so advancing can move focus to it — the
  // assistive-tech-friendly wizard behaviour.
  function buildStatementCard(st) {
    var card = el("div", "card question-card");
    card.id = statementCardId(st.number);
    card.appendChild(el("p", "lede question-number", "Utsagn " + st.number));
    var text = el("p", "statement-text question-statement", st.text);
    text.tabIndex = -1;
    card.appendChild(text);

    // Collapsed by default so it doesn't add weight to the card, but it sits
    // right where you're about to answer — someone unfamiliar with a specific
    // policy area should be able to judge "sterkt enig" vs "sterkt uenig" from
    // the two extremes alone, without reading the whole evidence chain first
    // (2026-09-25 review). Only rendered when the statement actually has one.
    if (st.practical_stakes_enig || st.practical_stakes_uenig) {
      var stakes = el("details", "disclosure stakes-disclosure");
      stakes.appendChild(el("summary", null, "Hva betyr et svar i praksis?"));
      if (st.practical_stakes_enig) {
        var pEnig = el("p", null);
        pEnig.appendChild(el("strong", null, "Sterkt enig betyr: "));
        pEnig.appendChild(document.createTextNode(st.practical_stakes_enig));
        stakes.appendChild(pEnig);
      }
      if (st.practical_stakes_uenig) {
        var pUenig = el("p", null);
        pUenig.appendChild(el("strong", null, "Sterkt uenig betyr: "));
        pUenig.appendChild(document.createTextNode(st.practical_stakes_uenig));
        stakes.appendChild(pUenig);
      }
      card.appendChild(stakes);
    }

    var scale = el("div", "scale");
    SCALE.forEach(function (opt) {
      var btn = el("button", null, opt.label);
      btn.type = "button";
      var isSelected = answers[st.id] === opt.value;
      if (isSelected) btn.classList.add("selected");
      // aria-pressed makes the toggle-button semantics real, not just visual
      // (2026-09-26 review; the site's "never colour alone" rule, WCAG 1.4.1).
      btn.setAttribute("aria-pressed", isSelected ? "true" : "false");
      btn.addEventListener("click", function () {
        answers[st.id] = opt.value;
        saveAnswers(answers);
        // Update in place — NOT renderQuiz() — so the button the user just
        // activated keeps focus and its visible pressed state. Choosing an
        // answer never auto-advances: the visitor must be able to change
        // their mind, and it works better for assistive tech.
        markAnswered(st, opt.value);
        updateProgress();
        updateDimensionOverview();
        announce("Utsagn " + st.number + ": " + opt.label + ".");
      });
      scale.appendChild(btn);
    });

    // Sixth choice: the off-scale "vet ikke". Rendered inside the same .scale
    // row but with its own class (.scale-skip, see style.css) so it reads as
    // a separate "not a position" action, not a sixth point on the scale.
    var skipBtn = el("button", "scale-skip", SKIP_LABEL);
    skipBtn.type = "button";
    var skipSelected = answers[st.id] === SKIP;
    if (skipSelected) skipBtn.classList.add("selected");
    skipBtn.setAttribute("aria-pressed", skipSelected ? "true" : "false");
    skipBtn.addEventListener("click", function () {
      answers[st.id] = SKIP;
      saveAnswers(answers);
      // In place, not renderQuiz(): the skip button keeps focus too.
      markAnswered(st, SKIP);
      updateProgress();
      updateDimensionOverview();
      announce("Utsagn " + st.number + ": " + SKIP_LABEL + ".");
    });
    scale.appendChild(skipBtn);

    card.appendChild(scale);
    card.appendChild(buildWeightControl(st));

    var link = el("p", "lede");
    var a = document.createElement("a");
    a.href = "/utsagn/" + st.number;
    a.textContent = "Se kilden bak dette utsagnet\u00a0";
    appendArrow(a);
    link.appendChild(a);
    card.appendChild(link);

    return card;
  }

  // buildProgressHeader is the small position indicator: "Spørsmål 6 av 60"
  // plus a discreet line. Not a dashboard, and it never claims a completion
  // time that has not been measured.
  function buildProgressHeader() {
    var wrap = el("div", "quiz-progressbar");
    wrap.id = "quiz-progress";
    var text = el("p", "quiz-progress-text");
    text.id = "quiz-progress-text";
    wrap.appendChild(text);
    var track = el("div", "quiz-progress-track");
    var fill = el("div", "quiz-progress-fill pct-0");
    fill.id = "quiz-progress-fill";
    track.appendChild(fill);
    wrap.appendChild(track);
    return wrap;
  }

  // buildSettingsRow: two plain, visible choices. "Midlertidig økt" stores
  // nothing on the device; "Utforskermodus" restores the full list. The help
  // line explains each the first time, in plain language.
  function buildSettingsRow() {
    var wrap = el("div", "quiz-settings");
    var fs = document.createElement("fieldset");
    fs.className = "quiz-settings-fieldset";
    var legend = document.createElement("legend");
    legend.className = "quiz-settings-legend";
    legend.textContent = "Innstillinger";
    fs.appendChild(legend);

    var ephWrap = el("label", "quiz-settings-item");
    var eph = document.createElement("input");
    eph.type = "checkbox";
    eph.id = "quiz-ephemeral";
    eph.checked = ephemeral;
    ephWrap.appendChild(eph);
    ephWrap.appendChild(el("span", "quiz-settings-label", "Midlertidig økt"));
    fs.appendChild(ephWrap);
    eph.addEventListener("change", function () {
      ephemeral = eph.checked;
      paintNote();
      if (ephemeral) {
        announce("Midlertidig økt er slått på. Nye svar lagres ikke på denne enheten.");
      } else {
        announce("Midlertidig økt er slått av. Svarene lagres igjen i denne nettleseren.");
        saveAnswers(answers);
        saveWeights(weights);
      }
    });

    var expWrap = el("label", "quiz-settings-item");
    var exp = document.createElement("input");
    exp.type = "checkbox";
    exp.id = "quiz-explorer";
    exp.checked = viewMode === "utforsker";
    expWrap.appendChild(exp);
    expWrap.appendChild(el("span", "quiz-settings-label", "Utforskermodus"));
    fs.appendChild(expWrap);
    exp.addEventListener("change", function () {
      noteState.resume = null;
      paintNote();
      viewMode = exp.checked ? "utforsker" : "enkel";
      persistViewMode(viewMode);
      renderQuiz();
      announce(
        viewMode === "utforsker"
          ? "Utforskermodus er slått på. Du ser hele listen."
          : "Utforskermodus er slått av. Du ser ett spørsmål om gangen."
      );
    });

    wrap.appendChild(fs);
    wrap.appendChild(
      el(
        "p",
        "quiz-settings-help",
        "Midlertidig økt lagrer ingenting på enheten, nyttig hvis flere skal bruke den. Utforskermodus viser alle utsagnene i én liste."
      )
    );
    return wrap;
  }

  // renderNav rebuilds the Forrige / Neste controls for the current state. It
  // never re-renders the question card itself, so answering never moves focus.
  function renderNav() {
    var nav = document.getElementById("quiz-nav");
    if (!nav) return;
    nav.innerHTML = "";

    if (showingResult) {
      var back = el("button", "quiz-btn quiz-back", "Tilbake til spørsmålene");
      back.type = "button";
      back.addEventListener("click", function () {
        showingResult = false;
        showQuestion(currentIndex, true);
      });
      nav.appendChild(back);
      return;
    }

    if (reviewing) {
      var backReview = el("button", "quiz-btn quiz-back", "Tilbake til spørsmålene");
      backReview.type = "button";
      backReview.addEventListener("click", function () {
        showQuestion(currentIndex, true);
      });
      nav.appendChild(backReview);
      if (ephemeral) {
        var show = el("button", "quiz-btn quiz-next", "Se resultatet her");
        show.type = "button";
        show.addEventListener("click", function () {
          renderInlineResult();
        });
        nav.appendChild(show);
      } else {
        var toResult = document.createElement("a");
        toResult.className = "quiz-btn quiz-next quiz-to-result";
        toResult.href = "/resultat";
        toResult.textContent = "Se resultatet ditt";
        nav.appendChild(toResult);
      }
      return;
    }

    var prev = el("button", "quiz-btn quiz-prev", "Forrige");
    prev.type = "button";
    prev.disabled = currentIndex === 0;
    prev.addEventListener("click", function () {
      showQuestion(currentIndex - 1, true);
    });
    nav.appendChild(prev);

    var isLast = currentIndex === DATA.statements.length - 1;
    var next = el("button", "quiz-btn quiz-next", isLast ? "Se over svarene dine" : "Neste");
    next.type = "button";
    next.addEventListener("click", function () {
      if (isLast) {
        renderReview();
      } else {
        showQuestion(currentIndex + 1, true);
      }
    });
    nav.appendChild(next);

    if (!ephemeral) {
      var openResult = document.createElement("a");
      openResult.className = "quiz-to-result";
      openResult.href = "/resultat";
      openResult.textContent = "Se resultatet ditt\u00a0";
      appendArrow(openResult);
      nav.appendChild(openResult);
    }
  }

  function focusRegionHeading(region) {
    var target = region.querySelector(".question-statement") || region.querySelector("h2");
    if (target) target.focus();
  }

  // showQuestion renders exactly one statement in focus, updates the position
  // indicator and the nav. It is the one-question view's single step.
  function showQuestion(idx, focus) {
    if (focus && noteState.resume) {
      noteState.resume = null;
      paintNote();
    }
    reviewing = false;
    showingResult = false;
    currentIndex = clampIndex(idx);
    var region = document.getElementById("quiz-region");
    if (!region) return;
    region.innerHTML = "";
    var card = buildStatementCard(DATA.statements[currentIndex]);
    region.appendChild(card);
    region.appendChild(el("p", "lede key-hint", "Tips: tast 1–5 for å svare, S for «vet ikke»."));
    updateProgress();
    renderNav();
    if (focus) {
      focusRegionHeading(region);
      announce("Spørsmål " + (currentIndex + 1) + " av " + DATA.statements.length + ".");
    }
  }

  // renderReview is the short, skippable "se over svarene dine" step before
  // the result: one line per statement with its current answer and an Endre
  // action, so a misread answer can be fixed. Nothing is gated behind it.
  function renderReview() {
    showingResult = false;
    reviewing = true;
    var region = document.getElementById("quiz-region");
    if (!region) return;
    region.innerHTML = "";
    var wrap = el("div", "card review-card");
    var heading = el("h2", null, "Se over svarene dine");
    heading.tabIndex = -1;
    wrap.appendChild(heading);
    wrap.appendChild(
      el(
        "p",
        "lede",
        "Her ser du hva du har svart. Du kan endre et svar hvis du oppdaget noe, eller gå videre til resultatet. Du trenger ikke svare på alle."
      )
    );
    var list = el("ul", "review-list");
    DATA.statements.forEach(function (st, i) {
      var li = el("li", "review-item");
      li.appendChild(el("span", "review-stmt", "Utsagn " + st.number));
      var answer = el("span", "review-answer", answerLabel(answers[st.id]));
      if (answers[st.id] === undefined) answer.classList.add("review-answer-empty");
      li.appendChild(answer);
      var edit = el("button", "review-edit", "Endre");
      edit.type = "button";
      edit.addEventListener("click", function () {
        showQuestion(i, true);
      });
      li.appendChild(edit);
      list.appendChild(li);
    });
    wrap.appendChild(list);
    region.appendChild(wrap);
    updateProgress();
    renderNav();
    heading.focus();
    announce("Se over svarene dine.");
  }

  // renderInlineResult is the result shown inside the quiz page for an
  // ephemeral session, where the separate /resultat page would find nothing
  // stored to read. It reuses the same renderResults() the /resultat page
  // uses.
  function renderInlineResult() {
    showingResult = true;
    reviewing = false;
    var region = document.getElementById("quiz-region");
    if (!region) return;
    region.innerHTML = "";
    region.appendChild(renderResults(countAnswered(), DATA.statements.length));
    region.appendChild(
      el(
        "p",
        "lede quiz-ephemeral-note",
        "Du er i midlertidig økt. Resultatet vises bare her og lagres ikke på enheten."
      )
    );
    updateProgress();
    renderNav();
    var heading = region.querySelector("h2");
    if (heading) {
      heading.tabIndex = -1;
      heading.focus();
    }
    announce("Resultatet vises. Ingenting er lagret.");
  }

  function renderSimplePage() {
    root.appendChild(buildProgressHeader());
    root.appendChild(buildSettingsRow());
    var region = el("div", "quiz-region");
    region.id = "quiz-region";
    root.appendChild(region);
    var nav = el("div", "quiz-nav");
    nav.id = "quiz-nav";
    root.appendChild(nav);
    root.appendChild(buildDeleteControl());
    showQuestion(clampIndex(currentIndex), false);
  }

  // renderExplorerPage is the full list, restored for someone who wants to
  // scan everything. The statements stand alone here too: no thematic heading
  // is placed over any question, in any mode.
  function renderExplorerPage() {
    root.appendChild(buildProgressHeader());
    root.appendChild(buildSettingsRow());
    var list = el("div", "quiz-explorer-list");
    list.id = "quiz-explorer-list";
    DATA.statements.forEach(function (st) {
      list.appendChild(buildStatementCard(st));
    });
    root.appendChild(list);
    root.appendChild(buildDeleteControl());
    updateProgress();
  }

  function renderQuizPage() {
    reviewing = false;
    showingResult = false;
    if (viewMode === "utforsker") {
      renderExplorerPage();
    } else {
      renderSimplePage();
    }
  }

  // If the URL has a #utsagn-N fragment (set by statement.html.tmpl's
  // "← Til oversikten" link), the browser's native fragment-scroll runs
  // before this script ever inserts the card into the DOM, so it always
  // misses — this is the exact "pressing back doesn't return you to where
  // you were" bug reported 2026-09-24. So the hash is resolved here instead.
  // With the one-question view there is only one card in the DOM, so a
  // #utsagn-N fragment makes that statement the question in focus; in
  // utforskermodus the card exists in the list and is scrolled to, with a
  // brief outline so the jump is visible. A #dim-<slug> fragment (the front
  // page's "Kjerneverdiene" dimension links) resolves to the first statement
  // of that dimension — the theme itself is never shown as a heading here.
  function statementIndexByNumber(n) {
    for (var i = 0; i < DATA.statements.length; i++) {
      if (DATA.statements[i].number === n) return i;
    }
    return -1;
  }

  function firstIndexOfDimension(slug) {
    for (var i = 0; i < DATA.statements.length; i++) {
      if (slugify(DATA.statements[i].dimension) === slug) return i;
    }
    return -1;
  }

  function scrollToCard(number) {
    var target = document.getElementById(statementCardId(number));
    if (!target) return;
    target.scrollIntoView({ block: "center" });
    target.classList.add("card-returned");
    setTimeout(function () {
      target.classList.remove("card-returned");
    }, 1800);
  }

  // resolveIndexForHash returns the statement index a fragment points at, or
  // -1 when it points at nothing here. Used on load and on hashchange (the
  // front page's dimension links are same-page links, so they fire
  // hashchange without a reload).
  function resolveIndexForHash(hash) {
    if (!hash) return -1;
    if (hash.indexOf("#utsagn-") === 0) {
      return statementIndexByNumber(parseInt(hash.slice("#utsagn-".length), 10));
    }
    if (hash.indexOf("#dim-") === 0) {
      return firstIndexOfDimension(hash.slice("#dim-".length));
    }
    return -1;
  }

  function handleHash() {
    var idx = resolveIndexForHash(window.location.hash);
    if (idx < 0) return;
    if (viewMode === "utforsker") {
      scrollToCard(DATA.statements[idx].number);
    } else {
      showQuestion(idx, true);
    }
  }


  // pct is defined in scoring.js (loaded before this file) and aliased at
  // the top: it returns null for an unanswered statement, a skipped
  // ("vet ikke") statement, or a party with no coded position, so none of
  // those ever becomes NaN% or a fabricated 0% on /resultat.

  // Rounds a 0-100 match percentage to the nearest 5 and returns the
  // matching CSS step class (style.css's .pct-N ladder). Bars are never
  // drawn with an inline style="" or el.style write: the CSP here has no
  // 'unsafe-inline' for style-src (see internal/server/middleware.go),
  // so that would be silently dropped by the browser, not just avoided
  // on principle.
  function pctStepClass(matchPct) {
    var step = Math.round(matchPct / 5) * 5;
    if (step < 0) step = 0;
    if (step > 100) step = 100;
    return "pct-" + step;
  }

  // computeAggregate averages match% per party across statementList,
  // counting only the (statement, party) pairs where that party
  // actually has a coded position — exactly the same "documented vs.
  // undocumented" split the per-statement table already uses, so a
  // party with fewer coded positions is never silently padded with a
  // guessed number. Returns rows sorted best-match first; a party with
  // zero documented statements in this list gets avg: null and sorts
  // last, rendered as "Ikke vurdert" rather than 0%.
  // useWeights lets a caller ask "what would this be under other weights" (the
  // "uten avveiningene" line) without swapping the module-level map, which
  // would break the moment anything ran in between. Omitted, it is the
  // visitor's own weights.
  function computeAggregate(statementList, useWeights) {
    var W = useWeights || weights;
    var sums = {}, counts = {};
    DATA.parties.forEach(function (party) {
      sums[party.id] = 0;
      counts[party.id] = 0;
    });
    var den = {};
    DATA.parties.forEach(function (party) { den[party.id] = 0; });
    statementList.forEach(function (st) {
      var cells = DATA.cells[st.id] || {};
      var userAnswer = answers[st.id];
      if (userAnswer === undefined) return;
      // The visitor's importance for this statement scales its pull on the
      // average and nothing else. pct() never sees it, so a weighted
      // result can never be read as a different answer. Unweighted, w is 1
      // and this is the plain mean exactly as before.
      var w = weightOf(W, st.id);
      DATA.parties.forEach(function (party) {
        var cell = cells[party.id] || {};
        var m = pct(userAnswer, cell.party_position);
        if (m === null) return;
        sums[party.id] += m * w;
        den[party.id] += w;
        counts[party.id] += 1;
      });
    });
    var rows = DATA.parties.map(function (party) {
      var n = counts[party.id];
      return {
        party: party,
        avg: den[party.id] > 0 ? Math.round(sums[party.id] / den[party.id]) : null,
        n: n,
        totalWeight: den[party.id],
      };
    });
    rows.sort(function (a, b) {
      if (a.avg === null && b.avg === null) return 0;
      if (a.avg === null) return 1;
      if (b.avg === null) return -1;
      return b.avg - a.avg;
    });
    return rows;
  }

  // renderBarChart draws one horizontal bar per row: party, match%
  // (or "Ikke vurdert"), a bar sized via a discrete CSS step class, and
  // how many of the statements in this scope that average actually
  // covers — the sample-size caveat stays attached to every single bar,
  // not just a footnote above the chart.
  function renderBarChart(rows) {
    var wrap = el("div", "bar-chart");
    rows.forEach(function (r) {
      var row = el("div", "bar-row");
      var label = el("div", "bar-label");
      label.appendChild(el("span", "bar-party", r.party.short));
      if (r.avg !== null) {
        label.appendChild(el("span", "bar-pct", r.avg + "% samsvar"));
      } else {
        label.appendChild(el("span", "bar-pct bar-pct-empty", "Ikke vurdert"));
      }
      row.appendChild(label);

      var track = el("div", "bar-track");
      if (r.avg !== null) {
        track.appendChild(el("div", "bar-fill " + pctStepClass(r.avg)));
      }
      row.appendChild(track);

      row.appendChild(el("div", "bar-n", r.n + " utsagn"));
      wrap.appendChild(row);
    });
    return wrap;
  }

  // renderSummary is the results screen's headline: an overall "who are
  // you most aligned with" bar chart across every statement you've
  // both answered AND that has at least one documented party position,
  // plus a per-dimension drill-down so the same ranking can be checked
  // one core value at a time ("hvor man blir mest forankret i ett parti
  // vs et annet basert på disse kjerneverdier", requested 2026-09-24).
  // Placed first inside the results card, which renders progressively
  // as soon as you've answered anything at all (see renderQuiz) — so
  // the "documented" set here must be intersected with `answers`,
  // never the dataset-wide documented count, or the sentence below
  // would claim a sample size the aggregate never actually used.
  // buildWeightControl is the "hvor viktig" row on a card. A separate axis
  // from the answer scale on purpose: it never changes what the visitor
  // answered, only how far that answer reaches in the result. aria-pressed
  // makes the toggle semantics real rather than visual, the same rule the
  // answer buttons follow. "Vanlig" is the default and is stored as an
  // absence, so an untouched quiz weights nothing.
  function buildWeightControl(st) {
    var wrap = el("div", "weight");
    wrap.appendChild(el("span", "weight-label", "Hvor viktig er dette for deg?"));
    var opts = el("div", "weight-options");
    var current = weightOf(weights, st.id);
    WEIGHT_LEVELS.forEach(function (lvl) {
      var btn = el("button", "weight-btn", lvl.label);
      btn.type = "button";
      var on = current === lvl.value;
      if (on) btn.classList.add("selected");
      btn.setAttribute("aria-pressed", on ? "true" : "false");
      btn.addEventListener("click", function () {
        // Ask this BEFORE saving. saveWeights re-checks the records against the
        // new weights (settleSave), which drops the record from memory the
        // moment this weight stops matching it. Asked afterwards, the answer is
        // always "no": the stored record survives, the card note stays and the
        // removal is never announced.
        var ended = AV.inRecord(records, st.id);
        if (lvl.value === 1) {
          delete weights[st.id];
        } else {
          weights[st.id] = lvl.value;
        }
        saveWeights(weights);
        markWeight(st, lvl.value);
        updateProgress();
        // The visitor's own hand-set weight wins over a trade-off that set
        // it for them: the record would otherwise explain weights that no
        // longer exist.
        if (ended) {
          records = AV.dropOnManualChange(records, st.id);
          saveRecords();
          removeWeightNote(ended.a);
          removeWeightNote(ended.b);
        }
        announce(
          "Utsagn " + st.number + ": vekt satt til " + lvl.label + "." +
            (ended ? " Avveiningen med dette utsagnet er fjernet." : "")
        );
      });
      opts.appendChild(btn);
    });
    wrap.appendChild(opts);
    // When a trade-off set this weight, say so on the card, or the control
    // would claim "Svært viktig" with no hint that the result page chose it.
    // A "like mye" decision changed no weight, so there is nothing to say.
    var rec = AV.inRecord(records, st.id);
    if (rec && rec.valg !== "lik") {
      var other = stById(rec.a === st.id ? rec.b : rec.a);
      var note = el(
        "p",
        "weight-note",
        "Satt i en avveining mot utsagn " + (other ? other.number : "") +
          " på resultatsiden. Endrer du vekten her, fjernes avveiningen."
      );
      note.id = "weight-note-" + st.number;
      wrap.appendChild(note);
    }
    return wrap;
  }

  function stById(id) {
    for (var i = 0; i < DATA.statements.length; i++) {
      if (DATA.statements[i].id === id) return DATA.statements[i];
    }
    return null;
  }

  // removeWeightNote drops the "satt i en avveining" line from a card that is
  // on screen (explorer mode shows every card at once).
  function removeWeightNote(id) {
    var st = stById(id);
    var n = st && document.getElementById("weight-note-" + st.number);
    if (n && n.parentNode) n.parentNode.removeChild(n);
  }

  // markWeight updates the control in place, like markAnswered does for the
  // scale, so the button the visitor pressed keeps focus.
  function markWeight(st, value) {
    var card = document.getElementById(statementCardId(st.number));
    if (!card) return;
    var btns = card.querySelectorAll(".weight-btn");
    for (var i = 0; i < btns.length; i++) {
      var want = WEIGHT_LEVELS[i] && WEIGHT_LEVELS[i].value === value;
      btns[i].classList.toggle("selected", !!want);
      btns[i].setAttribute("aria-pressed", want ? "true" : "false");
    }
  }

  // --- the trade-off step (avveining) -----------------------------------
  // Design and rationale: docs/design/2026-10-13-avveining.md. The logic is
  // pure and lives in avveining.js; this is only the rendering and the wiring.

  function stLabel(st) {
    return "Utsagn " + st.number + (st.short_name ? " (" + st.short_name + ")" : "");
  }

  // rerenderAfterAvveining re-draws the result and tells the explanation panel
  // (analyse.js, a separate script) which weights are now in force. The event
  // CARRIES them: re-reading storage would show the panel stale weights
  // whenever a save had failed, which is the "two top parties on one page"
  // bug (#117) coming back through the side door.
  //
  // ids are the two statements a decision or undo is about. Weights are
  // merged key by key and only keys THIS tab changed are written, so a
  // decision that sets a weight to the value this tab already held would be
  // invisible to the merge, and two stale tabs deciding the same pair would
  // end with the record of one and the weights of neither. Marking both keys
  // as changed makes the last decision win cleanly, weights and record alike.
  function rerenderAfterAvveining(message, changed, ids) {
    if (changed) {
      (ids || []).forEach(function (id) { savedWeights[id] = "touched"; });
      saveWeights(weights);
      saveRecords();
    }
    if (PAGE === "resultat") renderQuiz(); else renderInlineResult();
    if (changed) {
      document.dispatchEvent(
        new CustomEvent("ektevalgomat:vekt-endret", { detail: { weights: weights, answers: answers } })
      );
    }
    var h = document.getElementById("avveining-neste") || document.getElementById("avveining-heading");
    if (h) h.focus();
    announce(message);
  }

  function partyName(r) { return r.party.name; }

  // effectLine says what the trade-offs did to the top of the result. It is
  // always shown next to the log, so a decision can never hide its effect,
  // and a tie is named as a tie: payload order never picks a winner.
  function effectLine(before, now) {
    var L0 = leaders(before), L1 = leaders(now);
    if (!L0.length || !L1.length) return null;
    var n0 = L0.map(partyName), n1 = L1.map(partyName), a0 = L0[0].avg, a1 = L1[0].avg;
    if (L1.length > 1) {
      return "Med avveiningene ligger " + AV.namesList(n1) + " likt øverst, med " + a1 + " %. Uten dem lå " +
        (L0.length > 1 ? AV.namesList(n0) + " likt" : n0[0]) + " øverst (" + a0 + " %).";
    }
    if (L0.length === 1 && L0[0].party.id === L1[0].party.id) {
      return "Øverst ligger fortsatt " + n1[0] + ", " + a0 + " % uten avveiningene og " + a1 + " % med dem.";
    }
    if (L0.length > 1) {
      return "Uten avveiningene lå " + AV.namesList(n0) + " likt øverst (" + a0 + " %). Med dem ligger " +
        n1[0] + " øverst (" + a1 + " %).";
    }
    return "Uten avveiningene ville du vært mest på linje med " + n0[0] + " (" + a0 + " %). Med dem er det " +
      n1[0] + " (" + a1 + " %).";
  }

  // movedLine lists how every party's number changed. Each part is one unit
  // that never breaks across a line, and it uses words rather than an arrow
  // so a screen reader reads the same thing a sighted reader sees.
  function movedLine(before, now) {
    var prev = {};
    before.forEach(function (r) { prev[r.party.id] = r.avg; });
    var moved = [], same = [];
    now.forEach(function (r) {
      if (r.avg === null) return;
      if (prev[r.party.id] === r.avg) same.push(r.party.short);
      else moved.push(r.party.short + " fra " + prev[r.party.id] + " til " + r.avg + " %");
    });
    var p = el("p", "avveining-flytt");
    if (!moved.length) {
      p.textContent = "Ingen partier fikk et annet tall.";
      return p;
    }
    p.appendChild(document.createTextNode("Slik flyttet tallene seg: "));
    moved.forEach(function (m, i) {
      p.appendChild(el("span", "avveining-flytt-del", m));
      p.appendChild(document.createTextNode(i < moved.length - 1 ? ", " : "."));
    });
    if (same.length) p.appendChild(document.createTextNode(" Uendret: " + same.join(", ") + "."));
    return p;
  }

  function buildAvveiningLog(box, documented) {
    var h = el("h3", "avveining-title", "Avveiningene dine");
    h.id = "avveining-heading";
    h.tabIndex = -1;
    box.appendChild(h);
    var ul = el("ul", "avveining-logg");
    records.forEach(function (r, i) {
      var a = stById(r.a), b = stById(r.b);
      var li = el("li", "avveining-logg-rad");
      var txt;
      if (r.valg === "lik") {
        txt = "Utsagn " + a.number + " og utsagn " + b.number + " veier like mye. Vektene er ikke endret.";
      } else {
        var win = r.valg === "a" ? a : b, lose = r.valg === "a" ? b : a;
        txt = stLabel(win) + " veier tyngst, foran " + stLabel(lose).replace("Utsagn", "utsagn") +
          ". Teller " + AV.WIN + " mot " + AV.LOSE + ".";
      }
      li.appendChild(el("span", "avveining-logg-tekst", txt));
      if (!isPosition(answers[r.a]) || !isPosition(answers[r.b])) {
        li.appendChild(el("span", "avveining-logg-merk",
          "Du har senere svart «" + SKIP_LABEL + "» på et av dem, så det teller ikke i samsvaret."));
      }
      var undoBtn = el("button", "avveining-angre", "Angre");
      undoBtn.type = "button";
      undoBtn.setAttribute("aria-label", "Angre avveiningen mellom utsagn " + a.number + " og " + b.number);
      undoBtn.addEventListener("click", function () {
        var u = AV.undo(weights, records, i);
        weights = u.weights;
        records = u.records;
        removeWeightNote(r.a);
        removeWeightNote(r.b);
        rerenderAfterAvveining("Avveiningen er angret. Vektene er tilbake til det du satte selv.", true, [r.a, r.b]);
      });
      li.appendChild(undoBtn);
      ul.appendChild(li);
    });
    box.appendChild(ul);
    var before = computeAggregate(documented, AV.weightsWithout(weights, records));
    var now = computeAggregate(documented);
    var effect = effectLine(before, now);
    if (effect) box.appendChild(el("p", "avveining-effekt", effect));
    box.appendChild(movedLine(before, now));
  }

  // noConflictLine is the stated absence. Silence would read as "the feature
  // did not run"; this says what was checked and, for pairs it could not
  // judge, that it could not.
  function noConflictLine(p) {
    var none = p.judgedPairs > 0
      ? "De " + p.marked + " viktige utsagnene dine trekker ikke hver sin vei, så det er ingenting å avveie."
      : "Vi kan ikke si om de viktige utsagnene dine trekker hver sin vei. For få partier har en posisjon med kilde på begge.";
    if (p.judgedPairs > 0 && p.unknownPairs > 0) {
      none += " For " + p.unknownPairs + (p.unknownPairs === 1 ? " av parene" : " av parene") +
        " har for få partier en posisjon med kilde på begge til at vi kan si det.";
    }
    return el("p", "lede avveining-ingen", none);
  }

  function buildAvveiningStep(box, pair, total) {
    var step = el("div", "avveining-steg");
    if (total > 1) step.appendChild(el("p", "avveining-teller", "Avveining " + (records.length + 1) + " av " + total));
    var q = el("h3", "avveining-title", "To viktige utsagn trekker hver sin vei");
    q.id = records.length === 0 ? "avveining-heading" : "avveining-neste";
    q.tabIndex = -1;
    step.appendChild(q);
    var both = weightOf(weights, pair.a.id) === MAX_WEIGHT && weightOf(weights, pair.b.id) === MAX_WEIGHT;
    step.appendChild(el("p", "lede",
      "Du har merket begge som " + (both ? "svært viktige" : "viktige") + ". Partiene som står nærmest deg på det ene, står langt fra deg på det andre. " +
      "Da deler snittet forskjellen mellom dem. Du kan velge hvilket som skal veie tyngst."));
    step.appendChild(el("p", "avveining-spm", "Hvilket av disse veier tyngst for deg?"));

    // The two options are identical in style and always in statement-number
    // order, never ordered by which one helps the current top party. The
    // choices carry no party names: the visitor decides on the statements.
    var opts = el("div", "avveining-valg");
    [["a", pair.a], ["b", pair.b]].forEach(function (pr) {
      var st = pr[1];
      var o = el("div", "avveining-alt");
      o.appendChild(el("p", "avveining-alt-nr", "Utsagn " + st.number + " · Du svarte: " + answerLabel(answers[st.id])));
      var t = el("p", "avveining-alt-tekst", st.text);
      t.id = "avv-tekst-" + st.number;
      o.appendChild(t);
      var b = el("button", "avveining-velg", "Utsagn " + st.number + " veier tyngst");
      b.type = "button";
      b.setAttribute("aria-describedby", t.id);
      b.addEventListener("click", function () {
        var d = AV.decide(weights, records, pair.a.id, pair.b.id, pr[0], Date.now());
        weights = d.weights;
        records = d.records;
        rerenderAfterAvveining("Utsagn " + st.number + " veier nå tyngst. Resultatet er regnet ut på nytt.", true, [pair.a.id, pair.b.id]);
      });
      o.appendChild(b);
      opts.appendChild(o);
    });
    step.appendChild(opts);
    step.appendChild(el("p", "avveining-effektlofte",
      "Det du velger, teller " + AV.WIN + " i snittet, det andre " + AV.LOSE + ". Svarene dine endres ikke, og du kan angre."));

    var row = el("div", "avveining-rad");
    var lik = el("button", "avveining-sekundaer", "De veier like mye");
    lik.type = "button";
    lik.addEventListener("click", function () {
      var d = AV.decide(weights, records, pair.a.id, pair.b.id, "lik", Date.now());
      weights = d.weights;
      records = d.records;
      rerenderAfterAvveining("Begge veier like mye. Vektene er ikke endret.", true, [pair.a.id, pair.b.id]);
    });
    row.appendChild(lik);
    var later = el("button", "avveining-sekundaer", "Ikke nå");
    later.type = "button";
    later.addEventListener("click", function () {
      avvLater = true; // this page view only; a stored deferral would be a decision never made
      rerenderAfterAvveining("Avveiningen er lagt til side.", false);
    });
    row.appendChild(later);
    step.appendChild(row);

    // Why we ask: the evidence, collapsed. Supporting detail, not the decision.
    var why = el("details", "disclosure avveining-hvorfor");
    why.appendChild(el("summary", null, "Hvorfor spør vi om dette?"));
    why.appendChild(el("p", null,
      pair.judged.shared + " av " + DATA.parties.length + " partier har en posisjon med kilde på begge utsagnene. " +
      "Samsvaret ditt med dem på det ene utsagnet går motsatt vei av samsvaret på det andre."));
    var tbl = document.createElement("table");
    tbl.className = "party-overview avveining-bevis";
    var hr = document.createElement("tr");
    ["Parti", "Utsagn " + pair.a.number, "Utsagn " + pair.b.number].forEach(function (x) { hr.appendChild(el("th", null, x)); });
    tbl.appendChild(hr);
    pair.judged.evidence.forEach(function (e) {
      var tr = document.createElement("tr");
      tr.appendChild(el("td", "po-party", e.party.short));
      tr.appendChild(el("td", null, e.a + " % samsvar"));
      tr.appendChild(el("td", null, e.b + " % samsvar"));
      tbl.appendChild(tr);
    });
    why.appendChild(tbl);
    step.appendChild(why);
    box.appendChild(step);
  }

  function buildAvveining(documented) {
    var p = AV.plan(DATA, answers, weights, records);
    // Fewer than two marked and nothing decided: no step, and no line
    // advertising one, so nobody marks importance just to unlock it.
    if (p.marked < 2 && records.length === 0) return document.createDocumentFragment();

    if (p.pending.length === 0 && records.length === 0) return noConflictLine(p);

    var box = el("section", "avveining");
    box.id = "avveining";
    box.setAttribute("aria-labelledby", "avveining-heading");
    if (records.length > 0) buildAvveiningLog(box, documented);
    if (p.pending.length === 0) return box;

    if (avvLater) {
      var later = el("p", "avveining-utsatt");
      later.appendChild(document.createTextNode(
        p.pending.length === 1
          ? "Du har en avveining du ikke har tatt stilling til. "
          : "Du har " + p.pending.length + " avveininger du ikke har tatt stilling til. "
      ));
      var resume = el("button", "avveining-sekundaer", "Ta stilling nå");
      resume.type = "button";
      resume.addEventListener("click", function () {
        avvLater = false;
        rerenderAfterAvveining("Avveining vises.", false);
      });
      later.appendChild(resume);
      box.appendChild(later);
      return box;
    }
    buildAvveiningStep(box, p.pending[0], p.pending.length + records.length);
    return box;
  }

  function renderSummary() {
    var documented = DATA.statements.filter(function (st) {
      // Only a position on the scale enters the match. A skipped
      // ("vet ikke") statement is NOT counted here, even when the party
      // has a documented position: it is neither agreement nor
      // disagreement, so it must not inflate the documented sample size
      // the headline sentence quotes.
      if (!isPosition(answers[st.id])) return false;
      var cells = DATA.cells[st.id] || {};
      return DATA.parties.some(function (party) {
        var cell = cells[party.id] || {};
        return cell.party_position !== null && cell.party_position !== undefined;
      });
    });

    var wrap = el("div", "summary-card");
    // The headline names a ranking, so it only appears when there IS one.
    // Above a thin or empty sample it would ask a question the page then
    // refuses to answer (2026-10-23 review).
    if (documented.length === 0) {
      wrap.appendChild(el("h2", null, "Resultatet ditt"));
      wrap.appendChild(el("p", "lede", "Ingen av utsagnene du har svart på med en posisjon, har en partiposisjon med kilde ennå. Det finnes derfor ingenting å sammenligne svaret ditt mot."));
      return wrap;
    }

    var overall = computeAggregate(documented);
    // A ranking built on one or two answered statements is not a ranking.
    // It is a single data point dressed up as a percentage. Until the sample
    // is thick enough to mean anything, say so instead of printing a
    // headline "100% samsvar" next to a party the visitor has matched on
    // exactly one question. The per-statement tables below still show the
    // per-question match, which is honest at any N.
    var MIN_RANKING_N = 10;
    if (documented.length < MIN_RANKING_N) {
      wrap.appendChild(el("h2", null, "Resultatet ditt så langt"));
      wrap.appendChild(
        el(
          "p",
          "summary-lede",
          "Du har foreløpig svart på " + documented.length + " utsagn med en partiposisjon med kilde. " +
            "Det er for tynt grunnlag til å rangere partiene mot hverandre. " +
            "Svar på minst " + MIN_RANKING_N + " for å se en rangering. " +
            "Utsagnene under viser likevel samsvaret ditt per utsagn."
        )
      );
      return wrap;
    }
    wrap.appendChild(el("h2", null, "Hvem er du mest på linje med?"));
    // A tie at the top is a tie. The headline used to name overall[0], so the
    // order of the data picked a winner among parties with the same number.
    var lead = leaders(overall);
    if (lead.length > 0) {
      var tied = lead.length > 1;
      wrap.appendChild(
        el(
          "p",
          "summary-lede",
          "Du har svart på " + countAnswered() + " utsagn. " +
            documented.length + " av dem har en partiposisjon med kilde. Ut fra disse er du foreløpig " +
            (tied ? "like mye på linje med " : "mest på linje med ") +
            AV.namesList(lead.map(partyName)) + " (" + lead[0].avg + "% samsvar" + (tied ? " hver" : "") + ")."
        )
      );
    }

    wrap.appendChild(
      el(
        "p",
        "lede",
        "Dette er ikke «hele sannheten» om et parti. Det er snittet av samsvaret ditt på utsagnene som har en posisjon med kilde så langt. " +
          "Datagrunnlaget er fortsatt lite og bygges ut jevnlig. Tallet til høyre for hver bar viser hvor mange utsagn snittet " +
          "bygger på for det partiet."
      )
    );
    // Say the weighting out loud. An unexplained weighting would make the
    // percentage unauditable, and silence would let a visitor assume their
    // result is a plain mean when it is not.
    var weighted = countWeighted();
    if (weighted > 0) {
      wrap.appendChild(
        el(
          "p",
          "lede",
          "Du har merket " + weighted +
            " utsagn som viktigere enn de andre. De teller mer i snittet under. " +
            "Vektingen endrer ikke svarene dine, bare hvor mye hvert svar trekker."
        )
      );
    }

    wrap.appendChild(buildAvveining(documented));
    wrap.appendChild(renderBarChart(overall));

    var dims = [];
    var seenDim = {};
    documented.forEach(function (st) {
      if (!seenDim[st.dimension]) {
        seenDim[st.dimension] = true;
        dims.push(st.dimension);
      }
    });

    if (dims.length > 1) {
      var dimWrap = el("div", "dimension-breakdown");
      dimWrap.appendChild(el("h3", null, "Samsvar per kjerneverdi"));
      dims.forEach(function (dim) {
        var dimStatements = documented.filter(function (st) {
          return st.dimension === dim;
        });
        var details = document.createElement("details");
        details.className = "disclosure dimension-disclosure";
        var summaryEl = document.createElement("summary");
        summaryEl.textContent = dim + " (" + dimStatements.length + " utsagn)";
        details.appendChild(summaryEl);
        details.appendChild(renderBarChart(computeAggregate(dimStatements)));
        dimWrap.appendChild(details);
      });
      wrap.appendChild(dimWrap);
    }

    return wrap;
  }

  var SVG_NS = "http://www.w3.org/2000/svg";
  function svgEl(tag, attrs) {
    var e = document.createElementNS(SVG_NS, tag);
    for (var k in attrs) e.setAttribute(k, attrs[k]);
    return e;
  }

  // A small doodled companion for the empty Resultat tab — not a
  // mascot with a name, just something friendlier than blank space
  // while you haven't answered anything yet (2026-09-25, playful ask:
  // "kan vi drodle mystiske vesener"). Colours come from style.css's
  // own custom properties via CSS classes, never a style="" attribute
  // (the CSP here has no 'unsafe-inline' for style-src) — SVG
  // presentation attributes like fill/stroke set via .setAttribute()
  // are unaffected either way, but the class hook keeps the palette in
  // one place if this creature grows siblings later.
  function mysteryCreature() {
    var svg = svgEl("svg", {
      viewBox: "0 0 120 96",
      class: "mystery-creature",
      "aria-hidden": "true",
      focusable: "false",
    });
    svg.appendChild(
      svgEl("path", {
        class: "mystery-creature-body",
        d:
          "M60 14 C30 14 18 34 18 54 L18 70 " +
          "Q22 82 28 70 Q34 82 40 70 Q46 82 52 70 Q58 82 64 70 " +
          "Q70 82 76 70 Q82 82 88 70 Q94 82 100 70 " +
          "L102 54 C102 34 90 14 60 14 Z",
      })
    );
    svg.appendChild(svgEl("circle", { class: "mystery-creature-eye", cx: 44, cy: 46, r: 8 }));
    svg.appendChild(svgEl("circle", { class: "mystery-creature-eye", cx: 76, cy: 46, r: 8 }));
    svg.appendChild(svgEl("circle", { class: "mystery-creature-pupil", cx: 46, cy: 47, r: 3 }));
    svg.appendChild(svgEl("circle", { class: "mystery-creature-pupil", cx: 78, cy: 47, r: 3 }));
    svg.appendChild(
      svgEl("path", {
        class: "mystery-creature-spots",
        d: "M30 62 l3 3 M90 60 l3 3 M60 66 l0 4",
      })
    );
    return svg;
  }

  function renderResultsEmptyState() {
    var wrap = el("div", "card results empty-state results-empty");
    wrap.appendChild(mysteryCreature());
    wrap.appendChild(el("p", "empty-state-count", "Ingen svar registrert ennå"));
    wrap.appendChild(
      el(
        "p",
        null,
        "Svar på minst ett utsagn i valgomaten for å se resultatet ditt her. Det oppdateres etter hvert svar, du trenger ikke svare på alle."
      )
    );
    return wrap;
  }

  function renderResults(answeredCount, total) {
    var wrap = el("div", "card results");
    var complete = answeredCount === total;
    wrap.appendChild(
      el(
        "h2",
        null,
        complete
          ? "Resultat"
          : "Resultat så langt (" + answeredCount + " av " + total + " utsagn besvart)"
      )
    );
    wrap.appendChild(
      el(
        "p",
        "lede",
        (complete
          ? ""
          : "Dette oppdateres etter hvert svar du gir. Du trenger ikke svare på alle for å se hvor du står. ") +
          "Samsvar vises bare der partiet har en posisjon med kilde. " +
          "«Ikke vurdert» betyr at ingen av kildene vi har hentet inn, viser partiets standpunkt på dette utsagnet. Det betyr ikke at partiet er nøytralt."
      )
    );

    // The evidence-strength label appears in a table column below but
    // used to have no explanation anywhere on this page. It is defined on
    // the statement pages and in /metode, but this is where people read
    // it 119 times. A legend directly above the first table closes that
    // gap without sending anyone to another page (2026-10-23 review).
    var legend = document.createElement("details");
    legend.className = "disclosure evidens-legend";
    var legendSum = document.createElement("summary");
    legendSum.textContent = "Hva betyr evidensstyrken?";
    legend.appendChild(legendSum);
    var legendList = document.createElement("ul");
    [
      ["●●● Sterk", "minst én kilde i klasse A: en offisiell avstemning, en lov eller et register."],
      ["●●○ Moderat", "beste kilde er klasse B: partiets eget program eller en grundig redegjørelse."],
      ["●○○ Svak", "bare klasse C eller lavere: debatt, intervju eller sekundær omtale."],
      ["○○○ Ingen kodet posisjon", "kilder er listet, men ingen av dem koder et standpunkt. Vises som kontekst."],
      ["Ingen evidens innhentet ennå", "vi har ikke funnet noen kilde for denne cellen."]
    ].forEach(function (row) {
      var li = document.createElement("li");
      var strong = document.createElement("strong");
      strong.textContent = row[0];
      li.appendChild(strong);
      li.appendChild(document.createTextNode(" — " + row[1]));
      legendList.appendChild(li);
    });
    legend.appendChild(legendList);
    var metodeLink = document.createElement("p");
    metodeLink.className = "lede";
    var ml = document.createElement("a");
    ml.href = "/metode";
    ml.textContent = "Les mer om kildene og evidensstyrken";
    metodeLink.appendChild(ml);
    legend.appendChild(metodeLink);
    wrap.appendChild(legend);

    // Say the skip count out loud. A "vet ikke" is a real reply but not a
    // position, so it is held out of the match entirely — and a
    // percentage must never read as more complete than the answers
    // behind it (2026-10-12: "vet ikke" used to sit on the scale as 0 and
    // silently score as agreement with a neutral party).
    var skipped = countSkipped();
    if (skipped > 0) {
      wrap.appendChild(
        el(
          "p",
          "lede",
          "Du har svart «" + SKIP_LABEL + "» på " + skipped + " utsagn. " +
            "De teller verken som enighet eller uenighet og er holdt helt utenfor samsvaret."
        )
      );
    }

    wrap.appendChild(renderSummary());

    // The page is ~29 000 px tall with 60 statement tables. A compact
    // jump index at the top is the cheapest way to make the detail
    // reachable without a scroll marathon (2026-10-23 review). It is a
    // <details> so it stays out of the way until asked for.
    var toc = document.createElement("details");
    toc.className = "disclosure result-toc";
    var tocSum = document.createElement("summary");
    tocSum.textContent = "Hopp til et utsagn (" + DATA.statements.length + ")";
    toc.appendChild(tocSum);
    var tocGrid = document.createElement("div");
    tocGrid.className = "toc-grid";
    DATA.statements.forEach(function (st) {
      var a = document.createElement("a");
      a.href = "#resultat-utsagn-" + st.number;
      a.textContent = st.number + (st.short_name ? " " + st.short_name : "");
      tocGrid.appendChild(a);
    });
    toc.appendChild(tocGrid);
    wrap.appendChild(toc);

    DATA.statements.forEach(function (st) {

      var cells = DATA.cells[st.id] || {};
      var documented = [];
      var undocumented = [];
      DATA.parties.forEach(function (party) {
        var cell = cells[party.id] || {};
        if (cell.party_position === null || cell.party_position === undefined) {
          undocumented.push(party);
        } else {
          documented.push({ party: party, cell: cell });
        }
      });

      var section = el("div", "statement-result");
      // resultat-utsagn-, NOT utsagn-: that prefix is the quiz card's id,
      // owned solely by statementCardId() and treated by handleHash as
      // "open question N". Reusing it here would both duplicate the id
      // scheme (web/dom_id_guard_test.go fails on that) and make a TOC
      // click call showQuestion() on a page with no #quiz-region.
      section.id = "resultat-utsagn-" + st.number;
      var heading = el("h3", null, "Utsagn " + st.number + (st.short_name ? ": " + st.short_name : ""));
      section.appendChild(heading);

      if (documented.length === 0) {
        // Nothing coded yet: one compact line says exactly what the full
        // 9-row table used to say, instead of forcing a scan of nine
        // rows to discover they all say the same thing (2026-09-23
        // design review).
        var empty = el("div", "empty-state result-empty-state");
        empty.appendChild(el("p", "empty-state-count", "0 av " + DATA.parties.length + " partier dokumentert"));
        empty.appendChild(
          el("p", null, "Vi har ingen partiposisjoner med kilde for dette utsagnet ennå. Utsagnet teller derfor ikke i resultatet ditt.")
        );
        var link = document.createElement("a");
        link.href = "/utsagn/" + st.number;
        link.textContent = "Se kilden bak utsagnet\u00a0";
        appendArrow(link);
        empty.appendChild(link);
        section.appendChild(empty);
        wrap.appendChild(section);
        return;
      }

      var table = document.createElement("table");
      table.className = "party-overview";
      var thead = document.createElement("thead");
      var headRow = document.createElement("tr");
      ["Parti", "Samsvar", "Dokumentasjon"].forEach(function (label) {
        var th = document.createElement("th");
        th.scope = "col";
        th.textContent = label;
        headRow.appendChild(th);
      });
      thead.appendChild(headRow);
      table.appendChild(thead);
      var tbody = document.createElement("tbody");
      documented.forEach(function (d) {
        var userReply = answers[st.id];
        var matchPct = pct(userReply, d.cell.party_position);
        var row = document.createElement("tr");
        row.appendChild(el("td", "po-party", d.party.short));
        // Every party here has a coded position, so a null match means
        // the user gave no position: either "hoppet over" (a real reply,
        // held out of the match) or not answered at all.
        var scoreText;
        if (matchPct !== null) {
          scoreText = matchPct + "% samsvar";
        } else if (userReply === SKIP) {
          scoreText = "Hoppet over";
        } else {
          scoreText = "Ikke besvart";
        }
        row.appendChild(el("td", null, scoreText));
        // po-evidence, not po-value: .po-value is hidden under 520px on
        // the statement pages because the exact -2..+2 value repeats in
        // the per-party card below. Here the column IS the evidence
        // signal and has no repeat below, so it keeps its own class and
        // stays visible on a phone (2026-10-23 review).
        row.appendChild(el("td", "po-evidence", d.cell.evidensstyrke || ""));
        tbody.appendChild(row);
      });
      table.appendChild(tbody);
      section.appendChild(table);

      if (undocumented.length > 0) {
        var details = document.createElement("details");
        details.className = "disclosure undocumented-toggle";
        var summary = document.createElement("summary");
        summary.textContent = undocumented.length + (undocumented.length === 1 ? " parti" : " partier") + " uten posisjon med kilde";
        details.appendChild(summary);
        var uTable = document.createElement("table");
        uTable.className = "party-overview party-overview-empty";
        var uHead = document.createElement("thead");
        var uHeadRow = document.createElement("tr");
        ["Parti", "Status"].forEach(function (label) {
          var th = document.createElement("th");
          th.scope = "col";
          th.textContent = label;
          uHeadRow.appendChild(th);
        });
        uHead.appendChild(uHeadRow);
        uTable.appendChild(uHead);
        var uBody = document.createElement("tbody");
        undocumented.forEach(function (party) {
          var row = document.createElement("tr");
          row.appendChild(el("td", "po-party", party.short));
          row.appendChild(el("td", null, "Ikke vurdert"));
          uBody.appendChild(row);
        });
        uTable.appendChild(uBody);
        details.appendChild(uTable);
        section.appendChild(details);
      }

      var seeMore = el("p", "lede");
      var a2 = document.createElement("a");
      a2.href = "/utsagn/" + st.number;
      a2.textContent = "Se kilden bak dette utsagnet\u00a0";
      appendArrow(a2);
      seeMore.appendChild(a2);
      section.appendChild(seeMore);

      wrap.appendChild(section);
    });

    return wrap;
  }


  // resumePoint is where the one-question view should open. Answers were
  // always kept across a reload, but the PLACE was not: with 30 of 60
  // answered the quiz opened on question 1 again. It opens on the first
  // statement with no reply instead, or the last one when all are answered.
  function resumePoint() {
    var replied = 0, first = -1;
    DATA.statements.forEach(function (st, i) {
      if (answers[st.id] === undefined) {
        if (first < 0) first = i;
      } else {
        replied++;
      }
    });
    if (replied === 0) return null;
    return { index: first < 0 ? DATA.statements.length - 1 : first, replied: replied };
  }

  // Number-key shortcuts (2026-10-23 review): 1-5 pick a point on the
  // scale, S picks «vet ikke». Restricted to the one-question view —
  // in utforskermodus every card is on screen at once, so "the current
  // statement" would be whichever one happened to be in view. A modifier
  // key or a text field turns the shortcut off, so it can never eat a
  // normal keystroke. The action is a real .click() on the same button a
  // pointer would use, so focus, aria-pressed and the announce() call all
  // follow the existing path instead of a second implementation.
  document.addEventListener("keydown", function (e) {
    if (PAGE === "resultat") return;
    if (viewMode !== "enkel") return;
    if (reviewing || showingResult) return;
    if (e.ctrlKey || e.metaKey || e.altKey) return;
    var t = e.target;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.tagName === "SELECT" || t.isContentEditable)) return;
    var st = DATA.statements[currentIndex];
    if (!st) return;
    var card = document.getElementById(statementCardId(st.number));
    if (!card) return;
    var buttons = card.querySelectorAll(".scale button");
    var idx = -1;
    if (e.key >= "1" && e.key <= "5") {
      idx = Number(e.key) - 1;
    } else if (e.key === "s" || e.key === "S") {
      idx = SCALE.length; // the skip button is the last child of .scale
    } else {
      return;
    }
    var btn = buttons[idx];
    if (!btn) return;
    e.preventDefault();
    btn.click();
  });

  noteState.nosave = !Store.available(storage);
  noteState.dropped = Store.droppedMessage(loadReport.kept, loadReport.dropped);
  if (PAGE !== "resultat" && viewMode !== "utforsker") {
    var resume = resumePoint();
    if (resume) {
      currentIndex = resume.index;
      noteState.resume =
        "Du fortsetter der du slapp. Du har svart på " + resume.replied + " av " + DATA.statements.length + ".";
    }
  }
  paintNote();

  renderQuiz();
  handleHash();
  window.addEventListener("hashchange", handleHash);
})();
