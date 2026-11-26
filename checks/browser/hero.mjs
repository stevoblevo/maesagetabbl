import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const p = await (await browser.newContext({ viewport: { width: 1280, height: 800 }, deviceScaleFactor: 2 })).newPage();
await p.goto('http://127.0.0.1:8787/'); await p.waitForTimeout(1200);
await p.mouse.move(1150, 640); await p.waitForTimeout(7000); await p.mouse.move(1150, 600); await p.waitForTimeout(600);
await p.screenshot({ path: 'hero-lobby.png' });
await p.evaluate(() => __maesage.start('moles')); await p.mouse.move(420, 372); await p.waitForTimeout(3200); await p.screenshot({ path: 'hero-moles.png' });
await p.evaluate(() => __maesage.start('golf')); await p.mouse.move(700, 470); await p.waitForTimeout(1800); await p.screenshot({ path: 'hero-golf.png' });
await browser.close();
