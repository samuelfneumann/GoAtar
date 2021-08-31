package main

import (
	"fmt"
	"time"

	"github.com/samuelfneumann/goatar"
)

func main() {
	env, err := goatar.New(goatar.Freeway, 0.1, false, time.Now().UnixNano())
	if err != nil {
		panic(err)
	}

	state, _ := env.State()
	fmt.Println(state[3])
}
