# MKP 指令帮助索引

[English](./directives.en.md) | 中文

---

> 本文档根据当前仓库代码整理：`sf_serial_port.go`、`directive_parser.go`、`types_cli.go`、`types_m10.go`、`types_keycode.go`、`helper/helper.go`。  
> 目标是提供一份可快速查找的 **CLI 指令 / Go API / 参数 / 输出解析器** 索引。

## 目录

- [使用约定](#使用约定)
- [快速总览](#快速总览)
- [同步/异步发送规则](#同步异步发送规则)
- [录制与回放指令](#录制与回放指令)
- [设备、网络与文件系统指令](#设备网络与文件系统指令)
- [鼠标指令：m10](#鼠标指令m10)
- [键盘指令：kpad](#键盘指令kpad)
- [输出解析器索引](#输出解析器索引)
- [返回结构索引](#返回结构索引)
- [Go API 索引](#go-api-索引)
- [草案/未封装指令](#草案未封装指令)

## 使用约定

- 默认串口类型：`*mkpgo.SFSerialPort`，由 `mkpgo.NewSFSerialPort()` 创建。
- 默认端口标记：鼠标 `MousePortFlag = "1"`，键盘 `KeyboardPortFlag = "2"`。
- 所有底层指令发送时都会追加 `\r\n`。
- 同步输出开关字段名为 `SyncOuputEnabled`（代码中保留该拼写）。默认由 `NewSFSerialPort()` 置为 `true`。
- 使用同步读取时，需要启动读循环：`go sfport.Read()`。
- 内置解析器通过 `mkpgo.InitParsers()` 注册（`init.go` 中已自动调用）。

最小使用示例：

```go
sfport := mkpgo.NewSFSerialPort()
sfport.Name = "COM5"
if err := sfport.Open(); err != nil {
    panic(err)
}
defer sfport.Close()

go sfport.Read()

sn, err := helper.DeviceSN(sfport)
fmt.Println(sn, err)
```

## 快速总览

| 指令 | CLI 形式 | 说明 | Go API | 输出/解析 |
|---|---|---|---|---|
| `alog` | `alog [logName] [options...]` | 开始设备端录制；同步发送时可等待录制结束输出 | `StartRecording` / `helper.StartRecord` / `helper.Alog` | 文本，解析器 `RawDirective_alog` |
| `aplay` | `aplay [logName] [--delay ms]` | 开始回放录制日志 | `StartReplaying` | 无内置解析器，通常异步 |
| `astop` | `astop` | 停止当前录制 | `StopRecording` / `Stop` / `helper.Astop` | 文本，解析器 `RawDirective_astop` |
| `acancel` | `acancel` | 取消当前回放任务 | `CancelReplay` / `helper.Cancel` | JSON 解析器存在，但 API 默认异步 |
| `sn` | `sn` | 获取设备序列号 | `helper.DeviceSN` | JSON -> `SN` |
| `list_dir` | `list_dir <path>` | 列出目录内容 | `helper.ListDir` | JSON -> `FileSystem` |
| `clean_dir` | `clean_dir <path>` | 清空目录 | `helper.CleanDir` | 成功返回空；失败返回错误 |
| `delete_file` | `delete_file <path>` | 删除文件 | `helper.DeleteFile` | 成功返回空；失败返回错误 |
| `join` | `join [ssid password]` | 连接 Wi-Fi；无参数时使用最近保存配置 | `helper.Join` / `Controller.Join` | 文本，解析器 `RawDirective_join`；成功包含 `connect: Connected` |
| `wifi_auto` | `wifi_auto [0|1]` | 查看或设置 Wi-Fi 自动连接状态 | `helper.WifiAuto` / `Controller.WifiAuto` | 文本，解析器 `RawDirective_wifi_auto`；查询返回 `on` / `off` |
| `adumj` | `adumj <logPath>` | 将日志文件转为便于解读的 JSON | `helper.Adumj` / `Controller.Adumj` | JSON -> `ActionDump`；EOF 为 `cli>`，成功内容包含 `<EOF>` |
| `ahttpbase` | `ahttpbase [url]` | 查看或设置文件管理服务器 API endpoint base URL | `helper.AHTTPBase` / `Controller.AHTTPBase` | JSON -> `AHTTPBase`；EOF 为 `cli>` |
| `alive` | `alive` | 心跳/存活检测 | `helper.Alive` | JSON -> `Heartbeat` |
| `atime` | `atime <path>` | 获取日志时长 | `helper.Atime` | JSON -> `LogLength` |
| `aversion` | `aversion` | 获取版本信息 | `helper.Aversion` | JSON -> `MKPVersion` |
| `ainsp` | `ainsp <path>` | 获取日志基础信息 | `helper.AInspect` | JSON -> `LogInfo` |
| `m10` | `m10 --port <port> [--b n] [--x n] [--y n] [--w n]` | 鼠标控制 | `Mouse10` / `helper.M10` / controller 鼠标方法 | 无内置解析器；同步模式可忽略输出 |
| `kpad` | `kpad --port <port> [--s hex] [--x1 hex]... [--rel n] [--d n] [--v 1]` | 键盘控制 | `Keypad` / helper/controller 键盘方法 | 无内置解析器；同步模式可忽略输出 |

## 同步/异步发送规则

### 底层发送 API

| API | 行为 |
|---|---|
| `SendDirective(directive[, opts...])` / `SendDirectiveContext(ctx, directive[, opts...])` / `SendSyncDirective(...)` | 同步发送；若 `SyncOuputEnabled=true`，等待 `Read()` 捕获到该指令的结束标记并返回原始输出；可选 `DirectiveOption` 覆盖本次等待设置。 |
| `SendDirectiveIgnoreOutput(directive[, opts...])` / `SendDirectiveIgnoreOutputContext(ctx, directive[, opts...])` | 同步发送，但只等待完成标记，不返回输出内容；可选 `DirectiveOption` 覆盖本次等待设置。 |
| `SendDirectiveAsync(directive)` / `SendDirectiveAsyncContext(ctx, directive)` | 异步发送；不等待输出。 |

### 结束标记

- 已注册解析器的指令使用解析器自己的 `EOFFlag()`。
- 未注册解析器的指令默认使用 CLI 提示符 `cli>` 作为结束标记。
- 当前常量：
  - `EOFDefault = "<EOF>"`
  - `EOFCLI = "cli>"`
- `alog` 的同步匹配会被规范化为 `alog`，因为实际输出不一定以完整 CLI 文本开头。

### 同步输出超时

- 默认超时由 `SFSerialPort.SyncOutputTimeout` 控制，`NewSFSerialPort()` 默认是 `10 * time.Second`。
- `SendDirective` / `SendDirectiveContext` / `SendSyncDirective` / `SendSyncDirectiveContext` 支持传入可选 `DirectiveOption`，例如 `WithSyncOutputTimeout(timeout)` 只覆盖本次同步等待。
- 未传入 `WithSyncOutputTimeout` 时，使用默认 `SyncOutputTimeout`。
- `WithSyncOutputTimeout(0)` 表示本次不启用定时器，仅由 `context` 取消结束等待。

```go
out, err := sfport.SendSyncDirective("join ssid password", mkpgo.WithSyncOutputTimeout(30*time.Second))
out, err = sfport.SendDirectiveContext(ctx, "alive") // 使用默认 SyncOutputTimeout
```

### 常见建议

- 需要拿到结构化结果的指令：使用同步 helper（例如 `helper.DeviceSN`、`helper.ListDir`）。
- 键鼠实时控制：默认使用异步（`M10Option.Async=true`、`KpadOption.Async=true`）。
- 如果需要串行保证键鼠指令已完成，可设置 `WithAsync(false)`；当前实现会同步等待结束标记但忽略输出。

## 录制与回放指令

### `alog`：开始录制 / 获取录制输出

**CLI：**

```text
alog [logName] [--width n] [--heigh n] [--stposx n] [--stposy n]
```

> 代码中的 `LogOption.CliArgs()` 使用 `--heigh`，不是 `--height`。

**参数：**

| 参数 | 说明 |
|---|---|
| `logName` | 日志名称，可为空。helper 会把它作为第一个参数拼接。 |
| `--width n` | 录制区域宽度，`LogOption.Width > 0` 时输出。 |
| `--heigh n` | 录制区域高度，`LogOption.Height > 0` 时输出。 |
| `--stposx n` | 起始 X，`LogOption.StPos.X > -1` 时输出。 |
| `--stposy n` | 起始 Y，`LogOption.StPos.Y > -1` 时输出。 |

**Go API：**

```go
// 异步开始录制
err := sfport.StartRecording("demo")
err = helper.StartRecord(sfport, "demo", opt)

// 同步发送 alog 并等待输出
out, err := helper.Alog(sfport, "demo", opt)
```

**解析器：** `RawDirective_alog`

- 输出类型：文本。
- 结束标记：`cli>`。
- 解析逻辑：清理 `\r`，去掉 CLI 前缀；有内容返回文本，无内容返回空字符串。

### `aplay`：开始回放

**CLI：**

```text
aplay [logName] [--delay ms]
```

**参数：**

| 参数 | 说明 |
|---|---|
| `logName` | 要回放的日志名称；为空时不拼接。 |
| `--delay ms` | 回放延迟；`delay >= 0` 时拼接。 |

**Go API：**

```go
err := sfport.StartReplaying("demo", 0)
err = sfport.StartReplayingContext(ctx, "demo", 100)
```

**解析器：** 当前未注册内置解析器。`StartReplaying` 默认异步发送。

### `astop`：停止录制

**CLI：**

```text
astop
```

**Go API：**

```go
err := sfport.StopRecording()
err = sfport.Stop()
err = helper.StopRecord(sfport)
err = helper.Astop(sfport)
```

**解析器：** `RawDirective_astop`

- 输出类型：文本。
- 结束标记：`cli>`。
- 解析逻辑：有内容返回文本；无内容返回 `ErrRawDirectiveParseFailed`。

### `acancel`：取消回放

**CLI：**

```text
acancel
```

**Go API：**

```go
err := sfport.CancelReplay()
err = helper.Cancel(sfport)
```

**解析器：** `RawDirective_acancel`

- 输出类型：JSON 文本。
- 结束标记：`<EOF>`。
- 注意：`sfport.CancelReplay()` 和 `helper.Cancel()` 当前走异步发送，不会解析输出；如果需要输出，请直接使用 `SendDirective("acancel")` 再通过解析器解析。

## 设备、网络与文件系统指令

### `sn`：获取序列号

**CLI：**

```text
sn
```

**Go API：**

```go
sn, err := helper.DeviceSN(sfport)
fmt.Println(sn.SN)
```

**输出：** JSON -> `mkpgo.SN`

```json
{"sn":"..."}
```

### `list_dir`：列出目录内容

**CLI：**

```text
list_dir <path>
```

**Go API：**

```go
fs, err := helper.ListDir(sfport, "/eMMC/applog/mkpdemo")
```

**输出：** JSON -> `mkpgo.FileSystem`

```json
{
  "rootDir": {
    "name": "applog",
    "path": "/eMMC/applog",
    "type": "directory",
    "size": 0,
    "contents": []
  },
  "error": ""
}
```

### `clean_dir`：清空目录

**CLI：**

```text
clean_dir <path>
```

**Go API：**

```go
err := helper.CleanDir(sfport, "/eMMC/applog/demo")
```

**限制：**

- `helper.CleanDir` 要求 `path` 以 `/eMMC/applog` 开头，否则返回 `only can clean directory in working directory`。
- 可使用 `helper.ComposeLogDirctory(logDir)` 将相对目录补成 `/eMMC/applog/<logDir>`。

**解析器：** `RawDirective_clean_dir`

- 输出类型：文本/空。
- 结束标记：`cli>`。
- 若输出包含 `Failed to`，返回 `failed to clean directory`。

### `delete_file`：删除日志文件

**CLI：**

```text
delete_file <path>
```

**Go API：**

```go
err := helper.DeleteFile(sfport, "demo/record1")
// helper 会转换为 /eMMC/applog/demo/record1.log
```

**路径规则：**

- `helper.DeleteFile` 会调用 `ComposeLogFullpath`：
  - 没有 `.log` 后缀时自动追加 `.log`。
  - 不是 `/eMMC/applog/` 开头时自动补成 `/eMMC/applog/<path>`。
- 转换后仍要求路径以 `/eMMC/applog` 开头。

**解析器：** `RawDirective_delete_file`

- 输出类型：文本/空。
- 结束标记：`cli>`。
- 若输出包含 `Failed to remove`，返回 `failed to remove file`。

### `join`：连接 Wi-Fi

**CLI：**

```text
join [ssid password]
```

调用形式：

```text
join wifi-name password
join
```

- 带 `ssid` / `password` 参数时，尝试连接指定 Wi-Fi。
- 无参数时，使用设备最近保存的 Wi-Fi 配置。

**Go API：**

```go
// 连接指定 Wi-Fi
out, err := helper.Join(sfport, &mkpgo.JoinOption{
    SSID:     "ssid",
    Password: "password1234",
})

// 使用最近保存的 Wi-Fi 配置
out, err = helper.Join(sfport, nil)
```

`Controller` 代理：

```go
out, err := ctrl.Join(&mkpgo.JoinOption{SSID: "ssid", Password: "password1234"})
out, err = ctrl.Join(nil)
```

**解析器：** `RawDirective_join`

- 输出类型：文本，非 JSON。
- 结束标记：`cli>`。
- 成功判断：输出包含 `connect: Connected`。
- 失败判断：输出包含 `Command returned non-zero error code` / `error code` 时返回 `ErrRawDirecitveExecutionFailed`。

成功输出示例：

```text
join ssid password1234
I (29664) connect: Connecting to 'ssid'
W (29664) wifi:Password length matches WPA2 standards, authmode threshold changes from OPEN to WPA2
I (31648) esp_netif_handlers: sta ip: 192.168.71.79, mask: 255.255.255.0, gw: 192.168.71.1
I (31648) connect: Connected
cli>
```

失败输出示例：

```text
join ssid password1234
I (8368) connect: Connecting to 'ssid'
W (18376) connect: Connection timed out
Command returned non-zero error code: 0x1 (ERROR)
cli>
```

无参数成功输出示例：

```text
join
I (736848) connect: Connecting to ''
ssid ChinaNet-9Wfg pass password1234
W (736856) wifi:Password length matches WPA2 standards, authmode threshold changes from OPEN to WPA2
W (736856) wifi:sta is connected, disconnect before connecting to new ap
I (736872) connect: Connected
cli>
```

### `wifi_auto`：查看/设置 Wi-Fi 自动连接

**CLI：**

```text
wifi_auto [0|1]
```

调用形式：

```text
wifi_auto
wifi_auto 1
wifi_auto 0
```

- 无参数时，查询自动连接状态。
- `state == "1"` 时启用启动自动连接。
- `state == "0"` 时关闭自动连接。

**Go API：**

```go
// 查询当前状态，返回 "on" 或 "off"
out, err := helper.WifiAuto(sfport, nil)

// 启用自动连接
out, err = helper.WifiAuto(sfport, &mkpgo.WifiAutoOption{State: "1"})

// 禁用自动连接
out, err = helper.WifiAuto(sfport, &mkpgo.WifiAutoOption{State: "0"})
```

`Controller` 代理：

```go
out, err := ctrl.WifiAuto(nil)
out, err = ctrl.WifiAuto(&mkpgo.WifiAutoOption{State: "1"})
```

**解析器：** `RawDirective_wifi_auto`

- 输出类型：文本，非 JSON。
- 结束标记：`cli>`。
- 查询成功判断：输出包含 `auto: on` 或 `auto: off`，分别返回 `"on"` / `"off"`。
- 设置成功判断：`wifi_auto 1` / `wifi_auto 0` 输出为空且正常到达 `cli>` 时返回空字符串与 `nil` error。
- 失败判断：输出包含 `Command returned non-zero error code` / `error code` 时返回 `ErrRawDirecitveExecutionFailed`。

查询输出示例：

```text
wifi_auto
auto: on

cli>
```

设置输出示例：

```text
wifi_auto 1
cli>
```

```text
wifi_auto 0
cli>
```

### `adumj`：日志转 JSON

**CLI：**

```text
adumj <logPath>
```

**Go API：**

```go
dump, err := helper.Adumj(sfport, &mkpgo.AdumjOption{LogPath: "demo-log"})
fmt.Println(dump.Format, dump.Version, len(dump.Events))
```

`Controller` 代理：

```go
dump, err := ctrl.Adumj(&mkpgo.AdumjOption{LogPath: "demo-log"})
```

**解析器：** `RawDirective_adumj`

- 输出类型：JSON。
- 结束标记：`cli>`；成功输出中的 JSON 后仍可能包含固件输出的 `<EOF>` 行，解析器会剔除该行。
- 失败判断：解析前 `PreFlight` 会将 `Command returned non-zero error code` 转为 `ErrRawDirecitveExecutionFailed`。
- 解析时会跳过前置日志行，提取第一个 `{` 到最后一个 `}` 之间的 JSON。

成功输出示例：

```text
adumj demo-log
I (922627) alog: logfile /eMMC/applog/demo-log.log
I (922635) alog: v2 format
{
  "format": "mkp-action-v1",
  "version": "MKv2",
  "meta": { "width": 1920, "height": 1080, "startX": 0, "startY": 0 },
  "events": [
    { "MouseMove": { "x": 1, "y": 2, "ts": 1064 } }
  ]
}
<EOF>
cli>
```

失败输出示例：

```text
adumj missing-log
E (1084411) alog: Failed to open file /eMMC/applog/missing-log.log
Command returned non-zero error code: 0xffffffff (ESP_FAIL)
cli>
```

### `ahttpbase`：文件管理服务器 API endpoint base URL

**CLI：**

```text
ahttpbase [url]
```

**Go API：**

```go
base, err := helper.AHTTPBase(sfport, nil) // 查询
base, err = helper.AHTTPBase(sfport, &mkpgo.AHTTPBaseOption{URL: "http://localhost:3000"}) // 设置
fmt.Println(base.AHTTPBase)
```

`Controller` 代理：

```go
base, err := ctrl.AHTTPBase(&mkpgo.AHTTPBaseOption{URL: "http://localhost:3000"})
```

**解析器：** `RawDirective_ahttpbase`

- 输出类型：JSON。
- 结束标记：`cli>`。
- 查询成功时提取固件返回的 `{ "ahttpbase": "..." }` JSON。
- 设置成功时固件返回 `OK`；解析器会按传入 URL 合成 JSON，方便 helper 统一反序列化为 `AHTTPBase`。
- 失败判断：解析前 `PreFlight` 会将 `Command returned non-zero error code` 转为 `ErrRawDirecitveExecutionFailed`。

查询输出示例：

```text
ahttpbase
W (954891) setupnvs: Error reading 'ahttpbase' from NVS: ESP_ERR_NVS_NOT_FOUND
{ "ahttpbase": "" }

<EOF>
cli>
```

设置输出示例：

```text
ahttpbase http://localhost:3000
OK

cli>
```
### `alive`：心跳检测

**CLI：**

```text
alive
```

**Go API：**

```go
hb, err := helper.Alive(sfport)
fmt.Println(hb.Timetamp)
```

**输出：** JSON -> `mkpgo.Heartbeat`

```json
{"timetamp": 1234567890}
```

> 字段名按当前代码/固件使用 `timetamp`。

### `atime`：获取日志时长

**CLI：**

```text
atime <path>
```

**Go API：**

```go
length, err := helper.Atime(sfport, "demo/record1")
length, err = helper.Atime(sfport, "/eMMC/applog/demo/record1.log")
```

**路径：** 代码注释说明支持相对路径（可不带 `.log`，如 `mkpdemo/1129f40`）或绝对路径（如 `/eMMC/applog/mkpdemo/1129f40.log`）。

**输出：** JSON -> `mkpgo.LogLength`

```json
{"seconds": 1, "milsec": 230}
```

### `aversion`：获取版本信息

**CLI：**

```text
aversion
```

**Go API：**

```go
version, err := helper.Aversion(sfport)
fmt.Println(version.UVersion, version.AVersion)
```

**输出：** JSON -> `mkpgo.MKPVersion`

```json
{"uver":"...", "aver":"..."}
```

### `ainsp`：获取日志基础信息

**CLI：**

```text
ainsp <path>
```

**Go API：**

```go
info, err := helper.AInspect(sfport, "demo/record1")
info, err = helper.AInspect(sfport, "/eMMC/applog/demo/record1.log")
```

**输出：** JSON -> `mkpgo.LogInfo`，由 `LogOption` 与 `LogLength` 组合：

```json
{
  "width": 1920,
  "heigh": 1080,
  "stpos": {"x": 0, "y": 0},
  "seconds": 1,
  "milsec": 230
}
```

## 鼠标指令：`m10`

### CLI

```text
m10 --port <port> [--b button] [--x dx] [--y dy] [--w wheel]
```

`SFSerialPort.Mouse10` 固定拼接：

```text
m10 --port <sp.MousePortFlag> ...
```

默认 `MousePortFlag = "1"`。

### 参数

| 参数 | 来源 | 范围/说明 |
|---|---|---|
| `--port <port>` | `sp.MousePortFlag` | 鼠标端口标记，默认 `1`。 |
| `--b <button>` | `M10Option.Button` | 鼠标按键 bitmask，低 5 bit。 |
| `--x <dx>` | `M10Option.X` | 相对 X 位移，代码注释范围 `-2048~2047`。 |
| `--y <dy>` | `M10Option.Y` | 相对 Y 位移，代码注释范围 `-2048~2047`。 |
| `--w <wheel>` | `M10Option.Wheel` | 滚轮位移，代码注释范围 `-128~127`。 |

### 按键 bitmask

| 名称 | 值 | 说明 |
|---|---:|---|
| `ReleaseMouseButton` | `0` | 释放/无按键 |
| `LeftMouseButton` | `1` | 左键 |
| `RightMouseButton` | `2` | 右键 |
| `BothLeftRightMouseButton` | `3` | 左右键同时按下 |
| `MiddleMouseButton` | `4` | 中键（代码注释范围 `[4-7]`） |
| `BackMouseButton` | `8` | 后退键（代码注释范围 `[8-9]`） |
| `FowardMouseButton` | `16` | 前进键（代码注释范围 `[16-31]`） |

### Go API

```go
// 默认异步发送
err := sfport.Mouse10(mkpgo.NewM10Option().SetX(100).SetY(0))

// 左键按下
err = sfport.Mouse10(mkpgo.NewM10Option().WithLeftButton())

// 释放全部鼠标按键
err = sfport.MouseReleaseAll()

// 同步发送，等待完成但忽略输出
err = sfport.Mouse10(mkpgo.NewM10Option().SetX(10).WithAsync(false))
// 或
err = sfport.Mouse10(mkpgo.NewM10Option().SetX(10).WithSyncIgnoreOutput(true))
```

### `M10Option` 构建方法索引

| 方法 | 说明 |
|---|---|
| `NewM10Option()` | 创建选项，默认 `Async=true`。 |
| `SetButton(v)` / `WithButton(v)` | 直接设置按钮 bitmask。 |
| `WithoutButton()` / `NoButton()` | 不输出 `--b`。 |
| `WithLeftButton()` | 设置左键。 |
| `WithRightButton()` | 设置右键。 |
| `WithBothLeftRightButton()` | 设置左右键。 |
| `WithMiddleButton()` | 设置中键。 |
| `WithBackButton()` | 设置后退键。 |
| `WithFowardButton()` | 设置前进键（代码中拼写为 `Foward`）。 |
| `SetX(v)` | 设置 `--x`。 |
| `SetY(v)` | 设置 `--y`。 |
| `SetWheel(v)` | 设置 `--w`。 |
| `WithAsync(async)` | 控制同步/异步；`false` 时等价于同步忽略输出。 |
| `WithSyncIgnoreOutput(syncIgnoreOutput)` | 控制同步发送并忽略输出。 |
| `Reset()` | 清空 Button/X/Y/Wheel。 |
| `ToString()` | 输出 CLI 参数片段。 |

## 键盘指令：`kpad`

### CLI

```text
kpad --port <port> [--s status] [--x1 key] [--x2 key] [--x3 key] [--x4 key] [--x5 key] [--x6 key] [--rel n] [--d n] [--v 1]
```

`SFSerialPort.Keypad` 固定拼接：

```text
kpad --port <sp.KeyboardPortFlag> ...
```

默认 `KeyboardPortFlag = "2"`。

### 参数

| 参数 | 来源 | 说明 |
|---|---|---|
| `--port <port>` | `sp.KeyboardPortFlag` | 键盘端口标记，默认 `2`。 |
| `--s <hex>` | `KpadOption.ModKeys.ToStatus()` | 修饰键状态字节，例如 `0x01` 表示左 Ctrl。 |
| `--x1` ~ `--x6` | `KpadOption.Keys[0..5]` | 最多 6 个普通键扫描码。 |
| `--rel <n>` | `KpadOption.Release` | 释放模式：`0` 按住；`1` 自动释放；`>1` 表示持续时间，单位 ms。 |
| `--d <n>` | `KpadOption.Delay` | 指令延迟值，代码注释为秒；常规默认 `0`。 |
| `--v 1` | `KpadOption.Verbose=true` | 输出详情。 |

### 修饰键

| 名称 | 状态位 |
|---|---:|
| `MOD_LCTRL` | `0x01` |
| `MOD_LSHIFT` | `0x02` |
| `MOD_LALT` | `0x04` |
| `MOD_LMETA` | `0x08` |
| `MOD_RCTRL` | `0x10` |
| `MOD_RSHIFT` | `0x20` |
| `MOD_RALT` | `0x40` |
| `MOD_RMETA` | `0x80` |

支持短别名：`LCTRL`、`LSHIFT`、`LALT`、`LMETA`、`RCTRL`、`RSHIFT`、`RALT`、`RMETA`。  
常用通用别名会规范化为左侧修饰键：`CTRL -> MOD_LCTRL`、`SHIFT -> MOD_LSHIFT`、`ALT -> MOD_LALT`、`META -> MOD_LMETA`。

### 普通键名

普通键名通过 `KeyNameToHexCode` 转换为 HID 扫描码。完整映射位于 `types_keycode_maps.go`，类别包括：

- 基础键：`NONE`、`ERR_OVF`、`A`~`Z`
- 数字键：`1`~`0`
- 控制键：`ENTER`、`ESC`、`BACKSPACE`、`TAB`、`SPACE` 等
- 功能键：`F1`~`F24`
- 系统/导航键：方向键、Insert/Delete/Home/End/PageUp/PageDown 等
- 小键盘键：`KP*` 系列
- 特殊功能键、应用控制键、国际键盘键、左右独立控制键、媒体控制键

可打印符号在 `KeyDown`/`KeyUp` 辅助逻辑中会展开为组合键，例如：

| 输入 | 展开 |
|---|---|
| `!` | `MOD_LSHIFT` + `1` |
| `@` | `MOD_LSHIFT` + `2` |
| `_` | `MOD_LSHIFT` + `MINUS` |
| `+` | `MOD_LSHIFT` + `EQUAL` |
| 空格 | `SPACE` |
| `[` / `]` | `LEFTBRACE` / `RIGHTBRACE` |

### Go API

```go
// 单次点击 A（自动释放）
err := sfport.Keypad(mkpgo.NewKpadOption().WithKeys([]string{"A"}).WithAutoRelease())

// 按住 W
err = sfport.Keypad(mkpgo.NewKpadOption().WithKeys([]string{"W"}).WithHold())

// Ctrl + C
err = sfport.Keypad(mkpgo.NewKpadOption().WithKeys([]string{"CTRL", "C"}).WithAutoRelease())

// 释放当前槽位
err = sfport.Keypad(mkpgo.HidKpadRelease)

// 全释放
err = sfport.Keypad(mkpgo.HidKpadReleaseAll)
```

### `KpadOption` 构建方法索引

| 方法 | 说明 |
|---|---|
| `NewKpadOption()` | 创建选项，默认 `Release=0`、`Delay=0`、`Verbose=false`、`Async=true`。 |
| `WithKeys(keys)` | 设置最多 6 个键；自动拆分修饰键与普通键。 |
| `WithKey(key)` | 将单键写入第一个键槽。 |
| `WithModKeys(modKeys)` | 直接设置修饰键集合。 |
| `WithDelay(delay)` | 设置 `--d`。 |
| `WithRelease(release)` | 直接设置 `--rel`。 |
| `WithHold()` | 设置 `Release=0`，表示按住。 |
| `WithAutoRelease()` | 设置 `Release=1`，表示自动释放。 |
| `WithDuration(duration)` | `duration > 1` 时设置 `Release=duration`。 |
| `WithVerbose(verbose)` | 控制 `--v 1`。 |
| `WithAsync(async)` | 控制同步/异步；`false` 时等价于同步忽略输出。 |
| `WithSyncIgnoreOutput(syncIgnoreOutput)` | 控制同步发送并忽略输出。 |
| `KeyDown(key)` | 基于本地缓存生成“按下并保持”包，并在发送成功后提交缓存。 |
| `KeyUp(key)` | 生成“释放目标键”和“保持剩余键”两个包。 |
| `ToString()` | 输出 CLI 参数片段。 |

### 预置释放选项

| 变量 | 等价配置 | 说明 |
|---|---|---|
| `mkpgo.HidKpadRelease` | `NewKpadOption().WithDelay(0).WithKey("NONE")` | 将当前键位槽释放为 `NONE`。 |
| `mkpgo.HidKpadReleaseAll` | `NewKpadOption().WithDelay(0).WithRelease(0)` | 发送全释放键盘包。 |

## 输出解析器索引

| 指令 | 解析器 | JSON | EOF | 解析结果 |
|---|---|---:|---|---|
| `alog` | `RawDirective_alog` | 否 | `cli>` | 文本；可为空。 |
| `astop` | `RawDirective_astop` | 否 | `cli>` | 文本；空内容视为解析失败。 |
| `acancel` | `RawDirective_acancel` | 是 | `<EOF>` | JSON 文本；空内容视为解析失败。 |
| `sn` | `RawDirective_sn` | 是 | `<EOF>` | JSON 文本 -> `SN`。 |
| `list_dir` | `RawDirective_list_dir` | 是 | `<EOF>` | JSON 文本 -> `FileSystem`；空内容返回空字符串。 |
| `clean_dir` | `RawDirective_clean_dir` | 否 | `cli>` | 成功空返回；包含 `Failed to` 时返回错误。 |
| `delete_file` | `RawDirective_delete_file` | 否 | `cli>` | 成功空返回；包含 `Failed to remove` 时返回错误。 |
| `join` | `RawDirective_join` | 否 | `cli>` | 文本；成功需包含 `connect: Connected`，错误码输出返回执行失败。 |
| `wifi_auto` | `RawDirective_wifi_auto` | 否 | `cli>` | 文本；查询返回 `on` / `off`；设置成功返回空字符串。 |
| `adumj` | `RawDirective_adumj` | 是 | `cli>` | 提取 JSON 文本 -> `ActionDump`；成功输出中的 `<EOF>` 作为内容被解析器剔除。 |
| `ahttpbase` | `RawDirective_ahttpbase` | 是 | `cli>` | 查询时提取 JSON -> `AHTTPBase`；设置成功 `OK` 时合成 JSON。 |
| `alive` | `RawDirective_alive` | 是 | `<EOF>` | JSON 文本 -> `Heartbeat`。 |
| `atime` | `RawDirective_atime` | 是 | `<EOF>` | 查找包含 `"seconds"` 的 JSON 行 -> `LogLength`。 |
| `aversion` | `RawDirective_aversion` | 是 | `<EOF>` | JSON 文本 -> `MKPVersion`。 |
| `ainsp` | `RawDirective_ainsp` | 是 | `<EOF>` | 查找同时包含 `"seconds"` 和 `"width"` 的 JSON 行 -> `LogInfo`。 |
| `aplay` | 无 | - | 默认 `cli>` | API 默认异步。 |
| `m10` | 无 | - | 默认 `cli>` | 同步模式通常忽略输出。 |
| `kpad` | 无 | - | 默认 `cli>` | 同步模式通常忽略输出。 |

通用错误规则：所有解析器会先执行 `PreFlight`，若输出中包含 `command returned non-zero error code`（大小写不敏感），返回 `ErrRawDirecitveExecutionFailed`。

## 返回结构索引

```go
type SN struct {
    SN string `json:"sn"`
}

type FileSystem struct {
    RootDir FileNode `json:"rootDir,omitempty"`
    Error   string   `json:"error,omitempty"`
}

type FileNode struct {
    DisplayName string     `json:"displayName,omitempty"`
    Name        string     `json:"name"`
    Path        string     `json:"path"`
    Type        string     `json:"type"` // directory/file
    Size        int        `json:"size"`
    Contents    []FileNode `json:"contents,omitempty"`
}

type Heartbeat struct {
    Timetamp int64 `json:"timetamp"`
}

type LogLength struct {
    Seconds      int `json:"seconds"`
    Milliseconds int `json:"milsec"`
}

type MKPVersion struct {
    UVersion string `json:"uver"`
    AVersion string `json:"aver"`
}

type LogOption struct {
    Width  int `json:"width"`
    Height int `json:"heigh"`
    StPos  struct {
        X int `json:"x"`
        Y int `json:"y"`
    } `json:"stpos"`
}

type JoinOption struct {
    SSID     string `json:"ssid"`
    Password string `json:"password"`
}

type WifiAutoOption struct {
    State string `json:"state"`
}

type AHTTPBaseOption struct {
    URL string `json:"url"`
}

type AHTTPBase struct {
    AHTTPBase string `json:"ahttpbase"`
}

type AdumjOption struct {
    LogPath string `json:"logPath"`
}

type ActionDump struct {
    Format  string                       `json:"format"`
    Version string                       `json:"version"`
    Meta    ActionDumpMeta               `json:"meta"`
    Events  []map[string]json.RawMessage `json:"events"`
}

type ActionDumpMeta struct {
    Width  int `json:"width"`
    Height int `json:"height"`
    StartX int `json:"startX"`
    StartY int `json:"startY"`
}

type LogInfo struct {
    LogOption
    LogLength
}
```

## Go API 索引

### `SFSerialPort` 指令方法

| 方法 | 对应指令 | 说明 |
|---|---|---|
| `SendDirective` / `SendDirectiveContext` / `SendSyncDirective` / `SendSyncDirectiveContext` | 任意 | 同步发送并返回原始输出；可选 `DirectiveOption` 覆盖本次同步等待。 |
| `SendDirectiveIgnoreOutput` / `SendDirectiveIgnoreOutputContext` | 任意 | 同步发送，等待完成但忽略输出；可选 `DirectiveOption` 覆盖本次同步等待。 |
| `SendDirectiveAsync` / `SendDirectiveAsyncContext` | 任意 | 异步发送。 |
| `StartRecording` / `StartRecordingContext` | `alog` | 异步开始录制。 |
| `StartReplaying` / `StartReplayingContext` | `aplay` | 异步开始回放。 |
| `StopRecording` / `StopRecordingContext` | `astop` | 异步停止录制。 |
| `Stop` / `StopContext` | `astop` | `StopRecording` 别名。 |
| `CancelReplay` / `CancelReplayContext` | `acancel` | 异步取消回放。 |
| `Mouse10` / `Mouse10Context` | `m10` | 发送鼠标指令；`Async=false` 时同步忽略输出。 |
| `MouseReleaseAll` / `MouseReleaseAllContext` | `m10 --b 0` | 释放全部鼠标按键。 |
| `Keypad` / `KeypadContext` | `kpad` | 发送键盘指令；`Async=false` 时同步忽略输出。 |

### `helper` 包方法

需要同步输出的 helper（如 `Alog`、`Astop`、`Join`、`WifiAuto`、`AHTTPBase`、`Adumj`、`DeviceSN`、`ListDir`、`CleanDir`、`DeleteFile`、`Alive`、`Atime`、`Aversion`、`AInspect`）均支持可选 `mkpgo.DirectiveOption`，可用 `mkpgo.WithSyncOutputTimeout(...)` 覆盖本次等待超时。

| 方法 | 对应指令 | 返回 |
|---|---|---|
| `StartRecord` / `StartRecordContext` | `alog` | `error` |
| `Alog` / `AlogContext` | `alog` | `string, error` |
| `StopRecord` / `StopRecordContext` | `astop` | `error` |
| `Astop` / `AstopContext` | `astop` | `error` |
| `Cancel` / `CancelContext` | `acancel` | `error` |
| `DeviceSN` / `DeviceSNContext` | `sn` | `*SN, error` |
| `ListDir` / `ListDirContext` | `list_dir` | `*FileSystem, error` |
| `CleanDir` / `CleanDirContext` | `clean_dir` | `error` |
| `DeleteFile` / `DeleteFileContext` | `delete_file` | `error` |
| `Join` / `JoinContext` | `join` | `string, error` |
| `WifiAuto` / `WifiAutoContext` | `wifi_auto` | `string, error` |
| `AHTTPBase` / `AHTTPBaseContext` | `ahttpbase` | `*AHTTPBase, error` |
| `Adumj` / `AdumjContext` | `adumj` | `*ActionDump, error` |
| `Alive` / `AliveContext` | `alive` | `*Heartbeat, error` |
| `Atime` / `AtimeContext` | `atime` | `*LogLength, error` |
| `Aversion` / `AversionContext` | `aversion` | `*MKPVersion, error` |
| `AInspect` / `AInspectContext` | `ainsp` | `*LogInfo, error` |
| `KeyDown` / `KeyDownContext` | `kpad` | `error` |
| `KeyUp` / `KeyUpContext` | `kpad` | `error` |
| `KeyTap` / `KeyTapContext` | `kpad` | `error` |
| `KeyPresses` / `KeyPressesContext` | `kpad` | `error` |
| `KeypadRelease` / `KeypadReleaseContext` | `kpad` | `error` |
| `KeypadReleaseAll` / `KeypadReleaseAllContext` | `kpad` | `error` |
| `MouseReleaseAll` / `MouseReleaseAllContext` | `m10` | `error` |
| `M10` | `m10` | `error` |

### `controller.Controller` 常用代理

`Controller` 中对应的同步输出代理方法同样支持可选 `mkpgo.DirectiveOption`。

`controller.Controller` 封装了 helper 和键鼠高级操作：

- 录制/回放：`StartRecord`、`StopRecord`、`Alog`、`Astop`、`Cancel`
- 设备/文件/网络：`DeviceSN`、`ListDir`、`CleanDir`、`DeleteFile`、`Join`、`WifiAuto`、`AHTTPBase`、`Adumj`、`Alive`、`Atime`、`Aversion`、`AInspect`
- 键盘：`KeyDown`、`KeyUp`、`KeyTap`、`KeyPresses`、`KeypadRelease`、`KeypadReleaseAll`
- 鼠标：`MouseClick`、`MouseClickWithOption`、`MouseScroll`、`MouseScrollWithOption`、`MouseScrollWithButton`、`MouseDown`、`MouseReleaseAll`、`MouseUp`、`M10Move`、`MouseMove`

## 草案/未封装指令

`mmm/mkp_directive260417.md` 中记录了两个建议指令；当前主代码中未注册解析器，也没有正式 Go wrapper。

### `apause`：暂停回放

建议含义：暂停正在回放的 log。

建议参数：

1. `logName`：可选；为空表示当前正在回放的 log。
2. `paused_at`：可选；暂停到指定时间点，例如 `1.22 s`。若固件机制不方便实现，则收到指令时立即暂停。

备注：暂停时保存 keyboard/mouse 状态，便于后续 resume。

### `aresume`：恢复回放

建议含义：恢复 log 回放。

建议参数：

1. `logName`：可选；为空表示当前正在回放的 log。
2. `useStates`：`1/0`；默认 `1`。`1` 表示恢复前使用 `apause` 保存的键鼠状态；`0` 表示直接恢复回放。

备注：`apause` 保存的状态在 `alog`、`astop`、`acancel`、`aresume` 后直接清除。
