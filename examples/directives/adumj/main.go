package main

import (
	"context"
	"log"
	"os"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	c "github.com/elvuel/mkp-go/examples/commonpkg"
	"github.com/elvuel/mkp-go/helper"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <logPath>", os.Args[0])
	}

	c.InitSFport()

	dump, err := helper.AdumjContext(
		context.Background(),
		c.SFPort,
		&mkpgo.AdumjOption{LogPath: os.Args[1]},
		mkpgo.WithSyncOutputTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalln(err)
	}

	c.Log(dump)
}
