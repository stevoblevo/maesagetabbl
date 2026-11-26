import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const errs = [];
const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
const A = await ctx.newPage(), C = await ctx.newPage();
for (const [n, p] of [['A', A], ['C', C]]) { p.on('pageerror', e => errs.push(n + ': ' + e.message)); p.on('console', m => { if (m.type() === 'error' && !/Failed to load resource|fonts/.test(m.text())) errs.push(n + ' console: ' + m.text()); }); }
await C.setViewportSize({ width: 1600, height: 900 });
await A.goto('http://127.0.0.1:8787/'); await C.goto('http://127.0.0.1:8787/#watch'); await C.waitForTimeout(900);
console.log('C role', await C.evaluate(() => JSON.stringify({ sc: __maesage.S.screen, hint: document.querySelector('#hint').textContent })));
await C.screenshot({ path: 'watch-idle.png' });
// A plays moles; both pages stay alive (rAF needs a visible page, so alternate focus)
await A.bringToFront(); await A.click('.spot[data-game="moles"]'); await A.waitForTimeout(2600);
const q = await A.evaluate(() => { const c = __maesage.G.cues.find(c => !c.out && c.kind !== 'tease'); return __maesage.toS(c.z.x, c.z.y); });
await A.mouse.move(q[0], q[1]); await A.mouse.down(); await A.waitForTimeout(400);
const a1 = await A.evaluate(() => JSON.stringify({ cues: __maesage.G.cues.length, tuned: __maesage.G.tuned }));
await C.bringToFront(); await C.waitForTimeout(500);
console.log('moles A', a1, '| C', await C.evaluate(() => JSON.stringify({ g: __maesage.G && __maesage.G.kind, cues: __maesage.G && __maesage.G.cues.length, tuned: __maesage.G && __maesage.G.tuned, down: __maesage.P.down, title: document.querySelector('#title').textContent })));
await C.screenshot({ path: 'watch-moles.png' });
await A.bringToFront(); await A.mouse.up(); await A.click('#leave'); await A.click('.spot[data-game="hogs"]'); await A.waitForTimeout(700);
await C.bringToFront(); await C.waitForTimeout(500);
console.log('hogs C', await C.evaluate(() => JSON.stringify({ g: __maesage.G && __maesage.G.kind, hogs: __maesage.G && __maesage.G.hogs.length })));
await A.bringToFront(); await A.click('#leave'); await A.click('.spot[data-game="golf"]'); await A.waitForTimeout(600);
await C.bringToFront(); await C.waitForTimeout(500);
console.log('golf C', await C.evaluate(() => JSON.stringify({ g: __maesage.G && __maesage.G.kind, ball: __maesage.G && __maesage.G.ball })));
await A.bringToFront(); await A.click('#leave'); await A.click('.spot[data-game="call"]'); await A.waitForTimeout(600);
await C.bringToFront(); await C.waitForTimeout(600);
console.log('call C', await C.evaluate(() => JSON.stringify({ g: __maesage.G && __maesage.G.kind, title: document.querySelector('#title').textContent, eb: document.querySelector('#eyebrow').textContent })));
await C.screenshot({ path: 'watch-call.png' });
// stop on the player shows on the watcher
await A.bringToFront(); await A.click('#stop'); await A.waitForTimeout(300); await C.bringToFront(); await C.waitForTimeout(500);
console.log('C stopped', await C.evaluate(() => __maesage.S.stopped)); await A.bringToFront(); await A.click('#resume'); await A.waitForTimeout(200); await C.bringToFront(); await C.waitForTimeout(500);
console.log('C resumed', await C.evaluate(() => !__maesage.S.stopped));
// weather at four states, on the player
await A.bringToFront(); await A.click('#leave');
for (const [name, t] of [['rain', .9], ['peach', .45], ['mint', .15], ['night', .02]]) { await A.evaluate(t => { __maesage.S.tension = t; __maesage.S.read = t; }, t); await A.waitForTimeout(2600); await A.screenshot({ path: `wx-${name}.png` }); console.log(name, await A.textContent('#wx'), await A.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--accent'))); }
console.log('errors:', errs.length ? errs : 'none');
await browser.close();
