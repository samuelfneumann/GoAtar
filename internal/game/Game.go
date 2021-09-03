package game

import (
	"gonum.org/v1/gonum/mat"
)

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

// minInt retruns the minimum int in a group of ints
func MinInt(ints ...int) int {
	min := ints[0]
	for i := 1; i < len(ints); i++ {
		if ints[i] < min {
			min = ints[i]
		}
	}
	return min
}

// maxInt retruns the maximum int in a group of ints
func MaxInt(ints ...int) int {
	max := ints[0]
	for i := 1; i < len(ints); i++ {
		if ints[i] > max {
			max = ints[i]
		}
	}
	return max
}

// clipInt clips an integer to be in the interval [min, max]
func ClipInt(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

// containsNonZero returns whether a matrix contains any non-zero
// elements
func ContainsNonZero(matrix *mat.Dense) bool {
	for _, val := range matrix.RawMatrix().Data {
		if val != 0.0 {
			return true
		}
	}
	return false
}

// CountNonZero returns the number of nonzero elements in the matrix
func CountNonZero(matrix *mat.Dense) int {
	total := 0
	for _, elem := range matrix.RawMatrix().Data {
		if elem == 0.0 {
			total++
		}
	}
	return total
}

// Where returns the indices in slice where condition is true
func Where(vec mat.Vector, condition func(i float64) bool) []int {
	var indices []int
	for i := 0; i < vec.Len(); i++ {
		if condition(vec.AtVec(i)) {
			indices = append(indices, i)
		}
	}
	return indices
}

// RollRowsUp rolls the rows of the matrix upwards. Rows that would go
// off the matrix's top wrap around back to the bottom.
func RollRowsUp(matrix *mat.Dense) {
	r, c := matrix.Dims()
	tmp1 := make([]float64, c)
	tmp2 := make([]float64, c)

	copy(tmp1, matrix.RawRowView(r-1))
	for i := r - 1; i > 0; i-- {
		copy(tmp2, matrix.RawRowView(i-1))
		matrix.SetRow(i-1, tmp1)
		copy(tmp1, tmp2)
	}
	matrix.SetRow(r-1, tmp1)
}

// RollRowsDown rolls the rows of the matrix downwards. Rows that
// would go off the matrix's bottom wrap around back to the top.
func RollRowsDown(matrix *mat.Dense) {
	r, c := matrix.Dims()
	tmp1 := make([]float64, c)
	tmp2 := make([]float64, c)

	copy(tmp1, matrix.RawRowView(0))
	for i := 0; i < r-1; i++ {
		copy(tmp2, matrix.RawRowView(i+1))
		matrix.SetRow(i+1, tmp1)
		copy(tmp1, tmp2)
	}
	matrix.SetRow(0, tmp1)
}

// RollColsLeft rolls the columns of the matrix left. Columns that
// would go off the matrix's side wrap around back to the other side.
func RollColsLeft(matrix *mat.Dense) {
	r, c := matrix.Dims()
	tmp1 := make([]float64, r)
	tmp2 := make([]float64, r)

	vecToSlice := func(slice []float64, vec mat.Vector) {
		for i := 0; i < vec.Len(); i++ {
			slice[i] = vec.AtVec(i)
		}
	}

	vecToSlice(tmp1, matrix.ColView(c-1))
	for i := c - 1; i > 0; i-- {
		vecToSlice(tmp2, matrix.ColView(i-1))
		matrix.SetCol(i-1, tmp1)
		copy(tmp1, tmp2)
	}
	matrix.SetCol(c-1, tmp1)
}

// RollColsRight rolls the columns of the matrix right. Columns that
// would go off the matrix's side wrap around back to the other side.
func RollColsRight(matrix *mat.Dense) {
	r, c := matrix.Dims()
	tmp1 := make([]float64, r)
	tmp2 := make([]float64, r)

	vecToSlice := func(slice []float64, vec mat.Vector) {
		for i := 0; i < vec.Len(); i++ {
			slice[i] = vec.AtVec(i)
		}
	}

	vecToSlice(tmp1, matrix.ColView(0))
	for i := 0; i < c-1; i++ {
		vecToSlice(tmp2, matrix.ColView(i+1))
		matrix.SetCol(i+1, tmp1)
		copy(tmp1, tmp2)
	}
	matrix.SetCol(0, tmp1)
}
