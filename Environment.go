package goatar

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar/game"
	"github.com/samuelfneumann/goatar/game/freeway"
)

const NumActions int = 6 // All games have 6 actions

type gameName string

const (
	Asterix       gameName = "asterix"
	SpaceInvaders gameName = "space invaders"
	Freeway       gameName = "freeway"
)

func make(game gameName, difficultyRamping bool, seed int64) (game.Game, error) {
	switch game {
	case Freeway:
		// create game
		return freeway.New(difficultyRamping, seed)
	default:
		return nil, fmt.Errorf("make: no such game")
	}
}

// Wrapepr that uses Template
type Environment struct {
	game.Game
	gameName          gameName
	rng               *rand.Rand
	nChannels         int
	stickyActionsProb float64
	lastAction        int
	closed            bool
}

func New(name gameName, stickyActionsProb float64, difficultyRamping bool,
	seed int64) (*Environment, error) {
	game, err := make(name, difficultyRamping, seed)
	if err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(seed))

	return &Environment{
		Game:              game,
		gameName:          name,
		rng:               rng,
		nChannels:         game.NChannels(),
		stickyActionsProb: stickyActionsProb,
		lastAction:        -1,
		closed:            false,
	}, nil
}

func (e *Environment) Act(a int) (float64, bool, error) {
	if e.rng.Float64() < e.stickyActionsProb {
		a = e.lastAction
	}
	e.lastAction = a
	return e.Game.Act(a)
}

func (e *Environment) NumActions() int {
	return NumActions
}

func (e *Environment) GameName() string {
	return string(e.gameName)
}
