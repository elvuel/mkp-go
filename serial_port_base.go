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

type SerialPort struct {
	Name      string `json:"name"`
	VID       string `json:"vid"`
	PID       string `json:"pid"`
	SerialNum string `json:"serial_num"`
	Product   string `json:"product"`
	IsUSB     bool   `json:"is_usb"`

	MousePortFlag    string `json:"mouse_port_flag"`
	KeyboardPortFlag string `json:"keyboard_port_flag"`

	Verbose  bool         `json:"verbose"`
	OpenMode *serial.Mode `json:"-"`
	port     serial.Port  `json:"-"`

	locker            sync.Mutex
	syncStateMu       sync.RWMutex
	SyncDirective     string
	SyncOuputEnabled  bool `json:"sync_output_enabled"`
	SyncOutputChan    chan string
	SyncOutputTimeout time.Duration `json:"sync_output_timeout"`

	withAsyncMark bool
	asyncMarkAt   time.Time
	readerBuf     []byte
	readClosed    bool `json:"-"`
}

func NewSerialPort() *SerialPort {
	return &SerialPort{}
}

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

func (sp *SerialPort) String() string {
	return fmt.Sprintf("Name: %s, VID: %s, PID: %s, SerialNum: %s, Product: %s, IsUSB: %t", sp.Name, sp.VID, sp.PID, sp.SerialNum, sp.Product, sp.IsUSB)
}

func (sp *SerialPort) Write(data []byte) (int, error) {
	if sp.port == nil {
		return 0, fmt.Errorf("serial port %s is not open", sp.Name)
	}
	return sp.port.Write(data)
}

func (sp *SerialPort) Read_V0() (string, error) {
	return sp.Read()
}

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

		syncDirective := sp.GetSyncDirective()
		if syncDirective == "" {
			continue
		}

		resultCache = append(resultCache, sp.readerBuf[:n]...)

		hittedIdx := bytes.Index(resultCache, []byte(syncDirective))
		if hittedIdx < 0 {
			continue
		}

		cliIdx := bytes.Index(resultCache[hittedIdx:], []byte("cli>"))
		if cliIdx < 0 {
			continue
		}

		sp.SyncOutputChan <- string(resultCache[hittedIdx : hittedIdx+cliIdx])
		sp.EmptySyncDirective()
		resultCache = resultCache[:0]
	}
}

func (sp *SerialPort) SetSyncDirective(directive string) {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncDirective = directive
}

func (sp *SerialPort) GetSyncDirective() string {
	sp.syncStateMu.RLock()
	defer sp.syncStateMu.RUnlock()
	return sp.SyncDirective
}

func (sp *SerialPort) EmptySyncDirective() {
	sp.SetSyncDirective("")
}

func (sp *SerialPort) EnableSyncOutput() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncOuputEnabled = true
}

func (sp *SerialPort) DisableSyncOutput() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.SyncOuputEnabled = false
}

func (sp *SerialPort) IsSyncOutputEnabled() bool {
	sp.syncStateMu.RLock()
	defer sp.syncStateMu.RUnlock()
	return sp.SyncOuputEnabled
}

func (sp *SerialPort) markAsync() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.withAsyncMark = true
	sp.asyncMarkAt = time.Now()
}

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

func (sp *SerialPort) clearAsyncMark() {
	sp.syncStateMu.Lock()
	defer sp.syncStateMu.Unlock()
	sp.withAsyncMark = false
	sp.asyncMarkAt = time.Time{}
}
