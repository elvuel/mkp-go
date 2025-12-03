package mkpgo

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

type LogLength struct {
	Seconds    int `json:"seconds"`
	Milseconds int `json:"milsec"`
}
