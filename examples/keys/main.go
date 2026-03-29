package main

import (
	"log"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	"github.com/elvuel/mkp-go/helper"
)

var sfport *mkpgo.SFSerialPort

func initSFport() {
	sfport = mkpgo.NewSFSerialPort()
	comName, _ := mkpgo.CheckSFSerialPort()
	if comName == "" {
		log.Fatalln("lookup mkp device failed")
	}
	sfport.Name = comName

	err := sfport.Open()
	if err != nil {
		// panic(err)
		log.Fatalln(err)
	}
	// defer sfport.Close()
	// sfport.Verbose = true
	go sfport.Read()

	// sfport.StopRecording()
	helper.StopRecord(sfport)

	sfport.SyncOuputEnabled = true
}

func terminateSFPort() {
	if sfport != nil {
		sfport.Close()
	}
}

func main() {
	initSFport()
	defer terminateSFPort()

	log.Println("sleep 5s")
	time.Sleep(5 * time.Second)
	log.Println("--start--")

	// DemoKpadOptionKeyDownUpV2()
	// DemoOutputExclamation(" ")

	sfport.VerboseDirective = true

	// 请根据以下调用顺序的期望输出对keyDown进行调整
	// 按下MOD_ALT: keydown("w")
	// kpad --port 2 --s 0x04 --rel 0 --d 0
	// 紧接着按下q(MOD_ALT未释放):keydown("q")
	// kpad --port 2 --s 0x04 --x1 0x14 --rel 0 --d 0

	helper.KeyDown(sfport, "w")
	helper.KeyDown(sfport, "d")
	time.Sleep(3000 * time.Millisecond)
	helper.KeyUp(sfport, "d")
	helper.KeyUp(sfport, "w")

	helper.KeyDown(sfport, "alt")
	helper.KeyDown(sfport, "n")
	time.Sleep(200 * time.Millisecond)
	helper.KeyUp(sfport, "n")
	helper.KeyUp(sfport, "alt")
	time.Sleep(3 * time.Second)

	helper.KeyTap(sfport, "esc")
	time.Sleep(3 * time.Second)

	// KpadOption 调整 KeyDown/KeyUp 如果 key是mod keys, 则设定ModKeys。（modkeys需要考虑 其它全局缓存中的mod keys)

	// 请根据以下调用顺序的期望输出对keyDown进行调整
	// 按下MOD_ALT: keydown("MOD_LALT")
	// kpad --port 2 --s 0x04 --rel 0 --d 0
	// 紧接着按下q(MOD_ALT未释放):keydown("q")
	// kpad --port 2 --s 0x04 --x1 0x14 --rel 0 --d 0

	helper.KeyDown(sfport, "MOD_LALT")
	helper.KeyDown(sfport, "q")
	time.Sleep(600 * time.Millisecond)
	helper.KeyUp(sfport, "q")
	helper.KeyUp(sfport, "MOD_LALT")

	time.Sleep(3 * time.Second)

	helper.KeyDown(sfport, "space")
	time.Sleep(3 * time.Second)
	helper.KeyUp(sfport, "space")

	time.Sleep(500 * time.Millisecond)

	helper.KeyDown(sfport, "MOD_LALT")
	helper.KeyDown(sfport, "q")
	time.Sleep(200 * time.Millisecond)
	helper.KeyUp(sfport, "q")
	helper.KeyUp(sfport, "MOD_LALT")

	// helper.KeyTap(sfport, "!")
	// helper.KeyTap(sfport, "enter")
	// helper.KeyTap(sfport, "y")
	// helper.KeyTap(sfport, "!")

	// helper.KeyPresses(sfport, []string{"!", "enter", "@", " ", "y", "!"}, 50)

}

func DemoKpadOptionKeyDownUpV2() error {
	opt := mkpgo.NewKpadOption().WithDelay(0)

	if err := sfport.Keypad(opt.KeyDown("w")); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

	if err := sfport.Keypad(opt.KeyDown("space")); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	releaseOpt, remainHoldOpt := opt.KeyUp("space")
	if releaseOpt != nil {
		if err := sfport.Keypad(releaseOpt); err != nil {
			return err
		}
	}
	if remainHoldOpt != nil {
		if err := sfport.Keypad(remainHoldOpt); err != nil {
			return err
		}
	}

	if err := sfport.Keypad(opt.KeyDown("lshift")); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	releaseOpt, remainHoldOpt = opt.KeyUp("lshift")
	if releaseOpt != nil {
		if err := sfport.Keypad(releaseOpt); err != nil {
			return err
		}
	}
	if remainHoldOpt != nil {
		if err := sfport.Keypad(remainHoldOpt); err != nil {
			return err
		}
	}

	time.Sleep(5 * time.Second)

	releaseOpt, remainHoldOpt = opt.KeyUp("w")
	if releaseOpt != nil {
		if err := sfport.Keypad(releaseOpt); err != nil {
			return err
		}
	}
	if remainHoldOpt != nil {
		if err := sfport.Keypad(remainHoldOpt); err != nil {
			return err
		}
	}

	return nil
}

func DemoKpadOptionKeyDownUpV2WithHelper() error {
	if err := helper.KeyDown(sfport, "w"); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

	if err := helper.KeyDown(sfport, "space"); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	if err := helper.KeyUp(sfport, "space"); err != nil {
		return err
	}

	if err := helper.KeyDown(sfport, "lshift"); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	if err := helper.KeyUp(sfport, "lshift"); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	if err := helper.KeyUp(sfport, "w"); err != nil {
		return err
	}

	return nil
}

func DemoOutputExclamation(key string) error {
	opt := mkpgo.NewKpadOption().WithDelay(0)

	if err := sfport.Keypad(opt.KeyDown(key)); err != nil {
		return err
	}

	releaseOpt, remainHoldOpt := opt.KeyUp(key)
	if releaseOpt != nil {
		if err := sfport.Keypad(releaseOpt); err != nil {
			return err
		}
	}
	if remainHoldOpt != nil {
		if err := sfport.Keypad(remainHoldOpt); err != nil {
			return err
		}
	}

	return nil
}
