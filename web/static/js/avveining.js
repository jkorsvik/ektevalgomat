// The trade-off ("avveining") logic, as pure functions: no DOM, no storage,
// no network. quiz.js calls in; nothing here reaches out. Design:
// docs/design/2026-10-13-avveining.md.
//
// The model in one paragraph. A trade-off is a recorded decision between two
// statements the visitor marked as important, where their answers pull the
// party match in opposite directions. The decision reaches the score through
// ONE channel only, the existing weights: the chosen statement is set to 3,
// the other to 2. pct() and the weighted mean are unchanged, so a trade-off
// can never make the score mean something different. The record exists so
// the decision can be listed, explained and undone exactly: it holds the two
// weights the visitor had set BEFORE the trade-off.
(function (root, factory) {
  if (typeof module === "object" && module.exports) {
    module.exports = factory(require("./scoring.js"));
  } else {
    root.EkteAvveining = factory(root.EkteScoring);
  }
})(typeof self !== "undefined" ? self : this, function (S) {
  "use strict";

  // A pair is only judged when this many parties have a coded position on
  // BOTH statements. Below it, a correlation is noise, and the honest answer
  // is "we cannot tell", never "they agree".
  var MIN_SHARED = 4;
  // Pull-apart threshold on the Pearson correlation of per-party match.
  // Open question 1 in the design doc; -0.3 is the recommendation. It flags
  // 36 % of judgeable pairs for a visitor who answers "sterkt enig" to all.
  var CONFLICT_R = -0.3;
  var WIN = 3, LOSE = 2;

  function posOf(data, id, partyId) {
    var c = (data.cells[id] || {})[partyId];
    if (!c) return null;
    var p = c.party_position;
    return p === null || p === undefined ? null : p;
  }

  // pairEvidence: per-party match on each statement, over the parties coded
  // on both. This is what the "Hvorfor spør vi om dette?" disclosure shows.
  function pairEvidence(data, answers, a, b) {
    var rows = [];
    data.parties.forEach(function (p) {
      var ma = S.pct(answers[a], posOf(data, a, p.id));
      var mb = S.pct(answers[b], posOf(data, b, p.id));
      if (ma === null || mb === null) return;
      rows.push({ party: p, a: ma, b: mb });
    });
    return rows;
  }

  function pearson(xs, ys) {
    var n = xs.length, mx = 0, my = 0, i;
    for (i = 0; i < n; i++) { mx += xs[i]; my += ys[i]; }
    mx /= n; my /= n;
    var sxy = 0, sxx = 0, syy = 0;
    for (i = 0; i < n; i++) {
      sxy += (xs[i] - mx) * (ys[i] - my);
      sxx += (xs[i] - mx) * (xs[i] - mx);
      syy += (ys[i] - my) * (ys[i] - my);
    }
    if (sxx === 0 || syy === 0) return null; // every party the same: no pull
    return sxy / Math.sqrt(sxx * syy);
  }

  // judgePair: "conflict", "no-conflict" or "unknown" (too few shared parties).
  function judgePair(data, answers, a, b) {
    var ev = pairEvidence(data, answers, a, b);
    if (ev.length < MIN_SHARED) return { verdict: "unknown", shared: ev.length, r: null, evidence: ev };
    var r = pearson(ev.map(function (e) { return e.a; }), ev.map(function (e) { return e.b; }));
    if (r === null) return { verdict: "no-conflict", shared: ev.length, r: null, evidence: ev };
    return { verdict: r <= CONFLICT_R ? "conflict" : "no-conflict", shared: ev.length, r: r, evidence: ev };
  }

  // marked: statements weighted above 1 AND answered with a position. A
  // skipped statement is out of the match, so it can be in no trade-off.
  function marked(data, answers, weights) {
    return data.statements.filter(function (st) {
      return S.weightOf(weights, st.id) > 1 && S.isPosition(answers[st.id]);
    });
  }

  function inRecord(records, id) {
    for (var i = 0; i < records.length; i++) {
      if (records[i].a === id || records[i].b === id) return records[i];
    }
    return null;
  }

  // plan: everything the result page needs to decide which state to show.
  // Pending pairs are DISJOINT: a statement sits in at most one trade-off.
  // That is what keeps the weights-only model consistent (A over B, then
  // B over C, would otherwise silently undo the first decision).
  function plan(data, answers, weights, records) {
    var m = marked(data, answers, weights);
    var free = m.filter(function (st) { return !inRecord(records, st.id); });
    var pairs = [], unknown = 0, judged = 0;
    for (var i = 0; i < free.length; i++) {
      for (var j = i + 1; j < free.length; j++) {
        var jp = judgePair(data, answers, free[i].id, free[j].id);
        if (jp.verdict === "unknown") { unknown++; continue; }
        judged++;
        if (jp.verdict === "conflict") pairs.push({ a: free[i], b: free[j], judged: jp });
      }
    }
    pairs.sort(function (x, y) { return x.judged.r - y.judged.r; }); // strongest pull first
    var used = {}, pending = [];
    pairs.forEach(function (p) {
      if (used[p.a.id] || used[p.b.id]) return;
      used[p.a.id] = used[p.b.id] = true;
      pending.push(p);
    });
    return { marked: m.length, pending: pending, unknownPairs: unknown, judgedPairs: judged };
  }

  // decide returns NEW weights and records; inputs are not mutated.
  // choice: "a" | "b" | "lik". at is the time of the decision, passed in so
  // this stays pure; it only orders the log.
  function decide(weights, records, a, b, choice, at) {
    var w = Object.assign({}, weights);
    var rec = { a: a, b: b, valg: choice, foer: { a: S.weightOf(weights, a), b: S.weightOf(weights, b) }, at: at || 0 };
    if (choice === "a") { w[a] = WIN; w[b] = LOSE; }
    if (choice === "b") { w[b] = WIN; w[a] = LOSE; }
    return { weights: w, records: records.concat([rec]) };
  }

  function writeWeight(w, id, v) {
    if (v <= S.MIN_WEIGHT) delete w[id]; else w[id] = v;
  }

  // undo restores exactly the weights the visitor had before.
  function undo(weights, records, index) {
    var rec = records[index];
    if (!rec) return { weights: weights, records: records };
    var w = Object.assign({}, weights);
    writeWeight(w, rec.a, rec.foer.a);
    writeWeight(w, rec.b, rec.foer.b);
    return { weights: w, records: records.filter(function (_, i) { return i !== index; }) };
  }

  // A manual weight change on either statement ends the trade-off: the
  // visitor's new hand-set weight wins, and the record would otherwise
  // explain weights that no longer exist.
  function dropOnManualChange(records, id) {
    return records.filter(function (r) { return r.a !== id && r.b !== id; });
  }

  // weightsWithout: the weights as they would be with no trade-offs at all.
  // Used for the "uten avveiningene" line, so the effect is always visible.
  function weightsWithout(weights, records) {
    var w = Object.assign({}, weights);
    for (var i = records.length - 1; i >= 0; i--) {
      writeWeight(w, records[i].a, records[i].foer.a);
      writeWeight(w, records[i].b, records[i].foer.b);
    }
    return w;
  }

  // consistent keeps only the records whose weights are still in force. A
  // record is an EXPLANATION of the weights, so it may only be believed
  // while they match it. Weights and records are two separate writes, and two
  // tabs deciding at once could otherwise leave the log naming a winner the
  // score is not using, or two records claiming the same statement. Checking
  // at load makes both self-correcting: the later, hand-set truth wins.
  function consistent(weights, records) {
    return records.filter(function (r) {
      var wa = S.weightOf(weights, r.a), wb = S.weightOf(weights, r.b);
      if (r.valg === "a") return wa === WIN && wb === LOSE;
      if (r.valg === "b") return wb === WIN && wa === LOSE;
      return wa === r.foer.a && wb === r.foer.b; // "lik": nothing was changed
    });
  }

  // --- storage form --------------------------------------------------------
  // A record is stored under the pair it decides between ("utsagn-30|utsagn-55")
  // so store.js can merge it across tabs and drop it when EITHER statement is
  // reworded. The value is a small JSON string.
  function idsOf(key) { return String(key).split("|"); }
  function keyOf(rec) { return rec.a + "|" + rec.b; }

  function encode(rec) {
    return JSON.stringify({ valg: rec.valg, foer: { a: rec.foer.a, b: rec.foer.b }, at: rec.at || 0 });
  }
  function decode(key, value) {
    var ids = idsOf(key), v;
    try { v = JSON.parse(value); } catch (e) { return null; }
    if (ids.length !== 2 || !v || typeof v !== "object") return null;
    if (v.valg !== "a" && v.valg !== "b" && v.valg !== "lik") return null;
    var fa = v.foer && v.foer.a, fb = v.foer && v.foer.b;
    var ok = function (x) { return typeof x === "number" && x >= S.MIN_WEIGHT && x <= S.MAX_WEIGHT && x === Math.floor(x); };
    if (!ok(fa) || !ok(fb)) return null;
    return { a: ids[0], b: ids[1], valg: v.valg, foer: { a: fa, b: fb }, at: typeof v.at === "number" ? v.at : 0 };
  }
  function accept(value) {
    return typeof value === "string" && decode("x|y", value) !== null;
  }
  function toMap(records) {
    var m = {};
    records.forEach(function (r) { m[keyOf(r)] = encode(r); });
    return m;
  }
  function fromMap(map) {
    var out = [];
    Object.keys(map).forEach(function (k) {
      var rec = decode(k, map[k]);
      if (rec) out.push(rec);
    });
    out.sort(function (x, y) { return x.at - y.at || (keyOf(x) < keyOf(y) ? -1 : 1); });
    return out;
  }

  // namesList joins names the way a sentence does: "A", "A og B", "A, B og C".
  function namesList(names) {
    if (names.length <= 1) return names.join("");
    return names.slice(0, -1).join(", ") + " og " + names[names.length - 1];
  }

  return {
    MIN_SHARED: MIN_SHARED, CONFLICT_R: CONFLICT_R, WIN: WIN, LOSE: LOSE,
    pairEvidence: pairEvidence, judgePair: judgePair, marked: marked,
    inRecord: inRecord, plan: plan, decide: decide, undo: undo,
    dropOnManualChange: dropOnManualChange, weightsWithout: weightsWithout,
    consistent: consistent, toMap: toMap, fromMap: fromMap, namesList: namesList,
    storeOpts: { idsOf: idsOf, accept: accept },
  };
});
