package main

import (
	"context"
	"log"

	c "github.com/elvuel/mkp-go/examples/commonpkg"
	"github.com/elvuel/mkp-go/helper"
)

func main() {
	c.InitSFport()

	output, err := helper.AliveContext(context.Background(), c.SFPort)

	if err != nil {
		log.Fatalln(err)
	}

	c.Log(output)
}
