package breakout

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
	"gonum.org/v1/gonum/mat"
)

const (
	rows int = 10
	cols int = 10
)

type Breakout struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand

	ballY     int
	ballStart int
	ballX     int
	ballDir   int
	position  int
	brickMap  *mat.Dense
	strike    bool
	lastX     int
	lastY     int

	moveTimer      float64
	terminateTimer int
	terminal       bool
}

func New(_ bool, seed int64) (game.Game, error) {
	channels := map[string]int{
		"paddle": 0,
		"ball":   1,
		"trail":  2,
		"brick":  3,
	}
	actionMap := []rune{'n', 'l', 'u', 'r', 'd', 'f'}
	rng := rand.New(rand.NewSource(seed))

	breakout := &Breakout{
		channels:  channels,
		actionMap: actionMap,
		rng:       rng,
	}
	breakout.Reset()

	return breakout, nil
}

func (b *Breakout) Act(a int) (float64, bool, error) {
	if a >= len(b.actionMap) {
		return -1, false, fmt.Errorf("act: invalid action %v ∉ [0, )",
			len(b.actionMap))
	}

	reward := 0.0
	if b.terminal {
		return reward, b.terminal, nil
	}

	action := b.actionMap[a]
}

func (b *Breakout) State() ([]*mat.Dense, error) {
	state := make([]*mat.Dense, b.NChannels())
	for i := range state {
		state[i] = mat.NewDense(rows, cols, nil)
	}

	state[b.channels["ball"]].Set(b.ballY, b.ballX, 1.0)
	state[b.channels["paddle"]].Set(rows-1, b.position, 1.0)
	state[b.channels["trail"]].Set(b.lastY, b.lastX, 1.0)
	state[b.channels["brick"]] = b.brickMap

	return state, nil
}

func (b *Breakout) Reset() {
	b.ballY = 3
	b.ballStart = b.rng.Intn(2)
	b.ballX = [2]int{0, 9}[b.ballStart]
	b.ballDir = [2]int{2, 3}[b.ballStart]
	b.position = 4
	b.brickMap = mat.NewDense(rows, cols, nil)

	bricks := make([]float64, cols)
	for i := range bricks {
		bricks[i] = 1.0
	}
	for i := 0; i < 4*rows/10; i++ {
		b.brickMap.SetRow(i, bricks)
	}

	b.strike = false
	b.lastX = b.ballX
	b.lastY = b.ballY
	b.terminal = false
}

func (b *Breakout) NChannels() int {
	return len(b.channels)
}

func (b *Breakout) StateShape() []int {
	return []int{b.NChannels(), rows, cols}
}

// MinimalActionSet returns the actions which actually have an effect
// on the environment.
func (b *Breakout) MinimalActionSet() []int {
	minimalActions := []rune{'n', 'l', 'r'}
	minimalIntActions := make([]int, len(minimalActions))

	for i, minimalAction := range minimalActions {
		for j, action := range b.actionMap {
			if minimalAction == action {
				minimalIntActions[i] = j
			}
		}
	}
	return minimalIntActions
}
