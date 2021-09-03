// Package goatar implement minimal Atari games that run on a 10x10
// grid. This package was inspired by MinAtar, which can be found at:
// https://github.com/kenjyoung/MinAtar.
package goatar

import (
	"fmt"
	"math/rand"

	"github.com/samuelfneumann/goatar/internal/game"
	"github.com/samuelfneumann/goatar/internal/game/asterix"
	"github.com/samuelfneumann/goatar/internal/game/breakout"
	"github.com/samuelfneumann/goatar/internal/game/freeway"
	"github.com/samuelfneumann/goatar/internal/game/seaquest"
	"github.com/samuelfneumann/goatar/internal/game/spaceinvaders"
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
func make(game GameName, difficultyRamping bool, seed int64) (game.Game,
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
	game, err := make(name, difficultyRamping, seed)
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
