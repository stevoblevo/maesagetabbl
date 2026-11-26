import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const errs = [];
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
async function page(w, h, touch) { const p = await (await browser.newContext({ viewport: { width: w, height: h }, hasTouch: !!touch, deviceScaleFactor: 2 })).newPage(); p.on('pageerror', e => errs.push(e.message)); p.on('console', m => { if (m.type() === 'error' && !/Failed to load resource|fonts|ERR_/.test(m.text())) errs.push(m.text()); }); await p.goto('http://127.0.0.1:8787/'); await p.waitForTimeout(1500); return p; }
const A = await page(1280, 800);
await A.mouse.move(520, 430); await A.waitForTimeout(700); await A.screenshot({ path: 'suit-lobby.png' });
for (const g of ['moles', 'hogs', 'golf', 'call']) { await A.evaluate(g => __maesage.start(g), g); await A.mouse.move(600, 400); await A.waitForTimeout(g === 'moles' ? 2600 : 1500); await A.screenshot({ path: `suit-${g}.png` }); }
await A.evaluate(() => __maesage.enterTable()); await A.waitForTimeout(600); await A.screenshot({ path: 'suit-table.png' });
const B = await page(390, 800, true); await B.screenshot({ path: 'suit-phone.png' });
console.log('errors:', errs.length ? errs : 'none');
await browser.close();
