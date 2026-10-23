// Ekte valgomat — the "Hvorfor dette resultatet?" robustness panel on
// /resultat. Vanilla JS, no build step, no framework, no network (this is
// asserted by web/network_guard_test.go, which walks every .js under
// static/). The panel is deliberately collapsed behind one native
// <details>/<summary> so a first-time visitor still meets a clean result
// first.
//
// It answers four questions about how robust the headline number is:
//   1. which answered statements pulled most toward agreement, and most
//      toward disagreement;
//   2. which statements carried missing or contradictory documentation, and
//      therefore pulled little or no weight;
//   3. whether the ranking changes if every theme (dimension) weighs the
//      same instead of every statement;
//   4. how stable the top of the ranking is when one statement, or one whole
//      theme, is removed.
//
// The point is to show how much the conclusion depends on the sample, not
// to let someone tune the inputs until a preferred party wins. The panel's
// own copy says exactly that.
//
// Scoring (SKIP, isPosition, pct) is owned by scoring.js and loaded before
// this file on /resultat. This module USES it, it never copies it. In node
// it is required from the sibling file; in the browser it is read from the
// global EkteScoring that scoring.js already set. The pure functions are
// exported so web/static/js/analyse.test.js can call them directly.
(function (root, factory) {
  if (typeof module === "object" && module.exports) {
    module.exports = factory(require("./scoring.js"), require("./store.js"));
  } else {
    root.EkteAnalyse = factory(root.EkteScoring, root.EkteStore);
  }
})(typeof self !== "undefined" ? self : this, function (S, Store) {
  "use strict";

  // The weight store is quiz.js's own key and payload shape. Read here, never
  // written: the quiz owns the weights, the panel only reports on them.
  var WEIGHT_KEY = "ektevalgomat-vekt-v1";
  var STORAGE_KEY = "ektevalgomat-svar-v1";

  // --- reading the payload / the answers -------------------------------

  // The panel reads the visitor's replies and weights through the SAME
  // reader the quiz writes them with (store.js). It used to carry its own
  // copy of the snapshot rule, so the quiz and this panel could disagree
  // about what the visitor had answered. Read only: the quiz owns the data.
  function browserStorage() {
    try { return typeof localStorage === "undefined" ? null : localStorage; } catch (e) { return null; }
  }
  function loadAnswers(data) {
    return Store.read(browserStorage(), STORAGE_KEY, "answers", data).map;
  }
  function loadWeights(data) {
    return Store.read(browserStorage(), WEIGHT_KEY, "weights", data).map;
  }

  // positionOf returns a cell's coded party position (-2..2) or null when
  // the cell is undocumented. The quiz payload carries party_position as a
  // number or null; nothing else here reimplements the scoring contract.
  function positionOf(cell) {
    if (!cell) return null;
    var p = cell.party_position;
    return p === null || p === undefined ? null : p;
  }

  // docFlag classifies a cell whose own evidence is flagged as shaky. The
  // payload's konsistens is a full sentence from internal/scoring; these
  // prefixes are the ones that mean the party's record is not clean:
  // "Motsigelse" (a real contradiction), "Spenning" (a tension the
  // evidence itself names), and "Endret standpunkt" (an explicit change).
  // Positive labels ("Konsistent", "Ikke nok data", "Sakene ligner") are
  // not flags.
  function docFlag(cell) {
    var k = cell && cell.konsistens;
    if (!k) return null;
    if (k.indexOf("Motsigelse") === 0) return "motstridende";
    if (k.indexOf("Spenning") === 0) return "spenning";
    if (k.indexOf("Endret standpunkt") === 0) return "endret standpunkt";
    return null;
  }

  // --- pure arithmetic -------------------------------------------------

  // statementMatch scores one statement across every party that has a coded
  // position. avg is the mean match% over those parties (rounded), n is how
  // many pairs it rests on, documented is how many parties carry a coded
  // position. A statement the user answered with "vet-ikke" or did not
  // answer yields n: 0 and avg: null; so does a statement no party has
  // documented. A skip is never neutral, so it never enters here.
  function statementMatch(data, answers, st) {
    var reply = answers[st.id];
    var total = data.parties.length;
    if (!S.isPosition(reply)) return { avg: null, n: 0, documented: 0, total: total };
    var cells = data.cells[st.id] || {};
    var sum = 0, n = 0, documented = 0;
    data.parties.forEach(function (p) {
      var pos = positionOf(cells[p.id]);
      if (pos === null) return;
      documented++;
      var m = S.pct(reply, pos);
      if (m === null) return;
      sum += m;
      n++;
    });
    return { avg: n > 0 ? Math.round(sum / n) : null, n: n, documented: documented, total: total };
  }

  // aggregate averages match% per party over statementIds, counting only the
  // (statement, party) pairs where that party has a coded position and the
  // user gave a position. Same rule quiz.js's own headline uses. A party
  // with no coded positions in this scope gets avg: null (never a
  // fabricated 0%).
  // weights is optional: statementId -> 1..3, the importance the visitor
  // assigned (see scoring.js). It scales how much a statement pulls the
  // result and does nothing else — it never touches pct(), so a weighted
  // aggregate can never be mistaken for a different answer. Omitted, every
  // weight is 1 and the result is the plain mean, exactly as before.
  function aggregate(data, answers, statementIds, weights) {
    var num = {}, den = {}, counts = {}, wsum = {};
    data.parties.forEach(function (p) { num[p.id] = 0; den[p.id] = 0; counts[p.id] = 0; wsum[p.id] = 0; });
    statementIds.forEach(function (id) {
      var reply = answers[id];
      if (!S.isPosition(reply)) return;
      var cells = data.cells[id] || {};
      var w = S.weightOf(weights, id);
      if (w > 1) wsum[id] = w;
      data.parties.forEach(function (p) {
        var m = S.pct(reply, positionOf(cells[p.id]));
        if (m === null) return;
        num[p.id] += m * w;
        den[p.id] += w;
        counts[p.id] += 1;
      });
    });
    return data.parties.map(function (p) {
      var n = counts[p.id];
      return {
        partyId: p.id, party: p,
        avg: den[p.id] > 0 ? Math.round(num[p.id] / den[p.id]) : null,
        n: n,
        // totalWeight lets the result page say whether an answer was
        // weighted at all, rather than assuming the visitor used it.
        totalWeight: den[p.id],
      };
    });
  }

  // weightedIds lists the statements the visitor marked as mattering more,
  // in statement order. Empty is the normal case, and the result page must
  // say so rather than imply weighting happened.
  function weightedIds(data, weights) {
    return data.statements
      .filter(function (st) { return S.weightOf(weights, st.id) > 1; })
      .map(function (st) { return st; });
  }

  // ranked sorts aggregate rows best first, parties with no score last,
  // with a stable tie-break on the payload's own party order.
  function ranked(rows) {
    return rows
      .map(function (r, i) { return { r: r, i: i }; })
      .sort(function (a, b) {
        if (a.r.avg === null && b.r.avg === null) return a.i - b.i;
        if (a.r.avg === null) return 1;
        if (b.r.avg === null) return -1;
        if (b.r.avg !== a.r.avg) return b.r.avg - a.r.avg;
        return a.i - b.i;
      })
      .map(function (x) { return x.r; });
  }

  // topPartyId returns the best-ranked party, or null when nobody has a score.
  function topPartyId(rows) {
    var r = ranked(rows)[0];
    return r && r.avg !== null ? r.partyId : null;
  }

  // contributingIds is every statement the user answered with a position that
  // at least one party has a coded position on; the statements that actually
  // enter the aggregate.
  function contributingIds(data, answers) {
    return data.statements
      .filter(function (st) { return statementMatch(data, answers, st).n > 0; })
      .map(function (st) { return st.id; });
  }

  // contribution ranks answered statements that carried any weight by their
  // average match across documented parties: high = pulled toward agreement,
  // low = pulled toward disagreement.
  function contribution(data, answers) {
    var rows = [];
    data.statements.forEach(function (st) {
      var m = statementMatch(data, answers, st);
      if (m.n === 0) return;
      rows.push({ statement: st, avg: m.avg, n: m.n, documented: m.documented, total: m.total });
    });
    rows.sort(function (a, b) {
      return b.avg - a.avg || a.statement.number - b.statement.number;
    });
    return rows;
  }

  // documentationGaps lists answered statements whose documentation is
  // missing (some or all parties have no coded position) or flagged
  // (a documented cell whose own evidence reports a contradiction, a
  // tension, or a changed position). A fully undocumented statement pulled
  // no weight at all.
  function documentationGaps(data, answers) {
    var out = [];
    data.statements.forEach(function (st) {
      if (!S.isPosition(answers[st.id])) return;
      var cells = data.cells[st.id] || {};
      var documented = [], undocumented = [], flags = [];
      data.parties.forEach(function (p) {
        var cell = cells[p.id] || {};
        if (positionOf(cell) === null) undocumented.push(p);
        else documented.push(p);
        var f = docFlag(cell);
        if (f) flags.push({ party: p, flag: f });
      });
      if (undocumented.length === 0 && flags.length === 0) return;
      out.push({
        statement: st,
        documented: documented.length,
        total: data.parties.length,
        undocumented: undocumented,
        flags: flags,
      });
    });
    return out;
  }

  // dimensionIds groups the given statement ids by their dimension,
  // preserving first-appearance order.
  function dimensionIds(data, ids) {
    var set = {};
    ids.forEach(function (id) { set[id] = true; });
    var order = [], byDim = {};
    data.statements.forEach(function (st) {
      if (!set[st.id]) return;
      if (!byDim[st.dimension]) { byDim[st.dimension] = []; order.push(st.dimension); }
      byDim[st.dimension].push(st.id);
    });
    return order.map(function (d) { return { dimension: d, ids: byDim[d] }; });
  }

  // themeEqualWeight gives every theme the same weight regardless of how
  // many statements it holds: average each dimension's per-party average,
  // then average those dimension averages. A theme where the party has no
  // documented pair is left out of that party's own mean, never padded.
  function themeEqualWeight(data, answers) {
    var dims = dimensionIds(data, contributingIds(data, answers));
    return data.parties.map(function (p) {
      var perTheme = [];
      dims.forEach(function (d) {
        var row = aggregate(data, answers, d.ids).filter(function (r) { return r.partyId === p.id; })[0];
        if (row && row.avg !== null) perTheme.push(row.avg);
      });
      var avg = perTheme.length
        ? Math.round(perTheme.reduce(function (a, b) { return a + b; }, 0) / perTheme.length)
        : null;
      return { partyId: p.id, party: p, avg: avg, themes: perTheme.length };
    });
  }

  // summarizeStability compares a baseline top party against a set of
  // leave-one-out results (each {removed, top}).
  function summarizeStability(baseTop, results) {
    var unchanged = 0, skipped = 0, flips = {};
    results.forEach(function (r) {
      if (r.top === null) { skipped++; return; }
      if (r.top === baseTop) unchanged++;
      else flips[r.top] = (flips[r.top] || 0) + 1;
    });
    return { baseTop: baseTop, total: results.length, unchanged: unchanged, skipped: skipped, flips: flips };
  }

  // stability removes one statement at a time and reports how often the top
  // party holds.
  function stability(data, answers, ids, weights) {
    var baseTop = topPartyId(aggregate(data, answers, ids, weights));
    var results = ids.map(function (id) {
      var rest = ids.filter(function (x) { return x !== id; });
      return { removed: id, top: topPartyId(aggregate(data, answers, rest, weights)) };
    });
    return summarizeStability(baseTop, results);
  }

  // themeStability removes one whole theme (dimension) at a time.
  function themeStability(data, answers, ids, weights) {
    var baseTop = topPartyId(aggregate(data, answers, ids, weights));
    var results = dimensionIds(data, ids).map(function (d) {
      var rest = ids.filter(function (id) { return d.ids.indexOf(id) === -1; });
      return { removed: d.dimension, top: topPartyId(aggregate(data, answers, rest, weights)) };
    });
    return summarizeStability(baseTop, results);
  }

  // --- DOM rendering ---------------------------------------------------

  var FLAG_LABEL = {
    "motstridende": "motstridende",
    "spenning": "spenning",
    "endret standpunkt": "endret standpunkt",
  };

  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }

  // pctStepClass mirrors quiz.js's own helper: the site draws bars with a
  // discrete CSS step class (.pct-0 .. .pct-100 in style.css), never an
  // inline style="" — the CSP here has no 'unsafe-inline' for style-src.
  function pctStepClass(p) {
    var step = Math.round(p / 5) * 5;
    if (step < 0) step = 0;
    if (step > 100) step = 100;
    return "pct-" + step;
  }

  function partyNameMap(data) {
    var m = {};
    data.parties.forEach(function (p) { m[p.id] = p; });
    return m;
  }

  function lede(text) { return el("p", "lede", text); }

  function section(title) {
    var s = el("div", "analyse-section");
    s.appendChild(el("h3", null, title));
    return s;
  }

  function subhead(text) { return el("h4", "analyse-subhead", text); }

  // barRow reuses the site's bar treatment (.bar-track/.bar-fill/.pct-N).
  function barRow(name, pctText, pctValue, note) {
    var row = el("div", "analyse-row");
    var label = el("div", "analyse-row-label");
    label.appendChild(el("span", "analyse-row-name", name));
    if (pctText) label.appendChild(el("span", "analyse-row-pct", pctText));
    row.appendChild(label);
    var track = el("div", "bar-track");
    if (pctValue !== null && pctValue !== undefined) {
      track.appendChild(el("div", "bar-fill " + pctStepClass(pctValue)));
    }
    row.appendChild(track);
    row.appendChild(el("div", "analyse-row-n", note));
    return row;
  }

  function intro() {
    var w = el("div", "analyse-intro");
    w.appendChild(
      lede(
        "Resultatet ditt er et snitt av samsvaret mellom svarene dine og partienes " +
          "posisjoner med kilde. Hvor mye du kan feste lit til tallet, avhenger av to ting. " +
          "Hvor mange utsagn som har en posisjon med kilde, og hvor mye hvert utsagn veier " +
          "i snittet. Dette panelet viser begge deler."
      )
    );
    w.appendChild(
      lede(
        "Poenget er å vise hvor avhengig svaret er av hvilke utsagn som er med, ikke å " +
          "la deg justere inntil et bestemt parti vinner. Tallene under beskriver hvor " +
          "stødig resultatet er, ikke hvilket parti som er det riktige."
      )
    );
    return w;
  }

  function statementLabel(st) {
    return "Utsagn " + st.number + (st.short_name ? ": " + st.short_name : "");
  }

  function sectionContribution(data, answers) {
    var head = section("Hva som dro mest i hver retning");
    head.appendChild(
      lede(
        "Utsagnene du har svart på med en posisjon, sortert etter hvor godt de samsvarer " +
          "med partienes posisjoner med kilde. Øverst ligger utsagn der du er enig med " +
          "flest. Nederst der uenigheten er størst. «N av M» viser hvor mange av partiene " +
          "som har en posisjon med kilde på utsagnet, altså hvor mye det veier."
      )
    );
    var rows = contribution(data, answers);
    if (rows.length === 0) {
      head.appendChild(lede("Ingen av utsagnene du har besvart, har en dokumentert partiposisjon å måle mot."));
      return head;
    }
    function rowEl(r) {
      return barRow(
        statementLabel(r.statement),
        r.avg + " % samsvar",
        r.avg,
        r.documented + " av " + r.total + " partier"
      );
    }
    var top = rows.slice(0, 3);
    head.appendChild(subhead("Mest enighet"));
    top.forEach(function (r) { head.appendChild(rowEl(r)); });

    if (rows.length > 3) {
      head.appendChild(subhead("Mest uenighet"));
      rows
        .slice(-3)
        .forEach(function (r) {
          if (top.indexOf(r) !== -1) return;
          head.appendChild(rowEl(r));
        });
    }
    return head;
  }

  function sectionGaps(data, answers) {
    var head = section("Der dokumentasjonen mangler eller er motstridende");
    head.appendChild(
      lede(
        "Et utsagn teller bare i den grad partiene har en posisjon med kilde på det. " +
          "Mangler kildene, faller utsagnet ut av snittet uten å si fra. Listen viser " +
          "utsagn du har svart på der dette skjer, og hvor et partis eget grunnlag er " +
          "flagget som motstridende eller endret."
      )
    );
    var gaps = documentationGaps(data, answers);
    if (gaps.length === 0) {
      head.appendChild(
        lede("Alle utsagnene du har besvart med en posisjon, har dokumentasjon for alle partier, uten flagg.")
      );
      return head;
    }
    gaps.forEach(function (g) {
      var item = el("div", "analyse-gap");
      item.appendChild(el("p", "analyse-gap-title", statementLabel(g.statement)));
      var line;
      if (g.documented === 0) {
        line = "Ingen partiposisjon er dokumentert. Utsagnet telte ikke i noens resultat.";
      } else if (g.undocumented.length > 0) {
        line =
          "Dokumentert for " + g.documented + " av " + g.total + " partier. De øvrige er ikke " +
          "vurdert på dette utsagnet og telte ikke for dem.";
      } else {
        line = "Dokumentert for alle partier.";
      }
      item.appendChild(lede(line));
      if (g.flags.length > 0) {
        var names = g.flags
          .map(function (f) { return f.party.short + " (" + (FLAG_LABEL[f.flag] || f.flag) + ")"; })
          .join(", ");
        item.appendChild(
          el(
            "p",
            "analyse-flag",
            "Flagget i kilden: " + names + ". Partiet har en posisjon, men grunnlaget er merket som usikkert."
          )
        );
      }
      head.appendChild(item);
    });
    return head;
  }

  function sectionEqualWeight(data, answers, ids, weights) {
    var head = section("Om alle temaer hadde veid likt");
    var names = partyNameMap(data);
    var rows = ranked(themeEqualWeight(data, answers));
    var baseTop = topPartyId(aggregate(data, answers, ids, weights));
    var eqTop = rows[0] && rows[0].avg !== null ? rows[0].partyId : null;
    var msg;
    if (eqTop === null) {
      msg = "Det er ikke nok dokumenterte posisjoner til å regne ut et likt-vekt-resultat.";
    } else if (eqTop === baseTop) {
      msg =
        "Med lik vekt per tema er du fortsatt mest på linje med " + names[eqTop].name + ". Rangeringen holder.";
    } else {
      msg =
        "Med lik vekt per tema blir " + names[eqTop].name + " partiet på topp, ikke " +
        (baseTop ? names[baseTop].name : "partiet fra standardresultatet") +
        ". Det viser at resultatet avhenger av hvor mange utsagn hvert tema har.";
    }
    head.appendChild(
      lede(
        (weightedIds(data, weights).length
          ? "Resultatet ditt bruker vektene du har satt. Her får i stedet hvert tema samme "
          : "Standardresultatet vekter hvert utsagn likt. Her får i stedet hvert tema samme ") +
          "vekt, uansett hvor mange utsagn det rommer. " + msg
      )
    );
    rows.forEach(function (r) {
      head.appendChild(
        barRow(
          r.party.short,
          r.avg !== null ? r.avg + " % samsvar" : "Ikke vurdert",
          r.avg,
          r.themes + (r.themes === 1 ? " tema" : " temaer")
        )
      );
    });
    return head;
  }

  function stabilityLine(word, st, names) {
    var p = el("p", "lede");
    if (st.baseTop === null) {
      p.textContent = "Det er ikke nok dokumenterte posisjoner til å vurdere stabiliteten på " + word + ".";
      return p;
    }
    var text =
      "Ett " + word + " om gangen: partiet på topp (" + names[st.baseTop].name + ") holder seg i " +
      st.unchanged + " av " + st.total + " tilfeller.";
    var flipIds = Object.keys(st.flips);
    if (flipIds.length > 0) {
      text += " Ellers går toppen til " +
        flipIds.map(function (id) { return names[id].name + " (" + st.flips[id] + ")"; }).join(", ") + ".";
    } else {
      text += " Ingen enkelt fjerning endrer hvem som ligger øverst.";
    }
    p.textContent = text;
    return p;
  }

  function sectionStability(data, answers, ids, weights) {
    var head = section("Hvor stødig er rangeringen");
    var names = partyNameMap(data);
    head.appendChild(
      lede("Vi fjerner ett og ett utsagn, og ett og ett tema, og ser om partiet på topp holder seg.")
    );
    head.appendChild(stabilityLine("utsagn", stability(data, answers, ids, weights), names));
    head.appendChild(stabilityLine("tema", themeStability(data, answers, ids, weights), names));
    return head;
  }

  function render(data, answers, mountEl, weights) {
    mountEl.innerHTML = "";
    var details = document.createElement("details");
    details.className = "analyse-panel";
    var summary = document.createElement("summary");
    summary.textContent = "Hvorfor dette resultatet?";
    details.appendChild(summary);
    var body = el("div", "analyse-body");
    details.appendChild(body);

    body.appendChild(intro());

    var ids = contributingIds(data, answers);
    if (ids.length === 0) {
      body.appendChild(
        lede(
          "Svar på minst ett utsagn med en posisjon, og velg et utsagn der minst ett parti " +
            "har en posisjon med kilde, for å se denne analysen."
        )
      );
    } else {
      body.appendChild(sectionContribution(data, answers));
      body.appendChild(sectionGaps(data, answers));
      body.appendChild(sectionEqualWeight(data, answers, ids, weights));
      body.appendChild(sectionStability(data, answers, ids, weights));
    }
    mountEl.appendChild(details);
    return details;
  }

  function boot() {
    var mountEl = document.getElementById("analyse-root");
    if (!mountEl) return;
    var dataEl = document.getElementById("ektevalgomat-data");
    if (!dataEl) return;
    var data;
    try {
      data = JSON.parse(dataEl.textContent);
    } catch (e) {
      return;
    }
    var details = render(data, loadAnswers(data), mountEl, loadWeights(data));
    // A trade-off changes the weights while the page is open. The quiz says so
    // with an event that CARRIES the weights and answers now in force, so the
    // panel explains exactly the result on screen. Re-reading storage instead
    // would show it stale weights whenever a save had failed: two top parties
    // on one page, the bug #117 fixed, arriving through the side door.
    document.addEventListener("ektevalgomat:vekt-endret", function (ev) {
      var wasOpen = details && details.open;
      var d = (ev && ev.detail) || {};
      details = render(data, d.answers || loadAnswers(data), mountEl, d.weights || loadWeights(data));
      if (wasOpen) details.open = true;
    });
  }

  if (typeof document !== "undefined") {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", boot);
    } else {
      boot();
    }
  }

  return {
    loadAnswers: loadAnswers,
    positionOf: positionOf,
    docFlag: docFlag,
    statementMatch: statementMatch,
    aggregate: aggregate,
    weightedIds: weightedIds,
    ranked: ranked,
    topPartyId: topPartyId,
    contributingIds: contributingIds,
    contribution: contribution,
    documentationGaps: documentationGaps,
    dimensionIds: dimensionIds,
    themeEqualWeight: themeEqualWeight,
    summarizeStability: summarizeStability,
    stability: stability,
    themeStability: themeStability,
    pctStepClass: pctStepClass,
    render: render,
  };
});
