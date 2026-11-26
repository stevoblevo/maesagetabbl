import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const errs = [];
const b1 = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] }), b2 = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const A = await (await b1.newContext({ viewport: { width: 1280, height: 800 } })).newPage();
const B = await (await b2.newContext({ viewport: { width: 390, height: 800 }, ignoreHTTPSErrors: true, hasTouch: true })).newPage();
for (const [n, p] of [['A', A], ['B', B]]) { p.on('pageerror', e => errs.push(n + ': ' + e.message)); p.on('console', m => { if (m.type() === 'error' && !/Failed to load resource|fonts|ERR_/.test(m.text())) errs.push(n + ' console: ' + m.text()); }); }
await A.goto('http://127.0.0.1:8787/'); await B.goto('https://127.0.0.1:8788/#sense'); await B.waitForTimeout(900);
console.log('B before:', await B.textContent('#hint'), '| secure:', await B.evaluate(() => isSecureContext), '| disabled:', await B.evaluate(() => document.querySelector('#feel').disabled));
await B.screenshot({ path: 'sense-phone-idle.png' });
await B.click('#feel');
// a back breathing 11 times a minute under the phone, for as long as the page lives
await B.evaluate(() => { const g = 9.80665, t0 = performance.now(); setInterval(() => { const t = (performance.now() - t0) / 1000, tilt = 1.2 * Math.PI / 180 * Math.sin(2 * Math.PI * 11 / 60 * t), n = () => (Math.random() - .5) * .04; dispatchEvent(new DeviceMotionEvent('devicemotion', { accelerationIncludingGravity: { x: g * Math.sin(tilt) + n(), y: .3 + n(), z: g * Math.cos(tilt) + n() }, interval: 20 })); }, 20); });
await B.waitForTimeout(4000); console.log('B at 4 s:', await B.textContent('#hint'));
await A.click('.spot[data-game="moles"]');
await B.waitForTimeout(12000);
console.log('B at 16 s:', await B.textContent('#hint'), '| who:', await B.textContent('#who'), '| word:', await B.textContent('#word'));
await B.screenshot({ path: 'sense-phone-live.png' });
console.log('A:', await A.evaluate(() => JSON.stringify({ sensed: __maesage.S.sensed, tension: +__maesage.S.tension.toFixed(3), who: document.querySelector('#who').textContent, breath: document.querySelector('#breath').textContent, wx: document.querySelector('#wx').textContent })));
await A.screenshot({ path: 'sense-player.png' });
// the breath slows to 7; the player's screen should follow without anyone scoring a point
await B.evaluate(() => { for (let i = 1; i < 99999; i++) clearInterval(i); const g = 9.80665, t0 = performance.now(); setInterval(() => { const t = (performance.now() - t0) / 1000, tilt = 1.2 * Math.PI / 180 * Math.sin(2 * Math.PI * 7 / 60 * t); dispatchEvent(new DeviceMotionEvent('devicemotion', { accelerationIncludingGravity: { x: g * Math.sin(tilt), y: .3, z: g * Math.cos(tilt) }, interval: 20 })); }, 20); });
await B.waitForTimeout(45000);
console.log('A after slowing:', await A.evaluate(() => JSON.stringify({ bpm: __maesage.S.sensed && +__maesage.S.sensed.bpm.toFixed(1), settle: __maesage.S.sensed && +__maesage.S.sensed.settle.toFixed(2), word: document.querySelector('#word').textContent, wx: document.querySelector('#wx').textContent })));
await A.screenshot({ path: 'sense-player-slow.png' });
// the phone stops: the player falls back to the rehearsal and says so
await B.click('#feel'); await A.waitForTimeout(8000);
console.log('A after stop:', await A.evaluate(() => JSON.stringify({ sensed: __maesage.S.sensed, who: document.querySelector('#who').textContent })));
console.log('errors:', errs.length ? errs : 'none');
await b1.close(); await b2.close();
