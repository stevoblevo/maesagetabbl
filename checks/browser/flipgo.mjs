import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const errs = [];
const stub = () => { // a stand-in for the platform's room, over BroadcastChannel
  const id = Math.random().toString(36).slice(2), ch = new BroadcastChannel('room'), all = new Map(), subs = [];
  let mine = {};
  const fire = () => { const peers = [{ peer: id, presence: mine, sameTab: true, isMe: true, kind: 'viewer', guest: false, updatedAt: Date.now() }, ...all.values()]; subs.forEach(f => f({ peers, joined: [], left: [], updated: [] })); };
  ch.onmessage = e => { if (e.data.hello) { ch.postMessage({ id, presence: mine }); return; } all.set(e.data.id, { peer: e.data.id, presence: e.data.presence, sameTab: false, isMe: true, kind: 'viewer', guest: false, updatedAt: Date.now() }); fire(); };
  const room = { presence: async patch => { mine = { ...mine }; for (const k in patch) { if (patch[k] === null) delete mine[k]; else mine[k] = patch[k]; } ch.postMessage({ id, presence: mine }); fire(); }, onPeers: fn => { subs.push(fn); setTimeout(fire, 0); return () => {}; } };
  window.claude = { use: async n => n === 'room' ? room : null }; setTimeout(() => ch.postMessage({ hello: 1 }), 50);
};
const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });

const A = await ctx.newPage(), B = await ctx.newPage();
for (const [n, p] of [['A', A], ['B', B]]) p.on('pageerror', e => errs.push(n + ': ' + e.message));
await B.setViewportSize({ width: 390, height: 800 });
await A.goto('http://127.0.0.1:8787/'); await B.goto('http://127.0.0.1:8787/'); await A.waitForTimeout(500);
await A.screenshot({ path: 'go-flip-lobby.png' });
// solo rehearsal first
await A.click('.spot[data-game="flip"]'); await A.waitForTimeout(3200); await A.screenshot({ path: 'go-flip-rehearsal.png' });
console.log('rehearsal', await A.evaluate(() => JSON.stringify({ live: __maesage.G.live, cues: __maesage.G.cues.length, hint: document.querySelector('#hint').textContent })));
// B goes on the table and asks
await B.bringToFront(); await B.click('#sheetOpen'); await B.waitForTimeout(300); await B.click('#role'); await B.waitForTimeout(400);
await B.click('.pad[data-set="3"]');
const spot = await B.evaluate(() => __maesage.toS(-.44, .095)); await B.mouse.click(spot[0], spot[1]); await B.waitForTimeout(250);
await B.click('#trick'); const spot2 = await B.evaluate(() => __maesage.toS(.3, -.084)); await B.mouse.click(spot2[0], spot2[1]); await B.waitForTimeout(250);
const spine = await B.evaluate(() => __maesage.toS(-.3, 0)); await B.mouse.click(spine[0], spine[1]); await B.waitForTimeout(500);
await B.screenshot({ path: 'go-flip-table-phone.png' });
console.log('B placed', await B.evaluate(() => JSON.stringify(__maesage.T.cues.map(c => [c.id.slice(-2), c.kind, c.set]))), '| hint:', await B.textContent('#hint'));
await A.bringToFront(); await A.waitForTimeout(900);
console.log('A sees', await A.evaluate(() => JSON.stringify({ live: __maesage.G.live, cues: __maesage.G.cues.filter(c => c.id).map(c => [c.kind, c.set]) })));
const t = await A.evaluate(() => { const c = __maesage.G.cues.find(c => c.id && c.kind === 'real'); return __maesage.toS(c.z.x, c.z.y); });
await A.mouse.move(t[0], t[1]); await A.waitForTimeout(150); await A.mouse.down(); await A.waitForTimeout(500); await A.screenshot({ path: 'go-flip-play.png' }); await A.waitForTimeout(1300); await A.mouse.up(); await A.waitForTimeout(400);
console.log('A hit', await A.evaluate(() => JSON.stringify({ hit: __maesage.G.hit, hint: document.querySelector('#hint').textContent, who: document.querySelector('#who').textContent, word: document.querySelector('#word').textContent })));
await B.bringToFront(); await B.waitForTimeout(700);
console.log('B after', await B.evaluate(() => JSON.stringify(__maesage.T.cues.map(c => [c.kind, !!c.won]))));
// ceiling and stop travel from the table
await B.evaluate(() => { const i = document.querySelector('#ceilIn'); i.value = 2; i.dispatchEvent(new Event('input')); }); await B.click('#stop'); await B.waitForTimeout(300); await A.bringToFront(); await A.waitForTimeout(600);
await A.screenshot({ path: 'go-flip-stopped-remote.png' });
console.log('A stopped', await A.evaluate(() => JSON.stringify({ stopped: __maesage.S.stopped, by: __maesage.S.stopBy, ceil: __maesage.S.ceiling, resumeHidden: document.querySelector('#resume').hidden })));
await B.bringToFront(); await B.click('#resume'); await B.waitForTimeout(300); await A.bringToFront(); await A.waitForTimeout(600);
console.log('A resumed', await A.evaluate(() => !__maesage.S.stopped));
console.log('errors:', errs.length ? errs : 'none');
await browser.close();
