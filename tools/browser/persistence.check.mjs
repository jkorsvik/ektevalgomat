// Browser check for the persistence rules: replays every way a visitor used
// to lose progress, against a running server, in real Chrome.
//
//   cd ektevalgomat
//   PORT=8899 ADDR=127.0.0.1 go run ./cmd/server &
//   "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new \
//     --remote-debugging-port=9222 --user-data-dir="$(mktemp -d)" --no-first-run &
//   node tools/browser/persistence.check.mjs          # BASE=http://... to point elsewhere
//
// MANUAL, not part of CI: it needs a Chrome, and CI's node step only runs
// tools/*.test.js. tools/store.test.js is the part CI runs. This file exists
// because store.test.js cannot reach quiz.js (a browser-only script), and the
// bugs that mattered most lived in quiz.js's use of the store: a baseline that
// was reset after each save, a note that never reached a screen reader, a
// note scrolled out of sight. Each was found by running this, not by reading.
//
// ADDR is a PREFIX, not an address: the server listens on ADDR + ":" + PORT,
// so ADDR=:8899 becomes ":8899:8080" and fails. Change the port with PORT.
import { Page, sleep } from "./cdp.mjs";
const BASE = process.env.BASE || "http://127.0.0.1:8899";
const AK = "ektevalgomat-svar-v1", WK = "ektevalgomat-vekt-v1";
let pass = 0, fail = 0;
const t = (name, ok, detail = "") => { ok ? pass++ : fail++; console.log((ok ? "PASS " : "FAIL ") + name + (detail ? "  [" + detail + "]" : "")); };
const stored = (k = AK, f = "answers") => `(() => { const r = localStorage.getItem('${k}'); return r ? JSON.parse(r)['${f}'] : null; })()`;
const click = (sel, i = 0) => `(() => { const b = [...document.querySelectorAll(${JSON.stringify(sel)})][${i}]; if (!b) return false; b.click(); return true; })()`;
const clickText = (tx) => `(() => { const b = [...document.querySelectorAll('button')].find(x => x.textContent.trim().startsWith(${JSON.stringify(tx)})); if (!b) return false; b.click(); return true; })()`;
const note = `(() => { const c = document.getElementById('storage-note'); if (!c) return null; const p = c.querySelector('.storage-note');
  return { hidden: !p, text: p ? p.textContent.trim() : '', warn: !!(p && p.classList.contains('storage-note-warn')), role: c.getAttribute('role') }; })()`;
const prog = `(document.getElementById('quiz-progress-text') || {}).textContent || null`;
const errors = (p) => p.events.filter(e => e.method === "Runtime.exceptionThrown" || (e.method === "Runtime.consoleAPICalled" && e.params.type === "error"))
  .map(e => e.method === "Runtime.exceptionThrown" ? (e.params.exceptionDetails.exception?.description || e.params.exceptionDetails.text) : e.params.args.map(a => a.value).join(" "));
const seed = (n, extra = "") => `(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent);
  const a = {}; D.statements.slice(0, ${n}).forEach((s, i) => a[s.id] = [2, 1, 0, -1, -2][i % 5]);
  localStorage.setItem('${AK}', JSON.stringify({ snapshot_date: D.snapshot_date, direction_version: 2, answers: a, texts: EkteStore.prints(D) })); return Object.keys(a).length; })()`;
const fresh = async () => { const p = await Page.open(); await p.goto(BASE + "/"); await p.eval("localStorage.clear()"); await p.goto(BASE + "/"); return p; };

// ---- 1. two tabs ----
{ const a = await fresh(), b = await Page.open(); await b.goto(BASE + "/");
  await a.eval(click(".scale button", 0)); await sleep(300);
  await b.eval(clickText("Neste")); await sleep(300); await b.eval(click(".scale button", 0)); await sleep(300);
  const s = await a.eval(stored());
  t("two tabs: both answers survive", s && Object.keys(s).length === 2, JSON.stringify(Object.keys(s || {})));
  t("two tabs: no console errors", errors(a).length + errors(b).length === 0, errors(a).concat(errors(b)).join(" | "));
  await a.close(); await b.close(); }

// ---- 2. a failing write is reported, then recovers ----
{ const p = await fresh();
  t("working storage: no warning shown", (await p.eval(note)).hidden === true);
  await p.eval(`window.__set = Storage.prototype.setItem; Storage.prototype.setItem = function () { throw new DOMException('quota', 'QuotaExceededError'); }`);
  await p.eval(click(".scale button", 0)); await sleep(400);
  const n = await p.eval(note);
  t("failed write: the visitor is told", !n.hidden && /lar ikke siden lagre/.test(n.text) && n.warn, n.text.slice(0, 60));
  const c = await p.eval(`(() => { const c = document.getElementById('storage-note'); return { role: c.getAttribute('role'), hiddenAttr: c.hasAttribute('hidden'), text: c.textContent }; })()`);
  t("failed write: lands in its own role=status region, not the one the answer overwrites", c.role === "status" && !c.hiddenAttr && /lar ikke siden lagre/.test(c.text));
  const answerRegion = await p.eval(`[...document.querySelectorAll('[aria-live]')].filter(e => e.id !== 'storage-note').map(e => e.textContent).join('|')`);
  t("failed write: the answer's own announcement does not clobber it", /Sterkt uenig|Uenig|Enig|Nøytral/.test(answerRegion) && /lar ikke siden lagre/.test(c.text), answerRegion.slice(0, 40));
  await p.eval(`Storage.prototype.setItem = window.__set`);
  await p.eval(clickText("Neste")); await sleep(300); await p.eval(click(".scale button", 1)); await sleep(400);
  const s = await p.eval(stored());
  t("recovery: the answer made while failing is saved with the next one", s && Object.keys(s).length === 2, JSON.stringify(Object.keys(s || {})));
  t("recovery: the warning clears", (await p.eval(note)).hidden === true);
  t("failure path: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 2b. storage blocked from the very start: warned BEFORE answering ----
{ const p = await Page.open();
  await p.send("Page.addScriptToEvaluateOnNewDocument", { source: `Storage.prototype.setItem = function () { throw new DOMException('blocked', 'SecurityError'); };` });
  await p.goto(BASE + "/");
  const n = await p.eval(note);
  t("blocked storage: warned at load, before the first answer", !n.hidden && /lar ikke siden lagre/.test(n.text), n.text.slice(0, 50));
  t("blocked storage: quiz still works", await p.eval(click(".scale button", 2)));
  t("blocked storage: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 3. a data refresh keeps what did not change ----
{ const p = await fresh();
  await p.eval(`(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent);
    const ids = D.statements.slice(0, 10).map(s => s.id), a = {}; ids.forEach((id, i) => a[id] = 1 + (i % 2));
    const texts = EkteStore.prints(D); const t = {}; ids.forEach(id => t[id] = texts[id]); t[ids[2]] = 'was-different-wording';
    localStorage.setItem('${AK}', JSON.stringify({ snapshot_date: '2026-09-23', direction_version: 2, answers: a, texts: t })); })()`);
  await p.goto(BASE + "/");
  const n = await p.eval(note);
  t("refresh: 9 kept, 1 reworded dropped, and it says so", n && n.text === "Siden er oppdatert siden sist. 9 av svarene dine er beholdt. 1 gjaldt et utsagn som er endret, og er tatt bort.", n && n.text);
  t("refresh: resumes at the dropped statement", /Spørsmål 3 av 60/.test(await p.eval(prog)), await p.eval(prog));
  t("refresh: stored blob not destroyed before the visitor acts", Object.keys(await p.eval(stored())).length === 10);
  await p.eval(click(".scale button", 0)); await sleep(400);
  const s = await p.eval(stored());
  t("refresh: after the first new answer the 9 kept are still there", s && Object.keys(s).length === 10, String(Object.keys(s || {}).length));
  await p.goto(BASE + "/");
  t("refresh: the notice does not repeat once the blob is current", (await p.eval(note)).hidden !== false || !/oppdatert/.test((await p.eval(note)).text));
  // legacy blob (no fingerprints) from an older snapshot cannot be verified
  await p.eval(`localStorage.setItem('${AK}', JSON.stringify({ snapshot_date: '2026-09-23', direction_version: 2, answers: { 'utsagn-01': 2, 'utsagn-02': 1 } }))`);
  await p.goto(BASE + "/");
  const n2 = await p.eval(note);
  t("legacy blob from an old snapshot: says it could not carry answers over", /kunne ikke tas med/.test(n2.text), n2.text);
  t("refresh: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 4. reload keeps the place ----
{ const p = await fresh(); await p.eval(seed(30)); await p.goto(BASE + "/");
  t("reload: opens at question 31 after 30 answers", /Spørsmål 31 av 60/.test(await p.eval(prog)), await p.eval(prog));
  const n = await p.eval(note);
  t("reload: says where they left off", /Du fortsetter der du slapp. Du har svart på 30 av 60./.test(n.text), n.text);
  await p.eval(clickText("Neste")); await sleep(300);
  t("reload: the line goes away once they move on", (await p.eval(note)).hidden === true);
  await p.goto(BASE + "/#utsagn-5");
  t("a #utsagn-5 link still wins over resume", /Spørsmål 5 av 60/.test(await p.eval(prog)), await p.eval(prog));
  await p.eval(seed(0)); await p.goto(BASE + "/");
  t("nothing answered: opens at question 1, no resume line", /Spørsmål 1 av 60/.test(await p.eval(prog)) && (await p.eval(note)).hidden === true);
  await p.eval(seed(60)); await p.goto(BASE + "/");
  t("all answered: opens at the last question", /Spørsmål 60 av 60/.test(await p.eval(prog)), await p.eval(prog));
  t("resume: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 5. the result page reads the same store ----
{ const p = await fresh(); await p.eval(seed(60)); await p.goto(BASE + "/resultat", 2200);
  const rows = await p.eval(`document.querySelectorAll('.bar-row').length`);
  t("result page: draws a bar per party from the stored answers", rows >= 9, "bars=" + rows);
  t("result page: no resume line leaking onto /resultat", (await p.eval(note)).hidden === true);
  t("result page: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 6. ephemeral mode writes nothing; switching off saves answers AND weights ----
{ const p = await fresh();
  await p.eval(`(() => { const e = document.getElementById('quiz-ephemeral'); e.click(); })()`); await sleep(200);
  await p.eval(click(".scale button", 0)); await sleep(200); await p.eval(clickText("Viktig")); await sleep(200);
  t("ephemeral: nothing written for answers or weights", (await p.eval(stored())) === null && (await p.eval(stored(WK, "weights"))) === null);
  await p.eval(`document.getElementById('quiz-ephemeral').click()`); await sleep(400);
  const a = await p.eval(stored()), w = await p.eval(stored(WK, "weights"));
  t("ephemeral off: the answer is saved", a && Object.keys(a).length === 1);
  t("ephemeral off: the weight is saved too (was lost before)", w && Object.keys(w).length === 1, JSON.stringify(w));
  t("ephemeral: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 7. clear all still clears, and a stale tab does not bring it back ----
{ const a = await fresh(); await a.eval(seed(5)); await a.goto(BASE + "/"); const b = await Page.open(); await b.goto(BASE + "/");
  await a.eval(`document.getElementById('slett-svar-knapp').click()`); await sleep(300);
  await a.eval(`document.querySelector('.delete-confirm-yes').click()`); await sleep(500);
  t("clear all: the answers are gone", (await a.eval(stored())) === null || Object.keys(await a.eval(stored())).length === 0);
  await b.eval(click(".scale button", 0)); await sleep(400);
  const s = await b.eval(stored());
  t("clear all: a stale tab's next answer does not resurrect the old five", s && Object.keys(s).length === 1, JSON.stringify(Object.keys(s || {})));
  await a.close(); await b.close(); }

// ---- 8. a stale tab saving TWICE must not revert what the other tab changed in between ----
// (the first save refreshes the stale tab, so only the SECOND save exposes a wrong baseline)
{ const a = await fresh(), b = await Page.open(); await b.goto(BASE + "/");
  await a.eval(click(".scale button", 0)); await sleep(300);                       // A: q1 = -2
  await b.eval(clickText("Neste")); await sleep(300); await b.eval(click(".scale button", 0)); await sleep(300); // B: q2 = -2 (save 1)
  await a.eval(click(".scale button", 4)); await sleep(300);                       // A: q1 = +2 (A changes its mind)
  await b.eval(clickText("Neste")); await sleep(300); await b.eval(click(".scale button", 0)); await sleep(300); // B: q3 = -2 (save 2)
  const s8 = await a.eval(stored());
  t("stale tab's second save keeps the other tab's later change (q1 stays +2)", s8 && s8["utsagn-01"] === 2 && s8["utsagn-02"] === -2 && s8["utsagn-03"] === -2, JSON.stringify(s8));
  await a.close(); await b.close(); }

// ---- 9. a visitor who only LOOKS still gets their blob upgraded (so a later refresh keeps it) ----
{ const p = await fresh();
  await p.eval(`(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent);
    localStorage.setItem('${AK}', JSON.stringify({ snapshot_date: D.snapshot_date, direction_version: 2, answers: { 'utsagn-01': 2, 'utsagn-02': 1 } }));
    localStorage.setItem('${WK}', JSON.stringify({ snapshot_date: D.snapshot_date, weights: { 'utsagn-01': 3 } })); })()`);
  await p.goto(BASE + "/");                                   // load only, no click
  const raw = await p.eval(`({ a: JSON.parse(localStorage.getItem('${AK}')), w: JSON.parse(localStorage.getItem('${WK}')) })`);
  t("legacy answers gain fingerprints just by loading the page", raw.a.texts && Object.keys(raw.a.texts).length === 2 && Object.keys(raw.a.answers).length === 2, JSON.stringify(Object.keys(raw.a.texts || {})));
  t("legacy weights gain fingerprints too", raw.w.texts && Object.keys(raw.w.texts).length === 1 && raw.w.weights["utsagn-01"] === 3);
  t("upgrade changed no answer", raw.a.answers["utsagn-01"] === 2 && raw.a.answers["utsagn-02"] === 1);
  t("upgrade: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

console.log(`\n${pass} passed, ${fail} failed`);
process.exit(fail ? 1 : 0);
