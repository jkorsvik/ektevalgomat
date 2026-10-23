// Browser check for the trade-off step ("avveining"), in real Chrome, against
// a running server. Companion to persistence.check.mjs; same setup:
//
//   cd ektevalgomat
//   PORT=8899 ADDR=127.0.0.1 go run ./cmd/server &
//   "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new \
//     --remote-debugging-port=9222 --user-data-dir="$(mktemp -d)" --no-first-run &
//   node tools/browser/avveining.check.mjs            # BASE=http://... to point elsewhere
//
// MANUAL, not part of CI (it needs a Chrome). tools/avveining.test.js is the
// part CI runs; this file checks the wiring that test cannot reach: that the
// DOM, the module, the stored records and the explanation panel agree.
//
// State is seeded through the page's own EkteStore, so it is exactly the
// format the site writes. The visitor answers "sterkt enig" to every
// statement, which makes utsagn 30 and 55 a conflicting pair (7 of 9 parties
// have a position on both; checked against an independent implementation).
import { Page, sleep } from "./cdp.mjs";
const BASE = process.env.BASE || "http://127.0.0.1:8899";
const AK = "ektevalgomat-svar-v1", WK = "ektevalgomat-vekt-v1", RK = "ektevalgomat-avveining-v1";
let pass = 0, fail = 0;
const t = (name, ok, detail = "") => { ok ? pass++ : fail++; console.log((ok ? "PASS " : "FAIL ") + name + (detail ? "  [" + detail + "]" : "")); };
const errors = (p) => p.events.filter(e => e.method === "Runtime.exceptionThrown" || (e.method === "Runtime.consoleAPICalled" && e.params.type === "error"))
  .map(e => e.method === "Runtime.exceptionThrown" ? (e.params.exceptionDetails.exception?.description || e.params.exceptionDetails.text) : e.params.args.map(a => a.value).join(" "));
const J = JSON.stringify;

// seed: answers (all +2, or an object), weights, records, written through the site's own store
const seed = ({ answers = "all2", weights = {}, records = [] } = {}) => `(() => {
  const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent);
  localStorage.clear();
  const a = {}; const spec = ${J(answers)};
  if (spec === 'all2') D.statements.forEach(s => a[s.id] = 2); else Object.assign(a, spec);
  EkteStore.write(localStorage, '${AK}', 'answers', D, a, {}, { direction_version: 2 });
  EkteStore.write(localStorage, '${WK}', 'weights', D, ${J(weights)}, {});
  const recs = ${J(records)};
  if (recs.length) EkteStore.write(localStorage, '${RK}', 'avveininger', D, EkteAvveining.toMap(recs), {}, null, EkteAvveining.storeOpts);
})()`;
const stored = `(() => ({
  w: (JSON.parse(localStorage.getItem('${WK}') || '{}')).weights || null,
  r: (JSON.parse(localStorage.getItem('${RK}') || '{}')).avveininger || null,
}))()`;
const step = `(() => { const b = document.getElementById('avveining'); const s = b && b.querySelector('.avveining-steg'); if (!s) return null;
  return { title: s.querySelector('.avveining-title').textContent, counter: (s.querySelector('.avveining-teller') || {}).textContent || null,
    lede: s.querySelector('.lede').textContent, question: s.querySelector('.avveining-spm').textContent,
    promise: s.querySelector('.avveining-effektlofte').textContent,
    buttons: [...s.querySelectorAll('button')].map(x => x.textContent.trim()),
    optionText: s.querySelector('.avveining-valg').textContent,
    optionClasses: [...s.querySelectorAll('.avveining-velg')].map(x => x.className),
    optionNums: [...s.querySelectorAll('.avveining-alt-nr')].map(x => x.textContent) }; })()`;
const log = `(() => [...document.querySelectorAll('.avveining-logg-tekst')].map(x => x.textContent))()`;
const text = (sel) => `(() => { const e = document.querySelector(${J(sel)}); return e ? e.textContent.replace(/\\s+/g, ' ').trim() : null; })()`;
const clickText = (tx) => `(() => { const b = [...document.querySelectorAll('button')].find(x => x.textContent.trim().startsWith(${J(tx)})); if (!b) return false; b.click(); return true; })()`;
// Expected leaders, computed longhand from the payload (answers all +2): the
// weighted mean per party, no shared code with the site's scoring.
const expectLeaders = (w) => `(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent);
  const num = {}, den = {}; D.parties.forEach(p => { num[p.id] = 0; den[p.id] = 0; });
  const W = ${J(w)};
  D.statements.forEach(s => { const wt = Math.min(3, Math.max(1, W[s.id] || 1));
    D.parties.forEach(p => { const pos = (D.cells[s.id][p.id] || {}).party_position; if (pos === null || pos === undefined) return;
      num[p.id] += Math.round(100 - Math.abs(2 - pos) * 25) * wt; den[p.id] += wt; }); });
  const avg = {}; D.parties.forEach(p => avg[p.id] = den[p.id] ? Math.round(num[p.id] / den[p.id]) : null);
  const best = Math.max.apply(null, Object.values(avg).filter(x => x !== null));
  return D.parties.filter(p => avg[p.id] === best).map(p => p.name); })()`;
// the bar chart's first row, as a full party name
const barTop = `(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent); const r = document.querySelector('.bar-row');
  if (!r) return null; const sh = r.querySelector('.bar-party').textContent; const p = D.parties.find(x => x.short === sh); return p ? p.name : null; })()`;
const panelTop = `(() => { const m = (document.getElementById('analyse-root') || {textContent: ''}).textContent.match(/partiet på topp \\(([^)]+)\\)/); return m ? m[1] : null; })()`;
const open = async (path = "/resultat", wait = 2200) => { const p = await Page.open(); await p.goto(BASE + path); return p; };
const fresh = async (s, path = "/resultat") => { const p = await open("/", 1200); await p.eval(seed(s)); await p.goto(BASE + path, 2200); return p; };

const P3055 = { weights: { "utsagn-30": 3, "utsagn-55": 3 } };

// ---- 1. when there is nothing to ask, nothing is shown ----
{ let p = await fresh({ weights: {} });
  t("no marked statement: no trade-off section and no line advertising one", (await p.eval(`!document.getElementById('avveining') && !document.querySelector('.avveining-ingen')`)));
  await p.goto(BASE + "/resultat"); await p.eval(seed({ weights: { "utsagn-30": 3 } })); await p.goto(BASE + "/resultat", 2000);
  t("one marked statement: still nothing to ask", (await p.eval(`!document.getElementById('avveining') && !document.querySelector('.avveining-ingen')`)));
  t("nothing marked: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 2. a conflicting pair is asked, in the Designer's words ----
{ const p = await fresh(P3055);
  const s = await p.eval(step);
  t("the step is shown for utsagn 30 and 55", !!s && /Utsagn 30/.test(s.optionNums[0]) && /Utsagn 55/.test(s.optionNums[1]), s && J(s.optionNums));
  t("heading is the Designer's", s.title === "To viktige utsagn trekker hver sin vei", s.title);
  t("no counter when there is only one trade-off", s.counter === null);
  t("lede says both are 'svært viktige' when both are 3", /merket begge som svært viktige\./.test(s.lede));
  t("question is a ranking, not a sacrifice", s.question === "Hvilket av disse veier tyngst for deg?");
  t("the promise states the 3 : 2 rule and that answers do not change", s.promise === "Det du velger, teller 3 i snittet, det andre 2. Svarene dine endres ikke, og du kan angre.");
  t("the four buttons", J(s.buttons) === J(["Utsagn 30 veier tyngst", "Utsagn 55 veier tyngst", "De veier like mye", "Ikke nå"]), J(s.buttons));
  t("the two options are styled identically", s.optionClasses[0] === s.optionClasses[1], J(s.optionClasses));
  t("each option shows the visitor's own answer", s.optionNums.every(x => /Du svarte: Sterkt enig/.test(x)));
  const names = await p.eval(`JSON.parse(document.getElementById('ektevalgomat-data').textContent).parties.map(x => x.name)`);
  const leaked = names.filter(n => s.optionText.includes(n));
  t("the choice carries no party names (the visitor decides on the statements)", leaked.length === 0, leaked.join(","));
  t("evidence is behind a closed disclosure", await p.eval(`(() => { const d = document.querySelector('.avveining-hvorfor'); return !!d && !d.open && /Hvorfor spør vi om dette\\?/.test(d.textContent); })()`));
  await p.eval(`document.querySelector('.avveining-hvorfor summary').click()`);
  t("evidence: 7 of 9 parties, with a row per shared party", /7 av 9 partier har en posisjon med kilde på begge/.test(await p.eval(text(".avveining-hvorfor"))) && (await p.eval(`document.querySelectorAll('.avveining-bevis tr').length`)) === 8);
  t("no stored decision yet", (await p.eval(stored)).r === null);
  t("asking: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 3. choosing: weights, record, log, effect, panel, focus, persistence ----
{ const p = await fresh(P3055);
  await p.eval(clickText("Utsagn 30 veier tyngst")); await sleep(500);
  const st = await p.eval(stored);
  t("choosing 30: weights are 30=3, 55=2", st.w && st.w["utsagn-30"] === 3 && st.w["utsagn-55"] === 2, J(st.w));
  t("the record is stored under the pair it decides", st.r && Object.keys(st.r).join() === "utsagn-30|utsagn-55", J(Object.keys(st.r || {})));
  const lg = await p.eval(log);
  t("the log names the winner and the 3 : 2 rule", lg.length === 1 && /^Utsagn 30 \(.+\) veier tyngst, foran utsagn 55 \(.+\)\. Teller 3 mot 2\.$/.test(lg[0]), lg[0]);
  t("the effect line is shown next to the log", /^(Øverst ligger fortsatt|Uten avveiningene|Med avveiningene)/.test(await p.eval(text(".avveining-effekt"))), await p.eval(text(".avveining-effekt")));
  t("the moved-numbers line uses words, not an arrow", /^(Slik flyttet tallene seg: .* fra \d+ til \d+ %|Ingen partier fikk et annet tall\.)/.test(await p.eval(text(".avveining-flytt"))) && !(await p.eval(text(".avveining-flytt"))).includes("→"), await p.eval(text(".avveining-flytt")));
  t("the step is gone once decided (only one pair existed)", (await p.eval(step)) === null);
  t("focus moved to the log heading", await p.eval(`document.activeElement && document.activeElement.id === 'avveining-heading'`));
  t("announced", /Utsagn 30 veier nå tyngst\. Resultatet er regnet ut på nytt\./.test(await p.eval(`[...document.querySelectorAll('[aria-live]')].map(e => e.textContent).join('|')`)));
  const exp3 = await p.eval(expectLeaders({ "utsagn-30": 3, "utsagn-55": 2 })), bt3 = await p.eval(barTop);
  t("after the trade-off the bar chart's top party is the independently computed one", exp3.length === 1 ? bt3 === exp3[0] : exp3.includes(bt3), J(exp3) + " / bars " + bt3);
  const eff3 = await p.eval(text(".avveining-effekt"));
  t("the effect line names the party the independent computation puts on top", exp3.some(n => eff3.includes(n)), J(exp3) + " in: " + eff3);
  await p.goto(BASE + "/resultat", 2200);
  t("reload: the decision is still there (progress is not lost)", (await p.eval(log)).length === 1);
  const st2 = await p.eval(stored);
  t("reload: weights unchanged", st2.w["utsagn-30"] === 3 && st2.w["utsagn-55"] === 2);
  t("choosing: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 4. the panel explains the SAME result (the event carries the weights) ----
{ const p = await fresh({ weights: { "utsagn-30": 3, "utsagn-55": 3, "utsagn-11": 2 } });
  await p.eval(`document.querySelector('#analyse-root summary').click()`);
  await p.eval(clickText("Utsagn 55 veier tyngst")); await sleep(600);
  const after = await p.eval(text("#analyse-root"));
  t("the panel stays open across the re-render", await p.eval(`document.querySelector('#analyse-root details').open`));
  t("the panel re-rendered with the weights now in force", /Resultatet ditt bruker vektene du har satt/.test(after));
  // 55 won: 55=3, 30=2, 11=2
  const exp4 = await p.eval(expectLeaders({ "utsagn-55": 3, "utsagn-30": 2, "utsagn-11": 2 }));
  const bt = await p.eval(barTop), pt = await p.eval(panelTop);
  t("bar chart, panel and the independent computation all name the same party (no second winner on the page)",
    exp4.length === 1 ? (bt === exp4[0] && pt === exp4[0]) : (exp4.includes(bt)), J(exp4) + " / bars " + bt + " / panel " + pt);
  await p.close(); }

// ---- 4b. the panel trusts the event, not storage ----
// The event carries the weights in force. If the panel re-read storage instead,
// a failed save would leave it explaining stale weights beside a headline that
// uses the new ones. Prove it follows the event by sending one that disagrees
// with storage.
{ const p = await fresh({ weights: { "utsagn-30": 3, "utsagn-55": 3 } });
  const before = await p.eval(text("#analyse-root"));
  t("with weights in storage the panel says the result uses them", /Resultatet ditt bruker vektene du har satt/.test(before));
  await p.eval(`document.dispatchEvent(new CustomEvent('ektevalgomat:vekt-endret', { detail: { weights: {} } }))`); await sleep(300);
  const after = await p.eval(text("#analyse-root"));
  t("an event saying 'no weights' (storage still has them) makes the panel say so: it follows the event", /Standardresultatet vekter hvert utsagn likt/.test(after) && !/Resultatet ditt bruker vektene/.test(after));
  await p.close(); }

// ---- 5. undo restores exactly; "like mye"; "Ikke nå" is never stored ----
{ const p = await fresh({ weights: { "utsagn-30": 2, "utsagn-55": 3 } });
  const w0 = (await p.eval(stored)).w;
  await p.eval(clickText("Utsagn 30 veier tyngst")); await sleep(400);
  await p.eval(`document.querySelector('.avveining-angre').click()`); await sleep(500);
  const st = await p.eval(stored);
  t("undo restores the exact weights the visitor had set", J(st.w) === J(w0), J(st.w) + " vs " + J(w0));
  t("undo removes the record", !st.r || Object.keys(st.r).length === 0);
  t("after undo the pair can be asked again", (await p.eval(step)) !== null);
  await p.eval(clickText("De veier like mye")); await sleep(400);
  const lk = await p.eval(stored);
  t("'like mye' changes no weight", J(lk.w) === J(w0), J(lk.w));
  t("'like mye' is recorded, so the pair is not asked again", lk.r && Object.keys(lk.r).length === 1 && (await p.eval(step)) === null);
  t("'like mye' reads as such in the log", /^Utsagn 30 og utsagn 55 veier like mye\. Vektene er ikke endret\.$/.test((await p.eval(log))[0]), (await p.eval(log))[0]);
  await p.eval(`document.querySelector('.avveining-angre').click()`); await sleep(400);
  await p.eval(clickText("Ikke nå")); await sleep(400);
  t("'Ikke nå' leaves one line and a way back", /Du har en avveining du ikke har tatt stilling til\./.test(await p.eval(text(".avveining-utsatt"))) && await p.eval(`!!document.querySelector('.avveining-utsatt button')`));
  t("'Ikke nå' stores nothing", !(await p.eval(stored)).r || Object.keys((await p.eval(stored)).r).length === 0);
  await p.goto(BASE + "/resultat", 2200);
  t("'Ikke nå' does not survive a reload: the step is back (a stored deferral would be a decision never made)", (await p.eval(step)) !== null);
  await p.eval(clickText("Ikke nå")); await sleep(300); await p.eval(clickText("Ta stilling nå")); await sleep(300);
  t("'Ta stilling nå' brings the step back", (await p.eval(step)) !== null);
  t("undo / like / defer: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 6. a hand-set weight ends the trade-off, and the card says so beforehand ----
{ const p = await fresh(P3055);
  await p.eval(clickText("Utsagn 30 veier tyngst")); await sleep(500);
  await p.goto(BASE + "/#utsagn-55", 1800);
  const note = await p.eval(text("#weight-note-55"));
  t("the card says a trade-off set its weight", note === "Satt i en avveining mot utsagn 30 på resultatsiden. Endrer du vekten her, fjernes avveiningen.", note);
  await p.eval(clickText("Vanlig")); await sleep(500);
  t("changing it by hand removes the card note", (await p.eval(`!document.getElementById('weight-note-55')`)));
  const st = await p.eval(stored);
  t("...and removes the record", !st.r || Object.keys(st.r).length === 0, J(st.r));
  t("...and the hand-set weight wins", !st.w || st.w["utsagn-55"] === undefined, J(st.w));
  t("...with the right announcement", /Avveiningen med dette utsagnet er fjernet\./.test(await p.eval(`[...document.querySelectorAll('[aria-live]')].map(e => e.textContent).join('|')`)));
  await p.close(); }

// ---- 7. two tabs: nothing is lost, nothing lies ----
{ const a = await fresh(P3055), b = await open("/resultat");
  await a.eval(clickText("Utsagn 30 veier tyngst")); await sleep(500);
  await b.eval(clickText("Utsagn 55 veier tyngst")); await sleep(600);       // b was stale: it still showed the step
  const fin = await open("/resultat");
  const st = await fin.eval(stored), lines = await fin.eval(log);
  let lies = 0;
  lines.forEach(l => { const m = l.match(/^Utsagn (\d+).*veier tyngst, foran utsagn (\d+)/); if (m) { if (!(st.w["utsagn-" + m[1]] === 3 && st.w["utsagn-" + m[2]] === 2)) lies++; } });
  t("after two tabs decided the same pair differently, no log line names a winner the weights do not back", lies === 0, J(lines) + " weights " + J(st.w));
  // The earlier version of this scenario "passed" with an EMPTY log, i.e. the
  // decision silently vanished. The last decision must win, in log and weights.
  t("...and the LAST decision wins cleanly: one line, 55 ahead, weights 55=3 and 30=2", lines.length === 1 && /^Utsagn 55 .*veier tyngst, foran utsagn 30/.test(lines[0]) && st.w["utsagn-55"] === 3 && st.w["utsagn-30"] === 2, J(lines) + " " + J(st.w));
  t("two tabs: no console errors", errors(a).length + errors(b).length + errors(fin).length === 0, errors(a).concat(errors(b)).join(" | "));
  await a.close(); await b.close(); await fin.close(); }
{ // different pairs: both survive
  const pairs = await (async () => { const p = await open("/", 1200); const r = await p.eval(`(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent); const a = {}, w = {};
      D.statements.forEach(s => { a[s.id] = 2; w[s.id] = 3; }); return EkteAvveining.plan(D, a, w, []).pending.slice(0, 2).map(x => [x.a.id, x.b.id, x.a.number, x.b.number]); })()`); await p.close(); return r; })();
  const [P1, P2] = pairs;
  const weights = {}; [P1, P2].forEach(x => { weights[x[0]] = 3; weights[x[1]] = 3; });
  const a = await fresh({ weights }), b = await open("/resultat");
  t("four marked statements, two disjoint conflicting pairs: the counter says 'Avveining 1 av 2'", (await a.eval(step)).counter === "Avveining 1 av 2");
  await a.eval(clickText("Utsagn " + P1[2] + " veier tyngst")); await sleep(500);
  t("after the first, the counter says 'Avveining 2 av 2'", (await a.eval(step)).counter === "Avveining 2 av 2", J(await a.eval(step)));
  await b.eval(clickText("De veier like mye")); await sleep(400);              // stale tab settles pair 1 as equal, then is shown pair 2
  await b.eval(clickText("Utsagn " + P2[2] + " veier tyngst")); await sleep(600);
  const fin = await open("/resultat");
  const st = await fin.eval(stored);
  t("a second pair decided in the stale tab survives", st.r && Object.keys(st.r).includes(P2[0] + "|" + P2[1]), J(Object.keys(st.r || {})));
  t("the first pair keeps the stale tab's later 'like mye' (3 and 3, consistent with its record)", st.w[P1[0]] === 3 && st.w[P1[1]] === 3 && JSON.parse(st.r[P1[0] + "|" + P1[1]]).valg === "lik", J(st.w));
  t("its weights are in force: winner 3, other 2", st.w[P2[0]] === 3 && st.w[P2[1]] === 2, J(st.w));
  t("different pairs: no console errors", errors(a).length + errors(b).length + errors(fin).length === 0);
  await a.close(); await b.close(); await fin.close(); }

// ---- 8. a failed save is reported and the decision still stands for the session ----
{ const p = await fresh(P3055);
  await p.eval(`Storage.prototype.setItem = function () { throw new DOMException('quota', 'QuotaExceededError'); }`);
  await p.eval(clickText("Utsagn 30 veier tyngst")); await sleep(500);
  t("failed save: the decision is still shown for this session", (await p.eval(log)).length === 1);
  t("failed save: the visitor is told it will not survive", /lar ikke siden lagre/.test(await p.eval(text("#storage-note"))));
  t("failed save: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 9. a reworded statement drops the record, and says nothing false ----
{ const p = await fresh({ ...P3055, records: [{ a: "utsagn-30", b: "utsagn-55", valg: "a", foer: { a: 3, b: 3 }, at: 1 }] });
  await p.eval(`(() => { const w = JSON.parse(localStorage.getItem('${WK}')); w.weights['utsagn-30'] = 3; w.weights['utsagn-55'] = 2; localStorage.setItem('${WK}', JSON.stringify(w)); })()`);
  await p.goto(BASE + "/resultat", 2200);
  t("a record whose weights are in force is shown", (await p.eval(log)).length === 1);
  await p.eval(`(() => { const r = JSON.parse(localStorage.getItem('${RK}')); r.texts['utsagn-55'] = 'was-different-wording'; localStorage.setItem('${RK}', JSON.stringify(r)); })()`);
  await p.goto(BASE + "/resultat", 2200);
  t("after ONE of its two statements is reworded, the record is not shown", (await p.eval(log)).length === 0);
  await p.eval(seed({ ...P3055, records: [{ a: "utsagn-30", b: "utsagn-55", valg: "a", foer: { a: 3, b: 3 }, at: 1 }] }));
  await p.goto(BASE + "/resultat", 2200);
  t("a record whose weights were changed elsewhere (3/3 stored, record says 30 won) is not believed", (await p.eval(log)).length === 0);
  await p.close(); }

// ---- 10. the stated absences ----
{ const p = await open("/", 1200);
  const picks = await p.eval(`(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent); const a = {}; D.statements.forEach(s => a[s.id] = 2);
    const ids = D.statements.map(s => s.id); let agree = null, unk = null;
    for (let i = 0; i < ids.length && !(agree && unk); i++) for (let j = i + 1; j < ids.length; j++) { const v = EkteAvveining.judgePair(D, a, ids[i], ids[j]).verdict;
      if (v === 'no-conflict' && !agree) agree = [ids[i], ids[j]]; if (v === 'unknown' && !unk) unk = [ids[i], ids[j]]; }
    return { agree, unk }; })()`);
  await p.eval(seed({ weights: { [picks.agree[0]]: 3, [picks.agree[1]]: 3 } })); await p.goto(BASE + "/resultat", 2200);
  t("two important statements that do not pull apart: a stated absence", /^De 2 viktige utsagnene dine trekker ikke hver sin vei, så det er ingenting å avveie\./.test(await p.eval(text(".avveining-ingen"))), await p.eval(text(".avveining-ingen")));
  t("...and no step", (await p.eval(step)) === null);
  await p.eval(seed({ weights: { [picks.unk[0]]: 3, [picks.unk[1]]: 3 } })); await p.goto(BASE + "/resultat", 2200);
  t("a pair that cannot be judged says it cannot tell (never 'no conflict')", /^Vi kan ikke si om de viktige utsagnene dine trekker hver sin vei\./.test(await p.eval(text(".avveining-ingen"))), await p.eval(text(".avveining-ingen")));
  await p.close(); }

// ---- 11. a tie at the top is named as a tie ----
{ const p = await open("/", 1200);
  const pick = await p.eval(`(() => { const D = JSON.parse(document.getElementById('ektevalgomat-data').textContent); let tie = null, single = null;
    D.statements.forEach(st => [-2, -1, 0, 1, 2].forEach(r => { const m = D.parties.map(pt => EkteScoring.pct(r, (D.cells[st.id][pt.id] || {}).party_position)).filter(x => x !== null);
      if (!m.length) return; const best = Math.max.apply(null, m), n = m.filter(x => x === best).length;
      if (n >= 2 && !tie) tie = [st.id, r, n]; if (n === 1 && !single) single = [st.id, r]; })); return { tie, single }; })()`);
  await p.eval(seed({ answers: { [pick.tie[0]]: pick.tie[1] } })); await p.goto(BASE + "/resultat", 2200);
  const h = await p.eval(text(".summary-lede"));
  t("a tie: the headline names every tied party and says 'hver'", /like mye på linje med .+ \(\d+% samsvar hver\)\./.test(h) && (h.match(/ og | og|, /g) || []).length >= 1, h);
  await p.eval(seed({ answers: { [pick.single[0]]: pick.single[1] } })); await p.goto(BASE + "/resultat", 2200);
  const h2 = await p.eval(text(".summary-lede"));
  t("a single leader keeps the original wording", /mest på linje med [^,]+ \(\d+% samsvar\)\./.test(h2) && !/hver\)/.test(h2), h2);
  await p.close(); }

// ---- 12. ephemeral mode: the step works, nothing is written ----
{ const p = await fresh(P3055, "/");
  await p.eval(`document.getElementById('quiz-ephemeral').click()`); await sleep(300);
  await p.eval(clickText("Se over svarene dine")); await sleep(400);
  const hasBtn = await p.eval(clickText("Se resultatet her")); await sleep(600);
  t("ephemeral: the inline result shows the step", hasBtn && (await p.eval(step)) !== null);
  const before = await p.eval(`({ w: localStorage.getItem('${WK}'), r: localStorage.getItem('${RK}') })`);
  await p.eval(clickText("Utsagn 30 veier tyngst")); await sleep(500);
  t("ephemeral: the decision shows", (await p.eval(log)).length === 1);
  const after = await p.eval(`({ w: localStorage.getItem('${WK}'), r: localStorage.getItem('${RK}') })`);
  t("ephemeral: nothing was written (weights and records unchanged in storage)", J(before) === J(after), J(after).slice(0, 80));
  t("ephemeral: no console errors", errors(p).length === 0, errors(p).join(" | "));
  await p.close(); }

// ---- 13. clear all removes the records too ----
{ const p = await fresh({ ...P3055 }, "/resultat");
  await p.eval(clickText("Utsagn 30 veier tyngst")); await sleep(500);
  t("a decision is stored", (await p.eval(stored)).r !== null);
  await p.eval(`document.getElementById('slett-svar-knapp').click()`); await sleep(200);
  await p.eval(`document.querySelector('.delete-confirm-yes').click()`); await sleep(500);
  const st = await p.eval(`({ a: localStorage.getItem('${AK}'), w: localStorage.getItem('${WK}'), r: localStorage.getItem('${RK}') })`);
  t("Slett alle svar removes answers, weights AND trade-offs", st.a === null && st.w === null && st.r === null, J(st));
  await p.close(); }

// ---- 14. 375 px: stacks, tappable, nothing overflows ----
{ const p = await Page.open(); await p.viewport(375, 800, true); await p.goto(BASE + "/"); await p.eval(seed(P3055)); await p.goto(BASE + "/resultat", 2200);
  const m = await p.eval(`(() => { const box = document.getElementById('avveining'); const grid = getComputedStyle(box.querySelector('.avveining-valg')).gridTemplateColumns.split(' ').length;
    return { cols: grid, buttons: [...box.querySelectorAll('button')].map(b => Math.round(b.getBoundingClientRect().height)),
      overflow: document.documentElement.scrollWidth - window.innerWidth }; })()`);
  t("375 px: the two options stack in one column", m.cols === 1, "columns=" + m.cols);
  t("375 px: every button is at least 44 px tall", m.buttons.every(h => h >= 44), J(m.buttons));
  t("375 px: no horizontal scroll", m.overflow <= 0, "overflow=" + m.overflow);
  await p.close(); }

console.log(`\n${pass} passed, ${fail} failed`);
process.exit(fail ? 1 : 0);
