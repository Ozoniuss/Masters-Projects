package abstraction

import (
	"hw/queue"
	"hw/state"
)

func InitAbstractions(state *state.ProcState) map[string]Abstraction {

	abstractions := make(map[string]Abstraction, 10)
	return abstractions
}

func RegisterAbstraction(abstractions *map[string]Abstraction, key string, abs Abstraction) {
	(*abstractions)[key] = abs
}

func RegisterNnar(abstractions *map[string]Abstraction, key string, abs Abstraction, state *state.ProcState, queue *queue.Queue) {
	(*abstractions)[key] = abs

	// For this nnar, register beb and pl.
	RegisterAbstraction(abstractions, key+".beb", NewBeb(state, queue, key+".beb"))
	RegisterAbstraction(abstractions, key+".beb.pl", NewPl(state, queue, key+".beb.pl", abstractions))
	RegisterAbstraction(abstractions, key+".pl", NewPl(state, queue, key+".pl", abstractions))
}
