// Persistence for the visitor's own choices (answers, weights, and the
// trade-off records that come next). Pure: the storage object is passed in,
// so node can test it with a fake and two "tabs" over one backing store.
//
// Why this exists. Before 2026-10-16 each caller read and wrote localStorage
// by itself, and four ways of losing a visitor's progress had grown up:
//
//   1. Two tabs. save() wrote the whole blob from that tab's memory, so a tab
//      that had loaded earlier erased every answer the other tab had saved.
//   2. A failed write (private mode, quota) was swallowed. The live region
//      still said "Utsagn 1: Sterkt uenig." as if it had been kept.
//   3. A data refresh changed snapshot_date, the old rule dropped every
//      answer, and the first new answer then overwrote the stored ones for
//      good. A weekly evidence update cost every returning visitor everything.
//   4. quiz.js and analyse.js each carried their own copy of the load rule,
//      so they could (and did) disagree.
//
// The rules here, in one place:
//
//   - A stored entry is valid when its statement still exists AND its text
//     fingerprint matches. A refresh therefore keeps every answer whose
//     statement is unchanged, and drops only the ones that were reworded.
//     A blob written before fingerprints existed has none, so it falls back to
//     the old rule (same snapshot_date or nothing). Nothing is dropped
//     silently: read() reports kept and dropped so the page can say so.
//   - write() merges. It re-reads the latest stored state, applies only what
//     THIS tab changed since it last loaded or saved, and writes that. A stale
//     tab cannot erase another tab's answers, and a deletion made on purpose
//     is a change like any other.
//   - A failed write reports ok:false and leaves the changes "dirty", so the
//     next successful write persists them. It never claims success.
(function (root, factory) {
  if (typeof module === "object" && module.exports) {
    module.exports = factory();
  } else {
    root.EkteStore = factory();
  }
})(typeof self !== "undefined" ? self : this, function () {
  "use strict";

  var has = Object.prototype.hasOwnProperty;

  // fingerprint is FNV-1a over UTF-16 code units, base 36. It is a change
  // detector for statement wording, not a security hash.
  function fingerprint(text) {
    var h = 0x811c9dc5;
    var s = String(text);
    for (var i = 0; i < s.length; i++) {
      h ^= s.charCodeAt(i);
      h = Math.imul(h, 0x01000193) >>> 0;
    }
    return h.toString(36);
  }

  // prints maps every current statement id to the fingerprint of its text.
  function prints(data) {
    var m = {};
    (data.statements || []).forEach(function (st) {
      m[st.id] = fingerprint(st.text || "");
    });
    return m;
  }

  function validator(blob, data, current) {
    var texts = blob && blob.texts && typeof blob.texts === "object" ? blob.texts : null;
    var sameSnapshot = !!blob && blob.snapshot_date === data.snapshot_date;
    return function (id) {
      if (!has.call(current, id)) return false; // the statement is gone
      if (texts && has.call(texts, id)) return texts[id] === current[id];
      return sameSnapshot; // legacy blob: the old rule
    };
  }

  function single(k) { return [k]; }

  function plainValue(v) {
    return typeof v === "number" || typeof v === "string";
  }

  // read returns the entries of blob[field] that are still valid.
  //   { map, kept, dropped, blob, ok, corrupt, legacy }
  // legacy is true for a blob written before fingerprints existed. Its
  // entries are only trusted while snapshot_date matches, so the caller
  // should rewrite it (write() with no changes) the first time it is seen,
  // or a returning visitor who only LOOKS would lose it at the next refresh.
  // ok is false only when storage itself cannot be read. A missing or corrupt
  // blob is an empty result, not an error: there is nothing to lose.
  //
  // opts lets an entry stand for more than one statement. A trade-off record
  // is keyed by the PAIR of statements it decides between, so it is only
  // valid while BOTH are unchanged:
  //   opts.idsOf(entryKey) -> the statement ids the entry depends on
  //   opts.accept(value)   -> whether the stored value has the right shape
  // The defaults are one id per entry and a plain number or string.
  function read(storage, key, field, data, opts) {
    var out = { map: {}, kept: 0, dropped: 0, blob: null, ok: true, corrupt: false, legacy: false };
    if (!storage) { out.ok = false; return out; }
    var raw;
    try { raw = storage.getItem(key); } catch (e) { out.ok = false; return out; }
    if (raw === null || raw === undefined || raw === "") return out;
    var blob;
    try { blob = JSON.parse(raw); } catch (e) { out.corrupt = true; return out; }
    if (!blob || typeof blob !== "object" || Array.isArray(blob)) { out.corrupt = true; return out; }
    out.blob = blob;
    var src = blob[field];
    if (!src || typeof src !== "object" || Array.isArray(src)) return out;
    var valid = validator(blob, data, prints(data));
    var idsOf = (opts && opts.idsOf) || single;
    var accept = (opts && opts.accept) || plainValue;
    Object.keys(src).forEach(function (k) {
      var ids = idsOf(k);
      if (ids.length > 0 && ids.every(valid) && accept(src[k])) { out.map[k] = src[k]; out.kept++; } else { out.dropped++; }
    });
    out.legacy = !blob.texts && out.kept > 0;
    return out;
  }

  // write merges this tab's changes into whatever is stored now.
  //   current: this tab's map in memory
  //   base:    the map as this tab last loaded or saved it
  //   extra:   fields to carry on the blob (direction_version)
  // Returns { ok, map }. map is the merged result; on failure it is current.
  function write(storage, key, field, data, current, base, extra, opts) {
    if (!storage) return { ok: false, map: current };
    var idsOf = (opts && opts.idsOf) || single;
    var latest = read(storage, key, field, data, opts);
    if (!latest.ok) return { ok: false, map: current };
    var merged = Object.assign({}, latest.map);
    var seen = {};
    Object.keys(current).concat(Object.keys(base)).forEach(function (id) {
      if (seen[id]) return;
      seen[id] = true;
      if (current[id] === base[id]) return; // this tab did not touch it
      if (current[id] === undefined) delete merged[id]; else merged[id] = current[id];
    });
    var now = prints(data);
    Object.keys(merged).forEach(function (k) {
      if (!idsOf(k).every(function (id) { return has.call(now, id); })) delete merged[k];
    });
    var blob = Object.assign({}, extra || {}, { snapshot_date: data.snapshot_date });
    blob[field] = merged;
    blob.texts = {};
    Object.keys(merged).forEach(function (k) {
      idsOf(k).forEach(function (id) { blob.texts[id] = now[id]; });
    });
    try { storage.setItem(key, JSON.stringify(blob)); } catch (e) { return { ok: false, map: current }; }
    return { ok: true, map: merged };
  }

  // replaceInPlace makes target equal to source without changing the object's
  // identity, so every closure that holds a reference sees the merge.
  function replaceInPlace(target, source) {
    Object.keys(target).forEach(function (k) { if (!has.call(source, k)) delete target[k]; });
    Object.keys(source).forEach(function (k) { target[k] = source[k]; });
    return target;
  }

  // available probes whether a write can succeed right now, so a visitor in a
  // browser that blocks storage is told BEFORE they answer sixty questions.
  function available(storage) {
    if (!storage) return false;
    try {
      storage.setItem("ektevalgomat-probe", "1");
      storage.removeItem("ektevalgomat-probe");
      return true;
    } catch (e) {
      return false;
    }
  }

  function clear(storage, keys) {
    if (!storage) return false;
    var ok = true;
    keys.forEach(function (k) { try { storage.removeItem(k); } catch (e) { ok = false; } });
    return ok;
  }

  // droppedMessage is what the page says when a refresh could not carry
  // answers over. It never claims more than was measured.
  function droppedMessage(kept, dropped) {
    if (!dropped) return null;
    if (!kept) return "Siden er oppdatert siden sist, og svarene dine fra den gangen kunne ikke tas med.";
    return "Siden er oppdatert siden sist. " + kept + " av svarene dine er beholdt. " + dropped + (dropped === 1 ? " gjaldt et utsagn" : " gjaldt utsagn") + " som er endret, og er tatt bort.";
  }

  var NOSAVE_MESSAGE =
    "Nettleseren din lar ikke siden lagre svarene. De forsvinner hvis du laster siden på nytt eller lukker den.";

  return {
    fingerprint: fingerprint,
    prints: prints,
    read: read,
    write: write,
    replaceInPlace: replaceInPlace,
    available: available,
    clear: clear,
    droppedMessage: droppedMessage,
    NOSAVE_MESSAGE: NOSAVE_MESSAGE,
  };
});
