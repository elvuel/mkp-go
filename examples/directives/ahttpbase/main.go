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
	var opt *mkpgo.AHTTPBaseOption
	switch len(os.Args) {
	case 1:
		// No arguments: query current file-management API endpoint base URL.
	case 2:
		opt = &mkpgo.AHTTPBaseOption{URL: os.Args[1]}
	default:
		log.Fatalf("usage: %s [http-base-url]", os.Args[0])
	}

	c.InitSFport()

	base, err := helper.AHTTPBaseContext(
		context.Background(),
		c.SFPort,
		opt,
		mkpgo.WithSyncOutputTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalln(err)
	}

	c.Log(base)
}
