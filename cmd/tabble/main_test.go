package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func snap(t *testing.T, ch <-chan []byte) []Peer {
	t.Helper()
	select {
	case b := <-ch:
		var list []Peer
		if err := json.Unmarshal(b, &list); err != nil {
			t.Fatalf("snapshot is not JSON: %v", err)
		}
		return list
	case <-time.After(time.Second):
		t.Fatal("no snapshot within a second")
		return nil
	}
}

func TestRoomCarriesPresenceBetweenScreens(t *testing.T) {
	r := NewRoom()
	table, ok := r.Join("aaaaaaaa")
	if !ok {
		t.Fatal("first screen could not join")
	}
	snap(t, table) // itself
	play, _ := r.Join("bbbbbbbb")
	if got := snap(t, play); len(got) != 2 {
		t.Fatalf("second screen sees %d peers, want 2", len(got))
	}
	snap(t, table) // the join
	if !r.Set("aaaaaaaa", json.RawMessage(`{"role":"table","stopped":true}`)) {
		t.Fatal("table could not set presence")
	}
	got := snap(t, play)
	if string(got[0].Presence) != `{"role":"table","stopped":true}` {
		t.Fatalf("player did not hear the table's stop: %s", got[0].Presence)
	}
	r.Leave("aaaaaaaa")
	if got := snap(t, play); len(got) != 1 || got[0].Peer != "bbbbbbbb" {
		t.Fatalf("table left but the room still shows %+v", got)
	}
}

func TestPresenceNeedsAnOpenStream(t *testing.T) {
	r := NewRoom()
	if r.Set("cccccccc", json.RawMessage(`{}`)) {
		t.Fatal("a screen with no stream set presence")
	}
	if _, ok := r.Join("aaaaaaaa"); !ok {
		t.Fatal("join failed")
	}
	if _, ok := r.Join("aaaaaaaa"); ok {
		t.Fatal("the same id joined twice")
	}
}

func TestServerInjectsRoomAndRejectsBadPresence(t *testing.T) {
	web := t.TempDir()
	os.WriteFile(filepath.Join(web, "index.html"), []byte("<!DOCTYPE html><html><body>\n<p>table</p></body></html>"), 0o644)
	srv := httptest.NewServer(handler(web, NewRoom()))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	buf := new(strings.Builder)
	bufio.NewReader(res.Body).WriteTo(buf)
	res.Body.Close()
	if !strings.Contains(buf.String(), `<script src="/room.js"></script>`) {
		t.Fatal("the page was served without the room stand-in")
	}

	for name, tc := range map[string]struct {
		url, body string
		want      int
	}{
		"no stream yet": {"/room/presence?peer=aaaaaaaa", `{}`, http.StatusConflict},
		"not an object": {"/room/presence?peer=aaaaaaaa", `[1,2]`, http.StatusBadRequest},
		"bad peer id":   {"/room/presence?peer=../etc", `{}`, http.StatusBadRequest},
		"over four KiB": {"/room/presence?peer=aaaaaaaa", `{"x":"` + strings.Repeat("a", 5000) + `"}`, http.StatusBadRequest},
	} {
		res, err := http.Post(srv.URL+tc.url, "application/json", strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tc.want {
			t.Errorf("%s: got %d, want %d", name, res.StatusCode, tc.want)
		}
	}
}
