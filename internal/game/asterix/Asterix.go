package asterix

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
)

const (
	rows int = 10
	cols int = rows

	initSpawnSpeed   int = 10
	initMoveInterval int = 5
	shotCoolDown     int = 5
	rampInterval     int = 100

	maxEntities int = 8
)

type Asterix struct {
	channels  map[string]int
	actionMap []rune
	rng       *rand.Rand
	ramping   bool

	agent    *player
	entities []*Entity

	spawnSpeed int
	spawnTimer int
	moveSpeed  int
	rampTimer  int
	rampIndex  int
	terminal   bool
}

func New(ramping bool, seed int64) (game.Game, error) {
	channels := map[string]int{
		"player": 0,
		"enemy":  1,
		"trail":  2,
		"gold":   3,
	}
	actionMap := []rune{'n', 'l', 'u', 'r', 'd', 'f'}
	rng := rand.New(rand.NewSource(seed))

	asterix := &Asterix{
		channels:  channels,
		actionMap: actionMap,
		rng:       rng,
		ramping:   ramping,
	}
	asterix.Reset()

	return asterix, nil
}

func (a *Asterix) Reset() {
	a.entities = make([]*Entity, maxEntities)
	a.spawnSpeed = initSpawnSpeed
	a.spawnTimer = a.spawnSpeed
	a.moveSpeed = initMoveInterval
	a.agent = newPlayer(rows/2, cols/2, 0, a.moveSpeed)
	a.rampTimer = rampInterval
	a.rampIndex = 0
	a.terminal = false
}

func (a *Asterix) Act(act int) (float64, bool, error) {
	if act >= len(a.actionMap) || act < 0 {
		return -1, false, fmt.Errorf("act: invalid action %v ∉ [0, %v)",
			act, len(a.actionMap))
	}

	reward := 0.0
	if a.terminal {
		return reward, a.terminal, nil
	}

	// Spawn enemy if timer is up
	if a.spawnTimer <= 0 {
		a.spawnEntity()
		a.spawnTimer = a.spawnSpeed
	}

	// Resolve player action
	action := a.actionMap[act]
	switch action {
	case 'l':
		a.agent.moveLeft()

	case 'r':
		a.agent.moveRight()

	case 'u':
		a.agent.moveUp()

	case 'd':
		a.agent.moveDown()
	}

	// Update entities
	for i, entity := range a.entities {
		if entity == nil {
			continue
		}

		if entity.x() == a.agent.x() && entity.y() == a.agent.y() {
			if entity.isGold() {
				a.entities[i] = nil
				reward++
			} else {
				a.terminal = true
			}
		}
	}

	// Housekeeping when the agent can move
	if a.agent.canMove() {
		a.agent.setMoveTimer(a.moveSpeed)

		// Entities get updated and moved when the agent moves
		for i, entity := range a.entities {
			if entity == nil {
				continue
			}

			// Entities only move when the agent moves
			entity.move()

			if entity.x() < 0 || entity.x() > cols-1 {
				// Entity moves off the screen
				a.entities[i] = nil
			}

			if entity.x() == a.agent.x() || entity.y() == a.agent.y() {
				if entity.isGold() {
					a.entities[i] = nil
					reward++
				} else {
					a.terminal = true
				}
			}
		}
	}

	// Update timers
	if a.spawnTimer > 0 {
		a.spawnTimer--
	}

	if !a.agent.canMove() {
		a.agent.decrementMoveTimer()
	}

	// Update the difficulty
	if a.ramping && (a.spawnSpeed > 1 || a.moveSpeed > 1) {
		if a.rampTimer >= 0 {
			a.rampTimer--
		} else {
			if a.moveSpeed > 1 && a.rampIndex%2 == 1 {
				a.moveSpeed--
			}

			if a.spawnSpeed > 1 {
				a.spawnSpeed--
			}

			a.rampIndex++
			a.rampTimer = rampInterval
		}
	}

	return reward, a.terminal, nil
}

func (a *Asterix) State() ([]float64, error) {
	state := make([]float64, rows*cols*a.NChannels())

	// Set player location
	state[rows*cols+a.channels["player"]+a.agent.y()*cols+a.agent.x()] = 1.0

	// Set each entity
	for _, entity := range a.entities {
		if entity == nil {
			continue
		}

		// Get the channel for the entity
		ch := a.channels["enemy"]
		if entity.isGold() {
			ch = a.channels["gold"]
		}

		// Set the entity in the state observation tensor
		state[rows*cols*ch+entity.y()*cols+entity.x()] = 1.0

		// Set the trail for the entity, which denotes movement
		backX := entity.x() + 1
		if entity.orientedRight() {
			backX = entity.x() - 1
		}

		if backX >= 0 && backX <= cols-1 {
			state[rows*cols*a.channels["trail"]+entity.y()*cols*backX] = 1.0
		}
	}
	return state, nil
}

func (a *Asterix) Channel(i int) ([]float64, error) {
	if i >= a.NChannels() {
		return nil, fmt.Errorf("channel: index out of range [%v] with "+
			"length %v", i, a.NChannels())
	} else if i < 0 {
		return nil, fmt.Errorf("channel: invalid slice index %v (index "+
			"must be non-negative)", i)
	}

	state, err := a.State()
	if err != nil {
		return nil, fmt.Errorf("channel: %v", err)
	}

	return state[rows*cols*i : rows*cols*(i+1)], nil
}

func (a *Asterix) DifficultyRamp() int {
	return a.rampIndex
}

func (a *Asterix) NChannels() int {
	return len(a.channels)
}

func (a *Asterix) StateShape() []int {
	return []int{a.NChannels(), rows, cols}
}

// MinimalActionSet returns the actions which actually have an effect
// on the environment.
func (a *Asterix) MinimalActionSet() []int {
	minimalActions := []rune{'n', 'l', 'u', 'r', 'd'}
	minimalIntActions := make([]int, len(minimalActions))

	for i, minimalAction := range minimalActions {
		for j, action := range a.actionMap {
			if minimalAction == action {
				minimalIntActions[i] = j
			}
		}
	}
	return minimalIntActions
}

func (a *Asterix) spawnEntity() {
	lr := a.rng.Intn(2)
	isGold := a.rng.Intn(3) == 0

	var x int
	if lr == 1 {
		x = 0
	} else {
		x = cols - 1
	}

	// Get the non-nil slots for entities
	slotOptions := make([]int, 0, maxEntities)
	for i, entity := range a.entities {
		if entity == nil {
			slotOptions = append(slotOptions, i)
		}
	}

	if len(slotOptions) == 0 {
		// At maximum entity capacity
		return
	}

	// Get a random slot at which to add an entity
	slot := slotOptions[a.rng.Intn(len(slotOptions))]
	a.entities[slot] = newEntity(x, slot+1, lr == 1, isGold)
}
