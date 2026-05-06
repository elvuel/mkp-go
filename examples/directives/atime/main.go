package main

import (
	"context"
	"log"

	c "github.com/elvuel/mkp-go/examples/commonpkg"
	"github.com/elvuel/mkp-go/helper"
)

func main() {
	c.InitSFport()

	output, err := helper.AtimeContext(context.Background(), c.SFPort, "/eMMC/applog/uuu2.log")

	if err != nil {
		log.Fatalln(err)
	}

	c.Log(output)
}
