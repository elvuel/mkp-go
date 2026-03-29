package mkpgo

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

// SFSerialPort is an alias of SerialPort for SF device usage.
// SFSerialPort 是面向 SF 设备场景的 SerialPort 类型别名。
type SFSerialPort = SerialPort

// ErrSyncOutputTimeout indicates sync output wait timeout.
// ErrSyncOutputTimeout 表示等待同步输出超时。
var ErrSyncOutputTimeout = errors.New("timeout waiting for sync output")

// NewSFSerialPort creates a SerialPort with default SF-device settings.
// NewSFSerialPort 创建带默认 SF 设备参数的串口实例。
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

// SendDirective sends a directive and optionally waits for sync output.
// SendDirective 发送指令；在同步模式下等待并返回解析前原始输出。
func (sp *SFSerialPort) SendDirective(directive string) (string, error) {
	sp.locker.Lock()

	if sp.VerboseDirective {
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

// SendDirectiveAsync sends a directive without waiting for response.
// SendDirectiveAsync 异步发送指令，不阻塞等待输出。
func (sp *SFSerialPort) SendDirectiveAsync(directive string) error {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	if sp.VerboseDirective {
		log.Printf("preparing async directive: %s\n", directive)
	}

	sp.markAsync()
	_, err := sp.Write([]byte(directive + "\r\n"))
	return err
}

// GetRawDirective extracts directive verb from a full command.
// GetRawDirective 从完整命令中提取原始指令名。
func (sp *SFSerialPort) GetRawDirective(directive string) string {
	return strings.Split(directive, " ")[0]
}

// GetRawDirectiveOutputParser returns parser bound to directive verb.
// GetRawDirectiveOutputParser 返回与指令名对应的输出解析器。
func (sp *SFSerialPort) GetRawDirectiveOutputParser(directive string) RawDirectiveOutputParser {
	rawDirective := sp.GetRawDirective(directive)

	return RawDirectiveOutputParsers[rawDirective]
}

// GetSyncOutput waits for one sync output with configured timeout.
// GetSyncOutput 按配置超时等待一条同步输出结果。
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

// StartRecording starts device-side log recording.
// StartRecording 启动设备端录制（alog）。
func (sp *SFSerialPort) StartRecording(args string) error {
	if args == "" {
		return sp.SendDirectiveAsync("alog")
	}
	return sp.SendDirectiveAsync("alog " + args)
}

// StartReplaying starts replay from recorded log.
// StartReplaying 启动指定日志的回放任务。
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

// StopRecording stops current recording session.
// StopRecording 停止当前录制任务。
func (sp *SFSerialPort) StopRecording() error {
	return sp.SendDirectiveAsync("astop")
}

// Stop is an alias of StopRecording.
// Stop 是 StopRecording 的别名。
func (sp *SFSerialPort) Stop() error {
	return sp.StopRecording()
}

// CancelReplay cancels an ongoing replay task.
// CancelReplay 取消当前回放任务。
func (sp *SFSerialPort) CancelReplay() error {
	return sp.SendDirectiveAsync("acancel")
}

// Mouse10 sends m10 (mouse) directive.
// Mouse10 发送 m10 鼠标控制指令。
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
	return sp.SendDirectiveAsync(directive)
}

// MouseReleaseAll releases all mouse buttons.
// MouseReleaseAll 释放全部鼠标按键状态。
func (sp *SFSerialPort) MouseReleaseAll() error {
	btnReleaseOpt := &M10Option{}
	btnReleaseOpt.SetButton(0)
	return sp.Mouse10(btnReleaseOpt)
}

// Keypad sends kpad (keyboard) directive and commits local key cache on success.
// Keypad 发送 kpad 键盘指令，并在发送成功后提交本地按键缓存状态。
func (sp *SFSerialPort) Keypad(opt *KpadOption) error {
	directive := "kpad --port " + sp.KeyboardPortFlag
	if optStr := strings.TrimSpace(opt.ToString()); optStr != "" {
		directive += " " + optStr
	}
	if err := sp.SendDirectiveAsync(directive); err != nil {
		return err
	}
	opt.commitKpadState()
	return nil
}
