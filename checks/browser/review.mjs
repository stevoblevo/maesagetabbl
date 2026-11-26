import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const errs = [], out = {};
const launch = () => chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const b1 = await launch(), b2 = await launch(), b3 = await launch();
const watch = (n, p) => { p.on('pageerror', e => errs.push(n + ': ' + e.message)); p.on('console', m => { if (m.type() === 'error' && !/Failed to load resource|fonts|ERR_/.test(m.text())) errs.push(n + ' console: ' + m.text()); }); };
const A = await (await b1.newContext({ viewport: { width: 1280, height: 800 } })).newPage(); watch('A', A);
await A.addInitScript(() => { window.__osc = { start: 0, stop: 0 }; const P = OscillatorNode.prototype, s0 = P.start, s1 = P.stop; P.start = function (...a) { if (this.type === 'square') __osc.start++; return s0.apply(this, a); }; P.stop = function (...a) { if (this.type === 'square') __osc.stop++; return s1.apply(this, a); }; });
const B = await (await b2.newContext({ viewport: { width: 390, height: 800 }, ignoreHTTPSErrors: true, hasTouch: true })).newPage(); watch('B', B);
await A.goto('http://127.0.0.1:8787/'); await B.goto('https://127.0.0.1:8788/#sense'); await B.waitForTimeout(900);
const breatheAt = bpm => B.evaluate(bpm => { clearInterval(window.__br); const g = 9.80665, t0 = performance.now(); window.__br = setInterval(() => { const t = (performance.now() - t0) / 1000, tilt = 1.2 * Math.PI / 180 * Math.sin(2 * Math.PI * bpm / 60 * t); dispatchEvent(new DeviceMotionEvent('devicemotion', { accelerationIncludingGravity: { x: g * Math.sin(tilt), y: .3, z: g * Math.cos(tilt) } })); }, 20); }, bpm);
const sensed = p => p.evaluate(() => { const s = __maesage.S.sensed; return s ? { bpm: +s.bpm.toFixed(1), seconds: +s.seconds.toFixed(0) } : null; });

// 3. the gun's purr must stop when sound is switched off mid-hold
await A.click('[data-df-tool="sound"]'); await A.evaluate(() => __maesage.motor(true)); await A.waitForTimeout(150);
const running = await A.evaluate(() => ({ ...__osc, mot: !!__maesage.mot }));
await A.click('[data-df-tool="sound"]'); await A.waitForTimeout(150);
out.motor = { running, afterSoundOff: await A.evaluate(() => ({ ...__osc, mot: !!__maesage.mot })) };
await A.evaluate(() => __maesage.motor(true)); out.motor.onWhileMuted = await A.evaluate(() => ({ ...__osc, mot: !!__maesage.mot }));

// 1. a restart is a new session
await A.click('.spot[data-game="moles"]');
await B.click('#feel'); await breatheAt(9); await B.waitForTimeout(24000);
out.first = await sensed(A);
await B.click('#feel'); await B.waitForTimeout(1500); await B.click('#feel'); await breatheAt(19); await B.waitForTimeout(1200);
out.justRestarted = await sensed(A);
await B.waitForTimeout(11000); out.elevenIn = await B.evaluate(() => document.querySelector('#hint').textContent);
out.elevenInA = await sensed(A);
await B.waitForTimeout(16000); out.twentySevenIn = await sensed(A);

// 4. switching role from a sensing phone stops the sensing
out.role = { before: await B.evaluate(() => __maesage.feeling) };
await B.evaluate(() => __maesage.enterTable()); await B.waitForTimeout(300);
out.role.afterTable = await B.evaluate(() => ({ feeling: __maesage.feeling, button: document.querySelector('#feel').textContent, hash: location.hash }));

// 2. a screen that joins after the phone stopped is not handed the old reading as new
await A.waitForTimeout(2000);
const C = await (await b3.newContext({ viewport: { width: 1280, height: 800 } })).newPage(); watch('C', C);
await C.goto('http://127.0.0.1:8787/'); await C.waitForTimeout(900);
out.joinedTwoSecondsAfterStop = await sensed(C);
await C.waitForTimeout(4200); out.sameScreenFourSecondsLater = await sensed(C);
await C.reload(); await C.waitForTimeout(900); out.reloadedLater = await sensed(C);
out.A_afterStop = await A.evaluate(() => ({ sensed: __maesage.S.sensed, who: document.querySelector('#who').textContent }));
console.log(JSON.stringify(out, null, 1)); console.log('errors:', errs.length ? errs : 'none');
await b1.close(); await b2.close(); await b3.close();
