package benchkit

// BenchKit tracks metrics about your benchmark.
type BenchKit interface {
	// Setup must be called before doing any benchmark allocation.
	Setup()
	// Starting must be called once your benchmark data is ready,
	// and you're about to start the work you want to benchmark.
	Starting()
	// Teardown must be called once your benchmark is done.
	Teardown()
	// Each gives an object that tracks each step of your work.
	Each() BenchEach
}

// BenchEach tracks metrics about work units of your benchmark.
type BenchEach interface {
	// Before must be called _before_ starting a unit of work.
	Before(id int)
	// After must be called _after_ finishing a unit of work.
	After(id int)
}
