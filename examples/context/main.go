package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	"github.com/elvuel/mkp-go/helper"
)

func main() {
	sfport := mkpgo.NewSFSerialPort()
	sfport.Name = "COM5" // 修改为你的设备端口

	if err := sfport.Open(); err != nil {
		panic(err)
	}
	defer sfport.Close()

	// alog 需要同步模式并且 Read 循环在运行。
	sfport.SyncOuputEnabled = true
	go sfport.Read()

	logName := "mkpdemo/demo"

	fmt.Println("=== case 1: timeout cancel ===")
	runTimeoutCancel(sfport, logName)

	fmt.Println("=== case 2: manual cancel ===")
	runManualCancel(sfport, logName)
}

func runTimeoutCancel(sfport *mkpgo.SFSerialPort, logName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := helper.AlogContext(ctx, sfport, logName, nil)
	fmt.Printf("result=%q err=%v\n", result, err)
	fmt.Printf("is deadline exceeded: %v\n", errors.Is(err, context.DeadlineExceeded))
}

func runManualCancel(sfport *mkpgo.SFSerialPort, logName string) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	result, err := helper.AlogContext(ctx, sfport, logName, nil)
	fmt.Printf("result=%q err=%v\n", result, err)
	fmt.Printf("is canceled: %v\n", errors.Is(err, context.Canceled))
}

