package benchkit_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/aybabtme/benchkit"
	"github.com/aybabtme/benchkit/benchplot"
	"github.com/dustin/go-humanize"
	"github.com/dustin/randbo"
	"os"
	"runtime"
	"strconv"
	"time"
)

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

	// Plot the results !
	p, _ := benchplot.PlotMemory(
		fmt.Sprintf("archive/tar memory usage for %n files of %s", n, humanize.Bytes(uint64(size))),
		"files in archive",
		results,
	)
	_ = p.Save(8, 6, "tar_benchplot.png")

	// Output:
	// setup=4.1MB
	// starting=9.8MB
	//   0  before=9.8MB  after=11MB
	//   1  before=11MB  after=13MB
	//   2  before=13MB  after=19MB
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
