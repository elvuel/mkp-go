package mkpgo

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

type SFSerialPort = SerialPort

var ErrSyncOutputTimeout = errors.New("timeout waiting for sync output")

func NewSFSerialPort() *SFSerialPort {
	return &SFSerialPort{
		Name:             "COM3",
		VID:              "1A86",
		PID:              "55D3",
		SerialNum:        "5ABA089859", // "5A47027726"
		Product:          "USB-Enhanced-SERIAL CH343 (COM3)",
		MousePortFlag:    "1",
		KeyboardPortFlag: "2",
		IsUSB:            true,
		readerBuf:        make([]byte, 128),
		OpenMode:         &serial.Mode{BaudRate: 115200, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit},

		SyncOuputEnabled:  true,
		SyncOutputChan:    make(chan string, 1),
		SyncOutputTimeout: 10 * time.Second,
	}
}

func (sp *SFSerialPort) SendDirective(directive string) (string, error) {
	sp.locker.Lock()

	if sp.Verbose {
		log.Printf("preparing directive: %s\n", directive)
	}
	if sp.IsSyncOutputEnabled() {
		sp.clearAsyncMark()
		sp.EmptySyncDirective()
		if strings.HasPrefix(directive, "alog") { // as alog output does not start with full cli text.
			sp.SetSyncDirective("alog")
		} else {
			sp.SetSyncDirective(directive)
		}
		_, err := sp.Write([]byte(directive + "\r\n"))
		if err != nil {
			sp.EmptySyncDirective()
			sp.locker.Unlock()
			log.Printf("failed to execute directive %s: %v\n", directive, err)
			return "", err
		}

		sp.locker.Unlock()
		return sp.GetSyncOutput()
	}

	_, err := sp.Write([]byte(directive + "\r\n"))
	sp.locker.Unlock()
	return "", err
}

func (sp *SFSerialPort) SendDirectiveAsync(directive string) error {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	if sp.Verbose {
		log.Printf("preparing async directive: %s\n", directive)
	}

	sp.markAsync()
	_, err := sp.Write([]byte(directive + "\r\n"))
	return err
}

func (sp *SFSerialPort) GetRawDirective(directive string) string {
	return strings.Split(directive, " ")[0]
}

func (sp *SFSerialPort) GetRawDirectiveOutputParser(directive string) RawDirectiveOutputParser {
	rawDirective := sp.GetRawDirective(directive)

	return RawDirectiveOutputParsers[rawDirective]
}

func (sp *SFSerialPort) GetSyncOutput() (string, error) {
	if sp.SyncOutputTimeout > 0 {
		select {
		case output := <-sp.SyncOutputChan:
			return output, nil
		case <-time.After(sp.SyncOutputTimeout):
			sp.EmptySyncDirective()
			return "", ErrSyncOutputTimeout
		}
	}

	return <-sp.SyncOutputChan, nil
}

func (sp *SFSerialPort) StartRecording(args string) error {
	if args == "" {
		return sp.SendDirectiveAsync("alog")
	}
	return sp.SendDirectiveAsync("alog " + args)
}

func (sp *SFSerialPort) StartReplaying(logName string, delay int) error {
	directives := make([]string, 0, 4)
	directives = append(directives, "aplay")

	if logName != "" {
		directives = append(directives, logName)
	}

	if delay >= 0 {
		directives = append(directives, "--delay", strconv.Itoa(delay))
	}

	return sp.SendDirectiveAsync(strings.Join(directives, " "))
}

func (sp *SFSerialPort) StopRecording() error {
	return sp.SendDirectiveAsync("astop")
}

func (sp *SFSerialPort) Stop() error {
	return sp.StopRecording()
}

func (sp *SFSerialPort) CancelReplay() error {
	return sp.SendDirectiveAsync("acancel")
}

func (sp *SFSerialPort) Mouse10(opt *M10Option) error {
	// --p: port #
	// --b: botton
	// --x: x
	// --y: y
	// --w: wheel
	// m10 --port 1 --b xx --x xx --y xx --w xx
	directive := "m10 --port " + sp.MousePortFlag
	if optStr := strings.TrimSpace(opt.ToString()); optStr != "" {
		directive += " " + optStr
	}
	if sp.Verbose {
		log.Printf("preparing directive: %s\n", directive)
	}
	return sp.SendDirectiveAsync(directive)
}

func (sp *SFSerialPort) MouseReleaseAll() error {
	btnReleaseOpt := &M10Option{}
	btnReleaseOpt.SetButton(0)
	return sp.Mouse10(btnReleaseOpt)
}

func (sp *SFSerialPort) Keypad(opt *KpadOption) error {
	directive := "kpad --port " + sp.KeyboardPortFlag
	if optStr := strings.TrimSpace(opt.ToString()); optStr != "" {
		directive += " " + optStr
	}
	if sp.Verbose {
		log.Printf("preparing directive: %s\n", directive)
	}
	return sp.SendDirectiveAsync(directive)
}
