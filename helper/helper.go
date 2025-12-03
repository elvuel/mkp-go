package helper

import (
	"errors"
	"strings"

	usbcom "github.com/elvuel/mkp-go"
)

// DeviceSN 指令 返回设备序列号
func DeviceSN(sfport *usbcom.SFSerialPort) (string, error) {
	if !sfport.SyncOuputEnabled {
		return "", errors.New("please enable sync mode first")
	}

	directive := "sn"

	result, err := sfport.SendDirective(directive)

	if err != nil {
		return "", err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		return parser.Parse(directive, result)
	}

	return result, usbcom.ErrDirectiveParserMissing
}

// ListDir 指令 返回路径下的所有子目录及文件
func ListDir(sfport *usbcom.SFSerialPort, path string) (*usbcom.FileSystem, error) {
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
			fssys := &usbcom.FileSystem{}
			err = parser.UnmarshalTo(parsedResult, fssys)
			if fssys.Error != "" {
				return nil, errors.New(fssys.Error)
			}
			return fssys, err
		}

	}

	return nil, usbcom.ErrDirectiveParserMissing
}

func ComposeLogDirctory(logDir string) string {
	if !strings.HasPrefix(logDir, "/eMMC/applog") {
		return "/eMMC/applog/" + logDir
	}

	return logDir
}

func CleanDir(sfport *usbcom.SFSerialPort, path string) error {
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
		// 	fssys := &usbcom.FileSystem{}
		// 	err = parser.UnmarshalTo(parsedResult, fssys)
		// 	if fssys.Error != "" {
		// 		return nil, errors.New(fssys.Error)
		// 	}
		// 	return fssys, err
		// }
	}

	return usbcom.ErrDirectiveParserMissing
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
func DeleteFile(sfport *usbcom.SFSerialPort, path string) error {
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
		// 	fssys := &usbcom.FileSystem{}
		// 	err = parser.UnmarshalTo(parsedResult, fssys)
		// 	if fssys.Error != "" {
		// 		return nil, errors.New(fssys.Error)
		// 	}
		// 	return fssys, err
		// }
	}

	return usbcom.ErrDirectiveParserMissing
}

// Alive 指令 心跳时间戳
func Alive(sfport *usbcom.SFSerialPort) (*usbcom.Heartbeat, error) {
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
			hb := &usbcom.Heartbeat{}
			err = parser.UnmarshalTo(parsedResult, hb)
			if err != nil {
				return nil, err
			}
			return hb, nil
		}

	}

	return nil, usbcom.ErrDirectiveParserMissing
}

// Atime 指令 返回 日志时长。 path可以是相对路径(.log扩展 - mkpdemo/1129f40), 也可以是绝对路径(/eMMC/applog/mkpdemo/1129f40.log)
func Atime(sfport *usbcom.SFSerialPort, path string) (*usbcom.LogLength, error) {
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
			o := &usbcom.LogLength{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, usbcom.ErrDirectiveParserMissing
}

// Aversion 指令 返回 版本信息。
func Aversion(sfport *usbcom.SFSerialPort) (*usbcom.MKPVersion, error) {
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
			o := &usbcom.MKPVersion{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, usbcom.ErrDirectiveParserMissing
}

// AInspect 指令 返回 日志基础信息。 path可以是相对路径(.log扩展 - mkpdemo/1129f40), 也可以是绝对路径(/eMMC/applog/mkpdemo/1129f40.log)
func AInspect(sfport *usbcom.SFSerialPort, path string) (*usbcom.LogInfo, error) {
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
			o := &usbcom.LogInfo{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, usbcom.ErrDirectiveParserMissing
}
