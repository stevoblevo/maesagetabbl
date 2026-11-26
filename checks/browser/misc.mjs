import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const errs = [];
for (const [w, h, rm, name] of [[1600, 900, false, 'wide'], [820, 1180, true, 'tablet-still'], [844, 390, false, 'phone-land']]) {
  const ctx = await browser.newContext({ viewport: { width: w, height: h }, reducedMotion: rm ? 'reduce' : 'no-preference' });
  const p = await ctx.newPage(); p.on('pageerror', e => errs.push(name + ': ' + e.message));
  await p.goto('file:///home/claude/test/index.html'); await p.waitForTimeout(500);
  await p.screenshot({ path: `misc-${name}-lobby.png` });
  for (const g of ['moles', 'hogs', 'golf', 'call']) { await p.evaluate(g => __maesage.start(g), g); await p.waitForTimeout(g === 'moles' ? 2600 : 700); if (g === 'moles' || g === 'call') await p.screenshot({ path: `misc-${name}-${g}.png` }); }
  await ctx.close();
}
console.log('errors:', errs.length ? errs : 'none');
await browser.close();
