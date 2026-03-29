package mkpgo

import (
	"fmt"
	"log"
	"strings"

	"go.bug.st/serial/enumerator"
)

// GetDetailedPortsList prints detected serial ports with USB metadata.
// GetDetailedPortsList 打印本机串口列表及 USB 元信息。
func GetDetailedPortsList() {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
		return
	}
	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}
}

// CheckSFSerialPort finds the first candidate SF serial port name.
// CheckSFSerialPort 查找首个可用的 SF 串口并返回端口名。
func CheckSFSerialPort() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}

	if len(ports) == 0 {
		return "", fmt.Errorf("no serial ports found")
	}

	for _, port := range ports {
		if port.IsUSB {
			if strings.ToUpper(port.VID) == "1A86" || strings.ToUpper(port.PID) != "55D3" {
				return port.Name, nil
			}
		}
	}

	return "", fmt.Errorf("serial port not found")
}
