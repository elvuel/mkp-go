package main

import (
	"log"

	"github.com/elvuel/mkp-go/cmd/client/cmd"
)

func main() {
	cmd.SetVersion(Version)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
