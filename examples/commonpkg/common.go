package commonpkg

import (
	"log"

	mkpgo "github.com/elvuel/mkp-go"
	"github.com/elvuel/mkp-go/helper"
)

var SFPort *mkpgo.SFSerialPort

func InitSFport() {
	SFPort = mkpgo.NewSFSerialPort()
	comName, _ := mkpgo.CheckSFSerialPort()
	if comName == "" {
		log.Fatalln("lookup mkp device failed")
	}
	SFPort.Name = comName

	err := SFPort.Open()
	if err != nil {
		// panic(err)
		log.Fatalln(err)
	}
	// defer sfport.Close()
	// sfport.Verbose = true
	go SFPort.Read()

	// sfport.StopRecording()
	helper.StopRecord(SFPort)

	SFPort.SyncOuputEnabled = true
}

func Log(i interface{}) {
	log.Printf("%#v\n", i)
}
