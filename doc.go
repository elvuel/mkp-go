package mkpgo

/*

Example:

```go
package main

import (
	"fmt"
	"log"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	"github.com/elvuel/mkp-go/helper"
)

var sfport *mkpgo.SFSerialPort
var HidKpadRelease = mkpgo.NewKpadOption().WithDelay(0).WithKey("NONE")
var HidKpadReleaseAll = mkpgo.NewKpadOption().WithDelay(0).WithRelease(0)

func main() {
	sfport = mkpgo.NewSFSerialPort()
	sfport.Name = "COM5"

	err := sfport.Open()
	if err != nil {
		panic(err)
	}
	defer sfport.Close()
	// sfport.Verbose = true
	go sfport.Read()

	// sfport.StopRecording()
	helper.StopRecord(sfport)

	sfport.SyncOuputEnabled = true
	// fs, err := helper.ListDir(sfport, "/eMMC/applog/mkpdemo")
	// fmt.Println(fs, err)

	// hb, err := helper.Alive(sfport)
	// fmt.Println(hb, err)

	// o, err := helper.Atime(sfport, "/eMMC/applog/mkpdemo/1129f40.log")
	// fmt.Println(o, err)

	// o, err := helper.Aversion(sfport)
	// fmt.Println(o, err)

	// err = helper.DeleteFile(sfport, "/eMMC/applog/mpkdemo/draw3.log")
	// fmt.Println(err)

	// err = helper.CleanDir(sfport, "/eMMC/applog/r1203")
	// fmt.Println(err)

	// o, err := helper.AInspect(sfport, "/eMMC/applog/mkpdemo/1129f40.log")
	// o, err := helper.AInspect(sfport, "guijidashi/1210_102844")
	// fmt.Println(o, err)

	// sfport.Verbose = true
	// result, err := helper.Alog(sfport, "guijidashi/r111", nil)
	// log.Println("10s后开始。。。")
	// time.Sleep(10 * time.Second)
	// fmt.Println(result)

	// err = helper.StopRecord(sfport)
	// fmt.Println(err)

	// o, err := helper.AInspect(sfport, "guijidashi/r111")
	// fmt.Println(o, err)

	// for i := 0; i < 10; i++ {
	// 	KeyTap([]string{"w"})
	// 	time.Sleep(1000 * time.Millisecond)
	// }

	// 鼠标右键按下 释放
	m10Opt := mkpgo.NewM10Option().SetButton(1).SetX(0).SetY(0)
	log.Println("----")
	time.Sleep(5 * time.Second)
	sfport.Mouse10(m10Opt)
	log.Println("----")
	m10Opt.SetX(100).SetY(100).WithoutButton()
	sfport.Mouse10(m10Opt)
	// time.Sleep(5 * time.Second)
	log.Println("----")
	m10Opt.SetX(0).SetY(0).SetButton(0)
	sfport.Mouse10(m10Opt)
}

func KeyDown(key string) error {
	opt := mkpgo.NewKpadOption().WithKeys([]string{key}).WithDelay(0).WithHold()
	return sfport.Keypad(opt)
}

// 释放
func KeyUp(key string) error {
	opt := mkpgo.NewKpadOption().WithKeys([]string{key}).WithDelay(0).WithAutoRelease()
	return sfport.Keypad(opt)
}

// 按下释放
func KeyTap(keys []string) error {
	opt := mkpgo.NewKpadOption().WithKeys(keys).WithDelay(0).WithAutoRelease()
	return sfport.Keypad(opt)
}

func KeyPress(key string, sleep int) error {
	return KeyPresses([]string{key}, sleep)
}

func KeyPresses(keys []string, sleep int) error {
	opt := mkpgo.NewKpadOption().WithKeys(keys).WithDelay(0)
	if sleep > 0 {
		opt.WithRelease(sleep)
	}
	err := sfport.Keypad(opt)
	if err != nil {
		return err
	}
	fmt.Println(time.Now().Unix())

	if sleep > 0 {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
	fmt.Println(time.Now().Unix())

	return nil
}

func KeypadRelease() error {
	return sfport.Keypad(HidKpadRelease)
}

func KeypadReleaseAll() error {
	return sfport.Keypad(HidKpadReleaseAll)
}
```
*/
