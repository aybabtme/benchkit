package benchplot

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"github.com/aybabtme/benchkit"
	"image/color"
	"time"
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
		Width:  1,
		Color:  color.RGBA{69, 117, 180, 255},
	},
	{
		Name:   "p90",
		Filter: func(t benchkit.TimeStep) float64 { return float64(t.P(90)) },
		Width:  1,
		Color:  color.RGBA{252, 141, 89, 255},
	},
	{
		Name:   "p99",
		Filter: func(t benchkit.TimeStep) float64 { return float64(t.P(99)) },
		Width:  1,
		Color:  color.RGBA{215, 48, 39, 255},
	},
	{
		Name:   "average",
		Filter: func(t benchkit.TimeStep) float64 { return float64(t.Avg) },
		Width:  1,
		Color:  color.RGBA{254, 224, 144, 255},
	},
}

// PlotTime does stuff.
func PlotTime(title, xLabel string, results *benchkit.TimeResult, logscale bool) (*plot.Plot, error) {

	p, err := plot.New()
	if err != nil {
		return nil, err
	}

	p.Title.Text = title
	p.Y.Label.Text = "Duration (log10)"
	if logscale {
		p.Y.Scale = plot.LogScale
	}
	p.Y.Tick.Marker = readableDuration(plot.LogTicks)
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
	scatter.Color = color.RGBA{224, 243, 248, 255}
	scatter.GlyphStyle.Shape = plot.PlusGlyph{}
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

func readableDuration(marker func(min, max float64) []plot.Tick) func(float64, float64) []plot.Tick {
	return func(min, max float64) []plot.Tick {
		var out []plot.Tick
		for _, t := range marker(min, max) {
			t.Label = time.Duration(t.Value).String()
			out = append(out, t)
		}
		return out
	}
}
