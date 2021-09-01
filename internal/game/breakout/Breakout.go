package breakout

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
	"gonum.org/v1/gonum/mat"
)

const (
	rows int = 10
	cols int = rows
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

	terminal bool
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

	// Resolve player action
	action := b.actionMap[a]
	switch action {
	case 'l':
		b.position = maxInt(0, b.position-1)
	case 'r':
		b.position = maxInt(rows-1, b.position+1)
	}

	// Update ball position
	b.lastX = b.ballX
	b.lastY = b.ballY
	var newX, newY int
	switch b.ballDir {
	case 0:
		newX = b.ballX - 1
		newY = b.ballY - 1

	case 1:
		newX = b.ballX + 1
		newY = b.ballY - 1

	case 2:
		newX = b.ballX + 1
		newY = b.ballY + 1

	case 3:
		newX = b.ballX - 1
		newY = b.ballY + 1

	default:
		return 0, false, fmt.Errorf("act: no such ball direction %v", b.ballDir)
	}

	strikeToggle := false
	if newX < 0 || newX > rows-1 {
		newX = clipInt(newX, 0, rows-1)
		b.ballDir = [4]int{1, 0, 3, 2}[b.ballDir]
	}
	if newY < 0 {
		newY = 0
		b.ballDir = [4]int{3, 2, 1, 0}[b.ballDir]
	} else if b.brickMap.At(newY, newX) == 1.0 {
		strikeToggle = true
		if !b.strike {
			reward++
			b.strike = true
			b.brickMap.Set(newY, newX, 0.0)
			newY = b.lastY
			b.ballDir = [4]int{3, 2, 1, 0}[b.ballDir]
		}
	} else if newY == cols-1 {
		if containsNonZero(b.brickMap) {
			bricks := make([]float64, cols)
			for i := range bricks {
				bricks[i] = 1.0
			}
			for i := 0; i < 4*rows/10; i++ {
				b.brickMap.SetRow(i, bricks)
			}
		}

		if b.ballX == b.position {
			b.ballDir = [4]int{3, 2, 1, 0}[b.ballDir]
			newY = b.lastY
		} else if newX == b.position {
			b.ballDir = [4]int{2, 3, 0, 1}[b.ballDir]
			newY = b.lastY
		} else {
			b.terminal = true
		}
	}

	if !strikeToggle {
		b.strike = false
	}

	b.ballX = newX
	b.ballY = newY
	return reward, b.terminal, nil
}

func (b *Breakout) State() ([]float64, error) {
	state := make([]float64, rows*cols*b.NChannels())

	state[rows*cols*b.channels["ball"]+cols*b.ballY+b.ballX] = 1.0

	state[rows*cols*b.channels["paddle"]+(rows-1)*cols+b.position] = 1.0
	state[rows*cols*b.channels["trail"]+b.lastY*cols+b.lastX] = 1.0
	copy(state[rows*cols*b.channels["brick"]:], b.brickMap.RawMatrix().Data)

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

func (b *Breakout) DifficultyRamp() int {
	return 0
}

func (b *Breakout) StateShape() []int {
	return []int{rows, cols, b.NChannels()}
}

func (b *Breakout) Channel(i int) ([]float64, error) {
	if i >= b.NChannels() {
		return nil, fmt.Errorf("channel: index out of range [%v] with "+
			"length %v", i, b.NChannels())
	} else if i < 0 {
		return nil, fmt.Errorf("channel: invalid slice index %v (index "+
			"must be non-negative)", i)
	}

	state, err := b.State()
	if err != nil {
		return nil, fmt.Errorf("channel: %v", err)
	}

	return state[rows*cols*i : rows*cols*(i+1)], nil
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

func maxInt(ints ...int) int {
	max := ints[0]
	for i := 1; i < len(ints); i++ {
		if ints[i] > max {
			max = ints[i]
		}
	}
	return max
}

func clipInt(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func containsNonZero(matrix *mat.Dense) bool {
	for _, val := range matrix.RawMatrix().Data {
		if val != 0.0 {
			return true
		}
	}
	return false
}
