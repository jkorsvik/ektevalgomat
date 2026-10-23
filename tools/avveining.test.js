// Node check for the trade-off logic in web/static/js/avveining.js.
//
//   node tools/avveining.test.js
//
// It loads the REAL shipped module. Fixtures are hand-computed from the
// scoring contract, pct(reply, position) = round(100 - |reply - position| * 25),
// so the correlations below are known exactly, not copied from the code.
//
// Five parties. For a visitor who answers +2 on a statement:
//   "up"   positions  2, 1, 0,-1,-2  ->  match 100, 75, 50, 25,  0
//   "down" positions -2,-1, 0, 1, 2  ->  match   0, 25, 50, 75,100
//   "g"    positions  1, 2, 0,-2,-1  ->  match  75,100, 50,  0, 25
//   "h"    positions -1,-2, 0, 2, 1  ->  match  25,  0, 50,100, 75
//   "flat" positions  0, 0, 0, 0, 0  ->  match  50, 50, 50, 50, 50
// Pearson, by hand (means are all 50):
//   up  vs down = -1       up vs up = +1      up vs flat = undefined
//   up  vs g    = +0.8     up vs h  = -0.8    down vs h  = +0.8

"use strict";

var AV = require("../web/static/js/avveining.js");
var S = require("../web/static/js/scoring.js");

var failures = 0;
function canon(v) {
  if (Array.isArray(v)) return v.map(canon);
  if (v && typeof v === "object") {
    var o = {};
    Object.keys(v).sort().forEach(function (k) { o[k] = canon(v[k]); });
    return o;
  }
  return v;
}
function check(name, got, want) {
  var ok = JSON.stringify(canon(got)) === JSON.stringify(canon(want));
  if (!ok) {
    failures++;
    console.error("FAIL " + name + ": got " + JSON.stringify(got) + ", want " + JSON.stringify(want));
  } else {
    console.log("ok   " + name + " = " + JSON.stringify(got));
  }
}
function near(name, got, want) {
  var ok = typeof got === "number" && Math.abs(got - want) < 1e-9;
  if (!ok) { failures++; console.error("FAIL " + name + ": got " + got + ", want " + want); }
  else console.log("ok   " + name + " = " + got);
}

var PATTERNS = {
  up: [2, 1, 0, -1, -2],
  down: [-2, -1, 0, 1, 2],
  g: [1, 2, 0, -2, -1],
  h: [-1, -2, 0, 2, 1],
  flat: [0, 0, 0, 0, 0],
  three: [2, 1, 0, null, null], // only three parties coded
};
var PARTIES = ["p1", "p2", "p3", "p4", "p5"].map(function (id) { return { id: id, name: id, short: id }; });
function dataset(spec) {
  var d = { snapshot_date: "2026-10-12", parties: PARTIES, statements: [], cells: {} };
  Object.keys(spec).forEach(function (id, i) {
    d.statements.push({ id: id, number: i + 1, text: "Utsagn " + id, short_name: id });
    d.cells[id] = {};
    PARTIES.forEach(function (p, j) { d.cells[id][p.id] = { party_position: PATTERNS[spec[id]][j] }; });
  });
  return d;
}
function plus2(data) {
  var a = {};
  data.statements.forEach(function (st) { a[st.id] = 2; });
  return a;
}

// --- 1. judging a pair --------------------------------------------------
var D1 = dataset({ A: "up", B: "down", C: "up", D: "flat", E: "three", G: "g", H: "h" });
var ans = plus2(D1);
var j = AV.judgePair(D1, ans, "A", "B");
check("up vs down is a conflict", [j.verdict, j.shared], ["conflict", 5]);
near("up vs down: r = -1", j.r, -1);
j = AV.judgePair(D1, ans, "A", "C");
check("up vs up is no conflict", j.verdict, "no-conflict");
near("up vs up: r = +1", j.r, 1);
j = AV.judgePair(D1, ans, "A", "D");
check("a statement every party answers alike has no pull: no conflict, r undefined", [j.verdict, j.r], ["no-conflict", null]);
j = AV.judgePair(D1, ans, "A", "E");
check("3 shared parties is unknown, NEVER no-conflict", [j.verdict, j.shared], ["unknown", 3]);
j = AV.judgePair(D1, ans, "A", "G");
near("up vs g: r = +0.8", j.r, 0.8);
check("up vs g is no conflict", j.verdict, "no-conflict");
j = AV.judgePair(D1, ans, "A", "H");
near("up vs h: r = -0.8", j.r, -0.8);
check("up vs h is a conflict (-0.8 <= -0.3)", j.verdict, "conflict");
check("the thresholds are the documented ones", [AV.MIN_SHARED, AV.CONFLICT_R, AV.WIN, AV.LOSE], [4, -0.3, 3, 2]);
var skipped = Object.assign({}, ans, { B: "vet-ikke" });
check("a skipped statement shares no party, so it is unknown", AV.judgePair(D1, skipped, "A", "B").verdict, "unknown");

// --- 2. who is in a trade-off at all -----------------------------------
var W = { A: 2, B: 3, C: 1, F: 3 };
var D2 = dataset({ A: "up", B: "down", C: "up", F: "down" });
var a2 = plus2(D2); a2.F = "vet-ikke";
check("marked = weight above 1 AND answered with a position", AV.marked(D2, a2, W).map(function (s) { return s.id; }), ["A", "B"]);
var pl = AV.plan(D2, a2, W, []);
check("two marked, one conflicting pair: one pending", [pl.marked, pl.pending.length, pl.judgedPairs, pl.unknownPairs], [2, 1, 1, 0]);
check("fewer than two marked: nothing pending", AV.plan(D2, a2, { A: 3 }, []).pending.length, 0);
check("one marked reports marked=1", AV.plan(D2, a2, { A: 3 }, []).marked, 1);
var D3 = dataset({ A: "up", E: "three" });
pl = AV.plan(D3, plus2(D3), { A: 3, E: 2 }, []);
check("a pair that cannot be judged is counted as unknown, not as no-conflict", [pl.pending.length, pl.judgedPairs, pl.unknownPairs], [0, 0, 1]);
var D3b = dataset({ A: "up", C: "up" });
pl = AV.plan(D3b, plus2(D3b), { A: 3, C: 3 }, []);
check("judged and agreeing: nothing pending, judged=1", [pl.pending.length, pl.judgedPairs], [0, 1]);

// --- 3. pending pairs are disjoint, strongest first --------------------
var D4 = dataset({ A: "up", B: "down", A2: "up", B2: "down" });
var w4 = { A: 3, B: 3, A2: 3, B2: 3 };
pl = AV.plan(D4, plus2(D4), w4, []);
var seen = {}, dup = 0;
pl.pending.forEach(function (p) { [p.a.id, p.b.id].forEach(function (id) { if (seen[id]) dup++; seen[id] = true; }); });
check("four marked statements, four conflicting pairs: two pending, no statement twice", [pl.pending.length, dup], [2, 0]);
check("the first pending pair is the strongest available", pl.pending[0].a.id + "+" + pl.pending[0].b.id, "A+B");
var D5 = dataset({ A: "up", B: "down", A2: "up", H: "h" });
pl = AV.plan(D5, plus2(D5), { A: 3, B: 3, A2: 3, H: 3 }, []);
check("strongest pull first: r=-1 before r=-0.8", pl.pending.length === 2 && pl.pending[0].judged.r < pl.pending[1].judged.r, true);
near("   first r", pl.pending[0].judged.r, -1);
near("   second r", pl.pending[1].judged.r, -0.8);

// --- 4. deciding, and undoing exactly ----------------------------------
var w0 = { A: 2, B: 3 };
var frozen = JSON.stringify(w0);
var d = AV.decide(w0, [], "A", "B", "a", 100);
check("choosing A: A counts 3, B counts 2", d.weights, { A: 3, B: 2 });
check("the record holds the weights from BEFORE", d.records[0], { a: "A", b: "B", valg: "a", foer: { a: 2, b: 3 }, at: 100 });
check("decide does not mutate its inputs", JSON.stringify(w0), frozen);
d = AV.decide(w0, [], "A", "B", "b", 100);
check("choosing B: B counts 3, A counts 2", d.weights, { A: 2, B: 3 });
d = AV.decide(w0, [], "A", "B", "lik", 100);
check("'like mye' changes no weight", d.weights, { A: 2, B: 3 });
check("'like mye' still records the decision", d.records.length, 1);

// undo restores EXACTLY the earlier weights, including a key that was absent
var absent = { A: 2 }; // B is absent, meaning 1
d = AV.decide(absent, [], "A", "B", "a", 1);
check("with B absent: after choosing A", d.weights, { A: 3, B: 2 });
var u = AV.undo(d.weights, d.records, 0);
check("undo restores the exact input, B absent again (not stored as 1)", u.weights, { A: 2 });
check("undo removes the record", u.records, []);
check("undo of an index that is not there changes nothing", AV.undo(d.weights, d.records, 9).weights, d.weights);
check("undo then decide again is clean", AV.decide(u.weights, u.records, "A", "B", "b", 2).weights, { A: 2, B: 3 });

// a "lik" record settles its pair
var D6 = dataset({ A: "up", B: "down", A2: "up", B2: "down" });
var lik = AV.decide({ A: 3, B: 3, A2: 3, B2: 3 }, [], "A", "B", "lik", 1);
pl = AV.plan(D6, plus2(D6), lik.weights, lik.records);
check("a settled pair is not asked again; the other pair still is", pl.pending.map(function (p) { return p.a.id + "+" + p.b.id; }), ["A2+B2"]);

// a manual change ends the trade-off
var two = AV.decide({ A: 3, B: 3, C: 3, E: 3 }, [], "A", "B", "a", 1);
two = AV.decide(two.weights, two.records, "C", "E", "b", 2);
check("changing A by hand removes only the A/B record", AV.dropOnManualChange(two.records, "A").map(function (r) { return r.a + r.b; }), ["CE"]);
check("a statement in no record removes nothing", AV.dropOnManualChange(two.records, "Z").length, 2);

// the effect line compares against the weights without ANY trade-off
check("weightsWithout rewinds every record", AV.weightsWithout(two.weights, two.records), { A: 3, B: 3, C: 3, E: 3 });
var seq = AV.decide({ A: 2, B: 3, C: 3, D: 3 }, [], "A", "B", "a", 1);
seq = AV.decide(seq.weights, seq.records, "C", "D", "b", 2);
check("   ...two records, from distinct starting weights", AV.weightsWithout(seq.weights, seq.records), { A: 2, B: 3, C: 3, D: 3 });

// --- 5. a record is only believed while its weights are in force -------
var rec = AV.decide({ A: 3, B: 3 }, [], "A", "B", "a", 1);
check("consistent: weights match the choice", AV.consistent(rec.weights, rec.records).length, 1);
check("inconsistent: A was changed afterwards", AV.consistent({ A: 2, B: 2 }, rec.records).length, 0);
check("inconsistent: the loser was raised", AV.consistent({ A: 3, B: 3 }, rec.records).length, 0);
var recB = AV.decide({ A: 3, B: 3 }, [], "A", "B", "b", 1);
check("consistent for the other winner", AV.consistent(recB.weights, recB.records).length, 1);
check("a 'b' record is inconsistent with an 'a' outcome", AV.consistent({ A: 3, B: 2 }, recB.records).length, 0);
var recL = AV.decide({ A: 2, B: 3 }, [], "A", "B", "lik", 1);
check("'like mye' is consistent while the weights are as they were", AV.consistent({ A: 2, B: 3 }, recL.records).length, 1);
check("'like mye' is inconsistent once either weight differs", AV.consistent({ A: 3, B: 3 }, recL.records).length, 0);
// two tabs deciding overlapping pairs: the later outcome wins, the earlier record goes
var o1 = AV.decide({ A: 3, B: 3, C: 3 }, [], "A", "B", "a", 1);
var o2 = AV.decide(o1.weights, o1.records, "B", "C", "a", 2);
check("overlapping records: only the one the weights still explain survives",
  AV.consistent(o2.weights, o2.records).map(function (r) { return r.a + r.b; }), ["BC"]);

// --- 6. the stored form -----------------------------------------------
var recs = AV.decide({ A: 3, B: 3, C: 3, E: 3 }, [], "A", "B", "a", 20);
recs = AV.decide(recs.weights, recs.records, "C", "E", "lik", 10);
var map = AV.toMap(recs.records);
check("a record is stored under the pair it decides between", Object.keys(map).sort(), ["A|B", "C|E"]);
check("the stored value is a string", typeof map["A|B"], "string");
check("the store's idsOf names both statements", AV.storeOpts.idsOf("A|B"), ["A", "B"]);
check("round trip, ordered by time not by key (C|E at 10 before A|B at 20)", AV.fromMap(map).map(function (r) { return r.a + r.b + "@" + r.at; }), ["CE@10", "AB@20"]);
check("round trip keeps the weights from before", AV.fromMap(map)[1].foer, { a: 3, b: 3 });
check("ties in time are broken by key, so the order is stable", AV.fromMap(AV.toMap([
  { a: "B", b: "C", valg: "a", foer: { a: 3, b: 3 }, at: 5 }, { a: "A", b: "D", valg: "a", foer: { a: 3, b: 3 }, at: 5 }])).map(function (r) { return r.a; }), ["A", "B"]);
var accept = AV.storeOpts.accept;
check("accept: a good record", accept('{"valg":"a","foer":{"a":3,"b":2},"at":1}'), true);
check("accept: not a string", accept({ valg: "a" }), false);
check("accept: not JSON", accept("{nope"), false);
check("accept: unknown choice", accept('{"valg":"x","foer":{"a":3,"b":2}}'), false);
check("accept: a weight of 0 is not a weight", accept('{"valg":"a","foer":{"a":0,"b":2}}'), false);
check("accept: a weight of 4 is out of range", accept('{"valg":"a","foer":{"a":4,"b":2}}'), false);
check("accept: a fractional weight", accept('{"valg":"a","foer":{"a":2.5,"b":2}}'), false);
check("accept: missing foer", accept('{"valg":"a"}'), false);
check("fromMap skips a malformed entry instead of throwing", AV.fromMap({ "A|B": "garbage", "C|E": map["C|E"] }).length, 1);
check("fromMap skips a key that is not a pair", AV.fromMap({ "A": map["A|B"] }).length, 0);

// --- 7. naming parties in a sentence ----------------------------------
check("namesList: none", AV.namesList([]), "");
check("namesList: one", AV.namesList(["Sp"]), "Sp");
check("namesList: two", AV.namesList(["Sp", "Venstre"]), "Sp og Venstre");
check("namesList: three", AV.namesList(["Sp", "Venstre", "Rødt"]), "Sp, Venstre og Rødt");

if (failures > 0) {
  console.error("\n" + failures + " check(s) failed");
  process.exit(1);
}
console.log("\nall checks passed");
