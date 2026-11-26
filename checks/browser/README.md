# Browser checks (archive, not on `main`)

These are the scripts that drove a headless Chrome through the page while it was being built: every game end to end, keyboard play, Stop, two screens, the watcher, the breath path over https, Taste, and the review fixes. They are kept on this branch so the work is not lost.

They are **not** on `main` because the project's rule is Go, Rust or Erlang for tooling, and these are Node with Playwright. The plan is to port them to Go (see the issue "Port the browser checks to Go").

They assume the sandbox they were written in:

- Playwright at `/opt/npm-tools/node_modules/playwright`, Chrome at `/usr/bin/google-chrome`.
- The server built to `/tmp/tabble` and run as `/tmp/tabble -addr 127.0.0.1:8787 -tls 127.0.0.1:8788 -certs /tmp/certs -web <a copy of web/>`.
- That copy has one extra line before the page script's closing `})();`, which exposes the page's internals as `window.__maesage`. `mk.sh` makes the copy and adds the line (it uses Python for the edit, which is another reason this is an archive).

| Script | What it checks |
|---|---|
| `smoke.mjs` | The page loads with no errors. |
| `run.mjs` | Begin plays Moles, Hedgehogs and Golf to the end card; keyboard play; Stop and resume; the sound tool. |
| `misc.mjs`, `hum.mjs`, `faces.mjs` | Layouts, the hum, the night, day and contrast faces. |
| `flip.mjs`, `flipgo.mjs` | Table and player linked: a placed spot lights, a trick is not answered, remote Stop latches. |
| `watch.mjs` | The watching screen mirrors each game and the four weathers. |
| `sense.mjs` | A phone's motion over https becomes a breath reading on the player's screen, and falls back when it stops. |
| `review.mjs` | The four review findings on #11 and #12: session restart, stale reading, the gun's purr, role switch. |
| `taste.mjs` | `Depthfield.taste()` at desktop and phone widths. |
| `csp.mjs` | The published files under a content policy like the claude.ai artifact's. |
| `suit.mjs`, `ears.mjs`, `hero.mjs`, `shot.mjs` | Screenshots. |
