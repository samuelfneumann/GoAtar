package main

import (
	"fmt"

	"github.com/samuelfneumann/goatar"
)

func main() {
	env, err := goatar.New(goatar.Freeway, 0.1, false, 12)
	if err != nil {
		panic(err)
	}

	state, _ := env.State()
	fmt.Println(state[1])
}
