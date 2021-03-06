// Package spaceinvaders implements the SpaceInvaders game
//
//The player controls a cannon at the bottom of the screen and can
// shoot bullets upward at a cluster of aliens above. The aliens move
// across the screen until one of them hits the edge, at which point
// they all move down and switch directions. The current alien direction
// is indicated by 2 channels (one for left and one for right) one of
// which is active at the location of each alien. A reward of +1 is
// given each time an alien is shot, and that alien is also removed.
// The aliens will also shoot bullets back at the player. When few
// aliens are left, alien speed will begin to increase. When only one
// alien is left, it will move at one cell per frame. When a wave of
// aliens is fully cleared, a new one will spawn which moves at a
// slightly faster speed than the last. Termination occurs when an
// alien or bullet hits the player.
package spaceinvaders

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/samuelfneumann/goatar/internal/game"
	"gonum.org/v1/gonum/mat"
)

const (
	rows int = 10
	cols int = rows

	enemyMoveInterval = 12
	enemyShotInterval = 10
	shotCoolDown      = 5
)

// SpaceInvaders implements the SpaceInvaders game. In this game,
// the player must shoot all enemy aliens, while avoiding being
// shot by the enemies.
//
// See the package documentation for more details
//
// Underlying state is represented as a *player, denoting the player's
// position, and a *mat.Dense denoting the positions of the player's
// bullets, the enemies' bullets, and the aliens. Each element in these
// *mat.Dense represent a specific position on the screen.
//
// State observations consist of a 6 x rows x cols tensor. Each of the
// six channels represents:
//
//	1. Player's position (sometimes referred to as the cannon)
//	2. Positions of aliens
//	3. The trail behind the aliens, if they moved left last, else 0
//	4. The trail behind the aliens, if they moved right last, else 0
//	5. Positions of player's bullets
//	6. Positions of enemies' bullets
//
// The state observation tensor contains only 0's and 1's, where a 1
// indicates that a game element exists at the position and a 0
// indicates that no entity exists at that position. For example,
// if a 1 exists at row i and column j of channel 2, this means that
// an enemy alien is in position (j, i).
type SpaceInvaders struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand
	ramping   bool
	rampIndex int
	terminal  bool

	agent    *player
	fBullets *mat.Dense

	eBullets          *mat.Dense
	aliens            *mat.Dense
	alienDir          int
	enemyMoveInterval int
	alienMoveTimer    int
	alienShotTimer    int

	// currentState caches the last state of the environment to increase
	// computational efficiency if State() is called many times
	currentState []float64
}

// New returns a new SpaceInvaders game
func New(ramping bool, seed int64) (game.Game, error) {
	channels := map[string]int{
		"cannon":          0,
		"alien":           1,
		"alien_left":      2,
		"alien_right":     3,
		"friendly_bullet": 4,
		"enemy_bullet":    5,
	}
	actionMap := []rune{'n', 'l', 'u', 'r', 'd', 'f'}
	rng := rand.New(rand.NewSource(seed))

	spaceInvaders := &SpaceInvaders{
		channels:  channels,
		actionMap: actionMap,
		rng:       rng,
		ramping:   ramping,
	}
	spaceInvaders.Reset()

	return spaceInvaders, nil
}

// Act takes one environmental step, given some action a, and returns
// the reward for that action and whether the episode is finished.
func (s *SpaceInvaders) Act(a int) (float64, bool, error) {
	if a >= len(s.actionMap) || a < 0 {
		return -1, false, fmt.Errorf("act: invalid action %v ??? [0, %v)",
			a, len(s.actionMap))
	}

	reward := 0.0
	if s.terminal {
		return reward, s.terminal, nil
	}

	// Resolve player action
	action := s.actionMap[a]
	switch action {
	case 'f':
		if s.agent.canShoot() {
			s.fBullets.Set(rows-1, s.agent.x(), 1.0)
			s.agent.setShotTimer(shotCoolDown)
		}

	case 'l':
		s.agent.moveLeft()

	case 'r':
		s.agent.moveRight()
	}

	// Update friendly bullets
	game.RollRowsUp(s.fBullets)
	s.fBullets.SetRow(rows-1, make([]float64, cols))

	// Update enemy bullets
	game.RollRowsDown(s.eBullets)
	s.eBullets.SetRow(0, make([]float64, cols))
	if s.eBullets.At(rows-1, s.agent.x()) == 1.0 {
		s.terminal = true
	}

	// Update aliens
	if s.aliens.At(rows-1, s.agent.x()) == 1.0 {
		s.terminal = true
	}
	if s.alienMoveTimer == 0 {
		s.alienMoveTimer = game.MinInt(s.enemyMoveInterval,
			game.CountNonZero(s.aliens))

		if (mat.Sum(s.aliens.ColView(0)) > 0 && s.alienDir < 0) ||
			(mat.Sum(s.aliens.ColView(cols-1)) > 0 && s.alienDir > 0) {
			s.alienDir = -s.alienDir

			// Aliens have made it to the bottom of the screen
			if mat.Sum(s.aliens.RowView(rows-1)) > 0 {
				s.terminal = true
			}

			game.RollRowsDown(s.aliens)
		} else {
			// Move aliens left or right
			if s.alienDir < 0 {
				game.RollColsLeft(s.aliens)
			} else {
				game.RollColsRight(s.aliens)
			}
		}
		if s.aliens.At(rows-1, s.agent.x()) == 1.0 {
			s.terminal = true
		}
	}
	if s.alienShotTimer == 0 {
		// Shoot from the nearest alien
		s.alienShotTimer = enemyShotInterval
		nearestAlienX, nearestAlienY := s.nearestAlien(s.agent.x())
		if nearestAlienX > 0 && nearestAlienY > 0 {
			s.eBullets.Set(nearestAlienX, nearestAlienY, 1.0)
		}
	}

	// Find where the aliens were killed
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if s.fBullets.At(r, c) == 1.0 && s.aliens.At(r, c) == 1.0 {
				reward++
				s.aliens.Set(r, c, 0.0)
				s.fBullets.Set(r, c, 0.0)
			}
		}
	}

	// Update timers
	if !s.agent.canShoot() {
		s.agent.decrementShotTimer()
	}

	s.alienMoveTimer--
	s.alienShotTimer--

	// All aliens have been destroyed, reset them at the top and increase
	// the difficulty
	if game.CountNonZero(s.aliens) == 0 {
		if s.enemyMoveInterval > 0 && s.ramping { // MinAtar has > 6
			s.enemyMoveInterval--
			s.rampIndex++
		}
		// Set the aliens
		aliens := make([]float64, cols)
		for i := 2; i < cols-2; i++ {
			aliens[i] = 1
		}
		s.aliens = mat.NewDense(rows, cols, nil)
		for i := 0; i < 4*rows/10; i++ {
			s.aliens.SetRow(i, aliens)
		}
	}

	// Clear current state so next time State() is called it will be
	// recalculated and cached
	s.currentState = nil

	return reward, s.terminal, nil
}

// State returns the current state observation
func (s *SpaceInvaders) State() ([]float64, error) {
	if s.currentState != nil {
		return s.currentState, nil
	}

	state := make([]float64, rows*cols*s.NChannels())

	// Set the cannon at the bottom of the screen
	state[rows*cols*s.channels["cannon"]+(rows-1)*cols+s.agent.x()] = 1.0

	// Set the aliens channel
	start := rows * cols * (s.channels["alien"])
	end := rows * cols * (s.channels["alien"] + 1)
	copied := copy(state[start:end], s.aliens.RawMatrix().Data)
	if copied != rows*cols {
		return nil, fmt.Errorf("state: could not copy aliens channel " +
			"into state observation tensor")
	}

	// Set the alien movement direction channel
	if s.alienDir < 0 {
		start = rows * cols * (s.channels["alien_left"])
		end = rows * cols * (s.channels["alien_left"] + 1)
	} else {
		start = rows * cols * (s.channels["alien_right"])
		end = rows * cols * (s.channels["alien_right"] + 1)
	}
	copied = copy(state[start:end], s.aliens.RawMatrix().Data)
	if copied != rows*cols {
		return nil, fmt.Errorf("state: could not copy aliens direction " +
			"channel into state observation tensor")
	}

	// Set the friendly bullet channel
	start = rows * cols * (s.channels["friendly_bullet"])
	end = rows * cols * (s.channels["friendly_bullet"] + 1)
	copied = copy(state[start:end], s.fBullets.RawMatrix().Data)
	if copied != rows*cols {
		return nil, fmt.Errorf("state: could not copy friendly bullets " +
			"channel into state observation tensor")
	}

	// Set the enemy bullet channel
	start = rows * cols * (s.channels["enemy_bullet"])
	end = rows * cols * (s.channels["enemy_bullet"] + 1)
	copied = copy(state[start:end], s.eBullets.RawMatrix().Data)
	if copied != rows*cols {
		return nil, fmt.Errorf("state: could not copy enemy bullets " +
			"channel into state observation tensor")
	}

	// Cache the state observation
	s.currentState = state

	return state, nil
}

// Reset resets the environment to some starting state
func (s *SpaceInvaders) Reset() {
	start := s.rng.Intn(rows/4) + rows/2
	s.agent = newPlayer(start, 0)
	s.fBullets = mat.NewDense(rows, cols, nil)
	s.eBullets = mat.NewDense(rows, cols, nil)

	// Set the aliens
	aliens := make([]float64, cols)
	for i := 2; i < cols-2; i++ {
		aliens[i] = 1
	}
	s.aliens = mat.NewDense(rows, cols, nil)
	for i := 0; i < 4*rows/10; i++ {
		s.aliens.SetRow(i, aliens)
	}

	s.alienDir = -1
	s.enemyMoveInterval = enemyMoveInterval
	s.alienMoveTimer = s.enemyMoveInterval
	s.alienShotTimer = enemyShotInterval
	s.rampIndex = 0
	s.terminal = false

	s.currentState = nil
}

// Channel returns the channel at index i of the state observation
// tensor
func (s *SpaceInvaders) Channel(i int) ([]float64, error) {
	if i >= s.NChannels() {
		return nil, fmt.Errorf("channel: index out of range [%v] with "+
			"length %v", i, s.NChannels())
	} else if i < 0 {
		return nil, fmt.Errorf("channel: invalid slice index %v (index "+
			"must be non-negative)", i)
	}

	state, err := s.State()
	if err != nil {
		return nil, fmt.Errorf("channel: %v", err)
	}

	return state[rows*cols*i : rows*cols*(i+1)], nil
}

// NChannels returns the number of channels in the state observation
// tensor
func (s *SpaceInvaders) NChannels() int {
	return len(s.channels)
}

// DifficultyRamp returns the current difficulty level
func (s *SpaceInvaders) DifficultyRamp() int {
	return s.rampIndex
}

// StateShape returns the shape of state observation tensors
func (s *SpaceInvaders) StateShape() []int {
	return []int{s.NChannels(), rows, cols}
}

// MinimalActionSet returns the actions which actually have an effect
// on the environment.
func (s *SpaceInvaders) MinimalActionSet() []int {
	minimalActions := []rune{'n', 'l', 'r', 'f'}
	minimalIntActions := make([]int, len(minimalActions))

	for i, minimalAction := range minimalActions {
		for j, action := range s.actionMap {
			if minimalAction == action {
				minimalIntActions[i] = j
			}
		}
	}
	return minimalIntActions
}

// nearestAlien finds the alien closest to pos in terms of Manhattan
// distance. This is usually used to find the alien that will shoot
// next.
func (s *SpaceInvaders) nearestAlien(pos int) (x, y int) {
	searchOrder := make([]int, rows)
	for i := range searchOrder {
		searchOrder[i] = i
	}

	sort.Slice(searchOrder, func(i, j int) bool {
		return math.Abs(float64(i-pos)) < math.Abs(float64(j-pos))
	})

	for _, i := range searchOrder {
		if mat.Sum(s.aliens.ColView(i)) > 0. {
			aliensAt := game.Where(s.aliens.ColView(i), func(i float64) bool {
				return i != 0.0
			})
			x = game.MaxInt(aliensAt...)
			y = i
			return
		}
	}
	return -1, -1
}
