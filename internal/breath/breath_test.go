package breath

import (
	"math"
	"math/rand"
	"testing"
)

// phone feeds the estimator what a phone lying on a breathing back would report: gravity,
// tilted back and forth by the breath, plus sensor noise, a slow settling of the body, and
// whatever shaking the test asks for.
type phone struct {
	hz      float64 // the phone's own sample rate
	bpm     float64 // 0 for nobody breathing
	tiltDeg float64 // how far a breath tilts it
	noiseMG float64 // sensor noise, thousandths of gravity
	gunHz   float64 // a massage gun nearby, 0 for none
	gunMG   float64
	rng     *rand.Rand
	t       float64
}

func (p *phone) run(e *Estimator, seconds float64) {
	const g = 9.80665
	dt := 1 / p.hz
	for end := p.t + seconds; p.t < end; p.t += dt {
		tilt := 0.0
		if p.bpm > 0 {
			tilt = p.tiltDeg * math.Pi / 180 * math.Sin(2*math.Pi*p.bpm/60*p.t)
		}
		lean := 0.02 * p.t / 60 // the body sinking into the table: about a degree a minute
		n := func() float64 { return p.rng.NormFloat64() * p.noiseMG * g / 1000 }
		gun := p.gunMG * g / 1000 * math.Sin(2*math.Pi*p.gunHz*p.t)
		e.Add(dt, Sample{
			X: g*math.Sin(tilt+lean) + n() + gun*0.6,
			Y: g*0.05 + n() + gun*0.3,
			Z: g*math.Cos(tilt+lean) + n() + gun,
		})
	}
}

func newPhone(bpm float64) *phone {
	return &phone{hz: 60, bpm: bpm, tiltDeg: 1.2, noiseMG: 2, rng: rand.New(rand.NewSource(7))}
}

func TestFindsTheBreathRate(t *testing.T) {
	for _, bpm := range []float64{6, 7, 9, 12, 15, 19, 24} {
		var e Estimator
		newPhone(bpm).run(&e, 40)
		r := e.Read()
		if math.Abs(r.BPM-bpm) > 0.6 {
			t.Errorf("breathing at %v a minute: read %.2f", bpm, r.BPM)
		}
		if r.Sure < 0.7 {
			t.Errorf("breathing at %v a minute: only %.2f sure of a clean signal", bpm, r.Sure)
		}
	}
}

func TestSaysNothingBeforeTenSeconds(t *testing.T) {
	var e Estimator
	newPhone(12).run(&e, 8)
	if r := e.Read(); r.BPM != 0 || r.Sure != 0 {
		t.Fatalf("after 8 s it already claims %.1f a minute, %.2f sure", r.BPM, r.Sure)
	}
}

func TestIsUnsureOfAPhoneOnATable(t *testing.T) {
	// Nobody under it: sensor noise only. Whatever rate it guesses, it must not trust it.
	for seed := int64(1); seed <= 20; seed++ {
		var e Estimator
		p := newPhone(0)
		p.rng = rand.New(rand.NewSource(seed))
		p.run(&e, 40)
		if r := e.Read(); r.Sure >= 0.35 {
			t.Errorf("seed %d: %.2f sure of %.1f breaths a minute from an empty table", seed, r.Sure, r.BPM)
		}
	}
}

func TestSettleFollowsTheRate(t *testing.T) {
	read := func(bpm float64) Reading {
		var e Estimator
		newPhone(bpm).run(&e, 40)
		return e.Read()
	}
	if r := read(SettledBPM); r.Settle < 0.9 {
		t.Errorf("at %v a minute settle is %.2f, want a puddle", SettledBPM, r.Settle)
	}
	if r := read(TenseBPM); r.Settle > 0.1 {
		t.Errorf("at %v a minute settle is %.2f, want braced", TenseBPM, r.Settle)
	}
	if mid := read(13); mid.Settle < 0.4 || mid.Settle > 0.6 {
		t.Errorf("at 13 a minute settle is %.2f, want about half", mid.Settle)
	}
}

func TestAMassageGunShakesItButDoesNotFoolIt(t *testing.T) {
	var e Estimator
	p := newPhone(10)
	p.gunHz, p.gunMG = 28, 250 // knead, a hand's width away
	p.run(&e, 40)
	r := e.Read()
	if math.Abs(r.BPM-10) > 1 {
		t.Errorf("with a gun running it read %.2f breaths a minute, want 10", r.BPM)
	}
	if r.Move < 0.5 {
		t.Errorf("a 250 mg buzz reads as %.2f movement, want it noticed", r.Move)
	}
	var calm Estimator
	newPhone(10).run(&calm, 40)
	if c := calm.Read(); r.Sure >= c.Sure {
		t.Errorf("as sure with the gun (%.2f) as without (%.2f); it should say it is less sure", r.Sure, c.Sure)
	}
}

func TestFollowsABreathThatSlows(t *testing.T) {
	var e Estimator
	p := newPhone(18)
	p.run(&e, 40)
	before := e.Read()
	p.bpm = 8
	for i := 0; i < 60; i++ { // a minute, read once a second as the server does
		p.run(&e, 1)
		e.Read()
	}
	after := e.Read()
	if math.Abs(after.BPM-8) > 0.8 {
		t.Errorf("a minute after slowing to 8 it reads %.2f", after.BPM)
	}
	if after.Settle <= before.Settle+0.5 {
		t.Errorf("settle went %.2f to %.2f as the breath slowed from 18 to 8", before.Settle, after.Settle)
	}
}

func TestIgnoresGapsAndNonsense(t *testing.T) {
	var e Estimator
	e.Add(0.016, Sample{math.NaN(), 0, 9.8})
	e.Add(-1, Sample{0, 0, 9.8})
	e.Add(30, Sample{0, 0, 9.8})
	if e.n != 0 {
		t.Fatalf("bad samples were kept: %d", e.n)
	}
}
