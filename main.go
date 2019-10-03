package main

import (
	"log"

	"github.com/DeadLemon/crony/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		log.Fatal(err)
	}
}
