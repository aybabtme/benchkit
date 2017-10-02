package benchkit_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/aybabtme/benchkit"
	"github.com/dustin/go-humanize"
	"github.com/dustin/randbo"
)

func ExampleBench() {
	mem := benchkit.Bench(benchkit.Memory(10)).Each(func(each benchkit.BenchEach) {
		for i := 0; i < 10; i++ {
			each.Before(i)
			// do stuff
			each.After(i)
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
	size := 1000000
	buf := bytes.NewBuffer(nil)

	memkit, results := benchkit.Memory(n)

	memkit.Setup()
	files := GenTarFiles(n, size)
	memkit.Starting()

	each := memkit.Each()

	tarw := tar.NewWriter(buf)
	for i, file := range files {
		each.Before(i)
		_ = tarw.WriteHeader(file.TarHeader())
		_, _ = tarw.Write(file.Data())
		each.After(i)
	}
	_ = tarw.Close()

	memkit.Teardown()

	// Look at the results!
	fmt.Printf("setup=%s\n", effectMem(results.Setup))
	fmt.Printf("starting=%s\n", effectMem(results.Start))

	for i := 0; i < results.N; i++ {
		fmt.Printf("  %d  before=%s  after=%s\n",
			i,
			effectMem(results.BeforeEach[i]),
			effectMem(results.AfterEach[i]),
		)
	}
	fmt.Printf("teardown=%s\n", effectMem(results.Teardown))

	// Output:
	// setup=2.0 MB
	// starting=6.9 MB
	//   0  before=6.9 MB  after=8.0 MB
	//   1  before=8.0 MB  after=10 MB
	//   2  before=10 MB  after=15 MB
	//   3  before=15 MB  after=15 MB
	//   4  before=15 MB  after=15 MB
	// teardown=26 MB
}

func effectMem(mem *runtime.MemStats) string {
	effectMem := mem.Sys - mem.HeapReleased
	return humanize.Bytes(effectMem)
}

var rand = randbo.New()

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
