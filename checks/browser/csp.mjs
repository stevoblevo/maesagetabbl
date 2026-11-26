import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
import { readFileSync } from 'node:fs';
const WEB = '/home/claude/maesagetabbl/web/';
const CSP = "default-src 'none'; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com data:; img-src data: blob:; media-src data: blob:; connect-src 'none'; worker-src 'self' blob:";
const types = { html: 'text/html', css: 'text/css', js: 'text/javascript' };
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const p = await (await browser.newContext({ viewport: { width: 1280, height: 800 } })).newPage();
const out = { errors: [], console: [], outside: [], served: [] };
p.on('pageerror', e => out.errors.push(e.message));
p.on('console', m => { if (m.type() === 'error' || /Content Security|Refused/.test(m.text())) out.console.push(m.text().slice(0, 200)); });
await p.route('**/*', route => {
  const u = new URL(route.request().url());
  if (u.host !== 'artifact.test') { out.outside.push(u.host + u.pathname.slice(0, 40)); return route.abort(); }
  const path = u.pathname === '/' ? 'index.html' : u.pathname.slice(1);
  try { const body = readFileSync(WEB + path); out.served.push(path); route.fulfill({ status: 200, body, headers: { 'content-type': types[path.split('.').pop()] || 'application/octet-stream', ...(path === 'index.html' ? { 'content-security-policy': CSP } : {}) } }); }
  catch { out.served.push('404 ' + path); route.fulfill({ status: 404, body: '' }); }
});
await p.goto('https://artifact.test/'); await p.waitForTimeout(1200);
const kit = await p.evaluate(() => ({ df: !!window.Depthfield && Depthfield.version, theme: document.documentElement.dataset.theme, noJs: document.body.classList.contains('df-no-js'), word: document.querySelector('#word')?.textContent, spots: document.querySelectorAll('.spot').length, guardian: !!document.querySelector('.df-guardian'), icon: !!document.querySelector('[data-df-icon] svg') }));
await p.click('.spot[data-game="moles"]'); await p.waitForTimeout(900);
await p.mouse.move(600, 420); await p.mouse.down(); await p.waitForTimeout(500); await p.mouse.up();
const hint = await p.textContent('#hint');
await p.click('#sheetOpen'); await p.waitForTimeout(300);
const sheet = await p.evaluate(() => !document.querySelector('#sheet').hidden);
await p.screenshot({ path: 'csp-sheet.png' });
console.log(JSON.stringify({ kit, hint, sheet, ...out }, null, 1));
await browser.close();
