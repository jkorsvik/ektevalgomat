// Pure scoring helpers shared by the quiz UI (quiz.js) and the node check
// (tools/quiz-scoring.test.js). No DOM, no network, no framework: this file
// answers only two questions — is a reply a position on the -2..+2 scale,
// and how well does a position match a party's coded position.
//
// 2026-10-12: a user's "Vet ikke / hopp over" is stored as SKIP, a string,
// NOT the number 0. Keeping it off the numeric scale is the whole point:
// pct() returns null for it, so a skip is excluded from every party-match
// computation — neither agreement nor disagreement — while a genuine 0
// ("Nøytral") is still a position and still counts. This is the same
// distinction the site already makes for parties: "Ikke vurdert" does not
// mean "nøytralt" (see the resultat text in quiz.js).
(function (root, factory) {
  if (typeof module === "object" && module.exports) {
    module.exports = factory();
  } else {
    root.EkteScoring = factory();
  }
})(typeof self !== "undefined" ? self : this, function () {
  "use strict";

  // The off-scale "don't know" marker. Deliberately a string: the typeof
  // check in isPosition() is exactly what keeps it out of the match.
  var SKIP = "vet-ikke";

  // A reply is a position only when it is a number on the scale. SKIP is
  // not one, and neither is an unanswered statement (undefined).
  function isPosition(reply) {
    return typeof reply === "number";
  }

  function isSkip(reply) {
    return reply === SKIP;
  }

  // pct scores a user answer (-2..+2) against a party's coded position on
  // the same scale. Returns null when there is nothing to score: no
  // position from the user (unanswered OR skipped) or no coded party
  // position. Callers must render null as "not scored", never as 0%.
  function pct(reply, partyPosition) {
    if (!isPosition(reply)) return null;
    if (partyPosition === null || partyPosition === undefined) return null;
    var diff = Math.abs(reply - partyPosition);
    return Math.round(100 - (diff * 100) / 4);
  }

  // Importance weighting (2026-10-13). The user can mark a statement as
  // mattering more than the others. That is stored in its own map,
  // statementId -> 1..3, and NEVER written into the answer. This is the
  // whole design: a weight and a position are different kinds of thing.
  // A position says where the visitor stands (-2..+2). A weight says how
  // much that standing should pull the result, and it is invisible to the
  // per-statement match: pct() above does not know weights exist. So a
  // weighted result can never be mistaken for a different answer.
  var MIN_WEIGHT = 1;
  var MAX_WEIGHT = 3;

  // weightOf reads a weight safely. A missing, malformed or out-of-range
  // value is MIN_WEIGHT, i.e. "not weighted" — never a guess, and never 0
  // (a weight of 0 would silently drop the statement the user did answer,
  // which is what SKIP is for).
  function weightOf(weights, id) {
    if (!weights) return MIN_WEIGHT;
    var w = weights[id];
    if (typeof w !== "number" || !isFinite(w)) return MIN_WEIGHT;
    if (w < MIN_WEIGHT) return MIN_WEIGHT;
    if (w > MAX_WEIGHT) return MAX_WEIGHT;
    return w;
  }

  // weightedMean is the mean of {value, weight} pairs, or null when there
  // is nothing to average. Every caller counts only pairs it already
  // decided are scorable, so a skipped statement never reaches here.
  function weightedMean(pairs) {
    var num = 0, den = 0;
    for (var i = 0; i < pairs.length; i++) {
      var p = pairs[i];
      if (typeof p.value !== "number") continue;
      var w = (typeof p.weight === "number" && p.weight > 0) ? p.weight : MIN_WEIGHT;
      num += p.value * w;
      den += w;
    }
    return den > 0 ? Math.round(num / den) : null;
  }

  // leaders returns every row on the top figure, in the order given. A tie at
  // the top is a tie. The headline used to name rows[0], so payload order
  // picked a winner among parties with identical numbers: measured on
  // 2026-10-13, "mest på linje med Senterpartiet (73%)" with Venstre also at
  // 73 %. Rows with no figure (avg null) never lead.
  function leaders(rows) {
    var best = null;
    rows.forEach(function (r) {
      if (r.avg !== null && r.avg !== undefined && (best === null || r.avg > best)) best = r.avg;
    });
    if (best === null) return [];
    return rows.filter(function (r) { return r.avg === best; });
  }

  return {
    SKIP: SKIP, isPosition: isPosition, isSkip: isSkip, pct: pct,
    MIN_WEIGHT: MIN_WEIGHT, MAX_WEIGHT: MAX_WEIGHT,
    weightOf: weightOf, weightedMean: weightedMean, leaders: leaders,
  };
});
