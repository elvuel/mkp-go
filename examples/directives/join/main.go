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

	var opt *mkpgo.JoinOption
	switch len(os.Args) {
	case 1:
		// No arguments: use the most recently saved Wi-Fi configuration on the device.
	case 3:
		opt = &mkpgo.JoinOption{
			SSID:     os.Args[1],
			Password: os.Args[2],
		}
	default:
		log.Fatalf("usage: %s [ssid password]", os.Args[0])
	}

	c.SFPort.SyncOutputTimeout = 5 * time.Second

	output, err := helper.JoinContext(context.Background(), c.SFPort, opt)
	if err != nil {
		log.Fatalln(err)
	}

	c.Log(output)
}
