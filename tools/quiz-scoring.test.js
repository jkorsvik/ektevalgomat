// Heuristic check for the quiz's client-side scoring, run with the same
// node the site's plain-JS has no build step for:
//
//   node tools/quiz-scoring.test.js
//
// The server never scores a user's answers (see /resultat in
// internal/server/server.go), so this is the only executable check of the
// match arithmetic and of the rule this test exists for: a "Vet ikke /
// hopp over" answer is excluded from the match — neither agreement nor
// disagreement — while a genuine 0 ("Nøytral") is a position and counts.
//
// It loads the REAL shipped module (web/static/js/scoring.js), not a copy,
// so the check cannot drift from what the browser runs.

"use strict";

var S = require("../web/static/js/scoring.js");

var failures = 0;
function check(name, got, want) {
  var ok = got === want;
  if (!ok) {
    failures++;
    console.error("FAIL " + name + ": got " + JSON.stringify(got) + ", want " + JSON.stringify(want));
  } else {
    console.log("ok   " + name + " = " + JSON.stringify(got));
  }
}

// --- 1. Scale arithmetic (a -2..+2 distance normalised to 0..100). ---
check("pct(2, 2)", S.pct(2, 2), 100); // exact agreement at the positive end
check("pct(-2, -2)", S.pct(-2, -2), 100); // exact agreement at the negative end
check("pct(0, 0)", S.pct(0, 0), 100); // a genuine neutral position matches a neutral party
check("pct(2, -2)", S.pct(2, -2), 0); // maximum disagreement
check("pct(0, 2)", S.pct(0, 2), 50); // halfway
check("pct(1, 2)", S.pct(1, 2), 75);
check("pct(-1, 0)", S.pct(-1, 0), 75);

// --- 2. "Vet ikke" is off the scale and never scores. ---
check("isPosition(SKIP)", S.isPosition(S.SKIP), false);
check("isSkip(SKIP)", S.isSkip(S.SKIP), true);
check("isPosition(0)", S.isPosition(0), true); // Nøytral is still a position
check("isSkip(0)", S.isSkip(0), false); // ... and is NOT a skip
check("pct(SKIP, 0)", S.pct(S.SKIP, 0), null); // skip vs neutral party: no score
check("pct(SKIP, 2)", S.pct(S.SKIP, 2), null); // skip vs positive party: no score
check("pct(SKIP, -2)", S.pct(S.SKIP, -2), null); // skip vs negative party: no score
check("pct(undefined, 0)", S.pct(undefined, 0), null); // unanswered: no score

// --- 3. Missing data is never a fabricated 0%. ---
check("pct(0, null)", S.pct(0, null), null); // party position not coded
check("pct(0, undefined)", S.pct(0, undefined), null);

// --- 4. Aggregate heuristic: a skip is excluded exactly like an
// unanswered statement, and never scored. Mirrors computeAggregate in
// quiz.js over two parties and three statements, so the exclusion is
// proven at the level the UI actually averages, not only on a single cell.
function aggregate(answers, cells) {
  var partyIds = Object.keys(cells[Object.keys(cells)[0]]);
  var out = {};
  partyIds.forEach(function (pid) {
    var sum = 0, n = 0;
    Object.keys(cells).forEach(function (stId) {
      var reply = answers[stId];
      if (reply === undefined) return;
      var m = S.pct(reply, cells[stId][pid]);
      if (m === null) return;
      sum += m;
      n += 1;
    });
    out[pid] = { avg: n > 0 ? Math.round(sum / n) : null, n: n };
  });
  return out;
}

var cells = {
  s1: { a: 2, b: -2 },
  s2: { a: 0, b: 0 },
  s3: { a: 2, b: 2 },
};

// A SKIP on s3 must give the exact same result as leaving s3 unanswered:
// that is what "excluded from the match entirely" means.
var s3Skipped = aggregate({ s1: 2, s2: 0, s3: S.SKIP }, cells);
var s3Absent = aggregate({ s1: 2, s2: 0 }, cells);
var s3Answered = aggregate({ s1: 2, s2: 0, s3: 2 }, cells);

check("party a avg: skip == unanswered", JSON.stringify(s3Skipped.a), JSON.stringify(s3Absent.a));
check("party b avg: skip == unanswered", JSON.stringify(s3Skipped.b), JSON.stringify(s3Absent.b));
check("party a avg over {s1,s2}", s3Absent.a.avg, 100); // (100 + 100) / 2
check("party b sample drops when s3 is skipped", s3Skipped.b.n, s3Answered.b.n - 1);
check("party b avg over {s1,s2}", s3Absent.b.avg, 50); // (0 + 100) / 2

// A run that is ALL skips yields null for every party: nothing to score,
// not a 0% (which would read as maximum disagreement).
var allSkip = aggregate({ s1: S.SKIP, s2: S.SKIP, s3: S.SKIP }, cells);
check("all-skip party a avg", allSkip.a.avg, null);
check("all-skip party a sample", allSkip.a.n, 0);

// --- 6. Importance weighting is a separate axis from position. -------
// The point of these checks: a weight can never become an answer. weightOf
// reads its own map and pct() has no idea weights exist, so the two cannot
// be confused by any later edit that keeps this file green.
check("weightOf on nothing", S.weightOf(null, "s1"), 1);
check("weightOf on a missing id", S.weightOf({ s2: 2 }, "s1"), 1);
check("weightOf on a string", S.weightOf({ s1: "hoy" }, "s1"), 1);
check("weightOf on NaN", S.weightOf({ s1: NaN }, "s1"), 1);
check("weightOf reads a real weight", S.weightOf({ s1: 3 }, "s1"), 3);
check("weightOf clamps above MAX_WEIGHT", S.weightOf({ s1: 99 }, "s1"), S.MAX_WEIGHT);
check("weightOf clamps below MIN_WEIGHT", S.weightOf({ s1: -5 }, "s1"), S.MIN_WEIGHT);
check("weightOf never yields 0", S.weightOf({ s1: 0 }, "s1"), 1);

// A weight must not move the per-statement match, on any input.
check("pct ignores weights (agree)", S.pct(2, 2), 100);
check("pct still null for a skip with weights around", S.pct(S.SKIP, 2), null);

// weightedMean: the arithmetic itself, including the empty cases.
check("weightedMean of nothing", S.weightedMean([]), null);
check("weightedMean of non-numbers", S.weightedMean([{ value: null, weight: 3 }]), null);
check("weightedMean equal weights == plain mean", S.weightedMean([{ value: 100, weight: 1 }, { value: 0, weight: 1 }]), 50);
check("weightedMean pulls toward the heavy one", S.weightedMean([{ value: 100, weight: 1 }, { value: 0, weight: 3 }]), 25);
check("weightedMean treats a missing weight as 1", S.weightedMean([{ value: 100 }, { value: 0, weight: 3 }]), 25);
check("weightedMean rounds", S.weightedMean([{ value: 100, weight: 1 }, { value: 0, weight: 2 }]), 33);

// leaders: a tie at the top is named as a tie, never settled by row order.
// (check() here is strict ===, so arrays are compared as JSON strings.)
function row(id, avg) { return { party: { id: id }, avg: avg }; }
function ids(rows) { return JSON.stringify(rows.map(function (r) { return r.party.id; })); }
check("leaders: a single winner", ids(S.leaders([row("a", 80), row("b", 70)])), '["a"]');
check("leaders: a two-way tie names both, in the order given", ids(S.leaders([row("sp", 73), row("v", 73), row("h", 60)])), '["sp","v"]');
check("leaders: the same tie in the other order is still both", ids(S.leaders([row("v", 73), row("sp", 73), row("h", 60)])), '["v","sp"]');
check("leaders: a three-way tie", S.leaders([row("a", 50), row("b", 50), row("c", 50)]).length, 3);
check("leaders: a lower tie does not count", ids(S.leaders([row("a", 90), row("b", 40), row("c", 40)])), '["a"]');
check("leaders: no figures at all", S.leaders([row("a", null), row("b", null)]).length, 0);
check("leaders: null never leads, even beside a real zero", ids(S.leaders([row("a", null), row("b", 0)])), '["b"]');
check("leaders: an empty list", S.leaders([]).length, 0);

if (failures > 0) {
  console.error("\n" + failures + " check(s) failed");
  process.exit(1);
}
console.log("\nAll scoring checks passed.");
