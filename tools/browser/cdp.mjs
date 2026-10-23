// Dependency-free CDP helper: node's global fetch + WebSocket.
export const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
export class Page {
  static async open(port = 9222) {
    const t = await (await fetch(`http://127.0.0.1:${port}/json/new?about:blank`, { method: "PUT" })).json();
    const p = new Page(); p.tid = t.id; p.port = port; p.id = 0; p.pending = new Map(); p.events = [];
    p.ws = new WebSocket(t.webSocketDebuggerUrl);
    p.ws.addEventListener("message", (e) => {
      const m = JSON.parse(e.data);
      if (m.id && p.pending.has(m.id)) { const { res, rej } = p.pending.get(m.id); p.pending.delete(m.id);
        m.error ? rej(new Error(JSON.stringify(m.error))) : res(m.result); }
      else if (m.method) p.events.push(m);
    });
    await new Promise((r, j) => { p.ws.addEventListener("open", r, { once: true }); p.ws.addEventListener("error", j, { once: true }); });
    await p.send("Page.enable"); await p.send("Runtime.enable");
    return p;
  }
  send(method, params = {}) { const id = ++this.id; this.ws.send(JSON.stringify({ id, method, params }));
    return new Promise((res, rej) => this.pending.set(id, { res, rej })); }
  async goto(url, wait = 1500) { this.events = []; await this.send("Page.navigate", { url });
    const t0 = Date.now(); while (Date.now() - t0 < 8000) { if (this.events.some(e => e.method === "Page.loadEventFired")) break; await sleep(100); }
    await sleep(wait); }
  viewport(w, h, mobile = false) { return this.send("Emulation.setDeviceMetricsOverride", { width: w, height: h, deviceScaleFactor: 1, mobile }); }
  async eval(expression) { const r = await this.send("Runtime.evaluate", { expression, returnByValue: true, awaitPromise: true });
    if (r.exceptionDetails) throw new Error(r.exceptionDetails.exception?.description || r.exceptionDetails.text); return r.result.value; }
  async shot(path, full = false) { const r = await this.send("Page.captureScreenshot", { format: "png", captureBeyondViewport: full });
    (await import("node:fs")).writeFileSync(path, Buffer.from(r.data, "base64")); }
  async close() { try { this.ws.close(); await fetch(`http://127.0.0.1:${this.port}/json/close/${this.tid}`); } catch {} }
}
