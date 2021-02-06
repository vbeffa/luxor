package main

import (
	"log"

	"vbeffa/luxor"
)

func main() {
	s := luxor.Server{}
	log.Fatal(s.Start(1234, func() {}))
}
