package helper

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
)

func StopRecord(sfport *mkpgo.SFSerialPort) error {
	return sfport.StopRecording()
}

func StartRecord(sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption) error {
	args := make([]string, 0)
	args = append(args, logName)

	if opt != nil {
		args = append(args, opt.CliArgs()...)
	}
	return sfport.StartRecording(strings.Join(args, " "))
}

func Alog(sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption) (string, error) {
	if !sfport.SyncOuputEnabled {
		return "", errors.New("please enable sync mode first")
	}

	args := make([]string, 0)
	args = append(args, logName)

	if opt != nil {
		args = append(args, opt.CliArgs()...)
	}

	directive := "alog " + strings.Join(args, " ")
	fmt.Println(directive)

	result, err := sfport.SendDirective(directive)

	// log.Println("got ################ alog response:", result)

	if err != nil {
		return "", err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return "", err
		}

		return parsedResult, nil

	}

	return "", mkpgo.ErrDirectiveParserMissing
}

func Astop(sfport *mkpgo.SFSerialPort) error {
	if !sfport.SyncOuputEnabled {
		return errors.New("please enable sync mode first")
	}
	directive := "astop"

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		_, err := parser.Parse(directive, result)
		return err
	}

	return mkpgo.ErrDirectiveParserMissing
}

func Cancel(sfport *mkpgo.SFSerialPort) error {
	return sfport.CancelReplay()
}

// DeviceSN 指令 返回设备序列号
func DeviceSN(sfport *mkpgo.SFSerialPort) (*mkpgo.SN, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "sn"

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			sn := &mkpgo.SN{}
			err = parser.UnmarshalTo(parsedResult, sn)
			return sn, err
		}
	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// ListDir 指令 返回路径下的所有子目录及文件
func ListDir(sfport *mkpgo.SFSerialPort, path string) (*mkpgo.FileSystem, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "list_dir " + path

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			fssys := &mkpgo.FileSystem{}
			err = parser.UnmarshalTo(parsedResult, fssys)
			if fssys.Error != "" {
				return nil, errors.New(fssys.Error)
			}
			return fssys, err
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

func ComposeLogDirctory(logDir string) string {
	if !strings.HasPrefix(logDir, "/eMMC/applog") {
		return "/eMMC/applog/" + logDir
	}

	return logDir
}

func CleanDir(sfport *mkpgo.SFSerialPort, path string) error {
	if !strings.HasPrefix(path, "/eMMC/applog") {
		return errors.New("only can clean directory in working directory") // only can delete file within /eMMC/applog
	}

	if !sfport.SyncOuputEnabled {
		return errors.New("please enable sync mode first")
	}

	directive := "clean_dir " + path

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		_, err := parser.Parse(directive, result)

		if err != nil {
			return err
		}

		return nil

		// if parser.IsJSONOutput() {
		// 	fssys := &mkpgo.FileSystem{}
		// 	err = parser.UnmarshalTo(parsedResult, fssys)
		// 	if fssys.Error != "" {
		// 		return nil, errors.New(fssys.Error)
		// 	}
		// 	return fssys, err
		// }
	}

	return mkpgo.ErrDirectiveParserMissing
}

func ComposeLogFullpath(logPath string) string {
	if !strings.HasSuffix(logPath, ".log") {
		logPath += ".log"
	}

	if !strings.HasPrefix(logPath, "/eMMC/applog/") {
		return "/eMMC/applog/" + logPath
	}

	return logPath
}

// DeleteFile 指令 只能删除在/eMMC/applog下的文件(path 路径)
func DeleteFile(sfport *mkpgo.SFSerialPort, path string) error {
	path = ComposeLogFullpath(path)

	if !strings.HasPrefix(path, "/eMMC/applog") {
		return errors.New("only can delete file in working directory") // only can delete file within /eMMC/applog
	}

	if !sfport.SyncOuputEnabled {
		return errors.New("please enable sync mode first")
	}

	directive := "delete_file " + path

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		_, err := parser.Parse(directive, result)

		if err != nil {
			return err
		}

		return nil

		// if parser.IsJSONOutput() {
		// 	fssys := &mkpgo.FileSystem{}
		// 	err = parser.UnmarshalTo(parsedResult, fssys)
		// 	if fssys.Error != "" {
		// 		return nil, errors.New(fssys.Error)
		// 	}
		// 	return fssys, err
		// }
	}

	return mkpgo.ErrDirectiveParserMissing
}

// Alive 指令 心跳时间戳
func Alive(sfport *mkpgo.SFSerialPort) (*mkpgo.Heartbeat, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "alive"

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			hb := &mkpgo.Heartbeat{}
			err = parser.UnmarshalTo(parsedResult, hb)
			if err != nil {
				return nil, err
			}
			return hb, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// Atime 指令 返回 日志时长。 path可以是相对路径(.log扩展 - mkpdemo/1129f40), 也可以是绝对路径(/eMMC/applog/mkpdemo/1129f40.log)
func Atime(sfport *mkpgo.SFSerialPort, path string) (*mkpgo.LogLength, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "atime " + path

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			o := &mkpgo.LogLength{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// Aversion 指令 返回 版本信息。
func Aversion(sfport *mkpgo.SFSerialPort) (*mkpgo.MKPVersion, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "aversion"

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			o := &mkpgo.MKPVersion{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// AInspect 指令 返回 日志基础信息。 path可以是相对路径(.log扩展 - mkpdemo/1129f40), 也可以是绝对路径(/eMMC/applog/mkpdemo/1129f40.log)
func AInspect(sfport *mkpgo.SFSerialPort, path string) (*mkpgo.LogInfo, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "ainsp " + path

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			o := &mkpgo.LogInfo{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

func KeyDown(sfport *mkpgo.SFSerialPort, key string) error {
	return sendKeyDown(sfport, mkpgo.NewKpadOption().WithDelay(0), key)
}

// 释放
func KeyUp(sfport *mkpgo.SFSerialPort, key string) error {
	return sendKeyUp(sfport, mkpgo.NewKpadOption().WithDelay(0), key)
}

// 按下释放
func KeyTap(sfport *mkpgo.SFSerialPort, key string) error {
	sleep := rand.Intn(100) + 20
	if err := KeyDown(sfport, key); err != nil {
		return err
	}

	time.Sleep(time.Duration(sleep) * time.Millisecond)

	if err := KeyUp(sfport, key); err != nil {
		return err
	}

	return nil
}

func KeyPresses(sfport *mkpgo.SFSerialPort, keys []string, sleep int) error {
	for _, key := range keys {
		if err := KeyTap(sfport, key); err != nil {
			return err
		}
	}
	return nil
}

func sendKeyDown(sfport *mkpgo.SFSerialPort, opt *mkpgo.KpadOption, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}

	downOpt := opt.KeyDown(key)
	return sfport.Keypad(downOpt)
}

func sendKeyUp(sfport *mkpgo.SFSerialPort, opt *mkpgo.KpadOption, key string) error {
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

func KeypadRelease(sfport *mkpgo.SFSerialPort) error {
	return sfport.Keypad(mkpgo.HidKpadRelease)
}

func KeypadReleaseAll(sfport *mkpgo.SFSerialPort) error {
	return sfport.Keypad(mkpgo.HidKpadReleaseAll)
}

func MouseReleaseAll(sfport *mkpgo.SFSerialPort) error {
	return sfport.MouseReleaseAll()
}
