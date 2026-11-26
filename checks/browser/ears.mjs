import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const p = await (await browser.newContext({ viewport: { width: 1280, height: 800 }, deviceScaleFactor: 2 })).newPage();
const errs = []; p.on('pageerror', e => errs.push(e.message));
await p.goto('http://127.0.0.1:8787/'); await p.waitForTimeout(1200);
await p.evaluate(() => __maesage.start('hogs')); await p.mouse.move(900, 300);
for (const [n, v] of [['braced', .95], ['calm', .3], ['puddle', .02]]) {
  await p.evaluate(v => { __maesage.S.tension = v; __maesage.S.read = v; }, v); await p.waitForTimeout(2200);
  await p.screenshot({ path: `ears-${n}.png`, clip: { x: 60, y: 250, width: 520, height: 420 } }); const g = await p.evaluate(() => { const r = document.querySelector('.df-guardian').getBoundingClientRect(); return { x: r.x - 20, y: r.y - 20, width: r.width + 40, height: r.height + 40 }; }); await p.screenshot({ path: `berry-${n}.png`, clip: g });
  console.log(n, await p.evaluate(() => ({ word: document.querySelector('#word').textContent, doze: getComputedStyle(document.querySelector('.df-guardian')).getPropertyValue('--doze') })));
}
console.log('errors:', errs.length ? errs : 'none'); await browser.close();
