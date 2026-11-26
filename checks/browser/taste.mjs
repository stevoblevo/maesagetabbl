import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
for (const [w, h, name] of [[1280, 800, 'desk'], [390, 800, 'phone']]) {
  const p = await (await browser.newContext({ viewport: { width: w, height: h } })).newPage(); const errs = [];
  p.on('pageerror', e => errs.push(e.message)); p.on('console', m => { if (m.type() === 'error' && !/Failed to load resource|fonts|ERR_/.test(m.text())) errs.push('console: ' + m.text()); });
  await p.goto('file:///home/claude/test/index.html'); await p.waitForTimeout(1200);
  if (name === 'desk') await p.mouse.move(520, 430); await p.waitForTimeout(500);
  await p.screenshot({ path: `kit-lobby-${name}.png` });
  const lobby = await p.evaluate(() => Depthfield.taste({ log: false }).map(f => `${f.severity} ${f.rule} ×${f.count}: ${f.sample} -> ${f.fix}`));
  console.log(`--- ${name} lobby taste:`, lobby.length ? '\n  ' + lobby.join('\n  ') : 'clean');
  await p.evaluate(() => __maesage.start('moles')); await p.waitForTimeout(2800); await p.screenshot({ path: `kit-moles-${name}.png` });
  const moles = await p.evaluate(() => Depthfield.taste({ log: false }).map(f => `${f.severity} ${f.rule} ×${f.count}: ${f.sample}`));
  console.log(`--- ${name} moles taste:`, moles.length ? '\n  ' + moles.join('\n  ') : 'clean');
  if (name === 'desk') { await p.click('#leave'); await p.click('#sheetOpen'); await p.waitForTimeout(500); await p.screenshot({ path: 'kit-sheet.png' });
    const sheet = await p.evaluate(() => Depthfield.taste({ log: false }).map(f => `${f.severity} ${f.rule} ×${f.count}: ${f.sample}`)); console.log('--- sheet taste:', sheet.length ? '\n  ' + sheet.join('\n  ') : 'clean'); }
  console.log(name, 'errors:', errs.length ? errs : 'none', '| theme:', await p.evaluate(() => document.documentElement.dataset.theme), '| version:', await p.evaluate(() => Depthfield.version));
}
await browser.close();
