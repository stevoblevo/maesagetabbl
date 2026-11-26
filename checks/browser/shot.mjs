import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const file = 'file:///home/claude/test/index.html';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const errs = [];
async function page(w, h, touch=false) {
  const ctx = await browser.newContext({ viewport: { width: w, height: h }, deviceScaleFactor: 1, hasTouch: touch });
  const p = await ctx.newPage();
  p.on('pageerror', e => errs.push('pageerror: ' + e.message));
  p.on('console', m => { if (m.type() === 'error' && !/fonts\.g|ERR_|Failed to load resource/.test(m.text())) errs.push('console: ' + m.text()); });
  await p.goto(file); await p.waitForTimeout(900);
  return p;
}
const mode = process.argv[2] || 'lobby';
if (mode === 'lobby') {
  let p = await page(1280, 800); await p.mouse.move(520, 420); await p.waitForTimeout(700); await p.screenshot({ path: 'lobby-desk.png' });
  p = await page(390, 800, true); await p.screenshot({ path: 'lobby-phone.png' });
}
if (mode === 'moles') {
  const p = await page(1280, 800);
  await p.evaluate(() => __maesage.start('moles'));
  await p.waitForTimeout(2600);
  // hover the first live cue and hold
  const pos = await p.evaluate(() => { const c = __maesage.G.cues.find(c => !c.out && c.kind !== 'tease'); const [x, y] = __maesage.toS(c.z.x, c.z.y); return { x, y, kind: c.kind }; });
  await p.mouse.move(pos.x, pos.y); await p.waitForTimeout(200); await p.mouse.down(); await p.waitForTimeout(350);
  await p.screenshot({ path: 'moles-desk.png' });
  await p.waitForTimeout(1200); await p.mouse.up();
  console.log('moles state', await p.evaluate(() => JSON.stringify({ hit: __maesage.G.hit, fooled: __maesage.G.fooled, tension: __maesage.S.tension, kind: 0 })), pos.kind);
  // spine cut-out
  const sp = await p.evaluate(() => __maesage.toS(-.3, 0)); await p.mouse.move(sp[0], sp[1]); await p.mouse.down(); await p.waitForTimeout(400); await p.screenshot({ path: 'moles-spine.png' }); await p.mouse.up();
  // auto-play the rest quickly to reach the end card
  for (let i = 0; i < 60; i++) {
    const q = await p.evaluate(() => { const G = __maesage.G; if (!G || G.over || __maesage.S.screen !== 'moles') return null; const c = G.cues.find(c => !c.out && c.kind === 'real' && c.age > .3); if (!c) return 0; return __maesage.toS(c.z.x, c.z.y); });
    if (q === null) break; if (q === 0) { await p.waitForTimeout(250); continue; }
    await p.mouse.move(q[0], q[1]); await p.mouse.down(); await p.waitForTimeout(1250); await p.mouse.up();
  }
  await p.waitForTimeout(1800); await p.screenshot({ path: 'moles-end.png' });
  console.log('end', await p.evaluate(() => JSON.stringify({ screen: __maesage.S.screen, res: __maesage.S.results })));
}
if (mode === 'hogs') {
  const p = await page(1280, 800);
  await p.evaluate(() => __maesage.start('hogs')); await p.waitForTimeout(600);
  const h = await p.evaluate(() => { const h = __maesage.G.hogs[1]; return __maesage.toS(h.x, h.y); });
  await p.mouse.move(h[0] + 18, h[1]); await p.mouse.down();
  for (let i = 0; i < 70; i++) { const a = i * .28; await p.mouse.move(h[0] + Math.cos(a) * 18, h[1] + Math.sin(a) * 18); await p.waitForTimeout(30); if (i === 30) await p.screenshot({ path: 'hogs-desk.png' }); }
  await p.mouse.up(); await p.waitForTimeout(500); await p.screenshot({ path: 'hogs-ball.png' });
  console.log('hogs', await p.evaluate(() => JSON.stringify(__maesage.G.hogs.map(h => [h.state, +h.calm.toFixed(2)]))));
  // rush at hog 0
  const h0 = await p.evaluate(() => { const h = __maesage.G.hogs[0]; return __maesage.toS(h.x, h.y); });
  await p.mouse.move(h0[0] - 300, h0[1] - 100); await p.waitForTimeout(100); await p.mouse.move(h0[0], h0[1], { steps: 3 }); await p.waitForTimeout(250);
  await p.screenshot({ path: 'hogs-bristle.png' });
  console.log('bristled', await p.evaluate(() => __maesage.G.bristled));
  await p.waitForTimeout(4000);
  console.log('hogs later', await p.evaluate(() => JSON.stringify(__maesage.G.hogs.map(h => [h.state, +h.x.toFixed(2), +h.y.toFixed(2)]))));
}
if (mode === 'golf') {
  const p = await page(1280, 800);
  await p.evaluate(() => __maesage.start('golf')); await p.waitForTimeout(500);
  for (let n = 0; n < 14; n++) {
    const q = await p.evaluate(() => { const G = __maesage.G; if (!G || G.kind !== 'golf' || G.over) return null; if (G.ball.moving || G.wait) return 0; const b = G.ball, c = __maesage.HOLES[G.h].cup; const d = Math.hypot(c[0] - b.x, c[1] - b.y), len = Math.min(.5, d * 1.7 / 2.75 * 1.05); const a = __maesage.toS(b.x, b.y), z = __maesage.toS(b.x + (c[0] - b.x) / d * len, b.y + (c[1] - b.y) / d * len); return { a, z, h: G.h, strokes: G.strokes }; });
    if (q === null) break; if (q === 0) { await p.waitForTimeout(300); n--; continue; }
    await p.mouse.move(q.a[0], q.a[1]); await p.mouse.down(); await p.mouse.move((q.a[0] + q.z[0]) / 2, (q.a[1] + q.z[1]) / 2, { steps: 6 }); await p.waitForTimeout(250);
    if (n === 0) await p.screenshot({ path: 'golf-aim.png' });
    await p.mouse.move(q.z[0], q.z[1], { steps: 8 }); await p.waitForTimeout(350); await p.mouse.up(); await p.waitForTimeout(700);
    if (n === 1) await p.screenshot({ path: 'golf-roll.png' });
    console.log('stroke', JSON.stringify(q.h), q.strokes);
  }
  await p.waitForTimeout(2500); await p.screenshot({ path: 'golf-end.png' });
  console.log('golf end', await p.evaluate(() => JSON.stringify({ screen: __maesage.S.screen, res: __maesage.S.results.golf, t: __maesage.S.tension })));
}
if (mode === 'call') {
  let p = await page(390, 800, true);
  await p.evaluate(() => __maesage.start('call')); await p.waitForTimeout(1200); await p.screenshot({ path: 'call-phone.png' });
  p = await page(390, 800, true);
  await p.evaluate(() => __maesage.start('moles')); await p.waitForTimeout(3200); await p.screenshot({ path: 'moles-phone.png' });
  await p.evaluate(() => __maesage.stop()); await p.waitForTimeout(300); await p.screenshot({ path: 'stopped-phone.png' });
  p = await page(1024, 640);
  await p.evaluate(() => __maesage.start('call')); await p.waitForTimeout(1500); await p.screenshot({ path: 'call-desk.png' });
}
console.log('errors:', errs.length ? errs : 'none');
await browser.close();
