package benchplot

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg"
	"github.com/aybabtme/benchkit"
	"github.com/dustin/go-humanize"
	"image/color"
	"runtime"
)

var darkIdx = 0

func pickDark() color.Color {
	defer func() { darkIdx = (darkIdx + 1) % len(plotutil.DarkColors) }()
	return plotutil.DarkColors[darkIdx]
}

var softIdx = 0

func pickSoft() color.Color {
	defer func() { softIdx = (softIdx + 1) % len(plotutil.SoftColors) }()
	return plotutil.SoftColors[softIdx]
}

var memlines = []struct {
	Name   string
	Filter func(mem *runtime.MemStats) float64
	Width  float64
	Color  color.Color
}{
	{
		Name:   "current heap size",
		Filter: func(mem *runtime.MemStats) float64 { return float64(mem.HeapAlloc) },
		Width:  0.5,
		Color:  color.RGBA{69, 117, 180, 255},
	},
	{
		Name:   "total heap size",
		Filter: func(mem *runtime.MemStats) float64 { return float64(mem.HeapSys) },
		Width:  0.5,
		Color:  color.RGBA{215, 48, 39, 255},
	},
	{
		Name:   "memory allocated from OS",
		Filter: func(mem *runtime.MemStats) float64 { return float64(mem.Sys) },
		Width:  0.5,
		Color:  color.RGBA{254, 224, 144, 255},
	},
	{
		Name:   "effective memory consumption",
		Filter: func(mem *runtime.MemStats) float64 { return float64(mem.Sys - mem.HeapReleased) },
		Width:  0.5,
		Color:  color.RGBA{252, 141, 89, 255},
	},
}

// PlotMemory will create a line graph of AfterEach measurements. The lines
// plotted are:
//      current heap size            : HeapAlloc
//      total heap size              : HeapSys
//      memory allocated from OS     : Sys
//      effective memory consumption : Sys - HeapReleased
// The Y axis is implicitely measured in Bytes.
func PlotMemory(title, xLabel string, results *benchkit.MemResult, logscale bool) (*plot.Plot, error) {

	p, err := plot.New()
	if err != nil {
		return nil, err
	}

	p.Title.Text = title

	if logscale {
		p.Y.Label.Text = "Memory usage (log10)"
		p.Y.Scale = plot.LogScale
		p.Y.Tick.Marker = readableBytes(plot.LogTicks)
	} else {
		p.Y.Label.Text = "Memory usage"
		p.Y.Tick.Marker = readableBytes(p.Y.Tick.Marker)
	}
	p.X.Label.Text = xLabel

	p.Add(plotter.NewGrid())

	for _, data := range memlines {
		line, err := plotter.NewLine(mapResult(data.Filter, results.AfterEach))
		if err != nil {
			return nil, err
		}
		line.Width = vg.Points(data.Width)
		line.Color = data.Color
		p.Add(line)
		p.Legend.Add(data.Name, line)
	}

	return p, nil
}

func mapResult(f func(mem *runtime.MemStats) float64, mems []*runtime.MemStats) plotter.XYs {
	xys := make(plotter.XYs, len(mems))
	for i, mem := range mems {
		xys[i].X = float64(i)
		xys[i].Y = f(mem)
	}
	return xys
}

func readableBytes(marker func(min, max float64) []plot.Tick) func(float64, float64) []plot.Tick {
	return func(min, max float64) []plot.Tick {
		var out []plot.Tick
		for _, t := range marker(min, max) {
			t.Label = humanize.Bytes(uint64(t.Value))
			out = append(out, t)
		}
		return out
	}
}
