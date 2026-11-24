// Command tabble serves the Tabble games on your own network and links the screens that open them.
//
// The page is the same file the claude.ai artifact runs. There it uses the platform's room; here the
// server injects a small stand-in (room.js) and relays presence between screens: every screen posts
// its whole presence object, and every screen hears the whole room over one event stream.
// Nothing is stored. When a screen's stream drops, it leaves the room.
//
// One screen can be a sensor: a phone lying on the receiver's back posts its accelerometer, the
// server turns that into a breath reading (internal/breath), and the reading rides along with
// that phone in the room. Browsers only give a page its motion sensor over https, so the server
// also listens on a second port with a certificate it makes for itself.
package main

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/stevoblevo/maesagetabbl/internal/breath"
)

//go:embed room.js
var roomJS []byte

const (
	maxPeers    = 64
	maxPresence = 4 << 10 // bytes of JSON, the same bound the platform's room uses
	maxSamples  = 400     // accelerometer samples in one post: a few seconds at any phone's rate

	// quiet is how long a phone may say nothing before its signal counts as broken. One post covers
	// at most five seconds, so a longer silence is a hole no later post can fill: the reading it left
	// is withdrawn, and the next samples start a new estimate.
	quiet = 6 * time.Second
)

// Peer is one open screen.
type Peer struct {
	Peer      string          `json:"peer"`
	Presence  json.RawMessage `json:"presence"`
	UpdatedAt int64           `json:"updatedAt"`
	Reading   *breath.Reading `json:"reading,omitempty"` // set by the server, never by the screen
	// ReadingAge is how old Reading was, in milliseconds, when this snapshot was made. Screens keep
	// their own clocks, so the server says an age and not a time.
	ReadingAge int64 `json:"readingAge,omitempty"`

	readAt time.Time
}

// sensor is one phone's breath estimator, how much signal it has taken since its last reading,
// and when it last heard from the phone.
type sensor struct {
	est   breath.Estimator
	since float64
	last  time.Time
}

// Room is everyone who has the page open right now.
type Room struct {
	mu      sync.Mutex
	peers   map[string]*Peer
	streams map[string]chan []byte
	sensors map[string]*sensor
	now     func() time.Time
}

func NewRoom() *Room {
	return &Room{peers: map[string]*Peer{}, streams: map[string]chan []byte{}, sensors: map[string]*sensor{}, now: time.Now}
}

// snapshot returns the whole room as JSON, oldest peer id first so the order is stable.
// A reading whose phone has gone quiet is withdrawn here, so a screen that joins or reloads
// later is never handed an old reading as if it were new.
func (r *Room) snapshot() []byte {
	now := r.now()
	list := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		if p.Reading != nil {
			if age := now.Sub(p.readAt); age > quiet {
				p.Reading, p.ReadingAge = nil, 0
			} else {
				p.ReadingAge = age.Milliseconds()
			}
		}
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Peer < list[j].Peer })
	b, _ := json.Marshal(list)
	return b
}

// broadcast sends the room to every stream. A stream that is not keeping up skips this frame;
// the next one carries the whole state anyway.
func (r *Room) broadcast() {
	snap := r.snapshot()
	for _, ch := range r.streams {
		select {
		case ch <- snap:
		default:
		}
	}
}

// Join opens a stream for a peer. It reports false when the room is full or the id is taken.
func (r *Room) Join(id string) (<-chan []byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, taken := r.streams[id]; taken || len(r.streams) >= maxPeers {
		return nil, false
	}
	ch := make(chan []byte, 8)
	r.streams[id] = ch
	if _, ok := r.peers[id]; !ok {
		r.peers[id] = &Peer{Peer: id, Presence: json.RawMessage(`{}`), UpdatedAt: time.Now().UnixMilli()}
	}
	r.broadcast()
	return ch, true
}

// Leave removes a peer and tells the rest.
func (r *Room) Leave(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.streams, id)
	delete(r.peers, id)
	delete(r.sensors, id)
	r.broadcast()
}

// Set replaces a peer's presence. Only a peer with an open stream may set one.
func (r *Room) Set(id string, presence json.RawMessage) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, here := r.streams[id]; !here {
		return false
	}
	p := r.peers[id]
	p.Presence, p.UpdatedAt = presence, time.Now().UnixMilli()
	r.broadcast()
	return true
}

// Feel takes accelerometer samples from a phone, each covering dt seconds. After every second
// of signal it re-reads the breath and tells the room. Only a peer with an open stream may feel.
//
// A sensing session ends when the phone says so (fresh, on the first post after "Start feeling")
// or when it has been quiet too long. Either way the old estimate is thrown away with its reading:
// samples from before a pause, or from another person's back, are not this signal.
func (r *Room) Feel(id string, dt float64, samples []breath.Sample, fresh bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, here := r.streams[id]; !here {
		return false
	}
	now, p := r.now(), r.peers[id]
	s := r.sensors[id]
	if s == nil || fresh || now.Sub(s.last) > quiet {
		s = &sensor{}
		r.sensors[id] = s
		if p.Reading != nil {
			p.Reading, p.ReadingAge, p.UpdatedAt = nil, 0, now.UnixMilli()
			r.broadcast()
		}
	}
	s.last = now
	for _, v := range samples {
		s.est.Add(dt, v)
	}
	s.since += dt * float64(len(samples))
	if s.since >= 1 {
		s.since = 0
		reading := s.est.Read()
		p.Reading, p.readAt, p.UpdatedAt = &reading, now, now.UnixMilli()
		r.broadcast()
	}
	return true
}

func validID(id string) bool {
	if len(id) < 8 || len(id) > 32 {
		return false
	}
	for _, c := range id {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func (r *Room) events(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("peer")
	fl, ok := w.(http.Flusher)
	if !ok || !validID(id) {
		http.Error(w, "bad stream", http.StatusBadRequest)
		return
	}
	ch, ok := r.Join(id)
	if !ok {
		http.Error(w, "room is full", http.StatusServiceUnavailable)
		return
	}
	defer r.Leave(id)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	fl.Flush()
	keep := time.NewTicker(20 * time.Second)
	defer keep.Stop()
	for {
		select {
		case <-req.Context().Done():
			return
		case snap := <-ch:
			w.Write([]byte("data: "))
			w.Write(snap)
			w.Write([]byte("\n\n"))
			fl.Flush()
		case <-keep.C:
			w.Write([]byte(": keep\n\n"))
			fl.Flush()
		}
	}
}

func (r *Room) presence(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("peer")
	if req.Method != http.MethodPost || !validID(id) {
		http.Error(w, "bad presence", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, req.Body, maxPresence))
	var obj map[string]json.RawMessage
	if err != nil || json.Unmarshal(body, &obj) != nil {
		http.Error(w, "presence must be a JSON object of at most 4 KiB", http.StatusBadRequest)
		return
	}
	if !r.Set(id, body) {
		http.Error(w, "open the event stream first", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// samples takes one post of accelerometer readings: {"ms": how long they cover, "s": [[x, y, z], ...]} in m/s².
// "new": true marks the first post after the phone was told to start feeling.
// The span is measured by the phone's own clock because phones disagree about what "interval" means.
func (r *Room) samples(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("peer")
	if req.Method != http.MethodPost || !validID(id) {
		http.Error(w, "bad samples", http.StatusBadRequest)
		return
	}
	var in struct {
		Ms  float64      `json:"ms"`
		S   [][3]float64 `json:"s"`
		New bool         `json:"new"`
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, req.Body, 32<<10))
	if err != nil || json.Unmarshal(body, &in) != nil || len(in.S) == 0 || len(in.S) > maxSamples || !(in.Ms > 0 && in.Ms <= 5000) {
		http.Error(w, "samples must be {ms, s} with 1 to 400 readings covering at most 5 s", http.StatusBadRequest)
		return
	}
	out := make([]breath.Sample, len(in.S))
	for i, v := range in.S {
		out[i] = breath.Sample{X: v[0], Y: v[1], Z: v[2]}
	}
	if !r.Feel(id, in.Ms/1000/float64(len(in.S)), out, in.New) {
		http.Error(w, "open the event stream first", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// withRoom puts the room stand-in ahead of the page's own scripts: at the end of the head,
// or right after the body tag (whatever attributes it carries) when there is no head.
func withRoom(html []byte) []byte {
	tag := []byte("<script src=\"/room.js\"></script>\n")
	at := bytes.Index(html, []byte("</head>"))
	if at < 0 {
		if b := bytes.Index(html, []byte("<body")); b >= 0 {
			if end := bytes.IndexByte(html[b:], '>'); end >= 0 {
				at = b + end + 1
			}
		}
	}
	if at < 0 {
		at = 0
	}
	out := make([]byte, 0, len(html)+len(tag))
	out = append(out, html[:at]...)
	out = append(out, tag...)
	return append(out, html[at:]...)
}

// page serves index.html with the room stand-in loaded ahead of the page's own script.
func page(web string) http.HandlerFunc {
	files := http.FileServer(http.Dir(web))
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" && req.URL.Path != "/index.html" {
			files.ServeHTTP(w, req)
			return
		}
		html, err := os.ReadFile(filepath.Join(web, "index.html"))
		if err != nil {
			http.Error(w, "web/index.html is missing", http.StatusNotFound)
			return
		}
		html = withRoom(html)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(html)
	}
}

func handler(web string, room *Room) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/room/events", room.events)
	mux.HandleFunc("/room/presence", room.presence)
	mux.HandleFunc("/sense/samples", room.samples)
	mux.HandleFunc("/room.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Write(roomJS)
	})
	mux.HandleFunc("/", page(web))
	return mux
}

func main() {
	addr := flag.String("addr", ":8787", "address to listen on")
	secure := flag.String("tls", ":8788", "second address, over https with a self-made certificate; empty to turn off")
	web := flag.String("web", "web", "folder that holds index.html")
	certs := flag.String("certs", "", "folder to keep the certificate in (default: the user cache folder)")
	flag.Parse()

	h := handler(*web, NewRoom())
	hosts := lanHosts()
	_, port, _ := net.SplitHostPort(*addr)
	for _, host := range hosts {
		log.Printf("play, table, watch:  http://%s", net.JoinHostPort(host, port))
	}
	if *secure != "" {
		cert, err := certificate(certDir(*certs), hosts)
		if err != nil {
			log.Fatalf("could not make a certificate: %v", err)
		}
		_, sport, _ := net.SplitHostPort(*secure)
		for _, host := range hosts {
			log.Printf("the phone on their back:  https://%s/#sense  (accept the certificate warning once)", net.JoinHostPort(host, sport))
		}
		srv := &http.Server{Addr: *secure, Handler: h, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}}
		go func() { log.Fatal(srv.ListenAndServeTLS("", "")) }()
	}
	log.Fatal(http.ListenAndServe(*addr, h))
}
