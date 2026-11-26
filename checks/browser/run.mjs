import { chromium } from '/opt/npm-tools/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', args: ['--no-sandbox'] });
const errs = [];
const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
const p = await ctx.newPage();
p.on('pageerror', e => errs.push('pageerror: ' + e.message));
await p.goto('file:///home/claude/test/index.html'); await p.waitForTimeout(600);
await p.click('#begin'); await p.waitForTimeout(500);
console.log('1', await p.evaluate(() => __maesage.S.screen));
await p.evaluate(() => { const G = __maesage.G; G.i = G.seq.length; G.cues.length = 0; G.hit = 12; __maesage.S.tension = .4; });
await p.waitForTimeout(700); await p.screenshot({ path: 'run-between.png' }); await p.waitForTimeout(2200);
console.log('2', await p.evaluate(() => __maesage.S.screen));
await p.evaluate(() => { const G = __maesage.G; G.hogs.forEach(h => h.state = 'gone'); G.curled = 5; __maesage.S.tension = .2; });
await p.waitForTimeout(3000);
console.log('3', await p.evaluate(() => __maesage.S.screen));
await p.evaluate(() => { const G = __maesage.G; G.h = 2; G.total = 7; G.ball.x = .3; G.ball.y = .15; });
// real last stroke toward the cup
const q = await p.evaluate(() => ({ a: __maesage.toS(.3, .15), z: __maesage.toS(.3, .09) }));
await p.mouse.move(q.a[0], q.a[1]); await p.mouse.down(); await p.mouse.move(q.z[0], q.z[1], { steps: 10 }); await p.waitForTimeout(300); await p.mouse.up();
await p.waitForTimeout(900); await p.screenshot({ path: 'run-sunk.png' });
await p.waitForTimeout(3200); await p.screenshot({ path: 'run-end.png' });
console.log('4', await p.evaluate(() => JSON.stringify({ sc: __maesage.S.screen, res: __maesage.S.results, t: __maesage.S.tension })));
// keyboard: back to table, start moles by keyboard, move the hand with arrows, press space
await p.click('#back'); await p.waitForTimeout(300);
await p.focus('.spot[data-game="moles"]'); await p.keyboard.press('Enter'); await p.waitForTimeout(400);
await p.keyboard.down('ArrowRight'); await p.waitForTimeout(500); await p.keyboard.up('ArrowRight'); await p.keyboard.down(' '); await p.waitForTimeout(300);
console.log('kb', await p.evaluate(() => JSON.stringify({ sc: __maesage.S.screen, down: __maesage.P.down, inside: __maesage.P.inside, bx: +__maesage.P.bx.toFixed(2) })));
await p.keyboard.up(' ');
await p.keyboard.press('Escape'); await p.waitForTimeout(300); await p.screenshot({ path: 'run-stopped.png' });
console.log('stopped', await p.evaluate(() => __maesage.S.stopped));
await p.click('#resume'); await p.waitForTimeout(200); console.log('resumed', await p.evaluate(() => !__maesage.S.stopped));
// sound toggle should not throw
await p.click('[data-df-tool="sound"]'); await p.waitForTimeout(300); console.log('sound label', await p.getAttribute('[data-df-tool="sound"]', 'aria-pressed'));
console.log('errors:', errs.length ? errs : 'none');
await browser.close();
