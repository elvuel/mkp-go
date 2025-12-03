package mkpgo

import (
	"fmt"
)

type FileNode struct {
	DisplayName string     `json:"displayName,omitempty"` // 显示名称（友好的名称）
	Name        string     `json:"name"`                  // 节点名称（文件名/目录名）
	Path        string     `json:"path"`                  // 节点完整路径
	Type        string     `json:"type"`                  // 类型：directory（目录）/file（文件）
	Size        int        `json:"size"`                  // 大小（字节），目录通常为 0
	Contents    []FileNode `json:"contents,omitempty"`    // 子节点列表（仅目录有值）
}

// FileSystem 最外层结构体，对应 JSON 根对象
type FileSystem struct {
	RootDir FileNode `json:"rootDir,omitempty"` // 根目录节点（eMMC）
	Error   string   `json:"error,omitempty"`   // 错误信息
}

type Heartbeat struct {
	Timetamp int64 `json:"timetamp"`
}

type LogBasicOption struct {
	// "width":0, "heigh":0, "stpos": {"x":-1,"y":-1 }
	Width  int `json:"width"`
	Height int `json:"heigh"`
	StPos  struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"stpos"`
}

func (opt *LogBasicOption) CliArgs() []string {
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

type LogInfo struct {
	LogBasicOption
	LogLength
}

type LogLength struct {
}

type MKPVersion struct {
	UVersion string `json:"uver"`
	AVersion string `json:"aver"`
}
