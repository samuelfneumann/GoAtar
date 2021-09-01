package game

// Concrete implementations of games
type Game interface {
	Act(int) (float64, bool, error)

	// State returns the state observation in row-major order.
	// Since observations are of the form (rows, cols, channels),
	// the elements at n*rows*cols to (n+1)*rows*cols are the rows and
	// columns of channel n in row major order.
	State() ([]float64, error)

	Reset()

	// Returns the shape of the state observation in rows, columns,
	// chnnels
	StateShape() []int

	Channel(i int) ([]float64, error) // Returns the matrix at channel i
	NChannels() int

	MinimalActionSet() []int
	DifficultyRamp() int
}
