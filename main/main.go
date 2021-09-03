package main

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar"
	"gonum.org/v1/gonum/mat"
)

type Grid struct {
	*mat.Dense
	nchannels int
}

func (g *Grid) Min() float64 {
	return 0.0
}

func (g *Grid) Max() float64 {
	return float64(g.nchannels)
}

func (g *Grid) Z(c, r int) float64 {
	return g.Dense.At(r, c)
}

func (g *Grid) X(c int) float64 {
	_, cols := g.Dims()
	if c > cols {
		panic("too large")
	}
	if c < 0 {
		panic("too small")
	}
	return float64(c)
}

func (g *Grid) Y(r int) float64 {
	if rows, _ := g.Dims(); rows < r {
		panic("too large")
	}
	if r < 0 {
		panic("too small")
	}
	return float64(r)
}

// func loop(w *app.Window) error {
// 	th := material.NewTheme(gofont.Collection())
// 	var ops op.Ops
// 	for {
// 		e := <-w.Events()
// 		switch e := e.(type) {
// 		case system.DestroyEvent:
// 			return e.Err

// 		case system.FrameEvent:
// 			gtx := layout.NewContext(&ops, e)
// 			l := material.H1(th, "Hello, Gio")
// 			maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
// 			l.Color = maroon
// 			l.Alignment = text.Middle
// 			l.Layout(gtx)
// 			e.Frame(gtx.Ops)
// 		}
// 	}
// }

func main() {
	// go func() {
	// 	w := app.NewWindow()
	// 	if err := loop(w); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	os.Exit(0)
	// }()

	// data := mat.NewDense(3, 3, []float64{0, 1, 0, 2, 0, 0, 3, 3, 1})
	// p := plot.New()

	// colours, err := brewer.GetPalette(brewer.TypeQualitative, "Paired", 3)
	// if err != nil {
	// 	panic(err)
	// }
	// heatMap := plotter.NewHeatMap(&Grid{data}, colours)
	// p.Add(heatMap)

	// w, err := p.WriterTo(512, 512, "png")
	// if err != nil {
	// 	panic(err)
	// }

	// f, err := os.Create("out.png")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// w.WriteTo(f)

	// app.Main()

	env, err := goatar.New(goatar.SpaceInvaders, 0.1, true, 11)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 330; i++ {
		env.Act(rand.Intn(6))
		env.DisplayState(fmt.Sprint(i), 128, 128)
	}

	// 	state, err := env.State()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	size := env.StateShape()
	// 	r, c := size[1], size[2]
	// 	// chData, _ := env.Channel(1)
	// 	data := mat.NewDense(size[1], size[2], nil)
	// 	for ch := 0; ch < size[0]; ch++ {
	// 		chData := state[r*c*ch : r*c*(ch+1)]
	// 		fmt.Println(chData)
	// 		for row := 0; row < r; row++ {
	// 			for col := 0; col < c; col++ {
	// 				if chData[row*c+col] != 0 {
	// 					data.Set(row, col, chData[row*c+col]*float64(ch+1))
	// 				}
	// 			}
	// 		}
	// 	}

	// 	p := plot.New()

	// 	colours := cols{[]color.Color{
	// 		color.RGBA{30, 30, 30, 255},
	// 		color.RGBA{0, 63, 92, 255},
	// 		color.RGBA{88, 80, 141, 255},
	// 		color.RGBA{188, 80, 144, 255},
	// 		color.RGBA{255, 99, 97, 255},
	// 		color.RGBA{255, 166, 0, 255},
	// 		color.RGBA{72, 143, 49, 255},
	// 	}}
	// 	for env.NChannels() > len(colours.Colors()) {
	// 		rng := rand.New(rand.NewSource(10))
	// 		r := uint8(rng.Uint32() % 255)
	// 		g := uint8(rng.Uint32() % 255)
	// 		b := uint8(rng.Uint32() % 255)
	// 		colours.colours = append(colours.colours, color.RGBA{r, g, b, 255})
	// 	}
	// 	heatMap := plotter.NewHeatMap(&Grid{data, env.NChannels()}, colours)
	// 	p.Add(heatMap)

	// 	w, err := p.WriterTo(512, 512, "png")
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fnew, err := os.Create("state.png")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer fnew.Close()

	// 	w.WriteTo(fnew)
	// }

	// type cols struct {
	// 	colours []color.Color
	// }

	// func (c cols) Colors() []color.Color {
	// 	return c.colours
}
