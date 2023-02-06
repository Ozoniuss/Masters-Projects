package main

import (
	"fmt"
)

func Phi(n uint64) uint64 {
	result := n
	for i := uint64(2); i*i <= n; i++ {
		if n%i == 0 {
			for n%i == 0 {
				n /= i
			}
			result -= result / i
		}
	}
	if n > 1 {
		result -= result / n
	}
	return result
}

func main() {
	fmt.Println(Phi(66))
}
