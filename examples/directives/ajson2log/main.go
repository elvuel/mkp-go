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
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <jsonpath> <outputlogpath>", os.Args[0])
	}

	c.InitSFport()

	output, err := helper.AJSON2LogContext(
		context.Background(),
		c.SFPort,
		&mkpgo.AJSON2LogOption{JSONPath: os.Args[1], OutputLogPath: os.Args[2]},
		mkpgo.WithSyncOutputTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalln(err)
	}

	c.Log(output)
}
