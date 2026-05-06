package main

import (
	"context"
	"log"

	c "github.com/elvuel/mkp-go/examples/commonpkg"
	"github.com/elvuel/mkp-go/helper"
)

func main() {
	c.InitSFport()

	err := helper.CleanDirContext(context.Background(), c.SFPort, "/eMMC/applog/subdir1")

	if err != nil {
		log.Fatalln(err)
	}

	c.Log("succeed")
}
