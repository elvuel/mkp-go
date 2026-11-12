package mkpgo

import (
	"bytes"
	"context"
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
	locker             sync.Mutex
	directiveMu        sync.Mutex
	syncStateMu        sync.RWMutex
	SyncDirective      string
	SyncOuputEnabled   bool `json:"sync_output_enabled"`
	SyncOutputChan     chan string
	SyncOutputTimeout  time.Duration `json:"sync_output_timeout"`
	SyncOutputIgnored  bool          `json:"-"`
	syncReplyCh        chan string
	syncDirectiveLabel string    `json:"-"`
	pendingSyncBuffer  string    `json:"-"`
	lastSyncDirective  string    `json:"-"`
	lastSyncOutput     string    `json:"-"`
	lastSyncOutputAt   time.Time `json:"-"`

	// Async marker used to drop one unrelated echo frame.
	// 异步标记用于跳过一帧无关回显数据。
	withAsyncMark   bool
	asyncMarkAt     time.Time
	readerBuf       []byte
	readClosed      bool      `json:"-"`
	readLoopRunning bool      `json:"-"`
	lastReadAt      time.Time `json:"-"`
	lastReadErr     string    `json:"-"`
	lastReadErrAt   time.Time `json:"-"`
}

// SerialPortStatus is a snapshot for health reporting.
// SerialPortStatus 用于健康检查输出的串口状态快照。
type SerialPortStatus struct {
	Name                 string     `json:"name"`
	PortOpen             bool       `json:"port_open"`
	ReadClosed           bool       `json:"read_closed"`
	ReadLoopRunning      bool       `json:"read_loop_running"`
	SyncOutputEnabled    bool       `json:"sync_output_enabled"`
	SyncOutputTimeoutMS  int64      `json:"sync_output_timeout_ms"`
	CurrentSyncDirective string     `json:"current_sync_directive,omitempty"`
	SyncOutputIgnored    bool       `json:"sync_output_ignored"`
	AsyncMarked          bool       `json:"async_marked"`
	LastReadAt           *time.Time `json:"last_read_at,omitempty"`
	LastReadError        string     `json:"last_read_error,omitempty"`
	LastReadErrorAt      *time.Time `json:"last_read_error_at,omitempty"`
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
		sp.clearPendingSyncBuffer()
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
	sp.clearPendingSyncBuffer()
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
			sp.noteReadError(err)
			if strings.Contains(err.Error(), "Port has been closed") {
				sp.setReadClosed(true)
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

		sp.noteReadActivity()

		if sp.Verbose {
			log.Printf("received data: %q\n", sp.readerBuf[:n])
		}

		if !sp.IsSyncOutputEnabled() {
			continue
		}

		if sp.consumeAsyncMark() {
			resultCache = resultCache[:0]
			sp.clearPendingSyncBuffer()
			continue
		}

		syncDirective, syncOutputIgnored, syncReplyCh := sp.getSyncDirectiveState()
		if syncDirective == "" {
			continue
		}

		resultCache = append(resultCache, sp.readerBuf[:n]...)
		sp.setPendingSyncBuffer(string(resultCache))

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

		capturedOutput := string(resultCache[hittedIdx : hittedIdx+outputCompletedIdx])
		output := ""
		if !syncOutputIgnored {
			output = capturedOutput
		}

		sp.recordSyncOutput(syncDirective, capturedOutput)
		sp.clearPendingSyncBuffer()
		sp.deliverSyncOutput(output, syncReplyCh)
		sp.EmptySyncDirective()
		resultCache = resultCache[:0]
	}
}

// ReadLoopContext keeps the serial read loop alive until the port is closed or ctx is canceled.
// ReadLoopContext 持续维护串口读循环，直到端口关闭或 ctx 被取消。
func (sp *SerialPort) ReadLoopContext(ctx context.Context) error {
	const retryDelay = 200 * time.Millisecond
	sp.setReadLoopRunning(true)
	defer sp.setReadLoopRunning(false)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		_, err := sp.Read()
		if err == nil {
			if !sp.hasOpenPort() {
				return nil
			}
		} else {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !sp.hasOpenPort() || sp.isReadClosed() {
				return nil
			}
			log.Printf("serial read loop restarting after error: %v\n", err)
		}

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// SetSyncDirective sets the current sync directive token and captures output.
// SetSyncDirective 设置当前等待匹配的同步指令标记，并捕获输出。
func (sp *SerialPort) SetSyncDirective(directive string) {
	sp.setSyncDirectiveWithLabel(directive, directive, false, nil)
}

// SetSyncDirectiveIgnoreOutput sets the current sync directive token and ignores its output.
// SetSyncDirectiveIgnoreOutput 设置当前等待匹配的同步指令标记，但忽略输出内容。
func (sp *SerialPort) SetSyncDirectiveIgnoreOutput(directive string) {
	sp.setSyncDirectiveWithLabel(directive, directive, true, nil)
}

func (sp *SerialPort) setSyncDirective(directive string, ignoreOutput bool, replyCh chan string) {
	sp.setSyncDirectiveWithLabel(directive, directive, ignoreOutput, replyCh)
}

func (sp *SerialPort) setSyncDirectiveWithLabel(directive, label string, ignoreOutput bool, replyCh chan string) {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncDirective = directive
	sp.SyncOutputIgnored = ignoreOutput
	sp.syncReplyCh = replyCh
	sp.syncDirectiveLabel = label
	if directive != "" {
		sp.pendingSyncBuffer = ""
	}
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
	syncDirective, syncOutputIgnored, _ := sp.getSyncDirectiveState()
	return syncDirective, syncOutputIgnored
}

func (sp *SerialPort) getSyncDirectiveState() (string, bool, chan string) {
	directive, ignored, replyCh, _ := sp.getSyncDirectiveSnapshot()
	return directive, ignored, replyCh
}

func (sp *SerialPort) getSyncDirectiveSnapshot() (string, bool, chan string, string) {
	sp.syncStateMu.RLock()
	defer sp.syncStateMu.RUnlock()
	return sp.SyncDirective, sp.SyncOutputIgnored, sp.syncReplyCh, sp.syncDirectiveLabel
}

// EmptySyncDirective clears sync directive token.
// EmptySyncDirective 清空同步指令标记。
func (sp *SerialPort) EmptySyncDirective() {
	sp.setSyncDirective("", false, nil)
}

func (sp *SerialPort) setPendingSyncBuffer(buffer string) {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.pendingSyncBuffer = buffer
}

func (sp *SerialPort) clearPendingSyncBuffer() {
	sp.setPendingSyncBuffer("")
}

func (sp *SerialPort) pendingSyncBufferSnapshot() string {
	sp.syncStateMu.RLock()
	defer sp.syncStateMu.RUnlock()
	return sp.pendingSyncBuffer
}

func (sp *SerialPort) recordSyncOutput(directive, output string) {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.lastSyncDirective = directive
	sp.lastSyncOutput = output
	sp.lastSyncOutputAt = time.Now().UTC()
}

func (sp *SerialPort) currentSyncOutputChan() chan string {
	_, _, replyCh := sp.getSyncDirectiveState()
	if replyCh != nil {
		return replyCh
	}
	return sp.SyncOutputChan
}

func (sp *SerialPort) deliverSyncOutput(output string, replyCh chan string) {
	if replyCh != nil {
		select {
		case replyCh <- output:
		default:
		}
		return
	}
	if sp.SyncOutputChan == nil {
		return
	}
	select {
	case sp.SyncOutputChan <- output:
		return
	default:
	}
	select {
	case <-sp.SyncOutputChan:
	default:
	}
	select {
	case sp.SyncOutputChan <- output:
	default:
	}
}

func (sp *SerialPort) hasOpenPort() bool {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	return sp.port != nil
}

func (sp *SerialPort) isReadClosed() bool {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	return sp.readClosed
}

func (sp *SerialPort) setReadLoopRunning(running bool) {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	sp.readLoopRunning = running
}

func (sp *SerialPort) noteReadActivity() {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	now := time.Now()
	sp.lastReadAt = now
	sp.lastReadErr = ""
	sp.lastReadErrAt = time.Time{}
}

func (sp *SerialPort) noteReadError(err error) {
	if err == nil {
		return
	}
	sp.locker.Lock()
	defer sp.locker.Unlock()
	sp.lastReadErr = err.Error()
	sp.lastReadErrAt = time.Now()
}

func (sp *SerialPort) setReadClosed(closed bool) {
	sp.locker.Lock()
	defer sp.locker.Unlock()
	sp.readClosed = closed
}

// SnapshotStatus returns a concurrency-safe health snapshot.
// SnapshotStatus 返回线程安全的健康状态快照。
func (sp *SerialPort) SnapshotStatus() SerialPortStatus {
	var status SerialPortStatus

	sp.locker.Lock()
	status = SerialPortStatus{
		Name:                sp.Name,
		PortOpen:            sp.port != nil,
		ReadClosed:          sp.readClosed,
		ReadLoopRunning:     sp.readLoopRunning,
		SyncOutputTimeoutMS: sp.SyncOutputTimeout.Milliseconds(),
		LastReadError:       sp.lastReadErr,
	}
	if !sp.lastReadAt.IsZero() {
		lastReadAt := sp.lastReadAt
		status.LastReadAt = &lastReadAt
	}
	if !sp.lastReadErrAt.IsZero() {
		lastReadErrAt := sp.lastReadErrAt
		status.LastReadErrorAt = &lastReadErrAt
	}
	sp.locker.Unlock()

	sp.syncStateMu.RLock()
	status.SyncOutputEnabled = sp.SyncOuputEnabled
	status.CurrentSyncDirective = sp.SyncDirective
	status.SyncOutputIgnored = sp.SyncOutputIgnored
	status.AsyncMarked = sp.withAsyncMark
	sp.syncStateMu.RUnlock()

	return status
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
