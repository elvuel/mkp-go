package main

import (
	"log"

	"github.com/elvuel/mkp-go/cmd/client/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
