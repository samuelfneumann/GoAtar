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
	initSpawnSpeed   int = 20
	initMoveInterval int = 5
	shotCoolDown     int = 5

	enemyShotInterval int = 10
	enemyMoveInterval int = 5

	diverSpawnSpeed   int = 30
	diverMoveInterval int = 5
)

type SeaQuest struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand
	ramping   bool

	oxygen      int
	diverCount  int
	subX        int
	subY        int
	subOr       bool // Orientation - true if going right
	fBullets    []*bullet
	eBullets    []*bullet
	eFish       []*swimmer
	eSubs       []*submarine
	divers      []*swimmer
	eSpawnSpeed int
	eSpawnTimer int
	dSpawnTimer int
	moveSpeed   int
	rampIndex   int
	shotTimer   int
	surface     bool
	terminal    bool
}

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

func (s *SeaQuest) StateShape() []int {
	return []int{s.NChannels(), rows, cols}
}

func (s *SeaQuest) MinimalActionSet() []int {
	minActions := make([]int, len(s.actionMap))
	for i := range minActions {
		minActions[i] = i
	}

	return minActions
}

func (s *SeaQuest) DifficultyRamp() int {
	return s.rampIndex
}

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

func (s *SeaQuest) NChannels() int {
	return len(s.channels)
}

func (s *SeaQuest) Reset() {
	s.oxygen = maxOxygen
	s.diverCount = 0
	s.subX = 5
	s.subY = 0
	s.subOr = false

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
	s.surface = true
	s.terminal = false
}

func (s *SeaQuest) surfaceReward() float64 {
	var reward float64
	s.surface = true

	if s.diverCount == 6 {
		s.diverCount = 0
		reward = float64(s.oxygen * 10 / maxOxygen)
	} else {
		reward = 0
		s.oxygen = maxOxygen
		s.diverCount--

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

func (s *SeaQuest) Act(a int) (float64, bool, error) {
	if a >= len(s.actionMap) {
		return -1, false, fmt.Errorf("act: invalid action %v ∉ [0, )",
			len(s.actionMap))
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
		s.fBullets = append(s.fBullets, newBullet(s.subX, s.subY, s.subOr))
		s.shotTimer = shotCoolDown

	case 'l':
		s.subX = maxInt(0, s.subX-1)
		s.subOr = false

	case 'r':
		s.subX = minInt(rows-1, s.subX+1)
		s.subOr = true

	case 'u':
		s.subY = maxInt(0, s.subY-1)

	case 'd':
		s.subY = minInt(rows-2, s.subY+1)
	}

	// Update friendly bullets
	for i := len(s.fBullets) - 1; i > -1; i-- {
		bullet := s.fBullets[i]

		// Move bullet
		bullet.move()

		// Remove the bullet if it leaves the screen
		if bullet.x() < 0 || bullet.y() > rows-1 {
			s.fBullets = append(s.fBullets[:i], s.fBullets[i+1:]...)
		} else {
			removed := false
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
	}

	// Update divers
	for i := len(s.divers) - 1; i > -1; i-- {
		diver := s.divers[i]
		if diver.x() == s.subX && diver.y() == s.subY && s.diverCount < 6 {
			s.divers = append(s.divers[:i], s.divers[i+1:]...)
			s.diverCount++
		} else {
			if diver.canMove() {
				diver.setMoveTimer(diverMoveInterval)

				// Move diver
				diver.move()

				// Remove diver if leaving the screen
				if diver.x() < 0 || diver.x() > rows-1 {
					s.divers = append(s.divers[:i], s.divers[i+1:]...)
				} else if diver.x() == s.subX && diver.y() == s.subY &&
					s.diverCount < 6 {
					s.divers = append(s.divers[:i], s.divers[i+1:]...)
					s.diverCount++
				}
			} else {
				diver.decrementMoveTimer()
			}
		}
	}

	// Update enemy submarines
	for i := len(s.eSubs) - 1; i > -1; i-- {
		sub := s.eSubs[i]
		if sub.x() == s.subX && sub.y() == s.subY {
			s.terminal = true
		}

		if sub.canMove() {
			sub.setMoveTimer(s.moveSpeed)

			// Move submarine
			sub.move()

			// Remove submarine if leaving screen
			if sub.x() < 0 || sub.x() > rows-1 {
				s.eSubs = append(s.eSubs[:i], s.eSubs[i+1:]...)
			} else if sub.x() == s.subX && sub.y() == s.subY {
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
	}

	// Update enemy bullets
	for i := len(s.eBullets) - 1; i > -1; i-- {
		bullet := s.eBullets[i]
		if bullet.x() == s.subX && bullet.y() == s.subY {
			s.terminal = true
		}

		// Move bullet
		bullet.move()

		// Remove bullet if travelling off screen
		if bullet.x() < 0 || bullet.y() > rows-1 {
			s.eBullets = append(s.eBullets[:i], s.eBullets[i+1:]...)
		} else if bullet.x() == s.subX && bullet.y() == s.subY {
			s.terminal = true
		}
	}

	// Update enemy fish
	for i := len(s.eFish) - 1; i > -1; i++ {
		fish := s.eFish[i]
		if fish.x() == s.subX && fish.y() == s.subY {
			s.terminal = true
		}

		if fish.canMove() {
			fish.setMoveTimer(s.moveSpeed)

			// Move fish
			fish.move()

			// Remove fish if travelling off screen
			if fish.x() < 0 || fish.y() > rows-1 {
				s.eFish = append(s.eFish[:i], s.eFish[i+1:]...)
			} else if fish.x() == s.subX && fish.y() == s.subY {
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

	if s.oxygen < 0 {
		s.terminal = true
	}

	if s.subY > 0 {
		s.oxygen--
		s.surface = false
	} else if !s.surface {
		if s.diverCount == 0 {
			s.terminal = true
		} else {
			reward += s.surfaceReward()
		}
	}

	return reward, s.terminal, nil
}

func minInt(ints ...int) int {
	min := ints[0]
	for i := 1; i < len(ints); i++ {
		if ints[i] < min {
			min = ints[i]
		}
	}
	return min
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

func (s *SeaQuest) State() ([]float64, error) {
	state := make([]float64, rows*cols*s.NChannels())

	state[rows*cols*s.channels["sub_front"]+cols*s.subY+s.subX] = 1.0

	var backX int
	if s.subOr {
		backX = s.subX - 1
	} else {
		backX = s.subX + 1
	}
	state[rows*cols*s.channels["sub_back"]+cols*s.subY+backX] = 1.0

	// Fill oxygen guage
	for i := 0; i < s.oxygen*10/maxOxygen; i++ {
		state[rows*cols*s.channels["oxygen_guage"]+(rows-1)*cols+i] = 1.0
	}

	// Add the diver guage
	for i := (rows - 1) - s.diverCount; i < (rows - 1); i++ {
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
