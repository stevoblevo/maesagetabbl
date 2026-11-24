// Command tabble serves the Tabble games on your own network and links the screens that open them.
//
// The page is the same file the claude.ai artifact runs. There it uses the platform's room; here the
// server injects a small stand-in (room.js) and relays presence between screens: every screen posts
// its whole presence object, and every screen hears the whole room over one event stream.
// Nothing is stored. When a screen's stream drops, it leaves the room.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

//go:embed room.js
var roomJS []byte

const (
	maxPeers    = 64
	maxPresence = 4 << 10 // bytes of JSON, the same bound the platform's room uses
)

// Peer is one open screen.
type Peer struct {
	Peer      string          `json:"peer"`
	Presence  json.RawMessage `json:"presence"`
	UpdatedAt int64           `json:"updatedAt"`
}

// Room is everyone who has the page open right now.
type Room struct {
	mu      sync.Mutex
	peers   map[string]*Peer
	streams map[string]chan []byte
}

func NewRoom() *Room {
	return &Room{peers: map[string]*Peer{}, streams: map[string]chan []byte{}}
}

// snapshot returns the whole room as JSON, oldest peer id first so the order is stable.
func (r *Room) snapshot() []byte {
	list := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
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
	r.broadcast()
}

// Set replaces a peer's presence. Only a peer with an open stream may set one.
func (r *Room) Set(id string, presence json.RawMessage) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, here := r.streams[id]; !here {
		return false
	}
	r.peers[id] = &Peer{Peer: id, Presence: presence, UpdatedAt: time.Now().UnixMilli()}
	r.broadcast()
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
		html = bytes.Replace(html, []byte("<body>"), []byte("<body>\n<script src=\"/room.js\"></script>"), 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(html)
	}
}

func handler(web string, room *Room) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/room/events", room.events)
	mux.HandleFunc("/room/presence", room.presence)
	mux.HandleFunc("/room.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Write(roomJS)
	})
	mux.HandleFunc("/", page(web))
	return mux
}

func main() {
	addr := flag.String("addr", ":8787", "address to listen on")
	web := flag.String("web", "web", "folder that holds index.html")
	flag.Parse()
	log.Printf("Tabble on %s (open it on two screens; choose \"On the table\" on one)", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler(*web, NewRoom())))
}
