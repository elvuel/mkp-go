package main

import (
	"log"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	c "github.com/elvuel/mkp-go/examples/commonpkg"
)

func tap(keys ...string) error {
	return c.SFPort.Keypad(
		mkpgo.NewKpadOption().
			WithKeys(keys).
			WithAutoRelease().
			WithSyncIgnoreOutput(true),
	)
}

func main() {
	c.InitSFport()
	defer c.SFPort.Close()

	log.Println("3s 后开始输出文本: Hello MKP!")
	time.Sleep(3 * time.Second)

	sequence := [][]string{
		{"MOD_LSHIFT", "H"},
		{"E"},
		{"L"},
		{"L"},
		{"O"},
		{"SPACE"},
		{"MOD_LSHIFT", "M"},
		{"MOD_LSHIFT", "K"},
		{"MOD_LSHIFT", "P"},
		{"MOD_LSHIFT", "1"},
	}

	for _, keys := range sequence {
		if err := tap(keys...); err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("kpad 示例完成")
}
