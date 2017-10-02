package benchplot

import (
	"image/color"
	"time"

	"github.com/aybabtme/benchkit"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var timelines = []struct {
	Name   string
	Filter func(t benchkit.TimeStep) float64
	Width  float64
	Color  color.Color
}{
	{
		Name:   "p50",
		Filter: func(t benchkit.TimeStep) float64 { return float64(t.P(50)) },
		Width:  0.5,
		Color:  color.RGBA{43, 140, 190, 255},
	},
	// {
	// 	Name:   "p90",
	// 	Filter: func(t benchkit.TimeStep) float64 { return float64(t.P(90)) },
	// 	Width:  1,
	// 	Color:  color.RGBA{252, 141, 89, 255},
	// },
	// {
	// 	Name:   "p99",
	// 	Filter: func(t benchkit.TimeStep) float64 { return float64(t.P(99)) },
	// 	Width:  0.3,
	// 	Color:  color.RGBA{215, 48, 39, 255},
	// },
	// {
	// 	Name:   "average",
	// 	Filter: func(t benchkit.TimeStep) float64 { return float64(t.Avg) },
	// 	Width:  1,
	// 	Color:  color.RGBA{254, 224, 144, 255},
	// },
}

// PlotTime does stuff.
func PlotTime(title, xLabel string, results *benchkit.TimeResult, logscale bool) (*plot.Plot, error) {

	p, err := plot.New()
	if err != nil {
		return nil, err
	}

	p.Title.Text = title
	if logscale {
		p.Y.Label.Text = "Duration (log10)"
		p.Y.Scale = plot.LogScale{}
		p.Y.Tick.Marker = readableDuration(plot.LogTicks{})
	} else {
		p.Y.Label.Text = "Duration"
		p.Y.Tick.Marker = readableDuration(p.Y.Tick.Marker)
	}

	p.X.Label.Text = xLabel

	p.Add(plotter.NewGrid())

	scatter, err := plotter.NewScatter(func() plotter.XYs {
		var xys plotter.XYs
		for i, step := range results.Each {
			for _, dur := range step.PRange(1, 99) {
				xys = append(xys, struct{ X, Y float64 }{
					X: float64(i), Y: float64(dur),
				})
			}
		}
		return xys
	}())
	if err != nil {
		return nil, err
	}
	scatter.Color = color.RGBA{166, 189, 219, 255}
	scatter.GlyphStyle.Shape = draw.PlusGlyph{}
	scatter.GlyphStyle.Radius = vg.Points(1)
	p.Add(scatter)

	for _, data := range timelines {
		line, err := plotter.NewLine(mapSteps(data.Filter, results.Each))
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

func mapSteps(f func(step benchkit.TimeStep) float64, steps []benchkit.TimeStep) plotter.XYs {
	xys := make(plotter.XYs, len(steps))
	for i, step := range steps {
		xys[i].X = float64(i)
		xys[i].Y = f(step)
	}
	return xys
}

func readableDuration(marker plot.Ticker) plot.Ticker {
	return tickerFunc(func(min, max float64) []plot.Tick {
		var out []plot.Tick
		for _, t := range marker.Ticks(min, max) {
			t.Label = time.Duration(t.Value).String()
			out = append(out, t)
		}
		return out
	})
}
