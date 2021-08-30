package game

import "gonum.org/v1/gonum/mat"

// Concrete implementations of games
type Game interface {
	Act(int) (float64, bool, error)
	State() ([]*mat.Dense, error)
	Reset()
	StateShape() []int
	NChannels() int
	MinimalActionSet() []int
	DifficultyRamp() int
}
