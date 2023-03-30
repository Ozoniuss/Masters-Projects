package main

import (
	"flag"
)

func initFlags(port, index *int) {

	flag.IntVar(port, "p", 6000, "")
	flag.IntVar(index, "i", 0, "")

	flag.Parse()
}
