/*
Package benchkit is the lightweight, feather touch, benchmarking kit.

In comparison to the standard pprof utilities, this package is meant to
help generating graphs and other artifacts.

Usage

Get a benchmark kit:

    bench, result := benchkit.Memory(n)

A benchmark kit consists of 4 methods:

    - Setup():    call before doing any benchmark allocation.
    - Starting(): call once your benchmark data is ready, and you're about
                  to start the work you want to benchmark.
    - Each():     gives an object that tracks each step of your work.
    - Teardown(): call once your benchmark is done.

Here's an example:

    bench, result := benchkit.Memory(n)
    bench.Setup()
    // create benchmark data
    bench.Starting()
    doBenchmark(bench.Each())
    bench.Teardown()

Inside your benchmark, you will use the `BenchEach` object. This object
consists of 2 methods:

    - Before(i int): call it _before_ starting an atomic part of work.
    - After(i int): call it _after_ finishing an atomic part of work.

In both case, you must ensure that `0 <= i < n`, or you will panic.

    func doBenchmark(each BenchEach) {
        for i, job := range thingsToDoManyTimes {
            each.Before(i)
            job()
            each.After(i)
        }
    }

In the example above, you could use `defer each.After(i)`, however `defer` has
some overhead and thus, will reduce the precision of your benchmark results.

The `result` object given with your `bench` object will not be
populated before your call to `Teardown`.:

    bench, result := benchkit.Memory(n)
    // don't use `result`
    bench.Teardown()
    // now you can use `result`

Using `result` before `Teardown` will result in:

    panic: runtime error: invalid memory address or nil pointer dereference

So don't do that. =)

Memory kit

Collects memory allocation during the benchmark, using `runtime.ReadMemStats`.
The measurements are coarse.
*/
package benchkit
