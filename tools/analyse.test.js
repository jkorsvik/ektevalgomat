// Node check for the robustness panel's arithmetic, run with the same node
// the site's plain-JS has no build step for:
//
//   node tools/analyse.test.js
//
// It loads the REAL shipped module (web/static/js/analyse.js), not a copy, so
// the check cannot drift from what the browser runs. The fixtures are small
// and hand-computed, and they exercise the two cases the panel must get
// right: a "vet-ikke" reply that is excluded from every match (never
// neutral), and a statement nobody has documented (that pulls no weight).

"use strict";

var A = require("../web/static/js/analyse.js");
var S = require("../web/static/js/scoring.js");

var failures = 0;
function check(name, got, want) {
  var ok = JSON.stringify(got) === JSON.stringify(want);
  if (!ok) {
    failures++;
    console.error("FAIL " + name + ": got " + JSON.stringify(got) + ", want " + JSON.stringify(want));
  } else {
    console.log("ok   " + name + " = " + JSON.stringify(got));
  }
}

// --- Fixture. Two themes (dimensions), three parties. Party C has no coded
// position on s1 or s3; s4 is undocumented for everyone; s5 carries a coded
// position for every party but the user skipped it, so it must be excluded.
//
//   user answers: s1=2, s2=2, s3=1, s4=1, s5=vet-ikke
//
// Hand-computed statement matches:
//   s1: 2 vs A(2)=100, B(2)=100, C(null)      -> avg 100, n=2, documented 2
//   s2: 2 vs A(-2)=0,  B(2)=100, C(0)=50      -> avg 50,  n=3, documented 3
//   s3: 1 vs A(0)=75                           -> avg 75,  n=1, documented 1
//   s4: 1 vs all null                          -> avg null, n=0, documented 0
//   s5: skipped                                -> excluded entirely
//
// Aggregate over all ids (the aggregate skips non-positions and nulls):
//   A: 100 + 0 + 75 = 175, n=3 -> round(58.33) = 58
//   B: 100 + 100     = 200, n=2 -> 100
//   C: 50            =  50, n=1 -> 50
// Ranking: B(100), A(58), C(50). Top = B.
var data = {
  snapshot_date: "2026-10-12",
  parties: [
    { id: "a", name: "Parti A", short: "A" },
    { id: "b", name: "Parti B", short: "B" },
    { id: "c", name: "Parti C", short: "C" },
  ],
  statements: [
    { id: "s1", number: 1, dimension: "D1", short_name: "Ett" },
    { id: "s2", number: 2, dimension: "D1", short_name: "To" },
    { id: "s3", number: 3, dimension: "D2", short_name: "Tre" },
    { id: "s4", number: 4, dimension: "D2", short_name: "Fire" },
    { id: "s5", number: 5, dimension: "D2", short_name: "Fem" },
  ],
  cells: {
    s1: { a: { party_position: 2 }, b: { party_position: 2 }, c: { party_position: null } },
    s2: {
      a: { party_position: -2 },
      b: { party_position: 2 },
      c: {
        party_position: 0,
        konsistens: "Motsigelse: to sammenlignbare handlinger peker direkte motsatt",
      },
    },
    s3: { a: { party_position: 0 }, b: { party_position: null }, c: { party_position: null } },
    s4: { a: { party_position: null }, b: { party_position: null }, c: { party_position: null } },
    s5: { a: { party_position: 1 }, b: { party_position: 1 }, c: { party_position: 1 } },
  },
};
var answers = { s1: 2, s2: 2, s3: 1, s4: 1, s5: "vet-ikke" };

function st(id) {
  return data.statements.filter(function (s) { return s.id === id; })[0];
}
function byId(rows) {
  var m = {};
  rows.forEach(function (r) { m[r.partyId] = r; });
  return m;
}

// --- 1. Per-statement arithmetic. ------------------------------------
check("statementMatch s1 avg", A.statementMatch(data, answers, st("s1")).avg, 100);
check("statementMatch s1 n", A.statementMatch(data, answers, st("s1")).n, 2);
check("statementMatch s2 avg", A.statementMatch(data, answers, st("s2")).avg, 50);
check("statementMatch s2 documented", A.statementMatch(data, answers, st("s2")).documented, 3);
check("statementMatch s3 avg", A.statementMatch(data, answers, st("s3")).avg, 75);
check("statementMatch s4 avg (all undocumented)", A.statementMatch(data, answers, st("s4")).avg, null);
check("statementMatch s4 n", A.statementMatch(data, answers, st("s4")).n, 0);
// s5 has coded positions for everyone, but the user skipped it: excluded.
check("statementMatch s5 avg (skipped)", A.statementMatch(data, answers, st("s5")).avg, null);
check("statementMatch s5 n (skipped)", A.statementMatch(data, answers, st("s5")).n, 0);

// --- 2. Aggregate and ranking. ---------------------------------------
var agg = byId(A.aggregate(data, answers, ["s1", "s2", "s3", "s4", "s5"]));
check("aggregate A avg", agg.a.avg, 58);
check("aggregate A n", agg.a.n, 3);
check("aggregate B avg", agg.b.avg, 100);
check("aggregate B n", agg.b.n, 2);
check("aggregate C avg", agg.c.avg, 50);
check("aggregate C n", agg.c.n, 1);
check("top party", A.topPartyId(A.aggregate(data, answers, ["s1", "s2", "s3"])), "b");

// --- 2b. Importance weighting. ---------------------------------------
// Weights scale a statement's pull on the aggregate and nothing else.
// A: s1 100, s2 0, s3 75. Unweighted mean = 58. Weighting s2 (the one A
// disagrees on) by 3 pulls A down: (100*1 + 0*3 + 75*1) / (1+3+1) = 35.
var wAgg = byId(A.aggregate(data, answers, ["s1", "s2", "s3", "s4", "s5"], { s2: 3 }));
check("weighted aggregate A avg", wAgg.a.avg, 35);
check("weighted aggregate A n (pairs unchanged)", wAgg.a.n, 3);
check("weighted aggregate A totalWeight", wAgg.a.totalWeight, 5);
// B agrees on s1 and s2, so weighting s2 leaves B at 100.
check("weighted aggregate B avg", wAgg.b.avg, 100);
// Omitting weights must reproduce the plain mean exactly.
var unw = byId(A.aggregate(data, answers, ["s1", "s2", "s3"]));
var explicitOnes = byId(A.aggregate(data, answers, ["s1", "s2", "s3"], { s1: 1, s2: 1, s3: 1 }));
check("no weights == all weights 1", explicitOnes.a.avg, unw.a.avg);
// A weight can never be 0: that would drop a statement the user DID answer,
// which is what SKIP is for. weightOf clamps it to 1.
var zeroed = byId(A.aggregate(data, answers, ["s1", "s2", "s3"], { s2: 0 }));
check("weight 0 is treated as 1, not as a drop", zeroed.a.avg, unw.a.avg);
check("weight 0 leaves the pair count alone", zeroed.a.n, unw.a.n);
// A weight on a SKIPPED statement changes nothing: it is still excluded.
var skipWeighted = byId(A.aggregate(data, answers, ["s5"], { s5: 3 }));
check("weighting a skipped statement does not resurrect it", skipWeighted.a.avg, null);
check("weighting a skipped statement keeps n at 0", skipWeighted.a.n, 0);
// Out-of-range weights clamp, so a tampered localStorage cannot dominate.
var tampered = byId(A.aggregate(data, answers, ["s1", "s2", "s3"], { s2: 9999 }));
var clamped = byId(A.aggregate(data, answers, ["s1", "s2", "s3"], { s2: 3 }));
check("an absurd weight clamps to MAX_WEIGHT", tampered.a.avg, clamped.a.avg);
// weightedIds names only what the visitor actually raised.
var raised = A.weightedIds(data, { s2: 3, s1: 1 });
check("weightedIds counts only weights above 1", raised.length, 1);
check("weightedIds names the right statement", raised[0].id, "s2");
check("weightedIds is empty when nothing was raised", A.weightedIds(data, {}).length, 0);

// ranked: best first, nulls last.
check(
  "ranked order",
  A.ranked([
    { partyId: "x", avg: null },
    { partyId: "y", avg: 50 },
    { partyId: "z", avg: 90 },
  ]).map(function (r) { return r.partyId; }),
  ["z", "y", "x"]
);

// --- 3. Contribution ranking (Q1). -----------------------------------
var contrib = A.contribution(data, answers);
check("contribution order", contrib.map(function (r) { return r.statement.id; }), ["s1", "s3", "s2"]);
check("contribution top avg", contrib[0].avg, 100);
check("contribution bottom avg", contrib[contrib.length - 1].avg, 50);

// --- 4. Documentation gaps (Q2). -------------------------------------
var gaps = A.documentationGaps(data, answers);
check("gaps statement order", gaps.map(function (g) { return g.statement.id; }), ["s1", "s2", "s3", "s4"]);
var gapsById = {};
gaps.forEach(function (g) { gapsById[g.statement.id] = g; });
check("gap s1 documented", gapsById.s1.documented, 2);
check("gap s1 undocumented count", gapsById.s1.undocumented.length, 1);
check("gap s2 flagged count", gapsById.s2.flags.length, 1);
check("gap s2 flag kind", gapsById.s2.flags[0].flag, "motstridende");
check("gap s4 documented (fully undocumented)", gapsById.s4.documented, 0);
// s5 is skipped, so it is not a gap: it never entered a match at all.
check("gap list excludes skipped s5", gapsById.s5 === undefined, true);

// docFlag classification.
check("docFlag contradiction", A.docFlag({ konsistens: "Motsigelse: x" }), "motstridende");
check("docFlag tension", A.docFlag({ konsistens: "Spenning: x" }), "spenning");
check("docFlag changed", A.docFlag({ konsistens: "Endret standpunkt (eksplisitt)" }), "endret standpunkt");
check("docFlag consistent (not a flag)", A.docFlag({ konsistens: "Konsistent: x" }), null);
check("docFlag no data (not a flag)", A.docFlag({ konsistens: "Ikke nok data til å vurdere utvikling over tid" }), null);

// --- 5. Equal theme weight (Q3). -------------------------------------
// Contributing statements are s1,s2,s3 -> D1={s1,s2}, D2={s3}.
//   A: D1 mean (100,0)=50, D2 (75)=75 -> (50+75)/2 = 62.5 -> 63
//   B: D1 (100,100)=100, D2 none        -> 100
//   C: D1 (50)=50, D2 none              -> 50
var eq = byId(A.themeEqualWeight(data, answers));
check("equalWeight A", eq.a.avg, 63);
check("equalWeight A themes", eq.a.themes, 2);
check("equalWeight B", eq.b.avg, 100);
check("equalWeight C", eq.c.avg, 50);
check("equalWeight top unchanged", A.topPartyId(A.themeEqualWeight(data, answers)), "b");

// --- 6. Stability (Q4). ----------------------------------------------
// Leave-one-out over contributing statements s1,s2,s3: B stays top in all 3.
var stab = A.stability(data, answers, ["s1", "s2", "s3"]);
check("stability baseTop", stab.baseTop, "b");
check("stability unchanged", stab.unchanged, 3);
check("stability total", stab.total, 3);
check("stability flips", stab.flips, {});
// Leave-one-out over themes D1,D2: removing D1 makes A top (a flip).
var stabT = A.themeStability(data, answers, ["s1", "s2", "s3"]);
check("themeStability baseTop", stabT.baseTop, "b");
check("themeStability unchanged", stabT.unchanged, 1);
check("themeStability total", stabT.total, 2);
check("themeStability flips", stabT.flips, { a: 1 });

// --- 7. The all-undocumented case. -----------------------------------
var allNull = {
  snapshot_date: "2026-10-12",
  parties: data.parties,
  statements: [
    { id: "u1", number: 1, dimension: "D1" },
    { id: "u2", number: 2, dimension: "D1" },
  ],
  cells: {
    u1: { a: { party_position: null }, b: { party_position: null }, c: { party_position: null } },
    u2: { a: { party_position: null }, b: { party_position: null }, c: { party_position: null } },
  },
};
var ansAll = { u1: 2, u2: -1 };
check("all-undocumented contributingIds", A.contributingIds(allNull, ansAll), []);
check("all-undocumented top party", A.topPartyId(A.aggregate(allNull, ansAll, ["u1", "u2"])), null);
check("all-undocumented contribution", A.contribution(allNull, ansAll), []);
check("all-undocumented equalWeight top", A.topPartyId(A.themeEqualWeight(allNull, ansAll)), null);
var stNull = A.stability(allNull, ansAll, []);
check("all-undocumented stability baseTop", stNull.baseTop, null);

// --- 8. Snapshot mismatch empties the answer set (the stated contract). --
var realLS = global.localStorage;
global.localStorage = {
  getItem: function () {
    return JSON.stringify({ snapshot_date: "1999-01-01", answers: { s1: 2 } });
  },
};
check("loadAnswers on snapshot mismatch", A.loadAnswers(data), {});
global.localStorage = {
  getItem: function () {
    return JSON.stringify({ snapshot_date: "2026-10-12", answers: { s1: 2 } });
  },
};
check("loadAnswers on snapshot match", A.loadAnswers(data), { s1: 2 });
global.localStorage = { getItem: function () { return null; } };
check("loadAnswers with nothing stored", A.loadAnswers(data), {});
global.localStorage = realLS;

// --- 9. Bar step helper reuses the site ladder. ----------------------
check("pctStepClass(82)", A.pctStepClass(82), "pct-80");
check("pctStepClass(100)", A.pctStepClass(100), "pct-100");
check("pctStepClass(0)", A.pctStepClass(0), "pct-0");
check("pctStepClass(-5) clamps", A.pctStepClass(-5), "pct-0");
check("pctStepClass(120) clamps", A.pctStepClass(120), "pct-100");

// --- 10. Skip is never neutral: score uses scoring.js's contract. -----
check("skip excluded means n differs", S.pct("vet-ikke", 2), null);

// --- 11. The panel's baseline is the result the page shows. -----------
// Reported live on 2026-10-13: with weights set, the bar chart put one
// party on top and the "Hvor stodig er rangeringen" line named another,
// because stability() called aggregate() without the visitor's weights.
// Two top parties on one page, on the axis the page is about.
//
// This fixture is built so weighting FLIPS the top party, which the main
// fixture above cannot do (party B is at 100 on every statement there).
// Three parties, three statements, all documented, visitor answers 2 to all:
//   s1: A 100, B 0,   C 0
//   s2: A 0,   B 100, C 100
//   s3: A 0,   B 100, C 100
// Unweighted: A 33, B 67, C 67  -> top B.
// With s1 (A's strong statement) at 3: A 60, B 40, C 40 -> top A.
var flip = {
  snapshot_date: "2026-10-12",
  parties: [
    { id: "a", name: "Parti A", short: "A" },
    { id: "b", name: "Parti B", short: "B" },
    { id: "c", name: "Parti C", short: "C" },
  ],
  statements: [
    { id: "t1", number: 1, dimension: "D1", short_name: "Ett" },
    { id: "t2", number: 2, dimension: "D1", short_name: "To" },
    { id: "t3", number: 3, dimension: "D1", short_name: "Tre" },
  ],
  cells: {
    t1: { a: { party_position: 2 }, b: { party_position: -2 }, c: { party_position: -2 } },
    t2: { a: { party_position: -2 }, b: { party_position: 2 }, c: { party_position: 2 } },
    t3: { a: { party_position: -2 }, b: { party_position: 2 }, c: { party_position: 2 } },
  },
};
var flipAnswers = { t1: 2, t2: 2, t3: 2 };
var flipIds = ["t1", "t2", "t3"];
var flipWeights = { t1: 3 };

check("flip: unweighted top is B", A.topPartyId(A.aggregate(flip, flipAnswers, flipIds)), "b");
check("flip: weighted top is A", A.topPartyId(A.aggregate(flip, flipAnswers, flipIds, flipWeights)), "a");
// The panel's stability baseline must be the same top party as the bars.
check("flip: stability baseTop without weights is B",
  A.stability(flip, flipAnswers, flipIds).baseTop, "b");
check("flip: stability baseTop WITH weights is A (the bug was B here)",
  A.stability(flip, flipAnswers, flipIds, flipWeights).baseTop, "a");
check("flip: themeStability baseTop WITH weights is A",
  A.themeStability(flip, flipAnswers, flipIds, flipWeights).baseTop, "a");
// And the panel's baseline agrees with the aggregate it is commenting on.
check("flip: stability baseTop equals the weighted aggregate's top",
  A.stability(flip, flipAnswers, flipIds, flipWeights).baseTop,
  A.topPartyId(A.aggregate(flip, flipAnswers, flipIds, flipWeights)));
// Leave-one-out must carry the weights too, or the flips it reports are
// computed against a different result from the one it names.
check("flip: leave-one-out holds under weights",
  A.stability(flip, flipAnswers, flipIds, flipWeights).unchanged >= 1, true);

if (failures > 0) {
  console.error("\n" + failures + " check(s) failed");
  process.exit(1);
}
console.log("\nall checks passed");
