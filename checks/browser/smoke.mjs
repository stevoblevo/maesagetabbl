import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const errs = []; const p = await (await browser.newContext({ viewport: { width: 1280, height: 800 } })).newPage();
p.on('pageerror', e => errs.push(e.message));
await p.goto('file:///home/claude/test/index.html'); await p.waitForTimeout(400);
await p.click('#begin'); await p.waitForTimeout(2500); await p.mouse.move(600, 420); await p.mouse.down(); await p.waitForTimeout(600); await p.mouse.up();
await p.click('#leave'); await p.waitForTimeout(200);
for (const g of ['hogs', 'golf', 'call']) { await p.click(`.spot[data-game="${g}"]`); await p.waitForTimeout(500); await p.mouse.move(640, 430); await p.mouse.down(); await p.mouse.move(700, 440, { steps: 5 }); await p.mouse.up(); await p.waitForTimeout(300); await p.click('#leave'); }
console.log('title:', await p.title(), '| hint:', await p.textContent('#hint'), '| errors:', errs.length ? errs : 'none');
await browser.close();
