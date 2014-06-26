package benchplot

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

func ExamplePlotTime() {
	n := 100
	times := 100
	size := int(1e6)
	buf := bytes.NewBuffer(nil)

	timekit, results := benchkit.Time(n, times)

	timekit.Setup()
	files := GenTarFiles(n, size)
	timekit.Starting()

	each := timekit.Each()

	for i := 0; i < times; i++ {
		buf.Reset()
		tarw := tar.NewWriter(buf)
		for j, file := range files {
			each.Before(j)
			_ = tarw.WriteHeader(file.TarHeader())
			_, _ = tarw.Write(file.Data())
			each.After(j)
		}
		_ = tarw.Close()
	}

	timekit.Teardown()

	p, _ := PlotTime(
		fmt.Sprintf("archive/tar time usage for %d files, %s each, over %d measurements", n, humanize.Bytes(uint64(size)), times),
		"Files in archive",
		results,
		true,
	)
	_ = p.Save(6, 4, "tar_timeplot.svg")

	// Output:
	//
}

func ExamplePlotTime_bench() {
	n := 100
	times := 100
	size := int(1e6)
	buf := bytes.NewBuffer(nil)

	files := GenTarFiles(n, size)

	results := benchkit.Bench(benchkit.Time(n, times)).Each(func(each benchkit.BenchEach) {
		for repeat := 0; repeat < times; repeat++ {
			buf.Reset()
			tarw := tar.NewWriter(buf)
			for j, file := range files {
				each.Before(j)
				_ = tarw.WriteHeader(file.TarHeader())
				_, _ = tarw.Write(file.Data())
				each.After(j)
			}
			_ = tarw.Close()
		}

	}).(*benchkit.TimeResult)

	p, _ := PlotTime(
		fmt.Sprintf("archive/tar time usage for %d files, %s each, over %d measurements", n, humanize.Bytes(uint64(size)), times),
		"Files in archive",
		results,
		true,
	)
	_ = p.Save(6, 4, "tar_timeplot.png")

	// Output:
	//
}

func ExamplePlotMemory() {
	n := 100
	size := int(1e6)
	buf := bytes.NewBuffer(nil)

	memkit, results := benchkit.Memory(n)

	memkit.Setup()
	files := GenTarFiles(n, size)
	memkit.Starting()

	each := memkit.Each()

	buf.Reset()
	tarw := tar.NewWriter(buf)
	for j, file := range files {
		each.Before(j)
		_ = tarw.WriteHeader(file.TarHeader())
		_, _ = tarw.Write(file.Data())
		each.After(j)
	}
	_ = tarw.Close()

	memkit.Teardown()

	p, _ := PlotMemory(
		fmt.Sprintf("archive/tar memory usage for %d files, %s each", n, humanize.Bytes(uint64(size))),
		"Files in archive",
		results,
		false,
	)
	_ = p.Save(6, 4, "tar_memplot.svg")

	// Output:
	//
}

func ExamplePlotMemory_bench() {

	n := 100
	size := int(1e6)
	buf := bytes.NewBuffer(nil)

	files := GenTarFiles(n, size)

	results := benchkit.Bench(benchkit.Memory(n)).Each(func(each benchkit.BenchEach) {
		buf.Reset()
		tarw := tar.NewWriter(buf)
		for j, file := range files {
			each.Before(j)
			_ = tarw.WriteHeader(file.TarHeader())
			_, _ = tarw.Write(file.Data())
			each.After(j)
		}
		_ = tarw.Close()

	}).(*benchkit.MemResult)

	p, _ := PlotMemory(
		fmt.Sprintf("archive/tar memory usage for %d files, %s each", n, humanize.Bytes(uint64(size))),
		"Files in archive",
		results,
		false,
	)
	_ = p.Save(6, 4, "tar_memplot.png")

	// Output:
	//
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
