package main

import (
	"fmt"
	"time"

	"github.com/samuelfneumann/goatar"
)

func main() {
	env, err := goatar.New(goatar.SeaQuest, 0.1, false, time.Now().UnixNano())
	if err != nil {
		panic(err)
	}

	state, _ := env.State()
	fmt.Println(state, len(state), env.NChannels())
	fmt.Println()

	state, _ = env.Channel(0)
	fmt.Println(state, len(state), env.NChannels())
	fmt.Println()

	state, _ = env.Channel(0)
	fmt.Println(state, len(state), env.NChannels())
	fmt.Println()
}
