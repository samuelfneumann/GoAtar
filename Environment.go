package goatar

import (
	"fmt"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

const NumActions int = 6 // All games have 6 actions

type gameName string

const (
	Asterix       gameName = "asterix"
	SpaceInvaders gameName = "space invaders"
)

func make(game gameName, difficultyRamping bool, seed int64) (Game, error) {
	switch game {
	case Asterix:
		// create game
		return nil, fmt.Errorf("make: no such game")
	default:
		return nil, fmt.Errorf("make: no such game")
	}
}

// Concrete implementations of games
type Game interface {
	Act(int) (float64, bool)
	State() []*mat.Dense
	Reset()
	StateShape() []int
	NChannels() int
	MinimalActionSet() []int
}

// Wrapepr that uses Template
type Environment struct {
	Game
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

func (e *Environment) Act(a int) (float64, bool) {
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
