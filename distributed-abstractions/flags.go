package main

import (
	"flag"
)

func initFlags(port, index, loglevel *int) {

	flag.IntVar(port, "p", 6000, "listening port")
	flag.IntVar(index, "i", 0, "process index")

	flag.Parse()
}
