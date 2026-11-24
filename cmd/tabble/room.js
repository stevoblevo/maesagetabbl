// Stand-in for the claude.ai room when the page is served by cmd/tabble on your own network.
// Same two calls the page uses: presence(patch) and onPeers(fn). Presence is absolute state.
// One call more than the platform has: feel(ms, samples) posts a phone's accelerometer for the breath reading.
(() => {
  if (window.claude && window.claude.use) return;
  const peer = Array.from(crypto.getRandomValues(new Uint8Array(8)), b => b.toString(16).padStart(2, '0')).join('');
  let mine = {}, peers = [], timer = null;
  const subs = new Set();
  const tell = () => subs.forEach(fn => fn({ peers, joined: [], left: [], updated: [] }));
  const es = new EventSource('/room/events?peer=' + peer);
  es.onmessage = e => {
    try {
      // a reading is the server's own word about a phone's breath signal; it rides in that phone's presence,
      // with its age in milliseconds when the server sent it (screens do not share a clock)
      peers = JSON.parse(e.data).map(p => ({ peer: p.peer, by: null, isMe: p.peer === peer, sameTab: p.peer === peer, kind: 'viewer', guest: false, presence: p.reading ? { ...(p.presence || {}), reading: { ...p.reading, age: p.readingAge || 0 } } : p.presence || {}, updatedAt: p.updatedAt }));
      tell();
    } catch (err) {}
  };
  const send = () => { timer = null; fetch('/room/presence?peer=' + peer, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(mine) }).catch(() => {}); };
  es.onopen = send; // the server forgets a peer when its stream drops, so say it again on reconnect
  const room = {
    presence: async patch => { mine = { ...mine }; for (const k in patch) { if (patch[k] === null) delete mine[k]; else mine[k] = patch[k]; } if (!timer) timer = setTimeout(send, 40); },
    onPeers: fn => { subs.add(fn); queueMicrotask(() => fn({ peers, joined: [], left: [], updated: [] })); return () => subs.delete(fn); },
    peers: () => peers,
    connected: () => es.readyState === 1,
    // Not part of the claude.ai room: only this server can turn motion into breath.
    // fresh marks the first post of a sensing session, so the server starts a new estimate.
    feel: (ms, samples, fresh) => fetch('/sense/samples?peer=' + peer, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(fresh ? { ms, s: samples, new: true } : { ms, s: samples }) }).catch(() => {}),
  };
  window.claude = { use: async name => name === 'room' ? room : null };
})();
