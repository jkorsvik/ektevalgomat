// Node check for the persistence rules in web/static/js/store.js.
//
//   node tools/store.test.js
//
// It loads the REAL shipped module. Storage is a fake with switches, and two
// "tabs" are just two in-memory states over one backing object, which is
// exactly what the browser gives two open tabs of the same site.
//
// Each block below reproduces one way a visitor lost progress on the live
// site before 2026-10-16 (measured in a browser, see the PR), so that the
// fix is a tested property rather than a hope.

"use strict";

var E = require("../web/static/js/store.js");

var failures = 0;
// canon makes object comparison independent of key insertion order, which a
// merge is free to change and which means nothing here.
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

function fakeStorage() {
  var m = {};
  var s = {
    failWrite: false,
    failRead: false,
    getItem: function (k) {
      if (s.failRead) throw new Error("denied");
      return Object.prototype.hasOwnProperty.call(m, k) ? m[k] : null;
    },
    setItem: function (k, v) {
      if (s.failWrite) throw new Error("QuotaExceededError");
      m[k] = String(v);
    },
    removeItem: function (k) { delete m[k]; },
    raw: function (k) { return m[k]; },
  };
  return s;
}

function dataset(snapshot, texts) {
  var base = { s1: "Ett", s2: "To", s3: "Tre" };
  var t = Object.assign({}, base, texts || {});
  return {
    snapshot_date: snapshot,
    parties: [],
    cells: {},
    statements: Object.keys(t).map(function (id, i) { return { id: id, number: i + 1, text: t[id] }; }),
  };
}

var KEY = "k";
var D = dataset("2026-10-12");

// --- 1. fingerprint ----------------------------------------------------
check("fingerprint is stable", E.fingerprint("abc"), E.fingerprint("abc"));
check("fingerprint tells wordings apart", E.fingerprint("Ett") !== E.fingerprint("ett"), true);
check("fingerprint of the empty string is defined", typeof E.fingerprint(""), "string");
check("fingerprint matches the published FNV-1a vector for \"a\" (0xE40C292C)", E.fingerprint("a"), "1r9wi7g");

// --- 2. read: nothing to lose is not an error --------------------------
var st = fakeStorage();
var r = E.read(st, KEY, "answers", D);
check("read empty: ok and empty", [r.ok, r.map, r.kept, r.dropped], [true, {}, 0, 0]);
st.setItem(KEY, "{not json");
r = E.read(st, KEY, "answers", D);
check("read corrupt: empty, flagged, no throw", [r.ok, r.map, r.corrupt], [true, {}, true]);
st.setItem(KEY, JSON.stringify([1, 2]));
check("read non-object blob is corrupt", E.read(st, KEY, "answers", D).corrupt, true);
st.failRead = true;
check("read when storage throws: ok false", E.read(st, KEY, "answers", D).ok, false);
check("read with no storage: ok false", E.read(null, KEY, "answers", D).ok, false);

// --- 3. a refresh keeps what did not change ----------------------------
// Loss 3 on the old site: snapshot_date moved, every answer was ignored and
// the first new answer overwrote them. Now only reworded statements go.
var old = fakeStorage();
var oldData = dataset("2026-09-23");
var w = E.write(old, KEY, "answers", oldData, { s1: 2, s2: 1, s3: -1 }, {}, { direction_version: 2 });
check("write under the old snapshot succeeded", w.ok, true);
var fresh = dataset("2026-10-12", { s2: "To, omformulert" }); // s2 reworded, snapshot moved
r = E.read(old, KEY, "answers", fresh);
check("refresh keeps unchanged statements", r.map, { s1: 2, s3: -1 });
check("refresh drops the reworded one and counts it", [r.kept, r.dropped], [2, 1]);
var gone = { snapshot_date: "2026-10-12", parties: [], cells: {}, statements: [{ id: "s1", number: 1, text: "Ett" }] };
r = E.read(old, KEY, "answers", gone);
check("a statement that no longer exists is dropped", [r.map, r.dropped], [{ s1: 2 }, 2]);
r = E.read(old, KEY, "answers", dataset("2026-10-12"));
check("refresh with no wording change keeps everything", [r.kept, r.dropped], [3, 0]);

// --- 4. legacy blobs (no fingerprints) keep the old rule ---------------
var legacy = fakeStorage();
legacy.setItem(KEY, JSON.stringify({ snapshot_date: "2026-10-12", direction_version: 2, answers: { s1: 2, s2: 1 } }));
r = E.read(legacy, KEY, "answers", D);
check("legacy blob, same snapshot: kept", [r.map, r.dropped], [{ s1: 2, s2: 1 }, 0]);
r = E.read(legacy, KEY, "answers", dataset("2026-11-01"));
check("legacy blob, new snapshot: cannot be verified, dropped and counted", [r.map, r.kept, r.dropped], [{}, 0, 2]);
// The existence rule must stand on its own. A legacy blob has no fingerprint
// to fail, so "the statement is gone" is the ONLY thing that can drop it.
legacy.setItem(KEY, JSON.stringify({ snapshot_date: "2026-10-12", answers: { s1: 2, deleted: 1 } }));
r = E.read(legacy, KEY, "answers", D);
check("legacy blob, same snapshot: an answer to a deleted statement is dropped", [r.map, r.kept, r.dropped], [{ s1: 2 }, 1, 1]);
legacy.setItem(KEY, JSON.stringify({ snapshot_date: "2026-10-12", answers: { s1: { evil: true }, s2: 1 } }));
check("a value that is not a number or string is dropped", E.read(legacy, KEY, "answers", D).map, { s2: 1 });

// --- 4b. a legacy blob is upgraded the first time it is seen -----------
// Every visitor who answered before this release holds a legacy blob. If it
// only gained fingerprints on the next WRITE, someone who returned just to
// look would lose everything at the next data refresh.
var up = fakeStorage();
up.setItem(KEY, JSON.stringify({ snapshot_date: "2026-10-12", direction_version: 2, answers: { s1: 2, s2: 1 } }));
r = E.read(up, KEY, "answers", D);
check("legacy blob is flagged as legacy", r.legacy, true);
E.write(up, KEY, "answers", D, r.map, Object.assign({}, r.map), { direction_version: 2 });
var upBlob = JSON.parse(up.raw(KEY));
check("a write with no changes still stamps fingerprints", upBlob.texts, { s1: E.fingerprint("Ett"), s2: E.fingerprint("To") });
check("the upgrade changed none of the answers", upBlob.answers, { s1: 2, s2: 1 });
check("an upgraded blob is no longer legacy", E.read(up, KEY, "answers", D).legacy, false);
check("an upgraded blob now survives a data refresh", E.read(up, KEY, "answers", dataset("2026-11-30")).map, { s1: 2, s2: 1 });
check("an empty store is not legacy", E.read(fakeStorage(), KEY, "answers", D).legacy, false);

// --- 5. two tabs never erase each other --------------------------------
// Loss 1 on the old site, measured live: tab A saved utsagn-01, stale tab B
// saved utsagn-02, and only utsagn-02 remained.
var shared = fakeStorage();
var tabA = { answers: {}, base: {} };
var tabB = { answers: {}, base: {} }; // both loaded while storage was empty
function save(tab, extra) {
  var res = E.write(shared, KEY, "answers", D, tab.answers, tab.base, extra);
  if (res.ok) { E.replaceInPlace(tab.answers, res.map); tab.base = Object.assign({}, res.map); }
  return res;
}
tabA.answers.s1 = 2; save(tabA);
tabB.answers.s2 = -1; save(tabB);
check("two tabs: both answers survive", E.read(shared, KEY, "answers", D).map, { s1: 2, s2: -1 });
check("the stale tab picked up the other tab's answer", tabB.answers, { s1: 2, s2: -1 });
// A deletion made on purpose is still a change, and a stale tab does not undo it.
delete tabA.answers.s1; save(tabA);
tabB.answers.s3 = 1; save(tabB);
check("a deletion in one tab is not resurrected by a stale tab", E.read(shared, KEY, "answers", D).map, { s2: -1, s3: 1 });
// Two tabs changing the SAME statement: the later write wins, nothing else is lost.
tabA.answers.s2 = 2; save(tabA);
tabB.answers.s2 = -2; save(tabB);
check("same statement in two tabs: last write wins, the rest stays", E.read(shared, KEY, "answers", D).map, { s2: -2, s3: 1 });

// --- 6. a failed write is never reported as saved ----------------------
// Loss 2: the page said "Sterkt uenig" was kept while setItem threw.
var flaky = fakeStorage();
var tab = { answers: { s1: 2 }, base: {} };
flaky.failWrite = true;
var bad = E.write(flaky, KEY, "answers", D, tab.answers, tab.base);
check("failed write reports ok:false", bad.ok, false);
check("failed write stored nothing", flaky.raw(KEY), undefined);
// The change stays dirty because base was not advanced, so it lands later.
flaky.failWrite = false;
tab.answers.s2 = 1;
var good = E.write(flaky, KEY, "answers", D, tab.answers, tab.base);
check("after recovery both the old and the new change persist", [good.ok, good.map], [true, { s1: 2, s2: 1 }]);
check("write with no storage reports ok:false", E.write(null, KEY, "answers", D, {}, {}).ok, false);

// --- 7. what write puts on disk ----------------------------------------
var disk = fakeStorage();
E.write(disk, KEY, "answers", D, { s1: 2, ghost: 1 }, {}, { direction_version: 2 });
var blob = JSON.parse(disk.raw(KEY));
check("blob carries snapshot, direction_version and the answers", [blob.snapshot_date, blob.direction_version, blob.answers], ["2026-10-12", 2, { s1: 2 }]);
check("an id that is not a statement is never written", Object.keys(blob.answers), ["s1"]);
check("blob carries a fingerprint per stored answer", blob.texts, { s1: E.fingerprint("Ett") });
// weights go through the same code under their own key and field
E.write(disk, "w", "weights", D, { s1: 3 }, {});
check("weights use the same rules under their own field", E.read(disk, "w", "weights", D).map, { s1: 3 });

// --- 7b. an entry that depends on TWO statements ------------------------
// A trade-off record decides between a pair, so it is keyed "a|b" and is only
// valid while both statements are unchanged. Same merge, failure and refresh
// rules as an answer, through opts.
var pairOpts = {
  idsOf: function (k) { return k.split("|"); },
  accept: function (v) { return typeof v === "string" && v.charAt(0) === "{"; },
};
var pd = fakeStorage();
var pw = E.write(pd, "r", "avveininger", dataset("2026-09-23"), { "s1|s2": '{"valg":"a"}', "s2|s3": '{"valg":"b"}' }, {}, null, pairOpts);
check("a pair entry is written when both statements exist", [pw.ok, Object.keys(pw.map).sort()], [true, ["s1|s2", "s2|s3"]]);
var pb = JSON.parse(pd.raw("r"));
check("fingerprints are stamped for every statement the entry names", Object.keys(pb.texts).sort(), ["s1", "s2", "s3"]);
check("refresh, nothing reworded: both pair entries kept",
  E.read(pd, "r", "avveininger", dataset("2026-10-12"), pairOpts).kept, 2);
r = E.read(pd, "r", "avveininger", dataset("2026-10-12", { s3: "Tre, omformulert" }), pairOpts);
check("refresh, ONE statement of a pair reworded: that entry is dropped, the other kept", [Object.keys(r.map), r.dropped], [["s1|s2"], 1]);
r = E.read(pd, "r", "avveininger", dataset("2026-10-12", { s1: "Ett, omformulert" }), pairOpts);
check("refresh, the other end of the pair reworded: same", [Object.keys(r.map), r.dropped], [["s2|s3"], 1]);
var prune = fakeStorage();
var pp = E.write(prune, "r", "avveininger", D, { "s1|ghost": '{"valg":"a"}', "s1|s2": '{"valg":"a"}' }, {}, null, pairOpts);
check("a pair naming a statement that does not exist is never written", Object.keys(pp.map), ["s1|s2"]);
prune.setItem("r", JSON.stringify({ snapshot_date: "2026-10-12", avveininger: { "s1|s2": 42, "s2|s3": '{"valg":"a"}' } }));
check("a value of the wrong shape is dropped by opts.accept", Object.keys(E.read(prune, "r", "avveininger", D, pairOpts).map), ["s2|s3"]);
// two tabs deciding DIFFERENT pairs keep both; the same pair is last-write-wins
var ps = fakeStorage();
var pa = { cur: {}, base: {} }, pc = { cur: {}, base: {} };
pa.cur["s1|s2"] = '{"valg":"a"}'; var ra = E.write(ps, "r", "avveininger", D, pa.cur, pa.base, null, pairOpts);
pc.cur["s2|s3"] = '{"valg":"b"}'; var rc = E.write(ps, "r", "avveininger", D, pc.cur, pc.base, null, pairOpts);
check("two tabs deciding different pairs: both decisions survive", Object.keys(E.read(ps, "r", "avveininger", D, pairOpts).map).sort(), ["s1|s2", "s2|s3"]);
check("the default (no opts) still treats a pair key as a missing statement", Object.keys(E.read(ps, "r", "avveininger", D).map), []);

// --- 8. helpers --------------------------------------------------------
var tgt = { a: 1, b: 2 }, ref = tgt;
E.replaceInPlace(tgt, { b: 3, c: 4 });
check("replaceInPlace keeps identity and matches the source", [tgt === ref, tgt], [true, { b: 3, c: 4 }]);
check("available: working storage", E.available(fakeStorage()), true);
var blocked = fakeStorage(); blocked.failWrite = true;
check("available: storage that throws", E.available(blocked), false);
check("available: no storage", E.available(null), false);
check("clear removes the keys", (function () { var s = fakeStorage(); s.setItem("a", "1"); s.setItem("b", "2"); E.clear(s, ["a", "b"]); return [s.raw("a"), s.raw("b")]; })(), [undefined, undefined]);

// --- 9. the sentences the page may say ---------------------------------
check("nothing dropped: nothing said", E.droppedMessage(5, 0), null);
check("some dropped: says how many were kept and dropped",
  E.droppedMessage(40, 3),
  "Siden er oppdatert siden sist. 40 av svarene dine er beholdt. 3 gjaldt utsagn som er endret, og er tatt bort.");
check("one dropped uses the singular",
  E.droppedMessage(1, 1),
  "Siden er oppdatert siden sist. 1 av svarene dine er beholdt. 1 gjaldt et utsagn som er endret, og er tatt bort.");
check("all dropped: does not pretend anything was kept",
  E.droppedMessage(0, 12),
  "Siden er oppdatert siden sist, og svarene dine fra den gangen kunne ikke tas med.");

if (failures > 0) {
  console.error("\n" + failures + " check(s) failed");
  process.exit(1);
}
console.log("\nall checks passed");
