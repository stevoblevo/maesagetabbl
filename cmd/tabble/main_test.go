package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stevoblevo/maesagetabbl/internal/breath"
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

func TestAPhoneOnTheBackBecomesAReadingInTheRoom(t *testing.T) {
	r := NewRoom()
	if r.Feel("aaaaaaaa", 0.02, []breath.Sample{{Z: 9.8}}, false) {
		t.Fatal("a phone with no stream was felt")
	}
	phone, _ := r.Join("aaaaaaaa")
	snap(t, phone)
	r.Set("aaaaaaaa", json.RawMessage(`{"role":"sense"}`))
	snap(t, phone)

	// 30 s of a back breathing 9 times a minute, posted five times a second as the page does.
	const hz, g = 50.0, 9.80665
	var last []Peer
	for i := 0; i < 30*5; i++ {
		batch := make([]breath.Sample, 10)
		for j := range batch {
			tt := float64(i*10+j) / hz
			tilt := 1.2 * math.Pi / 180 * math.Sin(2*math.Pi*9/60*tt)
			batch[j] = breath.Sample{X: g * math.Sin(tilt), Y: 0.3, Z: g * math.Cos(tilt)}
		}
		if !r.Feel("aaaaaaaa", 1/hz, batch, false) {
			t.Fatal("the phone's samples were refused")
		}
		select {
		case b := <-phone:
			json.Unmarshal(b, &last)
		default:
		}
	}
	if len(last) != 1 || last[0].Reading == nil {
		t.Fatalf("no reading reached the room: %+v", last)
	}
	if got := last[0].Reading; math.Abs(got.BPM-9) > 0.6 || got.Sure < 0.7 {
		t.Fatalf("read %.2f breaths a minute, %.2f sure; want 9 and sure", got.BPM, got.Sure)
	}
	if string(last[0].Presence) != `{"role":"sense"}` {
		t.Fatalf("the reading overwrote the phone's own presence: %s", last[0].Presence)
	}
	r.Set("aaaaaaaa", json.RawMessage(`{"role":"sense","on":true}`))
	if got := snap(t, phone); got[0].Reading == nil {
		t.Fatal("setting presence dropped the reading")
	}
}

func TestSamplesAreChecked(t *testing.T) {
	room := NewRoom()
	srv := httptest.NewServer(handler(t.TempDir(), room))
	defer srv.Close()
	room.Join("aaaaaaaa")
	for name, tc := range map[string]struct {
		body string
		want int
	}{
		"good":             {`{"ms":200,"s":[[0,0,9.8],[0,0,9.8]]}`, http.StatusNoContent},
		"no samples":       {`{"ms":200,"s":[]}`, http.StatusBadRequest},
		"no span":          {`{"s":[[0,0,9.8]]}`, http.StatusBadRequest},
		"a minute at once": {`{"ms":60000,"s":[[0,0,9.8]]}`, http.StatusBadRequest},
		"not samples":      {`{"ms":200,"s":"x"}`, http.StatusBadRequest},
	} {
		res, err := http.Post(srv.URL+"/sense/samples?peer=aaaaaaaa", "application/json", strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tc.want {
			t.Errorf("%s: got %d, want %d", name, res.StatusCode, tc.want)
		}
	}
}

func TestCertificateIsMadeOnceAndNamesThisMachine(t *testing.T) {
	dir := t.TempDir()
	hosts := []string{"192.168.1.20"}
	first, err := certificate(dir, hosts)
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := x509.ParseCertificate(first.Certificate[0])
	for _, h := range []string{"192.168.1.20", "localhost", "127.0.0.1"} {
		if leaf.VerifyHostname(h) != nil {
			t.Errorf("the certificate does not name %s", h)
		}
	}
	again, _ := certificate(dir, hosts)
	if string(again.Certificate[0]) != string(first.Certificate[0]) {
		t.Error("a second start made a new certificate; a phone would be asked to accept it again")
	}
	moved, _ := certificate(dir, []string{"10.0.0.7"})
	if string(moved.Certificate[0]) == string(first.Certificate[0]) {
		t.Error("the machine's address changed and the old certificate was reused")
	}

	// And it really serves: a client that trusts exactly this certificate can fetch the page.
	web := t.TempDir()
	os.WriteFile(filepath.Join(web, "index.html"), []byte("<!DOCTYPE html><html><body></body></html>"), 0o644)
	srv := httptest.NewUnstartedServer(handler(web, NewRoom()))
	local, _ := certificate(t.TempDir(), []string{"127.0.0.1"})
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{local}}
	srv.StartTLS()
	defer srv.Close()
	pool := x509.NewCertPool()
	l, _ := x509.ParseCertificate(local.Certificate[0])
	pool.AddCert(l)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}
	res, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("https with the self-made certificate failed: %v", err)
	}
	res.Body.Close()
}

// clock is a room's clock that only moves when a test moves it.
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

// breathe posts seconds of a back breathing bpm times a minute, five posts a second as the page
// does, moving the clock with the signal. It returns the last snapshot the listener was sent.
func breathe(t *testing.T, r *Room, c *clock, listen <-chan []byte, id string, bpm, seconds float64, fresh bool) []Peer {
	t.Helper()
	const hz, g = 50.0, 9.80665
	var last []Peer
	for i := 0; i < int(seconds*5); i++ {
		batch := make([]breath.Sample, 10)
		for j := range batch {
			tt := float64(i*10+j) / hz
			tilt := 1.2 * math.Pi / 180 * math.Sin(2*math.Pi*bpm/60*tt)
			batch[j] = breath.Sample{X: g * math.Sin(tilt), Y: 0.3, Z: g * math.Cos(tilt)}
		}
		c.t = c.t.Add(200 * time.Millisecond)
		if !r.Feel(id, 1/hz, batch, fresh && i == 0) {
			t.Fatal("the phone's samples were refused")
		}
		for more := true; more; {
			select {
			case b := <-listen:
				last = nil
				json.Unmarshal(b, &last)
			default:
				more = false
			}
		}
	}
	return last
}

func TestANewSensingSessionStartsFromNothing(t *testing.T) {
	for name, pause := range map[string]struct {
		wait  time.Duration
		fresh bool
	}{
		"the phone says it started again": {time.Second, true},
		"the phone went quiet":            {30 * time.Second, false},
	} {
		c := &clock{t: time.Unix(1_800_000_000, 0)}
		r := NewRoom()
		r.now = c.now
		phone, _ := r.Join("aaaaaaaa")
		got := breathe(t, r, c, phone, "aaaaaaaa", 9, 40, true)
		if len(got) != 1 || got[0].Reading == nil || math.Abs(got[0].Reading.BPM-9) > 0.6 {
			t.Fatalf("%s: the first session did not read 9: %+v", name, got)
		}
		c.t = c.t.Add(pause.wait)
		got = breathe(t, r, c, phone, "aaaaaaaa", 19, 0.4, pause.fresh)
		if len(got) != 1 || got[0].Reading != nil {
			t.Fatalf("%s: the old session's reading was still in the room: %+v", name, got[0].Reading)
		}
		got = breathe(t, r, c, phone, "aaaaaaaa", 19, 9.6, false)
		if rd := got[0].Reading; rd == nil || rd.Seconds > 11 || (rd.BPM != 0 && math.Abs(rd.BPM-9) < 2) {
			t.Fatalf("%s: ten seconds into the new session it still stood on the old one: %+v", name, rd)
		}
		got = breathe(t, r, c, phone, "aaaaaaaa", 19, 20, false)
		if rd := got[0].Reading; rd == nil || math.Abs(rd.BPM-19) > 1.2 {
			t.Fatalf("%s: thirty seconds into the new session it did not read 19: %+v", name, rd)
		}
	}
}

func TestAnOldReadingIsNotHandedToAScreenThatJoinsLater(t *testing.T) {
	c := &clock{t: time.Unix(1_800_000_000, 0)}
	r := NewRoom()
	r.now = c.now
	phone, _ := r.Join("aaaaaaaa")
	breathe(t, r, c, phone, "aaaaaaaa", 9, 30, true)

	c.t = c.t.Add(2 * time.Second)
	soon, _ := r.Join("bbbbbbbb")
	got := snap(t, soon)
	if got[0].Reading == nil || got[0].ReadingAge < 2000 || got[0].ReadingAge > 3000 {
		t.Fatalf("two seconds on, want the reading with its age: %+v age %d", got[0].Reading, got[0].ReadingAge)
	}

	c.t = c.t.Add(time.Hour)
	late, _ := r.Join("cccccccc")
	for _, p := range snap(t, late) {
		if p.Reading != nil || p.ReadingAge != 0 {
			t.Fatalf("an hour on, %s was still handed a reading aged %d ms", p.Peer, p.ReadingAge)
		}
	}
}
