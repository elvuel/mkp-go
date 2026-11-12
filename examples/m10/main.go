package main

import (
	"log"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	c "github.com/elvuel/mkp-go/examples/commonpkg"
)

func main() {
	c.InitSFport()
	defer c.SFPort.Close()

	log.Println("3s 后开始移动鼠标：先向左 400，再向上 400")
	time.Sleep(3 * time.Second)

	if err := c.SFPort.Mouse10(mkpgo.NewM10Option().SetX(-400).WithSyncIgnoreOutput(true)); err != nil {
		log.Fatalln(err)
	}
	if err := c.SFPort.Mouse10(mkpgo.NewM10Option().SetY(-400).WithSyncIgnoreOutput(true)); err != nil {
		log.Fatalln(err)
	}

	log.Println("m10 示例完成")
}
