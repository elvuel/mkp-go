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
	c.InitSFport()

	var opt *mkpgo.WifiAutoOption
	switch len(os.Args) {
	case 1:
		// No arguments: query current Wi-Fi auto-connect state.
	case 2:
		state := os.Args[1]
		if state != "1" && state != "0" {
			log.Fatalf("usage: %s [0|1]", os.Args[0])
		}
		opt = &mkpgo.WifiAutoOption{State: state}
	default:
		log.Fatalf("usage: %s [0|1]", os.Args[0])
	}

	output, err := helper.WifiAutoContext(context.Background(), c.SFPort, opt, mkpgo.WithSyncOutputTimeout(30*time.Second))
	if err != nil {
		log.Fatalln(err)
	}

	c.Log(output)
}
