package mkpgo

import (
	"context"
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
	return sp.SendDirectiveContext(context.Background(), directive)
}

// SendDirectiveContext sends a directive and optionally waits for sync output.
// SendDirectiveContext 发送指令；支持通过 context 取消等待过程。
func (sp *SFSerialPort) SendDirectiveContext(ctx context.Context, directive string) (string, error) {
	return sp.sendDirectiveContext(ctx, directive, false)
}

// SendDirectiveIgnoreOutput sends a directive synchronously and discards its output.
// SendDirectiveIgnoreOutput 同步发送指令，但 Read 只等待完成标记并忽略输出内容。
func (sp *SFSerialPort) SendDirectiveIgnoreOutput(directive string) error {
	return sp.SendDirectiveIgnoreOutputContext(context.Background(), directive)
}

// SendDirectiveIgnoreOutputContext sends a directive synchronously and discards its output.
// SendDirectiveIgnoreOutputContext 同步发送指令并忽略输出，支持通过 context 取消等待过程。
func (sp *SFSerialPort) SendDirectiveIgnoreOutputContext(ctx context.Context, directive string) error {
	_, err := sp.sendDirectiveContext(ctx, directive, true)
	return err
}

func (sp *SFSerialPort) sendDirectiveContext(ctx context.Context, directive string, ignoreOutput bool) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	sp.locker.Lock()

	if sp.VerboseDirective {
		if ignoreOutput {
			log.Printf("preparing directive with ignored output: %s\n", directive)
		} else {
			log.Printf("preparing directive: %s\n", directive)
		}
	}
	if sp.IsSyncOutputEnabled() {
		sp.clearAsyncMark()
		sp.EmptySyncDirective()
		syncDirective := normalizeSFSyncDirective(directive)
		if ignoreOutput {
			sp.SetSyncDirectiveIgnoreOutput(syncDirective)
		} else {
			sp.SetSyncDirective(syncDirective)
		}
		_, err := sp.Write([]byte(directive + "\r\n"))
		if err != nil {
			sp.EmptySyncDirective()
			sp.locker.Unlock()
			log.Printf("failed to execute directive %s: %v\n", directive, err)
			return "", err
		}

		sp.locker.Unlock()
		return sp.GetSyncOutputContext(ctx)
	}

	_, err := sp.Write([]byte(directive + "\r\n"))
	sp.locker.Unlock()
	return "", err
}

func normalizeSFSyncDirective(directive string) string {
	if strings.HasPrefix(directive, "alog") { // as alog output does not start with full cli text.
		return "alog"
	}
	return directive
}

// SendDirectiveAsync sends a directive without waiting for response.
// SendDirectiveAsync 异步发送指令，不阻塞等待输出。
func (sp *SFSerialPort) SendDirectiveAsync(directive string) error {
	return sp.SendDirectiveAsyncContext(context.Background(), directive)
}

// SendDirectiveAsyncContext sends a directive without waiting for response.
// SendDirectiveAsyncContext 异步发送指令；在发送前可由 context 取消。
func (sp *SFSerialPort) SendDirectiveAsyncContext(ctx context.Context, directive string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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
	return sp.GetSyncOutputContext(context.Background())
}

// GetSyncOutputContext waits for one sync output with timeout and context cancellation.
// GetSyncOutputContext 按配置等待同步输出，支持 context 取消。
func (sp *SFSerialPort) GetSyncOutputContext(ctx context.Context) (string, error) {
	if sp.SyncOutputTimeout > 0 {
		timer := time.NewTimer(sp.SyncOutputTimeout)
		defer timer.Stop()

		select {
		case output := <-sp.SyncOutputChan:
			return output, nil
		case <-ctx.Done():
			sp.EmptySyncDirective()
			return "", ctx.Err()
		case <-timer.C:
			sp.EmptySyncDirective()
			return "", ErrSyncOutputTimeout
		}
	}

	select {
	case output := <-sp.SyncOutputChan:
		return output, nil
	case <-ctx.Done():
		sp.EmptySyncDirective()
		return "", ctx.Err()
	}
}

// StartRecording starts device-side log recording.
// StartRecording 启动设备端录制（alog）。
func (sp *SFSerialPort) StartRecording(args string) error {
	return sp.StartRecordingContext(context.Background(), args)
}

// StartRecordingContext starts device-side log recording.
// StartRecordingContext 启动设备端录制（alog），支持 context 取消。
func (sp *SFSerialPort) StartRecordingContext(ctx context.Context, args string) error {
	if args == "" {
		return sp.SendDirectiveAsyncContext(ctx, "alog")
	}
	return sp.SendDirectiveAsyncContext(ctx, "alog "+args)
}

// StartReplaying starts replay from recorded log.
// StartReplaying 启动指定日志的回放任务。
func (sp *SFSerialPort) StartReplaying(logName string, delay int) error {
	return sp.StartReplayingContext(context.Background(), logName, delay)
}

// StartReplayingContext starts replay from recorded log.
// StartReplayingContext 启动指定日志回放任务，支持 context 取消。
func (sp *SFSerialPort) StartReplayingContext(ctx context.Context, logName string, delay int) error {
	directives := make([]string, 0, 4)
	directives = append(directives, "aplay")

	if logName != "" {
		directives = append(directives, logName)
	}

	if delay >= 0 {
		directives = append(directives, "--delay", strconv.Itoa(delay))
	}

	return sp.SendDirectiveAsyncContext(ctx, strings.Join(directives, " "))
}

// StopRecording stops current recording session.
// StopRecording 停止当前录制任务。
func (sp *SFSerialPort) StopRecording() error {
	return sp.StopRecordingContext(context.Background())
}

// StopRecordingContext stops current recording session.
// StopRecordingContext 停止当前录制任务，支持 context 取消。
func (sp *SFSerialPort) StopRecordingContext(ctx context.Context) error {
	return sp.SendDirectiveAsyncContext(ctx, "astop")
}

// Stop is an alias of StopRecording.
// Stop 是 StopRecording 的别名。
func (sp *SFSerialPort) Stop() error {
	return sp.StopContext(context.Background())
}

// StopContext is an alias of StopRecordingContext.
// StopContext 是 StopRecordingContext 的别名。
func (sp *SFSerialPort) StopContext(ctx context.Context) error {
	return sp.StopRecordingContext(ctx)
}

// CancelReplay cancels an ongoing replay task.
// CancelReplay 取消当前回放任务。
func (sp *SFSerialPort) CancelReplay() error {
	return sp.CancelReplayContext(context.Background())
}

// CancelReplayContext cancels an ongoing replay task.
// CancelReplayContext 取消当前回放任务，支持 context 取消。
func (sp *SFSerialPort) CancelReplayContext(ctx context.Context) error {
	return sp.SendDirectiveAsyncContext(ctx, "acancel")
}

// Mouse10 sends m10 (mouse) directive.
// Mouse10 发送 m10 鼠标控制指令。
func (sp *SFSerialPort) Mouse10(opt *M10Option) error {
	return sp.Mouse10Context(context.Background(), opt)
}

// Mouse10Context sends m10 (mouse) directive.
// Mouse10Context 发送 m10 鼠标控制指令，支持 context 取消。
func (sp *SFSerialPort) Mouse10Context(ctx context.Context, opt *M10Option) error {
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
	if opt.IsAsync() {
		return sp.SendDirectiveAsyncContext(ctx, directive)
	}
	return sp.SendDirectiveIgnoreOutputContext(ctx, directive)
}

// MouseReleaseAll releases all mouse buttons.
// MouseReleaseAll 释放全部鼠标按键状态。
func (sp *SFSerialPort) MouseReleaseAll(opts ...*M10Option) error {
	return sp.MouseReleaseAllContext(context.Background(), opts...)
}

// MouseReleaseAllContext releases all mouse buttons.
// MouseReleaseAllContext 释放全部鼠标按键状态，支持 context 取消。
func (sp *SFSerialPort) MouseReleaseAllContext(ctx context.Context, opts ...*M10Option) error {
	btnReleaseOpt := NewM10Option().SetButton(0)
	if len(opts) > 0 && opts[0] != nil {
		btnReleaseOpt.Async = opts[0].Async
		btnReleaseOpt.SyncIgnoreOutput = opts[0].SyncIgnoreOutput
	}
	return sp.Mouse10Context(ctx, btnReleaseOpt)
}

// Keypad sends kpad (keyboard) directive and commits local key cache on success.
// Keypad 发送 kpad 键盘指令，并在发送成功后提交本地按键缓存状态。
func (sp *SFSerialPort) Keypad(opt *KpadOption) error {
	return sp.KeypadContext(context.Background(), opt)
}

// KeypadContext sends kpad directive and commits local key cache on success.
// KeypadContext 发送 kpad 指令，并在发送成功后提交本地按键缓存状态。
func (sp *SFSerialPort) KeypadContext(ctx context.Context, opt *KpadOption) error {
	directive := "kpad --port " + sp.KeyboardPortFlag
	if optStr := strings.TrimSpace(opt.ToString()); optStr != "" {
		directive += " " + optStr
	}
	var err error
	if opt.IsAsync() {
		err = sp.SendDirectiveAsyncContext(ctx, directive)
	} else {
		err = sp.SendDirectiveIgnoreOutputContext(ctx, directive)
	}
	if err != nil {
		return err
	}
	opt.commitKpadState()
	return nil
}
