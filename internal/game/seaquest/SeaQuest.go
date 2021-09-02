// Package seaquest implements the SeaQuest game.
//
// The player controls a submarine consisting of two cells, front
// and back, to allow direction to be determined. The player can also
// fire bullets from the front of the submarine. Enemies consist of
// submarines and fish, distinguished by the fact that submarines shoot
// bullets and fish do not. A reward of +1 is given each time an enemy
// is struck by one of the player's bullets, at which point the enemy
// is also removed. There are also divers which the player can move onto
// to pick up, doing so increments a bar indicated by another channel
// along the bottom of the screen. The player also has a limited supply
// of oxygen indicated by another bar in another channel. Oxygen
// degrades over time and is replenished whenever the player moves to
// the top of the screen as long as the player has at least one rescued
// diver on board. The player can carry a maximum of 6 divers. When
// surfacing with less than 6, one diver is removed. When surfacing
// with 6, all divers are removed and a reward is given for each active
// cell in the oxygen bar. Each time the player surfaces the difficulty
// is increased by increasing the spawn rate and movement speed of
// enemies. Termination occurs when the player is hit by an enemy fish,
// sub or bullet; or when oxygen reaches 0; or when the player attempts
// to surface with no rescued divers. Enemy and diver directions are
// indicated by a trail channel active in their previous location to
// reduce partial observability.
package seaquest

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
)

const (
	rows int = 10
	cols int = rows

	rampInterval     int = 100
	maxOxygen        int = 200
	maxDivers        int = 6
	initSpawnSpeed   int = 20
	initMoveInterval int = 5
	shotCoolDown     int = 5

	enemyShotInterval int = 10
	enemyMoveInterval int = 5

	diverSpawnSpeed   int = 30
	diverMoveInterval int = 5
)

// SeaQuest implements the SeaQuest game. In this game, the play must
// control a submarine to rescue as many divers as possible, while
// destroying or avoiding enemies.
//
// See the package documentation for more details.
//
// Underlying state is represented by slices of *bullet, *swimmer,
// and *submarine. The agent/player's position is implemented by a
// *player. Each of these structs hold the position of the
// corresponding entity in the state space, which is a 10 x rows x cols
// grid.
//
// State observations consist of a 10 x rows x cols tensor. Each of
// the 10 channels refers to the following entities:
//
//	1.  Agent/player ubmarine front
//	2.  Agent/player submarine back
//	3.  Bullets fired by agent/player
//	4.  Trails behind moving objects, used to infer movement direction
//	5.  Bullets fired by enemy submarines
//	6.  Enemy fish locations
//	7.  Enemy submarine locations
//	8.  Oxygen guage (indicates how much oxygen is left for the agent)
//	9.  Diver guage (indicates how many divers the agent has picked up)
//	10. Diver locations
//
// The state observation tensor contains only 0's and 1's, where a 1
// indicates that an entity exists at the position and a 0 indicates
// that no entity exists at that position. For example, if a 1 exists
// at row i and column j of channel 10, this means that a diver is in
// position (j, i).
type SeaQuest struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand
	ramping   bool

	agent     *player
	fBullets  []*bullet
	moveSpeed int
	shotTimer int
	atSurface bool

	eBullets    []*bullet
	eFish       []*swimmer
	eSubs       []*submarine
	eSpawnSpeed int
	eSpawnTimer int

	divers      []*swimmer
	dSpawnTimer int

	rampIndex int
	terminal  bool
}

// New returns a new SeaQuest game
func New(ramping bool, seed int64) (game.Game, error) {
	channels := map[string]int{
		"sub_front":       0,
		"sub_back":        1,
		"friendly_bullet": 2,
		"trail":           3,
		"enemy_bullet":    4,
		"enemy_fish":      5,
		"enemy_sub":       6,
		"oxygen_guage":    7,
		"diver_guage":     8,
		"diver":           9,
	}
	actionMap := []rune{'n', 'l', 'u', 'r', 'd', 'f'}
	rng := rand.New(rand.NewSource(seed))

	seaquest := &SeaQuest{
		channels:  channels,
		actionMap: actionMap,
		rng:       rng,
	}
	seaquest.Reset()

	return seaquest, nil
}

// Reset resets the environment to some starting state
func (s *SeaQuest) Reset() {
	s.agent = newPlayer(5, 0, false, initMoveInterval, 0, maxOxygen)

	s.fBullets = make([]*bullet, 0, 10)
	s.eBullets = make([]*bullet, 0, 10)
	s.eFish = make([]*swimmer, 0, 10)
	s.eSubs = make([]*submarine, 0, 10)
	s.divers = make([]*swimmer, 0, 10)
	s.eSpawnSpeed = initSpawnSpeed
	s.eSpawnTimer = s.eSpawnSpeed
	s.dSpawnTimer = diverSpawnSpeed
	s.moveSpeed = initMoveInterval
	s.rampIndex = 0
	s.shotTimer = 0
	s.atSurface = true
	s.terminal = false
}

// Act takes on environmental step given some action a and returns the
// reward for that action, as well as whether or not the episode is
// finished.
func (s *SeaQuest) Act(a int) (float64, bool, error) {
	if a >= len(s.actionMap) || a < 0 {
		return -1, false, fmt.Errorf("act: invalid action %v ∉ [0, %v)",
			a, len(s.actionMap))
	}

	reward := 0.
	if s.terminal {
		return reward, s.terminal, nil
	}

	// Spawn enemy if timer is up
	if s.eSpawnTimer == 0 {
		s.spawnEnemy()
		s.eSpawnTimer = s.eSpawnSpeed
	}

	// Spawn diver if timer is up
	if s.dSpawnTimer == 0 {
		s.spawnDiver()
		s.dSpawnTimer = diverSpawnSpeed
	}

	// Resolve action
	action := s.actionMap[a]
	switch action {
	case 'f':
		if s.shotTimer == 0 {
			s.fBullets = append(s.fBullets, newBullet(s.agent.x(),
				s.agent.y(), s.agent.orientedRight()))
			s.shotTimer = shotCoolDown
		}

	case 'l':
		s.agent.moveLeft()

	case 'r':
		s.agent.moveRight()

	case 'u':
		s.agent.moveUp()

	case 'd':
		s.agent.moveDown()
	}

	// Update friendly bullets
	for i := len(s.fBullets) - 1; i > -1; i-- {
		reward += s.updateFriendlyBullet(i)
	}

	// Update divers
	for i := len(s.divers) - 1; i > -1; i-- {
		s.updateDiver(i)
	}

	// Update enemy submarines
	for i := len(s.eSubs) - 1; i > -1; i-- {
		reward += s.updateEnemySubmarine(i)
	}

	// Update enemy bullets
	for i := len(s.eBullets) - 1; i > -1; i-- {
		s.updateEnemyBullet(i)
	}

	// Update enemy fish
	for i := len(s.eFish) - 1; i > -1; i-- {
		reward += s.updateEnemyFish(i)
	}

	// Update timers
	if s.eSpawnTimer > 0 {
		s.eSpawnTimer--
	}

	if s.dSpawnTimer > 0 {
		s.dSpawnTimer--
	}

	if s.shotTimer > 0 {
		s.shotTimer--
	}

	if s.agent.oxygen() < 0 {
		s.terminal = true
	}

	if s.agent.y() > 0 {
		s.agent.decrementOxygen()
		s.atSurface = false
	} else if !s.atSurface {
		if s.agent.divers() == 0 {
			s.terminal = true
		} else {
			reward += s.surface()
		}
	}

	return reward, s.terminal, nil
}

// State returns the current state observation
func (s *SeaQuest) State() ([]float64, error) {
	state := make([]float64, rows*cols*s.NChannels())

	state[rows*cols*s.channels["sub_front"]+cols*s.agent.y()+s.agent.x()] = 1.0

	var backX int
	if s.agent.orientedRight() {
		backX = s.agent.x() - 1
	} else {
		backX = s.agent.x() + 1
	}
	state[rows*cols*s.channels["sub_back"]+cols*s.agent.y()+backX] = 1.0

	// Fill oxygen guage
	for i := 0; i < s.agent.oxygen()*10/maxOxygen; i++ {
		state[rows*cols*s.channels["oxygen_guage"]+(rows-1)*cols+i] = 1.0
	}

	// Add the diver guage
	for i := (rows - 1) - s.agent.divers(); i < (rows - 1); i++ {
		state[rows*cols*s.channels["diver_guage"]+(rows-1)*cols+i] = 1.0
	}

	// Set friendly bullets
	for _, bullet := range s.fBullets {
		state[rows*cols*s.channels["friendly_bullet"]+bullet.y()*cols+
			bullet.x()] = 1.0
	}

	// Set enemy bullets
	for _, bullet := range s.eBullets {
		state[rows*cols*s.channels["enemy_bullet"]+bullet.y()*cols+
			bullet.x()] = 1.0
	}

	// Set the fish
	for _, fish := range s.eFish {
		state[rows*cols*s.channels["enemy_fish"]+fish.y()*cols+
			fish.x()] = 1.0

		// Set the trail behind fish, denoting direction of movement
		var backX int
		if fish.orientedRight() {
			backX = fish.x() - 1
		} else {
			backX = fish.x() + 1
		}

		if backX >= 0 && backX <= rows-1 {
			state[rows*cols*s.channels["trail"]+fish.y()*cols+backX] = 1.0
		}
	}

	// Set the submarines
	for _, sub := range s.eSubs {
		state[rows*cols*s.channels["enemy_sub"]+cols*sub.y()+sub.x()] = 1.0

		// Set the trail behind sub, denoting direction of movement
		var backX int
		if sub.orientedRight() {
			backX = sub.x() - 1
		} else {
			backX = sub.x() + 1
		}

		if backX >= 0 && backX <= rows-1 {
			state[rows*cols*s.channels["trail"]+sub.y()*cols+backX] = 1.0
		}
	}

	// Set the divers
	for _, diver := range s.divers {
		state[rows*cols*s.channels["diver"]+cols*diver.y()+diver.x()] = 1.0

		// Set the trail behind the diver, denoting direction of movement
		var backX int
		if diver.orientedRight() {
			backX = diver.x() - 1
		} else {
			backX = diver.x() + 1
		}

		if backX >= 0 && backX <= rows-1 {
			state[rows*cols*s.channels["trail"]+diver.y()*cols+backX] = 1.0
		}
	}

	return state, nil
}

// StateShape returns the shape of state observations
func (s *SeaQuest) StateShape() []int {
	return []int{s.NChannels(), rows, cols}
}

// MinimalActionSet returns the actions that actually affect the game
func (s *SeaQuest) MinimalActionSet() []int {
	minActions := make([]int, len(s.actionMap))
	for i := range minActions {
		minActions[i] = i
	}

	return minActions
}

// DifficultyRamp returns the current difficulty level of the game
func (s *SeaQuest) DifficultyRamp() int {
	return s.rampIndex
}

// Channel returns the state observation at channel i
func (s *SeaQuest) Channel(i int) ([]float64, error) {
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

// NChannels returns the number of channels in the state observations
func (s *SeaQuest) NChannels() int {
	return len(s.channels)
}

// surface performs the housekeeping when the agent reaches the surface
// of the water, and returns the reward for reaching the surface.
func (s *SeaQuest) surface() float64 {
	var reward float64
	s.atSurface = true

	if s.agent.divers() == maxDivers {
		s.agent.setDivers(0)
		reward = float64(s.agent.oxygen() * 10 / maxOxygen)
	} else {
		reward = 0
		s.agent.setOxygen(maxOxygen)
		s.agent.decrementDivers()

		if s.ramping && (s.eSpawnSpeed > 1 || s.moveSpeed > 2) {
			if s.moveSpeed > 2 && s.rampIndex%2 == 1 {
				s.moveSpeed--
			}
			if s.eSpawnSpeed > 1 {
				s.eSpawnSpeed--
			}
			s.rampIndex++
		}
	}
	return reward
}

// spawnEnemy spawns an enemy into the game at a random position
func (s *SeaQuest) spawnEnemy() {
	lr := s.rng.Intn(2)
	isSub := s.rng.Intn(3) == 0

	var x int
	if lr == 1 {
		x = 0
	} else {
		x = rows - 1
	}

	y := s.rng.Intn(rows-2) + 1

	// Don't spawn in a row already taken an enemy with opposite direction
	// to the enemy fish currently in the row
	for _, enemy := range s.eFish {
		if enemy.y() == y && enemy.direction() != lr {
			// Enemy has same row (y-position) and opposite direction
			// as current enemy in that row
			return
		}
	}
	for _, enemy := range s.eSubs {
		if enemy.y() == y && enemy.direction() != lr {
			// Enemy has same row (y-position) and opposite direction
			// to the enemy submarine currently in that row
			return
		}
	}

	// Spawn enemy
	orientedRight := lr == 1
	if isSub {
		s.eSubs = append(s.eSubs, newSubmarine(x, y, orientedRight,
			s.moveSpeed, enemyShotInterval))
	} else {
		s.eFish = append(s.eFish, newSwimmer(x, y, orientedRight, s.moveSpeed))
	}
}

// spawnDiver spawns a diver into the game at a random position
func (s *SeaQuest) spawnDiver() {
	lr := s.rng.Intn(2)

	var x int
	if lr == 1 {
		x = 0
	} else {
		x = rows - 1
	}

	y := s.rng.Intn(rows-2) + 1

	orientedRight := lr == 1
	s.divers = append(s.divers, newSwimmer(x, y, orientedRight,
		diverMoveInterval))
}

// updateFriendlyBullet updates the friendly bullet at location i in
// the s.fBullet slice and returns the reward for shooting any enemies.
func (s *SeaQuest) updateFriendlyBullet(i int) float64 {
	bullet := s.fBullets[i]
	reward := 0.

	// Move bullet
	bullet.move()

	// Remove the bullet if it leaves the screen
	if bullet.x() < 0 || bullet.y() > rows-1 {
		s.fBullets = append(s.fBullets[:i], s.fBullets[i+1:]...)
	} else {
		removed := false
		// Check if the player shot any enemy fishes
		for i, fish := range s.eFish {
			if bullet.x() == fish.x() && bullet.y() == fish.y() {
				// Remove fish if bullet hit it
				s.eFish = append(s.eFish[:i], s.eFish[i+1:]...)
				reward += 1
				removed = true
				break
			}
		}

		if !removed {
			// Check if the player shot any enemy submarines
			for i, sub := range s.eSubs {
				if bullet.x() == sub.x() && bullet.y() == sub.y() {
					// Remove fish if bullet hit it
					s.eSubs = append(s.eSubs[:i], s.eSubs[i+1:]...)
					reward += 1
					removed = true
					break
				}
			}
		}
	}
	return reward
}

// updateEnemyBullet updates the enemy bullet at location i in the
// s.eBullets slice and determines if the game has ended due to the
// agent being shot
func (s *SeaQuest) updateEnemyBullet(i int) {
	bullet := s.eBullets[i]
	if bullet.x() == s.agent.x() && bullet.y() == s.agent.y() {
		s.terminal = true
	}

	// Move bullet
	bullet.move()

	// Remove bullet if travelling off screen
	if bullet.x() < 0 || bullet.y() > rows-1 {
		s.eBullets = append(s.eBullets[:i], s.eBullets[i+1:]...)
	} else if bullet.x() == s.agent.x() && bullet.y() == s.agent.y() {
		s.terminal = true
	}
}

// updateDiver updates the diver at position i in the s.divers slice
func (s *SeaQuest) updateDiver(i int) {
	diver := s.divers[i]
	if diver.x() == s.agent.x() && diver.y() == s.agent.y() &&
		s.agent.divers() < maxDivers {
		s.divers = append(s.divers[:i], s.divers[i+1:]...)
		s.agent.incrementDivers()
	} else {
		if diver.canMove() {
			diver.setMoveTimer(diverMoveInterval)

			// Move diver
			diver.move()

			// Remove diver if leaving the screen
			if diver.x() < 0 || diver.x() > rows-1 {
				s.divers = append(s.divers[:i], s.divers[i+1:]...)
			} else if diver.x() == s.agent.x() &&
				diver.y() == s.agent.y() && s.agent.divers() < maxDivers {
				s.divers = append(s.divers[:i], s.divers[i+1:]...)
				s.agent.incrementDivers()
			}
		} else {
			diver.decrementMoveTimer()
		}
	}
}

// updateEnemySubmarine updates the enemy submarine at index i in the
// s.eSubs slice, determines if the game is over due to the enemy
// crashing into the player, and returns the reward for if the
// submarine was shot by the player
func (s *SeaQuest) updateEnemySubmarine(i int) float64 {
	sub := s.eSubs[i]
	reward := 0.

	if sub.x() == s.agent.x() && sub.y() == s.agent.y() {
		s.terminal = true
	}

	if sub.canMove() {
		sub.setMoveTimer(s.moveSpeed)

		// Move submarine
		sub.move()

		// Remove submarine if leaving screen
		if sub.x() < 0 || sub.x() > rows-1 {
			s.eSubs = append(s.eSubs[:i], s.eSubs[i+1:]...)
		} else if sub.x() == s.agent.x() && sub.y() == s.agent.y() {
			s.terminal = true
		} else {
			for j, bullet := range s.fBullets {
				if sub.x() == bullet.x() && sub.y() == bullet.y() {
					// Submarine is hit by bullet, remove it
					s.eSubs = append(s.eSubs[:i], s.eSubs[i+1:]...)
					s.fBullets = append(s.fBullets[:j],
						s.fBullets[j+1:]...)
					reward += 1
					break
				}
			}
		}
	} else {
		sub.decrementMoveTimer()
	}

	if sub.canShoot() {
		sub.setShotTimer(enemyShotInterval)
		bullet := newBullet(sub.x(), sub.y(), sub.orientedRight())
		s.eBullets = append(s.eBullets, bullet)
	} else {
		sub.decrementShotTimer()
	}
	return reward
}

// updateEnemyFish updates the fish at index i in the s.eFish slice,
// determines if the game has ended due to the fish crashing into the
// player and returns the reward if the enemy fish was shot
func (s *SeaQuest) updateEnemyFish(i int) float64 {
	fish := s.eFish[i]
	reward := 0.0

	if fish.x() == s.agent.x() && fish.y() == s.agent.y() {
		s.terminal = true
	}

	if fish.canMove() {
		fish.setMoveTimer(s.moveSpeed)

		// Move fish
		fish.move()

		// Remove fish if travelling off screen
		if fish.x() < 0 || fish.y() > rows-1 {
			s.eFish = append(s.eFish[:i], s.eFish[i+1:]...)
		} else if fish.x() == s.agent.x() && fish.y() == s.agent.y() {
			s.terminal = true
		} else {
			// Check if hit by friendly bullet
			for j, bullet := range s.fBullets {
				if fish.x() == bullet.x() && fish.y() == bullet.y() {
					s.eFish = append(s.eFish[:i], s.eFish[i+1:]...)
					s.fBullets = append(s.fBullets[:j],
						s.fBullets[j+1:]...)
					reward += 1
					break
				}
			}
		}
	} else {
		fish.decrementMoveTimer()
	}

	return reward
}

// minInt retruns the minimum int in a group of ints
func minInt(ints ...int) int {
	min := ints[0]
	for i := 1; i < len(ints); i++ {
		if ints[i] < min {
			min = ints[i]
		}
	}
	return min
}

// maxInt retruns the maximum int in a group of ints
func maxInt(ints ...int) int {
	max := ints[0]
	for i := 1; i < len(ints); i++ {
		if ints[i] > max {
			max = ints[i]
		}
	}
	return max
}
