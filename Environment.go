// Package goatar implement minimal Atari games that run on a 10x10
// grid. This package was inspired by MinAtar, which can be found at:
// https://github.com/kenjyoung/MinAtar.
package goatar

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"

	"github.com/samuelfneumann/goatar/internal/game"
	"github.com/samuelfneumann/goatar/internal/game/asterix"
	"github.com/samuelfneumann/goatar/internal/game/breakout"
	"github.com/samuelfneumann/goatar/internal/game/freeway"
	"github.com/samuelfneumann/goatar/internal/game/seaquest"
	"github.com/samuelfneumann/goatar/internal/game/spaceinvaders"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
)

const NumActions int = 6 // All games have 6 actions

//
type GameName struct {
	string // Hide the internals so that new GameNames can't be created
}

var (
	Asterix       GameName = GameName{"Asterix"}
	SpaceInvaders GameName = GameName{"Space Invaders"}
	Freeway       GameName = GameName{"Freeway"}
	Breakout      GameName = GameName{"Breakout"}
	SeaQuest      GameName = GameName{"SeaQuest"}
)

// make is a static factory for creating a game.Game for an environment
func makeEnv(game GameName, difficultyRamping bool, seed int64) (game.Game,
	error) {
	switch game {
	case Asterix:
		return asterix.New(difficultyRamping, seed)

	case Breakout:
		return breakout.New(difficultyRamping, seed)

	case Freeway:
		return freeway.New(difficultyRamping, seed)

	case SeaQuest:
		return seaquest.New(difficultyRamping, seed)

	case SpaceInvaders:
		return spaceinvaders.New(difficultyRamping, seed)

	default:
		return nil, fmt.Errorf("no such game")
	}
}

// Environment implements an environment that an agent can interact
// with.
type Environment struct {
	game.Game
	gameName          GameName
	rng               *rand.Rand
	nChannels         int
	stickyActionsProb float64
	lastAction        int // Is this action the first?
	firstAction       bool
	closed            bool
}

// New creates and returns a new Environment of the game specified
// by name.
func New(name GameName, stickyActionsProb float64, difficultyRamping bool,
	seed int64) (*Environment, error) {
	game, err := makeEnv(name, difficultyRamping, seed)
	if err != nil {
		return nil, fmt.Errorf("new: %v", err)
	}

	rng := rand.New(rand.NewSource(seed))

	return &Environment{
		Game:              game,
		gameName:          name,
		rng:               rng,
		nChannels:         game.NChannels(),
		stickyActionsProb: stickyActionsProb,
		firstAction:       true,
		lastAction:        -1,
		closed:            false,
	}, nil
}

// Act takes one environmental action
func (e *Environment) Act(a int) (float64, bool, error) {
	if e.firstAction {
		e.firstAction = false
	} else if e.rng.Float64() < e.stickyActionsProb {
		a = e.lastAction
	}
	e.lastAction = a
	return e.Game.Act(a)
}

// NumActions returns the total number of available actions
func (e *Environment) NumActions() int {
	return NumActions
}

// GameName returns the name of the game
func (e *Environment) GameName() string {
	return e.gameName.string
}

// Display state saves the current state as a png to a file
func (e *Environment) DisplayState(filename string, w, h float64) error {
	// Get current state
	state, err := e.State()
	if err != nil {
		return fmt.Errorf("displayState: %v", err)
	}
	size := e.StateShape()
	r, c := size[1], size[2]

	// Combine data to create heatmap
	data := mat.NewDense(size[1], size[2], nil)
	for ch := 0; ch < size[0]; ch++ {
		chData := state[r*c*ch : r*c*(ch+1)]
		for row := 0; row < r; row++ {
			for col := 0; col < c; col++ {
				if chData[row*c+col] != 0 {
					data.Set(row, col, chData[row*c+col]*float64(ch+1))
				}
			}
		}
	}

	// Set colours for heatmap
	colours := newColours([]color.Color{
		color.RGBA{30, 30, 30, 255},
		color.RGBA{0, 63, 92, 255},
		color.RGBA{88, 80, 141, 255},
		color.RGBA{188, 80, 144, 255},
		color.RGBA{255, 99, 97, 255},
		color.RGBA{255, 166, 0, 255},
		color.RGBA{72, 143, 49, 255},
	})

	// Generate random colours if above not enough
	for e.NChannels() > len(colours.Colors()) {
		rng := rand.New(rand.NewSource(10))
		r := uint8(rng.Uint32() % 255)
		g := uint8(rng.Uint32() % 255)
		b := uint8(rng.Uint32() % 255)
		colours.c = append(colours.c, color.RGBA{r, g, b, 255})
	}

	// Create the plot
	p := plot.New()
	p.HideAxes()

	// Create the heatmap
	heatMap := plotter.NewHeatMap(&Grid{data, e.NChannels()}, colours)
	p.Add(heatMap)

	// Create the writer to write the plot to
	writer, err := p.WriterTo(font.Length(w), font.Length(h), "png")
	if err != nil {
		return fmt.Errorf("displayState: %v", err)
	}

	// Create the file to save to
	fnew, err := os.Create(fmt.Sprintf("%v.png", filename))
	if err != nil {
		return fmt.Errorf("displayState: %v", err)
	}
	defer fnew.Close()

	// Write to file
	writer.WriteTo(fnew)
	return nil
}

type colours struct {
	c []color.Color
}

func newColours(cols []color.Color) *colours {
	return &colours{cols}
}

func (c *colours) Colors() []color.Color {
	return c.c
}

func (c *colours) Add(col color.Color) {
	c.c = append(c.c, col)
}

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
