package benchkit

import (
	"runtime"
)

// MemResult contains the memory measurements of a memory benchmark
// at each point of the benchmark.
type MemResult struct {
	N          int
	Setup      *runtime.MemStats
	Start      *runtime.MemStats
	Teardown   *runtime.MemStats
	BeforeEach []*runtime.MemStats
	AfterEach  []*runtime.MemStats
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
	m.results.BeforeEach = m.each.beforeEach
	m.results.AfterEach = m.each.afterEach

	sub(m.start, m.setup, m.start)
	sub(m.teardown, m.setup, m.teardown)

	for _, each := range m.results.BeforeEach {
		sub(each, m.results.Setup, each)
	}
	for _, each := range m.results.AfterEach {
		sub(each, m.results.Setup, each)
	}
}

type memEach struct {
	beforeEach []*runtime.MemStats
	afterEach  []*runtime.MemStats
}

func (m *memEach) Before(id int) { runtime.ReadMemStats(m.beforeEach[id]) }
func (m *memEach) After(id int)  { runtime.ReadMemStats(m.afterEach[id]) }

// Memory will track memory allocations using `runtime.ReadMemStats`.
func Memory(n int) (BenchKit, *MemResult) {

	bench := &memBenchKit{
		n:        n,
		setup:    &runtime.MemStats{},
		start:    &runtime.MemStats{},
		teardown: &runtime.MemStats{},
		each: &memEach{
			beforeEach: make([]*runtime.MemStats, n),
			afterEach:  make([]*runtime.MemStats, n),
		},
		results: &MemResult{},
	}

	for i := 0; i < n; i++ {
		bench.each.beforeEach[i] = &runtime.MemStats{}
		bench.each.afterEach[i] = &runtime.MemStats{}
	}

	return bench, bench.results
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
