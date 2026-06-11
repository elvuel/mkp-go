package mkpgo

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

// SerialPort is the core serial transport model used by the package.
// SerialPort 是本包使用的核心串口传输模型。
type SerialPort struct {
	// Device identity and USB metadata.
	// 设备标识与 USB 元信息。
	Name      string `json:"name"`
	VID       string `json:"vid"`
	PID       string `json:"pid"`
	SerialNum string `json:"serial_num"`
	Product   string `json:"product"`
	IsUSB     bool   `json:"is_usb"`

	// Mouse/keyboard directive target ports on firmware.
	// 固件侧鼠标/键盘指令端口标识。
	MousePortFlag    string `json:"mouse_port_flag"`
	KeyboardPortFlag string `json:"keyboard_port_flag"`

	// Runtime and I/O settings.
	// 运行时与 I/O 配置。
	VerboseDirective bool         `json:"verbose_directive"`
	Verbose          bool         `json:"verbose"`
	OpenMode         *serial.Mode `json:"-"`
	port             serial.Port  `json:"-"`

	// Internal synchronization and sync-output state.
	// 内部并发锁与同步输出状态。
	locker            sync.Mutex
	syncStateMu       sync.RWMutex
	SyncDirective     string
	SyncOuputEnabled  bool `json:"sync_output_enabled"`
	SyncOutputChan    chan string
	SyncOutputTimeout time.Duration `json:"sync_output_timeout"`
	SyncOutputIgnored bool          `json:"-"`

	// Async marker used to drop one unrelated echo frame.
	// 异步标记用于跳过一帧无关回显数据。
	withAsyncMark bool
	asyncMarkAt   time.Time
	readerBuf     []byte
	readClosed    bool `json:"-"`
}

// NewSerialPort creates an empty serial-port model.
// NewSerialPort 创建一个空的串口模型实例。
func NewSerialPort() *SerialPort {
	return &SerialPort{}
}

// Open opens the underlying serial device once.
// Open 打开底层串口设备（幂等）。
func (sp *SerialPort) Open() error {
	sp.locker.Lock()
	defer sp.locker.Unlock()

	if sp.port != nil {
		return nil
	}

	var err error
	sp.port, err = serial.Open(sp.Name, sp.OpenMode)
	if err == nil {
		sp.readClosed = false
	}
	return err
}

// Close closes the serial device and clears sync state.
// Close 关闭串口并重置同步状态。
func (sp *SerialPort) Close() error {
	sp.locker.Lock()
	defer sp.locker.Unlock()

	if sp.port == nil {
		return nil
	}

	err := sp.port.Close()
	sp.port = nil
	sp.readClosed = true
	sp.EmptySyncDirective()
	sp.clearAsyncMark()
	return err
}

// String returns human-readable port metadata.
// String 返回可读的串口元信息。
func (sp *SerialPort) String() string {
	return fmt.Sprintf("Name: %s, VID: %s, PID: %s, SerialNum: %s, Product: %s, IsUSB: %t", sp.Name, sp.VID, sp.PID, sp.SerialNum, sp.Product, sp.IsUSB)
}

// Write writes raw bytes to current serial port.
// Write 向当前串口写入原始字节数据。
func (sp *SerialPort) Write(data []byte) (int, error) {
	if sp.port == nil {
		return 0, fmt.Errorf("serial port %s is not open", sp.Name)
	}
	return sp.port.Write(data)
}

// Read_V0 is a compatibility alias of Read.
// Read_V0 是 Read 的兼容别名。
func (sp *SerialPort) Read_V0() (string, error) {
	return sp.Read()
}

// Read continuously reads serial output and resolves sync responses.
// Read 持续读取串口输出，并在同步模式下组装指令响应。
func (sp *SerialPort) Read() (string, error) {
	resultCache := make([]byte, 0)

	for {
		sp.locker.Lock()
		port := sp.port
		sp.locker.Unlock()
		if port == nil {
			return "", nil
		}

		n, err := port.Read(sp.readerBuf)
		if err != nil {
			if strings.Contains(err.Error(), "Port has been closed") {
				sp.readClosed = true
				if sp.Verbose {
					log.Println("serial port closed")
				}
			} else {
				log.Printf("read error: %v\n", err)
			}
			return "", err
		}

		if n <= 0 {
			continue
		}

		if sp.Verbose {
			log.Printf("received data: %q\n", sp.readerBuf[:n])
		}

		if !sp.IsSyncOutputEnabled() {
			continue
		}

		if sp.consumeAsyncMark() {
			resultCache = resultCache[:0]
			continue
		}

		syncDirective, syncOutputIgnored := sp.GetSyncDirectiveState()
		if syncDirective == "" {
			continue
		}

		resultCache = append(resultCache, sp.readerBuf[:n]...)

		hittedIdx := bytes.Index(resultCache, []byte(syncDirective))
		if hittedIdx < 0 {
			continue
		}

		// EOF
		var outputCompletedIdx int

		parser := GetRawDirectiveOutputParser(syncDirective)
		if parser == nil {
			outputCompletedIdx = bytes.Index(resultCache[hittedIdx:], []byte(EOFCLI))
		} else {
			outputCompletedIdx = bytes.Index(resultCache[hittedIdx:], []byte(parser.EOFFlag()))
		}

		if outputCompletedIdx < 0 {
			continue
		}

		output := ""
		if !syncOutputIgnored {
			output = string(resultCache[hittedIdx : hittedIdx+outputCompletedIdx])
		}

		sp.SyncOutputChan <- output
		sp.EmptySyncDirective()
		resultCache = resultCache[:0]
	}
}

// SetSyncDirective sets the current sync directive token and captures output.
// SetSyncDirective 设置当前等待匹配的同步指令标记，并捕获输出。
func (sp *SerialPort) SetSyncDirective(directive string) {
	sp.setSyncDirective(directive, false)
}

// SetSyncDirectiveIgnoreOutput sets the current sync directive token and ignores its output.
// SetSyncDirectiveIgnoreOutput 设置当前等待匹配的同步指令标记，但忽略输出内容。
func (sp *SerialPort) SetSyncDirectiveIgnoreOutput(directive string) {
	sp.setSyncDirective(directive, true)
}

func (sp *SerialPort) setSyncDirective(directive string, ignoreOutput bool) {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncDirective = directive
	sp.SyncOutputIgnored = ignoreOutput
}

// GetSyncDirective gets current sync directive token.
// GetSyncDirective 获取当前同步指令标记。
func (sp *SerialPort) GetSyncDirective() string {
	syncDirective, _ := sp.GetSyncDirectiveState()
	return syncDirective
}

// GetSyncDirectiveState gets current sync directive token and output-ignore flag.
// GetSyncDirectiveState 获取当前同步指令标记与是否忽略输出。
func (sp *SerialPort) GetSyncDirectiveState() (string, bool) {
	sp.syncStateMu.RLock()
	defer sp.syncStateMu.RUnlock()
	return sp.SyncDirective, sp.SyncOutputIgnored
}

// EmptySyncDirective clears sync directive token.
// EmptySyncDirective 清空同步指令标记。
func (sp *SerialPort) EmptySyncDirective() {
	sp.setSyncDirective("", false)
}

// EnableSyncOutput enables sync-output parsing mode.
// EnableSyncOutput 启用同步输出解析模式。
func (sp *SerialPort) EnableSyncOutput() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncOuputEnabled = true
}

// DisableSyncOutput disables sync-output parsing mode.
// DisableSyncOutput 禁用同步输出解析模式。
func (sp *SerialPort) DisableSyncOutput() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncOuputEnabled = false
}

// IsSyncOutputEnabled reports whether sync-output mode is enabled.
// IsSyncOutputEnabled 返回是否启用了同步输出模式。
func (sp *SerialPort) IsSyncOutputEnabled() bool {
	sp.syncStateMu.RLock()
	defer sp.syncStateMu.RUnlock()
	return sp.SyncOuputEnabled
}

// markAsync marks next frame as async side effect.
// markAsync 标记下一帧为异步副作用数据。
func (sp *SerialPort) markAsync() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.withAsyncMark = true
	sp.asyncMarkAt = time.Now()
}

// consumeAsyncMark consumes async marker if present.
// consumeAsyncMark 消费异步标记，若存在则返回 true。
func (sp *SerialPort) consumeAsyncMark() bool {
	const asyncMarkTTL = 2 * time.Second

	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()

	if sp.withAsyncMark && !sp.asyncMarkAt.IsZero() && time.Since(sp.asyncMarkAt) > asyncMarkTTL {
		sp.withAsyncMark = false
		sp.asyncMarkAt = time.Time{}
	}

	if !sp.withAsyncMark {
		return false
	}

	sp.withAsyncMark = false
	sp.asyncMarkAt = time.Time{}
	return true
}

// clearAsyncMark clears async marker state.
// clearAsyncMark 清空异步标记状态。
func (sp *SerialPort) clearAsyncMark() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.withAsyncMark = false
	sp.asyncMarkAt = time.Time{}
}
