package abstraction

import "hw/state"

func InitAbstractions(state *state.ProcState) map[string]Abstraction {

	abstractions := make(map[string]Abstraction, 10)
	return abstractions
}

func RegisterAbstraction(abstractions map[string]Abstraction, key string, abs Abstraction) {
	abstractions[key] = abs
}
