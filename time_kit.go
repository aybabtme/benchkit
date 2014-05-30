package benchkit

import (
	"math"
	"sort"
	"time"
)

// TimeResult contains the memory measurements of a memory benchmark
// at each point of the benchmark.
type TimeResult struct {
	N        int
	Setup    time.Time
	Start    time.Time
	Teardown time.Time
	Each     []TimeStep
}

// TimeStep contains statistics about a step of the benchmark.
type TimeStep struct {
	all         []time.Duration
	Significant []time.Duration
	Min         time.Duration
	Max         time.Duration
	Avg         time.Duration
	SD          time.Duration
}

// µ is the expected value. Greek letters because we can.
func (t *TimeStep) µ() time.Duration {
	// since all values are equaly probable, µ is sum/length
	var sum time.Duration
	for _, dur := range t.Significant {
		sum += dur
	}
	return sum / time.Duration(len(t.Significant))
}

// σ is the standard deviation. Greek letters because we can.
func (t *TimeStep) σ() time.Duration {
	var sum time.Duration
	for _, dur := range t.Significant {
		sum += ((dur - t.µ()) * (dur - t.µ()))
	}
	scaled := sum / time.Duration(len(t.Significant))

	σ := math.Sqrt(float64(scaled))

	return time.Duration(σ)
}

// P returns the percentile duration of the step, such as p50, p90, p99...
func (t *TimeStep) P(factor float64) time.Duration {
	if len(t.all) == 0 {
		return time.Duration(0)
	}
	pIdx := pIndex(len(t.all), factor)
	return t.all[pIdx-1]
}

// PRange returns the percentile duration of the step, such as p50, p90, p99...
func (t *TimeStep) PRange(from, to float64) []time.Duration {
	if len(t.all) == 0 {
		return t.all
	}
	fromIdx := pIndex(len(t.all), from)
	toIdx := pIndex(len(t.all), to)
	return t.all[fromIdx-1 : toIdx]
}

func pIndex(base int, factor float64) int {
	power := math.Log10(factor)
	closest := 10 * math.Pow10(int(power))
	idx := int(math.Ceil((factor * float64(base)) / closest))
	return idx
}

type timeBenchKit struct {
	n        int
	setup    time.Time
	start    time.Time
	teardown time.Time
	each     *timeEach

	results *TimeResult
}

func (t *timeBenchKit) Setup()          { t.setup = time.Now() }
func (t *timeBenchKit) Starting()       { t.start = time.Now() }
func (t *timeBenchKit) Each() BenchEach { return t.each }
func (t *timeBenchKit) Teardown() {
	t.teardown = time.Now()
	t.results.N = t.n
	t.results.Setup = t.setup
	t.results.Start = t.start
	t.results.Teardown = t.teardown
	t.results.Each = make([]TimeStep, len(t.each.after))
	for i, after := range t.each.after {
		d := durationSlice(after)
		sort.Sort(&d)
		step := TimeStep{all: d}
		step.Significant = step.PRange(0.5, 0.95)
		step.Min = step.Significant[0]
		step.Max = step.Significant[len(step.Significant)-1]
		step.Avg = step.µ()
		step.SD = step.σ()
		t.results.Each[i] = step
	}
}

type timeEach struct {
	before [][]time.Time
	after  [][]time.Duration
}

func (t *timeEach) Before(id int) {
	t.before[id] = append(t.before[id], time.Now())
}
func (t *timeEach) After(id int) {
	beforeIdx := max(len(t.before[id])-1, 0)
	before := t.before[id][beforeIdx]
	t.after[id] = append(t.after[id], time.Since(before))
}

// Time will track timings over exactly n steps, m times for each step.
// Memory is allocated in advance for m times per step, but you can record
// less than m times without effect, or more than m times with a loss of
// precision (due to extra allocation).
func Time(n, m int) (BenchKit, *TimeResult) {

	bench := &timeBenchKit{
		n: n,
		each: &timeEach{
			before: make([][]time.Time, n),
			after:  make([][]time.Duration, n),
		},
		results: &TimeResult{},
	}

	for i := 0; i < n; i++ {
		bench.each.before[i] = make([]time.Time, 0, m)
		bench.each.after[i] = make([]time.Duration, 0, m)
	}

	return bench, bench.results
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type durationSlice []time.Duration

func (d *durationSlice) Len() int           { return len(*d) }
func (d *durationSlice) Less(i, j int) bool { return (*d)[i] < (*d)[j] }
func (d *durationSlice) Swap(i, j int)      { (*d)[i], (*d)[j] = (*d)[j], (*d)[i] }
