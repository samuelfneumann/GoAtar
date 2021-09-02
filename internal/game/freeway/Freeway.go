// Package freeway implements the Freeway game.
//
// "The player begins at the bottom of the screen and the motion is
// restricted to travelling up and down. Player speed is also
// restricted such that the player can only move every 3 frames.
// A reward of +1 is given when the player reaches the top of the
// screen, at which point the player is returned to the bottom. Cars
// travel horizontally on the screen and teleport to the other side when
// the edge is reached. When hit by a car, the player is returned to the
// bottom of the screen. Car direction and speed is indicated by 5 trail
// channels. The location of the trail gives direction while the specific
// channel indicates how frequently the car moves (from once every frame
// to once every 5 frames). Each time the player successfully reaches
// the top of the screen, the car speeds are randomized. Termination
// occurs after 2500 frames have elapsed."
//		- MinAtar (https://github.com/kenjyoung/MinAtar)
package freeway

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
	"gonum.org/v1/gonum/mat"
)

const (
	playerSpeed float64 = 3.0
	timeLimit   int     = 2500

	// Rows and columns for underlying state matrix
	rows int = 8
	cols int = 4

	// Rows and columns for observation matrix
	observationRows int = rows + 2
	observationCols int = rows + 2
)

// Freeway implements the Freeway game. In this game, an agent must
// travel to the top of the screen without colliding with any cars.
//
// See the package documentation for more details.
//
// Underlying state is represented by an integer position of the agent
// (also termed "chicken") and a matrix of information on cars. Each
// row i of the matrix cars refers to the information for car at
// row i in the state observation (recall the game consists of cars
// with fixed Y positions - rows - travelling horizontally). The number
// of rows in this matrix (equivalently, the number of cars in the
// game) is determined by the rows constant. For row i, each column
// has the following meaning:
//
//	Column		Meaning
//	  1			X position of car i
//	  2			Y position of car i (this will be constant)
//	  3			Speed of car i
//	  4			Direction of movement of car i
//
// State observations are constructed based on this underlying state
// representation.
type Freeway struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand

	cars     *mat.Dense // Matrix representing info on each car
	position int        // Position of agent

	moveTimer      float64
	terminateTimer int
	terminal       bool
}

// New returns a new Freeway game
func New(_ bool, seed int64) (game.Game, error) {
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

// State returns the current state observation
func (f *Freeway) State() ([]float64, error) {
	r, c := observationRows, observationCols
	state := make([]float64, r*c*f.NChannels())

	// Set the agent's position in the observation matrix
	state[r*c*f.channels["chicken"]+f.position*c+4] = 1.0

	// Set each car's position in the observation matrix
	for i := 0; i < 8; i++ {
		car := f.cars.RowView(i)
		y, x := int(car.AtVec(1)), int(car.AtVec(0))
		state[r*c*f.channels["car"]+y*c+x] = 1.0

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

		// Find the channel at which to place the car. Each channel
		// refers to a different speed.
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

		backY := int(car.AtVec(1))
		state[r*c*trail+backY*c+backX] = 1.0
	}
	return state, nil
}

// DifficultyRamp returns the current difficulty level.
// In Freeway, difficulty ramping is not allowed, so this method
// always returns 0.
func (f *Freeway) DifficultyRamp() int {
	return 0
}

// Act takes a single environmental step given an action a.
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
		f.moveTimer = playerSpeed
		if 0 > f.position-1 {
			f.position = 0
		} else {
			f.position--
		}
	} else if action == 'd' && f.moveTimer == 0 {
		f.moveTimer = playerSpeed
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

// randomizeCars randomizes all the car directions and speed for the
// start of a new episode.
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

// Reset resets the environment to some starting state.
func (f *Freeway) Reset() {
	f.randomizeCars(true)
	f.position = 9
	f.moveTimer = playerSpeed
	f.terminateTimer = timeLimit
	f.terminal = false
}

// StateShape returns the shape of the state observations
func (f *Freeway) StateShape() []int {
	return []int{f.NChannels(), observationRows, observationCols}
}

// NChannels returns the number of channels in each state observation
func (f *Freeway) NChannels() int {
	return len(f.channels)
}

// MinimalActionSet returns the actions which actually have an effect
// on the environment.
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

// Channel returns the state observation channel at index i
func (f *Freeway) Channel(i int) ([]float64, error) {
	if i >= f.NChannels() {
		return nil, fmt.Errorf("channel: index out of range [%v] with "+
			"length %v", i, f.NChannels())
	} else if i < 0 {
		return nil, fmt.Errorf("channel: invalid slice index %v (index "+
			"must be non-negative)", i)
	}

	state, err := f.State()
	if err != nil {
		return nil, fmt.Errorf("channel: %v", err)
	}

	return state[rows*cols*i : rows*cols*(i+1)], nil
}
