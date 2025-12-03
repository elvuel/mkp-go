package helper

import (
	"errors"

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
