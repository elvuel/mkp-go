package mkpgo

import (
	"log"
	"strconv"
	"strings"

	"go.bug.st/serial"
)

type SFSerialPort = SerialPort

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

		SyncOuputEnabled: true,
		SyncOutputChan:   make(chan string),
	}
}

func (sp *SFSerialPort) SendDirective(directive string) (string, error) {
	sp.locker.Lock()
	defer sp.locker.Unlock()

	if sp.Verbose {
		log.Printf("准备执行指令: %s\n", directive)
	}
	if sp.SyncOuputEnabled {
		sp.SetSyncDirective(directive)
		_, err := sp.Write([]byte(directive + "\r\n"))
		if err != nil {
			log.Printf("执行指令: %s , 错误： %s\n", directive, err)
			return "", err
		}

		return sp.GetSyncOutput()
	}

	_, err := sp.Write([]byte(directive + "\r\n"))
	return "", err
}

func (sp *SFSerialPort) SendDirectiveAsync(directive string) error {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	if sp.Verbose {
		log.Printf("准备执行指令: %s\n", directive)
	}

	sp.withAsyncMark = true
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
	return <-sp.SyncOutputChan, nil
}

func (sp *SFSerialPort) StartRecording(args string) error {
	directives := make([]string, 0)
	directives = append(directives, "alog")

	directives = append(directives, args)

	return sp.SendDirectiveAsync(strings.Join(directives, " "))
}

func (sp *SFSerialPort) StartReplaying(logName string, delay int) error {
	directives := make([]string, 0)
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
	directives := make([]string, 0)
	directives = append(directives, "astop")
	return sp.SendDirectiveAsync(strings.Join(directives, " "))
}

func (sp *SFSerialPort) CancelReplay() error {
	directives := make([]string, 0)
	directives = append(directives, "acancel")
	return sp.SendDirectiveAsync(strings.Join(directives, " "))
}

func (sp *SFSerialPort) Mouse10(opt *M10Option) error {
	// --p: port #
	// --b: botton
	// --x: x
	// --y: y
	// --w: wheel
	// m10 --port 1 --b xx --x xx --y xx --w xx
	directives := make([]string, 0)
	directives = append(directives, "m10")
	directives = append(directives, "--port", sp.MousePortFlag)
	directives = append(directives, opt.ToString())
	_, err := sp.SendDirective(strings.Join(directives, " "))
	return err
}

func (sp *SFSerialPort) Keypad(opt *KpadOption) error {
	directives := make([]string, 0)
	directives = append(directives, "kpad")
	directives = append(directives, "--port", sp.KeyboardPortFlag)
	directives = append(directives, opt.ToString())
	_, err := sp.SendDirective(strings.Join(directives, " "))
	return err
}
