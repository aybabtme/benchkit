package benchkit

// BUG(aybabtme): not sure I like the Bench func.

// Bench is a helper func that will call Starting/Teardown for you.
func Bench(kit BenchKit, results interface{}) *eacher {
	kit.Setup()
	kit.Starting()
	return &eacher{kit: kit, results: results}
}

type eacher struct {
	kit     BenchKit
	results interface{}
}

func (e *eacher) Each(f func(each BenchEach)) interface{} {
	f(e.kit.Each())
	e.kit.Teardown()
	return e.results
}
