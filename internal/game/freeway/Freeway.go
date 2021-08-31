package freeway

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
	"gonum.org/v1/gonum/mat"
)

const (
	PlayerSpeed float64 = 3.0
	TimeLimit   int     = 2500

	// Rows and columns for underlying state matrix
	rows int = 8
	cols int = 4

	// Rows and columns for observation matrix
	observationRows int = 10
	observationCols int = 10
)

type Freeway struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand
	cars      *mat.Dense

	position       int
	moveTimer      float64
	terminateTimer int
	terminal       bool
}

func New(difficultyRamping bool, seed int64) (game.Game, error) {
	channels := map[string]int{
		"chicken": 0,
		"car":     1,
		"speed1":  2,
		"speed2":  3,
		"speed3":  4,
		"speed4":  5,
		"speed5":  6,
	}
	actionMap := []rune{'n', 'l', 'u', 'r', 'd', 'f'}
	rng := rand.New(rand.NewSource(seed))

	freeway := &Freeway{
		channels:  channels,
		actionMap: actionMap,
		rng:       rng,
	}
	freeway.Reset()

	return freeway, nil
}

func (f *Freeway) State() ([]*mat.Dense, error) {
	state := make([]*mat.Dense, f.NChannels())
	for i := 0; i < f.NChannels(); i++ {
		state[i] = mat.NewDense(observationRows, observationCols, nil)
	}

	state[f.channels["chicken"]].Set(f.position, 4, 1.)

	for i := 0; i < 8; i++ {
		car := f.cars.RowView(i)
		state[f.channels["car"]].Set(int(car.AtVec(1)), int(car.AtVec(0)), 1.)

		var backX int
		if car.AtVec(3) > 0 {
			backX = int(car.AtVec(0)) - 1
		} else {
			backX = int(car.AtVec(0)) + 1
		}

		if backX < 0 {
			backX = 9
		} else if backX > 9 {
			backX = 0
		}

		var trail int
		switch int(math.Abs(car.AtVec(3))) {
		case 1:
			trail = f.channels["speed1"]

		case 2:
			trail = f.channels["speed2"]

		case 3:
			trail = f.channels["speed3"]

		case 4:
			trail = f.channels["speed4"]

		case 5:
			trail = f.channels["speed5"]

		default:
			return nil, fmt.Errorf("state: no such speed value %v",
				int(math.Abs(car.AtVec(3))))
		}

		state[trail].Set(int(car.AtVec(1)), backX, 1)
	}
	return state, nil
}

func (f *Freeway) DifficultyRamp() int {
	return 0
}

func (f *Freeway) Act(a int) (float64, bool, error) {
	reward := 0.0
	if f.terminal {
		return reward, f.terminal, nil
	}

	if a >= len(f.actionMap) {
		return -1, false, fmt.Errorf("act: invalid action %v ∉ [0, )",
			len(f.actionMap))
	}

	// Update the environment with respect to the action
	action := f.actionMap[a]
	if action == 'u' && f.moveTimer == 0 {
		f.moveTimer = PlayerSpeed
		if 0 > f.position-1 {
			f.position = 0
		} else {
			f.position--
		}
	} else if action == 'd' && f.moveTimer == 0 {
		f.moveTimer = PlayerSpeed
		if 9 < f.position {
			f.position = 9
		} else {
			f.position++
		}
	}

	// Win condition
	if f.position == 0 {
		reward += 1
		f.randomizeCars(false)
		f.position = 9
	}

	r, _ := f.cars.Dims()
	for i := 0; i < r; i++ {
		if f.cars.At(i, 0) == 4 && f.cars.At(i, 1) == float64(f.position) {
			f.position = 9
		}
		if f.cars.At(i, 2) == 0.0 {
			f.cars.Set(i, 2, math.Abs(f.cars.At(i, 3)))

			if f.cars.At(i, 3) > 0 {
				f.cars.Set(i, 0, f.cars.At(i, 0)+1)
			} else {
				f.cars.Set(i, 0, 9)
			}

			if f.cars.At(i, 0) > 9 {
				f.cars.Set(i, 0, 0)
			}

			if f.cars.At(i, 0) == 4.0 &&
				f.cars.At(i, 1) == float64(f.position) {
				f.position = 9
			}
		} else {
			f.cars.Set(i, 2, f.cars.At(i, 2)-1)
		}
	}

	// Update various timers
	if f.moveTimer > 0 {
		f.moveTimer--
	}
	f.terminateTimer -= 1
	if f.terminateTimer < 0 {
		f.terminal = true
	}

	return reward, f.terminal, nil
}

func (f *Freeway) randomizeCars(init bool) {
	var directions [rows]float64
	for i := range directions {
		if float64(f.rng.Intn(2)-1) == 0 {
			directions[i] = -1.0
		} else {
			directions[i] = 1.0
		}
	}

	var speeds [rows]float64
	for i := range speeds {
		speeds[i] = directions[i] * float64(f.rng.Intn(4)+1)
	}

	if init {
		cars := make([]float64, rows*cols)
		for i := 0; i < rows; i++ {
			cars[cols*i] = 0.0
			cars[cols*i+1] = float64(i + 1)
			cars[cols*i+2] = math.Abs(speeds[i])
			cars[cols*i+3] = speeds[i]
		}
		f.cars = mat.NewDense(rows, cols, cars)
	} else {
		for i := 0; i < rows; i++ {
			f.cars.Set(i, 2, math.Abs(speeds[i]))
			f.cars.Set(i, 3, speeds[i])
		}
	}
}

func (f *Freeway) Reset() {
	f.randomizeCars(true)
	f.position = 9
	f.moveTimer = PlayerSpeed
	f.terminateTimer = TimeLimit
	f.terminal = false
}

func (f *Freeway) StateShape() []int {
	return []int{observationRows, observationCols, f.NChannels()}
}

func (f *Freeway) NChannels() int {
	return len(f.channels)
}

func (f *Freeway) MinimalActionSet() []int {
	minimalActions := []rune{'n', 'u', 'd'}
	minimalIntActions := make([]int, len(minimalActions))

	for i, minimalAction := range minimalActions {
		for j, action := range f.actionMap {
			if minimalAction == action {
				minimalIntActions[i] = j
			}
		}
	}

	return minimalIntActions
}
