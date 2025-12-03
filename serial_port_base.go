package mkpgo

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"

	"go.bug.st/serial"
)

// - v20250926-5
// - v20250930

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

	locker           sync.Mutex
	SyncDirective    string
	SyncOuputEnabled bool `json:"sync_output_enabled"`
	SyncOutputChan   chan string

	withAsyncMark bool // 同步输出状态下，任然支持“异步”

	readerBuf  []byte
	readClosed bool `json:"-"`
}

func NewSerialPort() *SerialPort {
	return &SerialPort{}
}

func (sp *SerialPort) Open() error {
	if sp.port != nil {
		return nil
	}
	var err error
	sp.port, err = serial.Open(sp.Name, sp.OpenMode)
	return err
}

func (sp *SerialPort) Close() error {
	if sp.port != nil {
		return sp.port.Close()
	}

	return nil
}

func (sp *SerialPort) String() string {
	return fmt.Sprintf("Name: %s, VID: %s, PID: %s, SerialNum: %s, Product: %s, IsUSB: %t", sp.Name, sp.VID, sp.PID, sp.SerialNum, sp.Product, sp.IsUSB)
}

func (sp *SerialPort) Write(data []byte) (int, error) {
	return sp.port.Write(data)
}

func (sp *SerialPort) Read_V0() (string, error) {

	if sp.SyncOuputEnabled {
		result := make([]byte, 0)
		var n int
		var err error
		for {
			// 从串口读取数据
			n, err = sp.port.Read(sp.readerBuf)
			if err != nil {
				if strings.Contains(err.Error(), "Port has been closed") {
					sp.readClosed = true
					log.Println("串口已关闭")
				} else {
					log.Printf("读取错误: %v\n", err)
				}
				break // 遇到EOF或其他错误时退出循环
			}

			if n > 0 {
				// 处理接收到的数据
				if sp.Verbose {
					fmt.Printf("收到数据: %q\n", sp.readerBuf[:n])
				}

				result = append(result, sp.readerBuf[:n]...)

				hittedIdx := bytes.Index(result, []byte(sp.SyncDirective))

				if hittedIdx >= 0 {
					// looking for cli> after hittedIdx
					cliIdx := bytes.Index(result[hittedIdx:], []byte("cli>"))
					if cliIdx >= 0 {
						sp.SyncOutputChan <- string(result[hittedIdx : hittedIdx+cliIdx])

						sp.EmptySyncDirective()
						result = result[:0]
					}
				}
			}
		}
	}

	result := ""
	var n int
	var err error
	for {
		// 从串口读取数据
		n, err = sp.port.Read(sp.readerBuf)
		if err != nil {
			if strings.Contains(err.Error(), "Port has been closed") {
				sp.readClosed = true
				log.Println("串口已关闭")
			} else {
				log.Printf("读取错误: %v\n", err)
			}
			break // 遇到EOF或其他错误时退出循环
		}

		if n > 0 {
			// 处理接收到的数据
			if sp.Verbose {
				fmt.Printf("收到数据: %q\n", sp.readerBuf[:n])
			}
		}
	}

	return result, err
}

func (sp *SerialPort) Read() (string, error) {

	resultCache := make([]byte, 0)
	var n int
	var err error
	for {
		// 从串口读取数据
		n, err = sp.port.Read(sp.readerBuf)
		if err != nil {
			if strings.Contains(err.Error(), "Port has been closed") {
				sp.readClosed = true
				log.Println("串口已关闭")
			} else {
				log.Printf("读取错误: %v\n", err)
			}
			break // 遇到EOF或其他错误时退出循环
		}

		if n > 0 {
			// 处理接收到的数据
			if sp.Verbose {
				fmt.Printf("收到数据: %q\n", sp.readerBuf[:n])
			}

			if sp.SyncOuputEnabled {
				if !sp.withAsyncMark {
					resultCache = append(resultCache, sp.readerBuf[:n]...)

					hittedIdx := bytes.Index(resultCache, []byte(sp.SyncDirective))

					if hittedIdx >= 0 {
						// looking for cli> after hittedIdx
						cliIdx := bytes.Index(resultCache[hittedIdx:], []byte("cli>"))
						if cliIdx >= 0 {
							sp.SyncOutputChan <- string(resultCache[hittedIdx : hittedIdx+cliIdx])
							// clear the result to empty
							sp.SyncDirective = ""
							resultCache = resultCache[:0]
						}
					}
				} else { // 如果是“异步”标识， 无条件直清空cache
					resultCache = resultCache[:0]
					sp.withAsyncMark = false // 重置
				}
			}
		}
	}

	return "", err
}

func (sp *SerialPort) SetSyncDirective(directive string) {
	sp.SyncDirective = directive
}

func (sp *SerialPort) EmptySyncDirective() {
	sp.SetSyncDirective("")
}

func (sp *SerialPort) EnableSyncOutput() {
	sp.SyncOuputEnabled = true
}

func (sp *SerialPort) DisableSyncOutput() {
	sp.SyncOuputEnabled = false
}
