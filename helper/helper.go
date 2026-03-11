package helper

import (
	"errors"
	"strings"

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

	if !strings.HasPrefix(logPath, "/eMMC/applog") {
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
	opt := mkpgo.NewKpadOption().WithKeys([]string{key}).WithDelay(0).WithHold()
	return sfport.Keypad(opt)
}

// 释放
func KeyUp(sfport *mkpgo.SFSerialPort, key string) error {
	opt := mkpgo.NewKpadOption().WithKeys([]string{key}).WithDelay(0).WithAutoRelease()
	return sfport.Keypad(opt)
}

// 按下释放
func KeyTap(sfport *mkpgo.SFSerialPort, keys []string) error {
	opt := mkpgo.NewKpadOption().WithKeys(keys).WithDelay(0).WithAutoRelease()
	return sfport.Keypad(opt)
}

func KeyPress(sfport *mkpgo.SFSerialPort, key string, sleep int) error {
	return KeyPresses(sfport, []string{key}, sleep)
}

func KeyPresses(sfport *mkpgo.SFSerialPort, keys []string, sleep int) error {
	applyKeys := make([]string, 0)

	for _, key := range keys {
		switch key {
		case "!":
			applyKeys = append(applyKeys, "mod_lshift", "1")
		case "@":
			applyKeys = append(applyKeys, "mod_lshift", "2")
		case "#":
			applyKeys = append(applyKeys, "mod_lshift", "3")
		case "$":
			applyKeys = append(applyKeys, "mod_lshift", "4")
		case "%":
			applyKeys = append(applyKeys, "mod_lshift", "5")
		case "^":
			applyKeys = append(applyKeys, "mod_lshift", "6")
		case "&":
			applyKeys = append(applyKeys, "mod_lshift", "7")
		case "*":
			applyKeys = append(applyKeys, "mod_lshift", "8")
		case "(":
			applyKeys = append(applyKeys, "mod_lshift", "9")
		case ")":
			applyKeys = append(applyKeys, "mod_lshift", "0")
		case " ":
			applyKeys = append(applyKeys, "space")
		case "-":
			applyKeys = append(applyKeys, "minus")
		case "_":
			applyKeys = append(applyKeys, "mod_lshift", "minus")
		case "=":
			applyKeys = append(applyKeys, "equal")
		case "+":
			applyKeys = append(applyKeys, "mod_lshift", "equal")
		case "[":
			applyKeys = append(applyKeys, "leftbracket")
		case "{":
			applyKeys = append(applyKeys, "mod_lshift", "leftbracket")
		case "]":
			applyKeys = append(applyKeys, "rightbracket")
		case "}":
			applyKeys = append(applyKeys, "mod_lshift", "rightbracket")
		case "\\":
			applyKeys = append(applyKeys, "backslash")
		case "|":
			applyKeys = append(applyKeys, "mod_lshift", "backslash")
		// case "~":
		// 	applyKeys = append(applyKeys, "mod_lshift", "hashtilde")
		case ";":
			applyKeys = append(applyKeys, "semicolon")
		case ":":
			applyKeys = append(applyKeys, "mod_lshift", "semicolon")
		case "'":
			applyKeys = append(applyKeys, "apostrophe")
		case "\"":
			applyKeys = append(applyKeys, "mod_lshift", "apostrophe")
		case "`":
			applyKeys = append(applyKeys, "grave")
		case "~":
			applyKeys = append(applyKeys, "mod_lshift", "grave")
		case ",":
			applyKeys = append(applyKeys, "comma")
		case "<":
			applyKeys = append(applyKeys, "mod_lshift", "comma")
		case ".":
			applyKeys = append(applyKeys, "dot")
		case ">":
			applyKeys = append(applyKeys, "mod_lshift", "dot")
		case "/":
			applyKeys = append(applyKeys, "slash")
		case "?":
			applyKeys = append(applyKeys, "mod_lshift", "slash")
		default:
			applyKeys = append(applyKeys, key)
		}
	}

	opt := mkpgo.NewKpadOption().WithKeys(applyKeys).WithDelay(0)
	if sleep > 0 {
		opt.WithRelease(sleep)
	}
	err := sfport.Keypad(opt)
	if err != nil {
		return err
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
