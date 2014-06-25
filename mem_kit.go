package benchkit

import (
	"math"
	"runtime"
	"sort"
)

// MemResult contains the memory measurements of a memory benchmark
// at each point of the benchmark.
type MemResult struct {
	N        int
	Setup    *runtime.MemStats
	Start    *runtime.MemStats
	Teardown *runtime.MemStats
	Each     []MemStep
}

// MemStep contains statistics about a step of the benchmark.
type MemStep struct {
	all         []*runtime.MemStats
	Significant []*runtime.MemStats
	Min         runtime.MemStats
	Max         runtime.MemStats
	Avg         runtime.MemStats
	SD          runtime.MemStats
}

// µ is the expected value. Greek letters because we can.
func (m *MemStep) µ() runtime.MemStats {
	// since all values are equaly probable, µ is sum/length
	sum := &runtime.MemStats{}
	for _, dur := range m.Significant {
		add(dur, sum, sum)
	}

	scale(sum, uint64(len(m.Significant)), sum)
	return *sum
}

// σ is the standard deviation. Greek letters because we can.
func (m *MemStep) σ() runtime.MemStats {
	sum := &runtime.MemStats{}
	exp := m.µ()
	diff := &runtime.MemStats{}
	pow2 := &runtime.MemStats{}
	for _, dur := range m.Significant {
		sub(dur, &exp, diff)
		times(diff, diff, pow2)
		add(pow2, sum, sum)
	}
	scaled := &runtime.MemStats{}

	scale(sum, uint64(len(m.Significant)), scaled)

	σ := &runtime.MemStats{}
	sqrt(scaled, σ)
	return *σ
}

// P returns the percentile duration of the step, such as p50, p90, p99...
func (m *MemStep) P(factor float64) runtime.MemStats {
	if len(m.all) == 0 {
		return runtime.MemStats{}
	}
	pIdx := pIndex(len(m.all), factor)
	return *m.all[pIdx-1]
}

// PRange returns the percentile duration of the step, such as p50, p90, p99...
func (m *MemStep) PRange(from, to float64) []*runtime.MemStats {
	if len(m.all) == 0 {
		return m.all
	}
	fromIdx := pIndex(len(m.all), from)
	toIdx := pIndex(len(m.all), to)
	return m.all[fromIdx-1 : toIdx]
}

type memBenchKit struct {
	n        int
	setup    *runtime.MemStats
	start    *runtime.MemStats
	teardown *runtime.MemStats
	each     *memEach

	results *MemResult
}

func (m *memBenchKit) Setup()          { runtime.ReadMemStats(m.setup) }
func (m *memBenchKit) Starting()       { runtime.ReadMemStats(m.start) }
func (m *memBenchKit) Each() BenchEach { return m.each }
func (m *memBenchKit) Teardown() {
	runtime.ReadMemStats(m.teardown)
	m.results.N = m.n
	m.results.Setup = m.setup
	m.results.Start = m.start
	m.results.Teardown = m.teardown
	m.results.Each = make([]MemStep, len(m.each.afterEach))
	for i, after := range m.each.afterEach {
		mslice := memstatSlice(after)
		sort.Sort(&mslice)
		step := MemStep{all: mslice}
		step.Significant = step.PRange(0.5, 0.95)
		step.Min = *step.Significant[0]
		step.Max = *step.Significant[len(step.Significant)-1]
		step.Avg = step.µ()
		step.SD = step.σ()
		m.results.Each[i] = step
	}
}

type memEach struct {
	beforeEach [][]*runtime.MemStats
	afterEach  [][]*runtime.MemStats
}

func (m *memEach) Before(id int) {

	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	m.beforeEach[id] = append(m.beforeEach[id], mem)

	start := m.beforeEach[0][len(m.beforeEach[0])-1]
	// dirty, changes `mem` in place after having put it in the slice
	sub(mem, start, mem)
}
func (m *memEach) After(id int) {
	start := m.beforeEach[0][len(m.beforeEach[0])-1]

	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	sub(mem, start, mem)
	m.afterEach[id] = append(m.afterEach[id], mem)
}

// Memory will track memory allocations using `runtime.ReadMemStats`.
func Memory(n, m int) (BenchKit, *MemResult) {

	bench := &memBenchKit{
		n:        n,
		setup:    &runtime.MemStats{},
		start:    &runtime.MemStats{},
		teardown: &runtime.MemStats{},
		results:  &MemResult{},
	}

	bench.each = &memEach{
		beforeEach: make([][]*runtime.MemStats, n),
		afterEach:  make([][]*runtime.MemStats, n),
	}

	for i := 0; i < n; i++ {
		bench.each.beforeEach[i] = make([]*runtime.MemStats, 0, m)
		bench.each.afterEach[i] = make([]*runtime.MemStats, 0, m)
	}

	return bench, bench.results
}

type memstatSlice []*runtime.MemStats

func (m *memstatSlice) Len() int      { return len(*m) }
func (m *memstatSlice) Swap(i, j int) { (*m)[i], (*m)[j] = (*m)[j], (*m)[i] }
func (m *memstatSlice) Less(i, j int) bool {
	iMemUsed := (*m)[i].Sys - (*m)[i].HeapReleased
	jMemUsed := (*m)[j].Sys - (*m)[j].HeapReleased
	return iMemUsed < jMemUsed
}

func add(a, b, out *runtime.MemStats) {
	out.Alloc = a.Alloc + b.Alloc
	out.TotalAlloc = a.TotalAlloc + b.TotalAlloc
	out.Sys = a.Sys + b.Sys
	out.Lookups = a.Lookups + b.Lookups
	out.Mallocs = a.Mallocs + b.Mallocs
	out.Frees = a.Frees + b.Frees
	out.HeapAlloc = a.HeapAlloc + b.HeapAlloc
	out.HeapSys = a.HeapSys + b.HeapSys
	out.HeapIdle = a.HeapIdle + b.HeapIdle
	out.HeapInuse = a.HeapInuse + b.HeapInuse
	out.HeapReleased = a.HeapReleased + b.HeapReleased
	out.HeapObjects = a.HeapObjects + b.HeapObjects
	out.StackInuse = a.StackInuse + b.StackInuse
	out.StackSys = a.StackSys + b.StackSys
	out.MSpanInuse = a.MSpanInuse + b.MSpanInuse
	out.MSpanSys = a.MSpanSys + b.MSpanSys
	out.MCacheInuse = a.MCacheInuse + b.MCacheInuse
	out.MCacheSys = a.MCacheSys + b.MCacheSys
	out.BuckHashSys = a.BuckHashSys + b.BuckHashSys
	out.GCSys = a.GCSys + b.GCSys
	out.OtherSys = a.OtherSys + b.OtherSys
	out.NextGC = a.NextGC + b.NextGC
	out.LastGC = a.LastGC + b.LastGC
	out.PauseTotalNs = a.PauseTotalNs + b.PauseTotalNs

	for i := range out.PauseNs {
		out.PauseNs[i] = a.PauseNs[i] + b.PauseNs[i]
	}

	for i := range out.BySize {
		out.BySize[i].Size = a.BySize[i].Size + b.BySize[i].Size
		out.BySize[i].Mallocs = a.BySize[i].Mallocs + b.BySize[i].Mallocs
		out.BySize[i].Frees = a.BySize[i].Frees + b.BySize[i].Frees
	}
}

func sub(a, b, out *runtime.MemStats) {
	out.Alloc = a.Alloc - b.Alloc
	out.TotalAlloc = a.TotalAlloc - b.TotalAlloc
	out.Sys = a.Sys - b.Sys
	out.Lookups = a.Lookups - b.Lookups
	out.Mallocs = a.Mallocs - b.Mallocs
	out.Frees = a.Frees - b.Frees
	out.HeapAlloc = a.HeapAlloc - b.HeapAlloc
	out.HeapSys = a.HeapSys - b.HeapSys
	out.HeapIdle = a.HeapIdle - b.HeapIdle
	out.HeapInuse = a.HeapInuse - b.HeapInuse
	out.HeapReleased = a.HeapReleased - b.HeapReleased
	out.HeapObjects = a.HeapObjects - b.HeapObjects
	out.StackInuse = a.StackInuse - b.StackInuse
	out.StackSys = a.StackSys - b.StackSys
	out.MSpanInuse = a.MSpanInuse - b.MSpanInuse
	out.MSpanSys = a.MSpanSys - b.MSpanSys
	out.MCacheInuse = a.MCacheInuse - b.MCacheInuse
	out.MCacheSys = a.MCacheSys - b.MCacheSys
	out.BuckHashSys = a.BuckHashSys - b.BuckHashSys
	out.GCSys = a.GCSys - b.GCSys
	out.OtherSys = a.OtherSys - b.OtherSys
	out.NextGC = a.NextGC - b.NextGC
	out.LastGC = a.LastGC - b.LastGC
	out.PauseTotalNs = a.PauseTotalNs - b.PauseTotalNs

	for i := range out.PauseNs {
		out.PauseNs[i] = a.PauseNs[i] - b.PauseNs[i]
	}

	for i := range out.BySize {
		out.BySize[i].Size = a.BySize[i].Size - b.BySize[i].Size
		out.BySize[i].Mallocs = a.BySize[i].Mallocs - b.BySize[i].Mallocs
		out.BySize[i].Frees = a.BySize[i].Frees - b.BySize[i].Frees
	}
}

func times(a, b, out *runtime.MemStats) {
	out.Alloc = a.Alloc * b.Alloc
	out.TotalAlloc = a.TotalAlloc * b.TotalAlloc
	out.Sys = a.Sys * b.Sys
	out.Lookups = a.Lookups * b.Lookups
	out.Mallocs = a.Mallocs * b.Mallocs
	out.Frees = a.Frees * b.Frees
	out.HeapAlloc = a.HeapAlloc * b.HeapAlloc
	out.HeapSys = a.HeapSys * b.HeapSys
	out.HeapIdle = a.HeapIdle * b.HeapIdle
	out.HeapInuse = a.HeapInuse * b.HeapInuse
	out.HeapReleased = a.HeapReleased * b.HeapReleased
	out.HeapObjects = a.HeapObjects * b.HeapObjects
	out.StackInuse = a.StackInuse * b.StackInuse
	out.StackSys = a.StackSys * b.StackSys
	out.MSpanInuse = a.MSpanInuse * b.MSpanInuse
	out.MSpanSys = a.MSpanSys * b.MSpanSys
	out.MCacheInuse = a.MCacheInuse * b.MCacheInuse
	out.MCacheSys = a.MCacheSys * b.MCacheSys
	out.BuckHashSys = a.BuckHashSys * b.BuckHashSys
	out.GCSys = a.GCSys * b.GCSys
	out.OtherSys = a.OtherSys * b.OtherSys
	out.NextGC = a.NextGC * b.NextGC
	out.LastGC = a.LastGC * b.LastGC
	out.PauseTotalNs = a.PauseTotalNs * b.PauseTotalNs

	for i := range out.PauseNs {
		out.PauseNs[i] = a.PauseNs[i] * b.PauseNs[i]
	}

	for i := range out.BySize {
		out.BySize[i].Size = a.BySize[i].Size * b.BySize[i].Size
		out.BySize[i].Mallocs = a.BySize[i].Mallocs * b.BySize[i].Mallocs
		out.BySize[i].Frees = a.BySize[i].Frees * b.BySize[i].Frees
	}
}

func scale(a *runtime.MemStats, b uint64, out *runtime.MemStats) {
	out.Alloc = a.Alloc / b
	out.TotalAlloc = a.TotalAlloc / b
	out.Sys = a.Sys / b
	out.Lookups = a.Lookups / b
	out.Mallocs = a.Mallocs / b
	out.Frees = a.Frees / b
	out.HeapAlloc = a.HeapAlloc / b
	out.HeapSys = a.HeapSys / b
	out.HeapIdle = a.HeapIdle / b
	out.HeapInuse = a.HeapInuse / b
	out.HeapReleased = a.HeapReleased / b
	out.HeapObjects = a.HeapObjects / b
	out.StackInuse = a.StackInuse / b
	out.StackSys = a.StackSys / b
	out.MSpanInuse = a.MSpanInuse / b
	out.MSpanSys = a.MSpanSys / b
	out.MCacheInuse = a.MCacheInuse / b
	out.MCacheSys = a.MCacheSys / b
	out.BuckHashSys = a.BuckHashSys / b
	out.GCSys = a.GCSys / b
	out.OtherSys = a.OtherSys / b
	out.NextGC = a.NextGC / b
	out.LastGC = a.LastGC / b
	out.PauseTotalNs = a.PauseTotalNs / b

	for i := range out.PauseNs {
		out.PauseNs[i] = a.PauseNs[i] / b
	}

	for i := range out.BySize {
		out.BySize[i].Size = a.BySize[i].Size / uint32(b)
		out.BySize[i].Mallocs = a.BySize[i].Mallocs / b
		out.BySize[i].Frees = a.BySize[i].Frees / b
	}
}

func sqrt(a *runtime.MemStats, out *runtime.MemStats) {
	out.Alloc = uint64(math.Sqrt(float64(a.Alloc)))
	out.TotalAlloc = uint64(math.Sqrt(float64(a.TotalAlloc)))
	out.Sys = uint64(math.Sqrt(float64(a.Sys)))
	out.Lookups = uint64(math.Sqrt(float64(a.Lookups)))
	out.Mallocs = uint64(math.Sqrt(float64(a.Mallocs)))
	out.Frees = uint64(math.Sqrt(float64(a.Frees)))
	out.HeapAlloc = uint64(math.Sqrt(float64(a.HeapAlloc)))
	out.HeapSys = uint64(math.Sqrt(float64(a.HeapSys)))
	out.HeapIdle = uint64(math.Sqrt(float64(a.HeapIdle)))
	out.HeapInuse = uint64(math.Sqrt(float64(a.HeapInuse)))
	out.HeapReleased = uint64(math.Sqrt(float64(a.HeapReleased)))
	out.HeapObjects = uint64(math.Sqrt(float64(a.HeapObjects)))
	out.StackInuse = uint64(math.Sqrt(float64(a.StackInuse)))
	out.StackSys = uint64(math.Sqrt(float64(a.StackSys)))
	out.MSpanInuse = uint64(math.Sqrt(float64(a.MSpanInuse)))
	out.MSpanSys = uint64(math.Sqrt(float64(a.MSpanSys)))
	out.MCacheInuse = uint64(math.Sqrt(float64(a.MCacheInuse)))
	out.MCacheSys = uint64(math.Sqrt(float64(a.MCacheSys)))
	out.BuckHashSys = uint64(math.Sqrt(float64(a.BuckHashSys)))
	out.GCSys = uint64(math.Sqrt(float64(a.GCSys)))
	out.OtherSys = uint64(math.Sqrt(float64(a.OtherSys)))
	out.NextGC = uint64(math.Sqrt(float64(a.NextGC)))
	out.LastGC = uint64(math.Sqrt(float64(a.LastGC)))
	out.PauseTotalNs = uint64(math.Sqrt(float64(a.PauseTotalNs)))

	for i := range out.PauseNs {
		out.PauseNs[i] = uint64(math.Sqrt(float64(a.PauseNs[i])))
	}

	for i := range out.BySize {
		out.BySize[i].Size = uint32(math.Sqrt(float64(a.BySize[i].Size)))
		out.BySize[i].Mallocs = uint64(math.Sqrt(float64(a.BySize[i].Mallocs)))
		out.BySize[i].Frees = uint64(math.Sqrt(float64(a.BySize[i].Frees)))
	}
}
