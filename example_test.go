package benchkit_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/aybabtme/benchkit"
	"github.com/dustin/go-humanize"
	"github.com/dustin/randbo"
	"os"
	"runtime"
	"strconv"
	"time"
)

func ExampleBench() {
	mem := benchkit.Bench(benchkit.Memory(10, 10)).Each(func(each benchkit.BenchEach) {
		for repeat := 0; repeat < 10; repeat++ {
			for i := 0; i < 10; i++ {
				each.Before(i)
				// do stuff
				each.After(i)
			}
		}
	}).(*benchkit.MemResult)

	_ = mem

	times := benchkit.Bench(benchkit.Time(10, 100)).Each(func(each benchkit.BenchEach) {
		for repeat := 0; repeat < 100; repeat++ {
			for i := 0; i < 10; i++ {
				each.Before(i)
				// do stuff
				each.After(i)
			}
		}

	}).(*benchkit.TimeResult)

	_ = times
}

func ExampleMemory() {
	n := 5
	m := 10
	size := 1000000
	buf := bytes.NewBuffer(nil)

	memkit, results := benchkit.Memory(n, m)

	memkit.Setup()
	files := GenTarFiles(n, size)

	each := memkit.Each()

	for repeat := 0; repeat < m; repeat++ {
		buf.Reset()
		tarw := tar.NewWriter(buf)
		for i, file := range files {
			each.Before(i)
			_ = tarw.WriteHeader(file.TarHeader())
			_, _ = tarw.Write(file.Data())
			each.After(i)
		}
		_ = tarw.Close()
	}

	memkit.Teardown()

	// Look at the results!
	fmt.Printf("setup=%s\n", effectMem(results.Setup))
	fmt.Printf("starting=%s\n", effectMem(results.Start))

	for i := 0; i < results.N; i++ {
		fmt.Printf("  %d  before=%s  after=%s\n",
			i,
			effectMem(&results.Each[i].Avg),
			effectMem(&results.Each[i].Avg),
		)
	}
	fmt.Printf("teardown=%s\n", effectMem(results.Teardown))

	// Output:
	// setup=4.1MB
	// starting=0B
	//   0  before=11MB  after=11MB
	//   1  before=13MB  after=13MB
	//   2  before=19MB  after=19MB
	//   3  before=19MB  after=19MB
	//   4  before=19MB  after=19MB
	// teardown=29MB
}

func effectMem(mem *runtime.MemStats) string {
	effectMem := mem.Sys - mem.HeapReleased
	return humanize.Bytes(effectMem)
}

var rand = randbo.NewFast()

func GenTarFiles(n, size int) []TarFile {
	files := make([]TarFile, n)
	for i := range files {
		files[i] = GenTarFile(i, size)
	}
	return files
}

func GenTarFile(id, size int) TarFile {
	data := make([]byte, size)
	_, _ = rand.Read(data)
	return TarFile{
		Name:    strconv.Itoa(id),
		LastMod: time.Now(),
		data:    *bytes.NewBuffer(data),
	}
}

type TarFile struct {
	Name    string
	LastMod time.Time
	data    bytes.Buffer
}

func (t *TarFile) Data() []byte { return t.data.Bytes() }

func (t *TarFile) TarHeader() *tar.Header {
	return &tar.Header{
		Name:       t.Name,
		Size:       int64(t.data.Len()),
		Mode:       int64(0644),
		AccessTime: time.Now(),
		ChangeTime: t.LastMod,
		ModTime:    t.LastMod,
		Typeflag:   tar.TypeReg,
		Uid:        os.Getuid(),
		Gid:        os.Getgid(),
	}
}
