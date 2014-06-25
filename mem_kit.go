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
