package helper

import (
	"errors"

	usbcom "github.com/elvuel/mkp-go"
)

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
