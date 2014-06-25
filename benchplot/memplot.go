package benchplot

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"github.com/aybabtme/benchkit"
	"github.com/dustin/go-humanize"
	"image/color"
)

var memlines = []struct {
	Name   string
	Filter func(mem benchkit.MemStep) float64
	Width  float64
	Color  color.Color
}{
	{
		Name:   "current heap size (p50)",
		Filter: func(mem benchkit.MemStep) float64 { return float64(mem.P(50).HeapAlloc) },
		Width:  0.5,
		Color:  color.RGBA{202, 0, 32, 255},
	},
	{
		Name:   "effective memory consumption (p50)",
		Filter: func(mem benchkit.MemStep) float64 { return float64(mem.P(50).Sys - mem.P(50).HeapReleased) },
		Width:  0.5,
		Color:  color.RGBA{5, 113, 176, 255},
	},
}

// PlotMemory will create a line graph of AfterEach measurements. The lines
// plotted are:
//
//      current heap size            : HeapAlloc
//      effective memory consumption : Sys - HeapReleased
//
// The Y axis is implicitely measured in Bytes.
func PlotMemory(title, xLabel string, results *benchkit.MemResult, logscale bool) (*plot.Plot, error) {

	p, err := plot.New()
	if err != nil {
		return nil, err
	}

	p.Title.Text = title

	if logscale {
		p.Y.Label.Text = "Memory usage (log10)"
		p.Y.Tick.Marker = readableBytes(plot.LogTicks)
		p.Y.Scale = plot.LogScale
		p.Y.Min = 1.0
	} else {
		p.Y.Tick.Marker = readableBytes(p.Y.Tick.Marker)
		p.Y.Label.Text = "Memory usage"
		// p.Y.Min = 0.0
	}
	p.X.Label.Text = xLabel

	p.Add(plotter.NewGrid())

	// Scatters

	heapAlloc, err := plotter.NewScatter(func() plotter.XYs {
		var xys plotter.XYs
		for i, step := range results.Each {
			for _, mem := range step.PRange(1, 99) {
				xys = append(xys, struct{ X, Y float64 }{
					X: float64(i), Y: float64(mem.HeapAlloc),
				})
			}
		}
		return xys
	}())
	if err != nil {
		return nil, err
	}
	heapAlloc.Color = color.RGBA{244, 165, 130, 255}
	heapAlloc.GlyphStyle.Shape = plot.PlusGlyph{}
	heapAlloc.GlyphStyle.Radius = vg.Points(0.5)
	p.Add(heapAlloc)

	effectUsage, err := plotter.NewScatter(func() plotter.XYs {
		var xys plotter.XYs
		for i, step := range results.Each {
			for _, mem := range step.PRange(1, 99) {
				xys = append(xys, struct{ X, Y float64 }{
					X: float64(i), Y: float64(mem.Sys - mem.HeapReleased),
				})
			}
		}
		return xys
	}())
	if err != nil {
		return nil, err
	}

	effectUsage.Color = color.RGBA{146, 197, 222, 255}
	effectUsage.GlyphStyle.Shape = plot.PlusGlyph{}
	effectUsage.GlyphStyle.Radius = vg.Points(0.5)
	p.Add(effectUsage)

	// Lines

	for _, data := range memlines {
		line, err := plotter.NewLine(mapResult(data.Filter, results.Each))
		if err != nil {
			return nil, err
		}
		line.Width = vg.Points(data.Width)
		line.Color = data.Color
		p.Add(line)
		p.Legend.Add(data.Name, line)
		p.Legend.Top = true
	}

	return p, nil
}

func mapResult(f func(mem benchkit.MemStep) float64, mems []benchkit.MemStep) plotter.XYs {
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
