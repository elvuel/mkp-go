package mkpgo

import (
	"encoding/json"
	"fmt"
)

// SN is serial-number response model.
// SN 是序列号响应结构。
type SN struct {
	// SN is device serial number string.
	// SN 是设备序列号。
	SN string `json:"sn"`
}

// FileNode describes one file-system node returned by firmware.
// FileNode 描述固件返回的文件系统节点。
type FileNode struct {
	DisplayName string     `json:"displayName,omitempty"` // 显示名称（友好的名称）
	Name        string     `json:"name"`                  // 节点名称（文件名/目录名）
	Path        string     `json:"path"`                  // 节点完整路径
	Type        string     `json:"type"`                  // 类型：directory（目录）/file（文件）
	Size        int        `json:"size"`                  // 大小（字节），目录通常为 0
	Contents    []FileNode `json:"contents,omitempty"`    // 子节点列表（仅目录有值）
}

// FileSystem is root response model for filesystem queries.
// FileSystem 是文件系统查询的根响应结构。
type FileSystem struct {
	RootDir FileNode `json:"rootDir,omitempty"` // 根目录节点（eMMC）
	Error   string   `json:"error,omitempty"`   // 错误信息
}

// Heartbeat represents alive response payload.
// Heartbeat 表示存活检测返回结构。
type Heartbeat struct {
	Timetamp int64 `json:"timetamp"`
}

// LogOption defines optional arguments for alog command.
// LogOption 定义 alog 命令的可选参数。
type LogOption struct {
	// "width":0, "heigh":0, "stpos": {"x":-1,"y":-1 }
	Width  int `json:"width"`
	Height int `json:"heigh"`
	StPos  struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"stpos"`
}

// CliArgs converts log options to CLI argument list.
// CliArgs 将日志参数转换为命令行参数数组。
func (opt *LogOption) CliArgs() []string {
	// alog --width 1920 --heigh 1024 --stposx 300 --stposy 300 circle3
	args := make([]string, 0)
	if opt.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", opt.Width))
	}
	if opt.Height > 0 {
		args = append(args, "--heigh", fmt.Sprintf("%d", opt.Height))
	}
	if opt.StPos.X > -1 {
		args = append(args, "--stposx", fmt.Sprintf("%d", opt.StPos.X))
	}
	if opt.StPos.Y > -1 {
		args = append(args, "--stposy", fmt.Sprintf("%d", opt.StPos.Y))
	}
	return args
}

// JoinOption defines optional arguments for join Wi-Fi command.
// JoinOption 定义 join Wi-Fi 命令的可选参数。
type JoinOption struct {
	SSID     string `json:"ssid"`     // Wi-Fi name/SSID
	Password string `json:"password"` // Wi-Fi password
}

// CliArgs converts join options to CLI argument list.
// CliArgs 将 join 参数转换为命令行参数数组。
func (opt *JoinOption) CliArgs() []string {
	if opt == nil || (opt.SSID == "" && opt.Password == "") {
		return nil
	}
	args := make([]string, 0, 2)
	if opt.SSID != "" {
		args = append(args, opt.SSID)
	}
	if opt.Password != "" {
		args = append(args, opt.Password)
	}
	return args
}

// WifiAutoOption defines optional argument for wifi_auto command.
// WifiAutoOption 定义 wifi_auto 命令的可选参数。
type WifiAutoOption struct {
	State string `json:"state"` // Auto-connect switch: "1" enables, "0" disables.
}

// CliArgs converts wifi_auto options to CLI argument list.
// CliArgs 将 wifi_auto 参数转换为命令行参数数组。
func (opt *WifiAutoOption) CliArgs() []string {
	if opt == nil || opt.State == "" {
		return nil
	}
	return []string{opt.State}
}

// AGetOption defines arguments for aget command.
// AGetOption 定义 aget 命令参数。
type AGetOption struct {
	FilePath string `json:"filepath"` // File path under file server root.
}

// CliArgs converts aget options to CLI argument list.
// CliArgs 将 aget 参数转换为命令行参数数组。
func (opt *AGetOption) CliArgs() []string {
	if opt == nil || opt.FilePath == "" {
		return nil
	}
	return []string{opt.FilePath}
}

// APutOption defines arguments for aput command.
// APutOption 定义 aput 命令参数。
type APutOption struct {
	FilePath string `json:"filepath"` // File path under MKP /eMMC root.
}

// CliArgs converts aput options to CLI argument list.
// CliArgs 将 aput 参数转换为命令行参数数组。
func (opt *APutOption) CliArgs() []string {
	if opt == nil || opt.FilePath == "" {
		return nil
	}
	return []string{opt.FilePath}
}

// AHTTPBaseOption defines optional argument for ahttpbase command.
// AHTTPBaseOption 定义 ahttpbase 命令的可选参数。
type AHTTPBaseOption struct {
	URL string `json:"url"` // HTTP base URL, without trailing slash.
}

// CliArgs converts ahttpbase options to CLI argument list.
// CliArgs 将 ahttpbase 参数转换为命令行参数数组。
func (opt *AHTTPBaseOption) CliArgs() []string {
	if opt == nil || opt.URL == "" {
		return nil
	}
	return []string{opt.URL}
}

// AHTTPBase is the JSON output model returned by ahttpbase.
// AHTTPBase 是 ahttpbase 返回的 JSON 结构。
type AHTTPBase struct {
	AHTTPBase string `json:"ahttpbase"`
}

// AdumjOption defines arguments for adumj command.
// AdumjOption 定义 adumj 命令参数。
type AdumjOption struct {
	LogPath string `json:"logPath"` // Log file path/name.
}

// CliArgs converts adumj options to CLI argument list.
// CliArgs 将 adumj 参数转换为命令行参数数组。
func (opt *AdumjOption) CliArgs() []string {
	if opt == nil || opt.LogPath == "" {
		return nil
	}
	return []string{opt.LogPath}
}

// ActionDump is the JSON output model returned by adumj.
// ActionDump 是 adumj 返回的动作日志 JSON 结构。
type ActionDump struct {
	Format  string                       `json:"format"`
	Version string                       `json:"version"`
	Meta    ActionDumpMeta               `json:"meta"`
	Events  []map[string]json.RawMessage `json:"events"`
}

// ActionDumpMeta describes action log capture metadata.
// ActionDumpMeta 描述动作日志录制元信息。
type ActionDumpMeta struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	StartX int `json:"startX"`
	StartY int `json:"startY"`
}

// LogInfo contains both log option metadata and duration.
// LogInfo 同时包含日志参数与时长信息。
type LogInfo struct {
	LogOption
	LogLength
}

// LogLength is duration model returned by firmware.
// LogLength 是固件返回的时长结构。
type LogLength struct {
	Seconds      int `json:"seconds"`
	Milliseconds int `json:"milsec"`
}

// MKPVersion is firmware version payload.
// MKPVersion 表示固件版本信息。
type MKPVersion struct {
	UVersion string `json:"uver"`
	AVersion string `json:"aver"`
}
