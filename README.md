# MKP Go

[English](./README.en.md) | 中文 | [指令帮助索引](./directives.md) | [Controller.MouseMove 使用说明](./doc/controller_mouse_move.md)

---

### MKP 是什么？

**MKP** 在本项目中指一套通过串口控制的 **鼠标 / 键盘录制、回放与文件管理设备及其 CLI 指令协议**。

它通常运行在带 USB HID 能力的外部设备 / 固件上：上位机通过串口向 MKP 设备发送 `m10`、`kpad`、`alog`、`aplay`、`list_dir` 等指令，设备再以 USB HID 鼠标 / 键盘的形式向目标电脑输出输入事件，或在设备端录制、回放、检查和管理输入日志 / 文件。

`mkp-go` 是 MKP 协议的 Go SDK / 客户端库，主要负责：

- 打开和管理 MKP 设备串口；
- 发送原始 CLI 指令；
- 封装鼠标、键盘、录制、回放、文件管理等常用操作；
- 解析设备返回的 JSON 或文本输出；
- 提供更高层的 helper / controller API，方便直接实现点击、移动、按键、录制和回放。

> 简单来说：**MKP 是硬件 / 固件侧的鼠标键盘自动化与回放协议；MKP Go 是用 Go 操作它的库。**

### 项目定位

`mkp-go` 不是单纯的系统级鼠标键盘模拟库。它的核心链路是：

```text
Go 程序 -> 串口 -> MKP 设备 / 固件 -> USB HID 鼠标 / 键盘事件 -> 目标系统
```

这种方式适合：

- 需要通过外部 HID 设备输出鼠标 / 键盘事件的场景；
- 需要在设备侧录制、保存、检查和回放输入日志的场景；
- 需要把低层 MKP CLI 指令封装成 Go API、Controller 能力或工具链的自动化场景。

### 主要能力

| 能力 | 说明 | 相关指令/API |
|---|---|---|
| 串口连接 | 自动 / 手动配置串口参数并读写设备 | `NewSFSerialPort`、`Open`、`Read`、`Close` |
| 原始指令 | 同步 / 异步发送任意 MKP CLI 指令 | `SendDirective`、`SendDirectiveAsync` |
| 鼠标控制 | 相对移动、按键、滚轮、释放 | `m10`、`Mouse10`、`MouseReleaseAll` |
| 键盘控制 | 按下、释放、点击、组合键、全释放 | `kpad`、`Keypad`、`helper.KeyTap` |
| 录制回放 | 在设备端录制、停止、取消、回放日志 | `alog`、`astop`、`acancel`、`aplay` |
| 文件管理 | 目录浏览、删文件、清目录、日志导入导出 | `list_dir`、`delete_file`、`clean_dir`、`aget`、`aput`、`ajson2log` |
| 设备信息 | SN、心跳、版本、日志长度与详情 | `sn`、`alive`、`aversion`、`atime`、`ainsp` |
| 网络配置 | Wi‑Fi、HTTP 基址等设备侧网络能力 | `join`、`wifi_auto`、`ahttpbase` |
| 高层封装 | helper / controller 高层鼠标键盘动作 | `helper`、`controller` |

## 安装

```bash
go get github.com/elvuel/mkp-go
```

本仓库 `go.mod` 当前声明：

```text
go 1.25.0
```

## 快速开始

```go
package main

import (
    "fmt"

    mkpgo "github.com/elvuel/mkp-go"
    "github.com/elvuel/mkp-go/helper"
)

func main() {
    sfport := mkpgo.NewSFSerialPort()
    sfport.Name = "COM5" // 按实际设备端口修改

    if err := sfport.Open(); err != nil {
        panic(err)
    }
    defer sfport.Close()

    // 同步读取输出时必须启动读循环。
    go sfport.Read()

    sn, err := helper.DeviceSN(sfport)
    if err != nil {
        panic(err)
    }

    fmt.Println("device sn:", sn.SN)
}
```

## 示例

下表按“源码路径 / 运行命令 / 示例行为”整理 Go 版示例。

| 示例 | 源码路径 | 运行命令 | 示例行为 |
|---|---|---|---|
| `m10` | `examples/m10/main.go` | `go run .\examples\m10` | 演示鼠标相对移动、按键按下与释放 |
| `kpad` | `examples/kpad/main.go` | `go run .\examples\kpad` | 演示键盘字符、组合键与自动释放 |
| `keys` | `examples/keys/main.go` | `go run .\examples\keys` | 演示 `helper.KeyTap`、`KeyDown`、`KeyUp` |
| `context` | `examples/context/main.go` | `go run .\examples\context` | 演示带 `context` 的超时取消与手动取消 |
| `commonpkg` | `examples/commonpkg/common.go` | 被其它示例复用 | 提供串口初始化、停止录制与日志输出辅助 |

常见 directive 示例同样按表整理：

| 指令示例 | 源码路径 | 运行命令 | 用途 |
|---|---|---|---|
| `sn` / `aversion` / `alive` | `examples/directives/<name>/main.go` | `go run .\examples\directives\sn` | 查询设备序列号、版本与在线状态 |
| `list_dir` / `clean_dir` / `delete_file` | `examples/directives/<name>/main.go` | `go run .\examples\directives\list_dir /eMMC/applog` | 浏览目录、清理目录、删除文件 |
| `join` / `wifi_auto` / `ahttpbase` | `examples/directives/<name>/main.go` | `go run .\examples\directives\join <ssid> <password>` | 配置 Wi‑Fi、自动联网与 HTTP 基地址 |
| `ainsp` / `atime` | `examples/directives/<name>/main.go` | `go run .\examples\directives\atime /eMMC/applog/spike.log` | 查看日志信息与时长 |
| `adumj` / `aget` / `aput` / `ajson2log` | `examples/directives/<name>/main.go` | `go run .\examples\directives\ajson2log D:\tmp\demo.json /eMMC/applog/demo.log` | 上传、下载与 JSON/日志互转 |
| `alog` / `astop` | `examples/directives/<name>/main.go` | `go run .\examples\directives\alog mkpdemo/demo` | 开始录制与停止录制 |

## 录制与回放

```go
// 异步开始录制
err := sfport.StartRecording("demo")

// 停止录制
err = sfport.StopRecording()

// 异步回放 demo，delay 为 0
err = sfport.StartReplaying("demo", 0)

// 取消回放
err = sfport.CancelReplay()
```

如需同步发送 `alog` 并等待输出，可使用：

```go
out, err := helper.Alog(sfport, "demo", nil)
fmt.Println(out, err)
```

## 文件与设备信息

```go
fs, err := helper.ListDir(sfport, "/eMMC/applog")
version, err := helper.Aversion(sfport)
hb, err := helper.Alive(sfport)
length, err := helper.Atime(sfport, "demo/record1")
info, err := helper.AInspect(sfport, "demo/record1")
```

## 指令模式

mkp-go 支持三类发送模式：

| 模式 | API | 说明 |
|---|---|---|
| 同步并返回输出 | `SendDirective` | 等待设备输出结束并返回原始输出。 |
| 同步但忽略输出 | `SendDirectiveIgnoreOutput` | 等待设备完成，但返回空输出；适合 `m10`、`kpad`。 |
| 异步 | `SendDirectiveAsync` | 只写入串口，不等待输出；适合高频键鼠操作。 |

## 仓库结构

| 路径 | 说明 |
|---|---|
| `serial_port_base.go` | 串口基础读写与同步输出处理。 |
| `sf_serial_port.go` | 面向 MKP/SF 设备的串口封装和核心指令方法。 |
| `directive_parser.go` | 内置指令输出解析器。 |
| `types_cli.go` | 设备信息、日志信息、文件系统等返回结构。 |
| `types_m10.go` | 鼠标 `m10` 参数模型。 |
| `types_keycode*.go` | 键盘 HID 键码和 `kpad` 参数模型。 |
| `helper/` | 面向业务的便捷函数。 |
| `controller/` | 更高层的鼠标/键盘控制器。 |
| `directives.md` | 完整 MKP 指令帮助索引。 |

## 注意事项

- 使用同步指令时，请确保 `SyncOuputEnabled=true` 且已执行 `go sfport.Read()`。
- `m10` / `kpad` 默认异步发送；如需串行保证完成，可通过 `WithAsync(false)` 或 `WithSyncIgnoreOutput(true)` 切换为同步等待且忽略输出。
- 文件删除/清理 helper 默认限制在 `/eMMC/applog` 下，避免误操作设备其他路径。
- 默认串口配置来自 `NewSFSerialPort()`，实际使用时通常需要修改 `Name`、`VID`、`PID`、`SerialNum` 等字段。
- 具体固件是否支持某个指令，以设备固件版本为准。
- 同步超时或解析失败时，会自动写入串口调试快照到 `.mkp-debug/`；可通过环境变量 `MKP_DEBUG_SERIAL_DIR` 覆盖目录。

---












