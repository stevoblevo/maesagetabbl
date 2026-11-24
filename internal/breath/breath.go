// Package breath reads a person's breathing from a phone lying on their back.
//
// A breath tilts the phone a degree or two, which moves gravity across its accelerometer axes.
// The estimator keeps forty seconds of that, finds the one direction the signal moves along,
// and looks for the period at which the signal best repeats itself.
//
// It reports how sure it is. Hands and a massage gun shake the phone too, so the reading is
// least sure while someone is being worked on, and that is said rather than hidden.
package breath

import "math"

const (
	// Rate is the internal sample rate in hertz. Phones deliver 50 to 100; averaging down to this
	// rate is what removes a massage gun's buzz.
	Rate = 10.0

	window   = 400 // samples kept: 40 s
	minRead  = 100 // samples before the first reading: 10 s
	slowMean = 101 // samples in the mean that is subtracted: 10.1 s, passes 5 breaths a minute and up
	smooth   = 7   // samples in the mean that is kept: 0.7 s, passes 42 breaths a minute and down

	minLag = 14  // 1.4 s: 43 breaths a minute
	maxLag = 120 // 12 s: 5 breaths a minute

	// TenseBPM and SettledBPM are the ends of the settle scale. They are the same two numbers the
	// game draws: 19 breaths a minute braced, 7 a puddle.
	TenseBPM   = 19.0
	SettledBPM = 7.0

	mg = 9.80665 / 1000 // one thousandth of gravity, in m/s²
)

// Sample is one accelerometer reading in m/s², gravity included.
type Sample struct{ X, Y, Z float64 }

// Reading is what the estimator believes right now.
type Reading struct {
	BPM     float64 `json:"bpm"`     // breaths a minute; 0 until there is one
	Sure    float64 `json:"sure"`    // 0 to 1: how cleanly the signal repeats
	Depth   float64 `json:"depth"`   // size of the breath movement, in thousandths of gravity
	Move    float64 `json:"move"`    // 0 to 1: shaking that is not breath (hands, a gun, a flinch)
	Settle  float64 `json:"settle"`  // 0 to 1, from BPM alone, held while unsure
	Wave    float64 `json:"wave"`    // -1 to 1: where in the breath they are now
	Seconds float64 `json:"seconds"` // how much signal this reading stands on
}

// Estimator turns a stream of samples into readings. The zero value is ready to use.
// It is not safe for concurrent use.
type Estimator struct {
	buf   [window][3]float64 // ring of averaged samples at Rate
	n     int                // samples written, ever
	sum   [3]float64         // the bin being filled
	sumSq float64
	count int
	t     float64
	shake [window]float64 // spread inside each bin, m/s²: what the averaging removed

	axis   [3]float64 // the direction breath moves along, kept between readings so Wave keeps its sign
	settle float64
	have   bool
}

// Add takes one sample that covers dt seconds.
func (e *Estimator) Add(dt float64, s Sample) {
	if dt <= 0 || dt > 1 || math.IsNaN(s.X+s.Y+s.Z) || math.IsInf(s.X+s.Y+s.Z, 0) {
		return
	}
	e.sum[0] += s.X
	e.sum[1] += s.Y
	e.sum[2] += s.Z
	e.sumSq += s.X*s.X + s.Y*s.Y + s.Z*s.Z
	e.count++
	e.t += dt
	if e.t < 1/Rate {
		return
	}
	c := float64(e.count)
	m := [3]float64{e.sum[0] / c, e.sum[1] / c, e.sum[2] / c}
	spread := e.sumSq/c - (m[0]*m[0] + m[1]*m[1] + m[2]*m[2])
	i := e.n % window
	e.buf[i] = m
	e.shake[i] = math.Sqrt(math.Max(0, spread))
	e.n++
	e.sum, e.sumSq, e.count = [3]float64{}, 0, 0
	e.t -= 1 / Rate
	if e.t > 1/Rate { // a gap in the stream: do not invent samples for it
		e.t = 0
	}
}

// Read reports the current reading. Call it about once a second.
func (e *Estimator) Read() Reading {
	w := e.n
	if w > window {
		w = window
	}
	r := Reading{Seconds: float64(w) / Rate, Settle: e.settle}
	if w < minRead {
		return r
	}
	raw := make([][3]float64, w)
	for i := range raw {
		raw[i] = e.buf[(e.n-w+i)%window]
	}

	// Keep what is slower than 0.7 s and faster than 10 s.
	low := mean3(raw, smooth)
	slow := mean3(raw, slowMean)
	band := make([][3]float64, w)
	for i := range band {
		for k := 0; k < 3; k++ {
			band[i][k] = low[i][k] - slow[i][k]
		}
	}

	// Shaking: what the 0.7 s mean removed, and what the averaging into bins removed, over 3 s.
	var hf float64
	last := 30
	for i := w - last; i < w; i++ {
		var d float64
		for k := 0; k < 3; k++ {
			x := raw[i][k] - low[i][k]
			d += x * x
		}
		s := e.shake[(e.n-w+i)%window]
		hf += d + s*s
	}
	r.Move = clamp((math.Sqrt(hf/float64(last))/mg-6)/60, 0, 1)

	// The one direction the breath moves along.
	axis := principal(band, e.axis)
	e.axis = axis
	b := make([]float64, w)
	var power float64
	for i := range b {
		b[i] = band[i][0]*axis[0] + band[i][1]*axis[1] + band[i][2]*axis[2]
		power += b[i] * b[i]
	}
	power /= float64(w)
	rms := math.Sqrt(power)
	r.Depth = rms / mg
	if rms > 0 {
		r.Wave = clamp(b[w-1-smooth/2]/(2*rms), -1, 1) // a third of a second back, where the means are not lopsided
	}
	if power == 0 {
		return r
	}

	// Where does it best repeat itself?
	top := maxLag
	if top > w/2 {
		top = w / 2
	}
	ac := make([]float64, top+2)
	for lag := minLag - 1; lag <= top+1; lag++ {
		var s float64
		for i := 0; i+lag < w; i++ {
			s += b[i] * b[i+lag]
		}
		ac[lag] = s / float64(w-lag) / power
	}
	best, bestLag := 0.0, 0
	for lag := minLag; lag <= top; lag++ {
		if ac[lag] > ac[lag-1] && ac[lag] >= ac[lag+1] && ac[lag] > best {
			best, bestLag = ac[lag], lag
		}
	}
	if bestLag == 0 {
		return r
	}
	// Twice the true period repeats as well as the period does. Take the shortest that is nearly as good.
	for lag := minLag; lag < bestLag; lag++ {
		if ac[lag] > ac[lag-1] && ac[lag] >= ac[lag+1] && ac[lag] >= 0.85*best {
			bestLag = lag
			break
		}
	}
	// Between samples: fit a parabola through the peak and its neighbours.
	a, c, d := ac[bestLag-1], ac[bestLag], ac[bestLag+1]
	period := float64(bestLag)
	if den := a - 2*c + d; den < 0 {
		period += 0.5 * (a - d) / den
	}
	r.BPM = 60 * Rate / period
	r.Sure = clamp(ac[bestLag], 0, 1)
	if r.Depth < 0.5 { // smaller than the sensor's own noise: not a breath
		r.Sure *= r.Depth / 0.5
	}
	r.Sure *= 1 - 0.5*r.Move

	if r.Sure >= 0.35 {
		want := clamp((TenseBPM-r.BPM)/(TenseBPM-SettledBPM), 0, 1)
		if !e.have {
			e.settle, e.have = want, true
		} else {
			e.settle += (want - e.settle) * 0.15
		}
	}
	r.Settle = e.settle
	return r
}

// mean3 is a centred moving mean over n samples, shortened at the ends.
func mean3(x [][3]float64, n int) [][3]float64 {
	out := make([][3]float64, len(x))
	half := n / 2
	var sum [3]float64
	lo, hi := 0, 0 // the mean covers x[lo:hi]
	for i := range x {
		for hi < len(x) && hi <= i+half {
			for k := 0; k < 3; k++ {
				sum[k] += x[hi][k]
			}
			hi++
		}
		for lo < i-half {
			for k := 0; k < 3; k++ {
				sum[k] -= x[lo][k]
			}
			lo++
		}
		c := float64(hi - lo)
		out[i] = [3]float64{sum[0] / c, sum[1] / c, sum[2] / c}
	}
	return out
}

// principal finds the direction of greatest spread by power iteration, signed to agree with prev.
func principal(x [][3]float64, prev [3]float64) [3]float64 {
	var c [3][3]float64
	for _, v := range x {
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				c[i][j] += v[i] * v[j]
			}
		}
	}
	v := [3]float64{0.577, 0.577, 0.577}
	if prev != ([3]float64{}) {
		v = prev
	}
	for it := 0; it < 40; it++ {
		var nv [3]float64
		for i := 0; i < 3; i++ {
			nv[i] = c[i][0]*v[0] + c[i][1]*v[1] + c[i][2]*v[2]
		}
		l := math.Sqrt(nv[0]*nv[0] + nv[1]*nv[1] + nv[2]*nv[2])
		if l == 0 {
			return v
		}
		v = [3]float64{nv[0] / l, nv[1] / l, nv[2] / l}
	}
	if prev != ([3]float64{}) && v[0]*prev[0]+v[1]*prev[1]+v[2]*prev[2] < 0 {
		v = [3]float64{-v[0], -v[1], -v[2]}
	}
	return v
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
