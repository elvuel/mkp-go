# MKP Command Reference

English | [中文](./directives.md)

---

> This document is based on the current repository code: `sf_serial_port.go`, `directive_parser.go`, `types_cli.go`, `types_m10.go`, `types_keycode.go`, and `helper/helper.go`.  
> It provides a quick reference for **MKP CLI commands / Go APIs / options / output parsers**.

## Table of Contents

- [Conventions](#conventions)
- [Quick Overview](#quick-overview)
- [Synchronous and Asynchronous Sending](#synchronous-and-asynchronous-sending)
- [Recording and Replay Commands](#recording-and-replay-commands)
- [Device, Network, and Filesystem Commands](#device-network-and-filesystem-commands)
- [Mouse Command: m10](#mouse-command-m10)
- [Keyboard Command: kpad](#keyboard-command-kpad)
- [Output Parser Index](#output-parser-index)
- [Response Model Index](#response-model-index)
- [Go API Index](#go-api-index)
- [Draft / Unwrapped Commands](#draft--unwrapped-commands)

## Conventions

- Default serial type: `*mkpgo.SFSerialPort`, created by `mkpgo.NewSFSerialPort()`.
- Default port flags: mouse `MousePortFlag = "1"`, keyboard `KeyboardPortFlag = "2"`.
- All low-level commands are written to the serial port with `\r\n` appended.
- The synchronous output switch is named `SyncOuputEnabled` in code. `NewSFSerialPort()` sets it to `true` by default.
- When using synchronous output, start the read loop first: `go sfport.Read()`.
- Built-in parsers are registered by `mkpgo.InitParsers()`; `init.go` calls it automatically.

Minimal example:

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

## Quick Overview

| Command | CLI Form | Description | Go API | Output / Parser |
|---|---|---|---|---|
| `alog` | `alog [logName] [options...]` | Start device-side recording; synchronous send can wait for recording output | `StartRecording` / `helper.StartRecord` / `helper.Alog` | Text, parser `RawDirective_alog` |
| `aplay` | `aplay [logName] [--delay ms]` | Start replaying a recorded log | `StartReplaying` | No built-in parser; usually async |
| `astop` | `astop` | Stop current recording | `StopRecording` / `Stop` / `helper.Astop` | Text, parser `RawDirective_astop` |
| `acancel` | `acancel` | Cancel current replay task | `CancelReplay` / `helper.Cancel` | JSON parser exists, but wrapper sends async by default |
| `sn` | `sn` | Get device serial number | `helper.DeviceSN` | JSON -> `SN` |
| `list_dir` | `list_dir <path>` | List directory contents | `helper.ListDir` | JSON -> `FileSystem` |
| `clean_dir` | `clean_dir <path>` | Clean a directory | `helper.CleanDir` | Empty on success; error on failure |
| `delete_file` | `delete_file <path>` | Delete a file | `helper.DeleteFile` | Empty on success; error on failure |
| `join` | `join [ssid password]` | Connect Wi-Fi; without args, use the most recently saved config | `helper.Join` / `Controller.Join` | Text parser `RawDirective_join`; success contains `connect: Connected` |
| `alive` | `alive` | Heartbeat / liveness check | `helper.Alive` | JSON -> `Heartbeat` |
| `atime` | `atime <path>` | Get log duration | `helper.Atime` | JSON -> `LogLength` |
| `aversion` | `aversion` | Get firmware/application version | `helper.Aversion` | JSON -> `MKPVersion` |
| `ainsp` | `ainsp <path>` | Inspect basic log metadata | `helper.AInspect` | JSON -> `LogInfo` |
| `m10` | `m10 --port <port> [--b n] [--x n] [--y n] [--w n]` | Mouse control | `Mouse10` / `helper.M10` / controller mouse methods | No built-in parser; sync mode may ignore output |
| `kpad` | `kpad --port <port> [--s hex] [--x1 hex]... [--rel n] [--d n] [--v 1]` | Keyboard control | `Keypad` / helper/controller keyboard methods | No built-in parser; sync mode may ignore output |

## Synchronous and Asynchronous Sending

### Low-level send APIs

| API | Behavior |
|---|---|
| `SendDirective(directive[, opts...])` / `SendDirectiveContext(ctx, directive[, opts...])` / `SendSyncDirective(...)` | Synchronous send. If `SyncOuputEnabled=true`, it waits until `Read()` captures the command completion marker and returns raw output; optional `DirectiveOption` values override this wait. |
| `SendDirectiveIgnoreOutput(directive[, opts...])` / `SendDirectiveIgnoreOutputContext(ctx, directive[, opts...])` | Synchronous send, but only waits for completion and returns no output; optional `DirectiveOption` values override this wait. |
| `SendDirectiveAsync(directive)` / `SendDirectiveAsyncContext(ctx, directive)` | Asynchronous send. It writes the command and does not wait for output. |

### EOF markers

- Registered commands use their parser's `EOFFlag()`.
- Unregistered commands use CLI prompt `cli>` as the default completion marker.
- Current constants:
  - `EOFDefault = "<EOF>"`
  - `EOFCLI = "cli>"`
- `alog` synchronous matching is normalized to `alog`, because actual output may not start with the full CLI text.

### Sync output timeout

- The default timeout is controlled by `SFSerialPort.SyncOutputTimeout`; `NewSFSerialPort()` defaults it to `10 * time.Second`.
- `SendDirective` / `SendDirectiveContext` / `SendSyncDirective` / `SendSyncDirectiveContext` accept optional `DirectiveOption` values, such as `WithSyncOutputTimeout(timeout)`, that only override the current synchronous wait.
- If `WithSyncOutputTimeout` is omitted, the default `SyncOutputTimeout` is used.
- `WithSyncOutputTimeout(0)` disables the timer for this wait, so only `context` cancellation can stop it.

```go
out, err := sfport.SendSyncDirective("join ssid password", mkpgo.WithSyncOutputTimeout(30*time.Second))
out, err = sfport.SendDirectiveContext(ctx, "alive") // uses default SyncOutputTimeout
```

### Recommendations

- For commands that return structured results, use synchronous helpers such as `helper.DeviceSN` and `helper.ListDir`.
- Mouse/keyboard real-time control defaults to async (`M10Option.Async=true`, `KpadOption.Async=true`).
- To enforce command ordering for mouse/keyboard operations, set `WithAsync(false)`; the current implementation waits for completion while ignoring output.

## Recording and Replay Commands

### `alog`: start recording / get recording output

**CLI:**

```text
alog [logName] [--width n] [--heigh n] [--stposx n] [--stposy n]
```

> `LogOption.CliArgs()` uses `--heigh` in the current code, not `--height`.

**Options:**

| Option | Description |
|---|---|
| `logName` | Log name; optional. The helper appends it as the first argument. |
| `--width n` | Recording region width. Emitted when `LogOption.Width > 0`. |
| `--heigh n` | Recording region height. Emitted when `LogOption.Height > 0`. |
| `--stposx n` | Start X. Emitted when `LogOption.StPos.X > -1`. |
| `--stposy n` | Start Y. Emitted when `LogOption.StPos.Y > -1`. |

**Go API:**

```go
// Start recording asynchronously.
err := sfport.StartRecording("demo")
err = helper.StartRecord(sfport, "demo", opt)

// Send alog synchronously and wait for output.
out, err := helper.Alog(sfport, "demo", opt)
```

**Parser:** `RawDirective_alog`

- Output type: text.
- EOF marker: `cli>`.
- Parsing behavior: removes `\r`, trims CLI prefix, returns text when non-empty; otherwise returns an empty string.

### `aplay`: start replay

**CLI:**

```text
aplay [logName] [--delay ms]
```

**Options:**

| Option | Description |
|---|---|
| `logName` | Log name to replay. Not appended when empty. |
| `--delay ms` | Replay delay. Appended when `delay >= 0`. |

**Go API:**

```go
err := sfport.StartReplaying("demo", 0)
err = sfport.StartReplayingContext(ctx, "demo", 100)
```

**Parser:** no built-in parser. `StartReplaying` sends asynchronously by default.

### `astop`: stop recording

**CLI:**

```text
astop
```

**Go API:**

```go
err := sfport.StopRecording()
err = sfport.Stop()
err = helper.StopRecord(sfport)
err = helper.Astop(sfport)
```

**Parser:** `RawDirective_astop`

- Output type: text.
- EOF marker: `cli>`.
- Parsing behavior: returns text when non-empty; empty content is treated as `ErrRawDirectiveParseFailed`.

### `acancel`: cancel replay

**CLI:**

```text
acancel
```

**Go API:**

```go
err := sfport.CancelReplay()
err = helper.Cancel(sfport)
```

**Parser:** `RawDirective_acancel`

- Output type: JSON text.
- EOF marker: `<EOF>`.
- Note: `sfport.CancelReplay()` and `helper.Cancel()` currently send asynchronously and do not parse output. If output is required, call `SendDirective("acancel")` directly and then parse it with the registered parser.

## Device, Network, and Filesystem Commands

### `sn`: get serial number

**CLI:**

```text
sn
```

**Go API:**

```go
sn, err := helper.DeviceSN(sfport)
fmt.Println(sn.SN)
```

**Output:** JSON -> `mkpgo.SN`

```json
{"sn":"..."}
```

### `list_dir`: list directory contents

**CLI:**

```text
list_dir <path>
```

**Go API:**

```go
fs, err := helper.ListDir(sfport, "/eMMC/applog/mkpdemo")
```

**Output:** JSON -> `mkpgo.FileSystem`

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

### `clean_dir`: clean a directory

**CLI:**

```text
clean_dir <path>
```

**Go API:**

```go
err := helper.CleanDir(sfport, "/eMMC/applog/demo")
```

**Restrictions:**

- `helper.CleanDir` requires `path` to start with `/eMMC/applog`; otherwise it returns `only can clean directory in working directory`.
- Use `helper.ComposeLogDirctory(logDir)` to convert a relative directory to `/eMMC/applog/<logDir>`.

**Parser:** `RawDirective_clean_dir`

- Output type: text/empty.
- EOF marker: `cli>`.
- If output contains `Failed to`, the parser returns `failed to clean directory`.

### `delete_file`: delete a log file

**CLI:**

```text
delete_file <path>
```

**Go API:**

```go
err := helper.DeleteFile(sfport, "demo/record1")
// helper converts it to /eMMC/applog/demo/record1.log
```

**Path rules:**

- `helper.DeleteFile` calls `ComposeLogFullpath`:
  - appends `.log` when missing;
  - prepends `/eMMC/applog/` when the path does not already start with `/eMMC/applog/`.
- The converted path must still start with `/eMMC/applog`.

**Parser:** `RawDirective_delete_file`

- Output type: text/empty.
- EOF marker: `cli>`.
- If output contains `Failed to remove`, the parser returns `failed to remove file`.

### `join`: connect Wi-Fi

**CLI:**

```text
join [ssid password]
```

Call forms:

```text
join wifi-name password
join
```

- With `ssid` / `password`, the device attempts to connect to the specified Wi-Fi network.
- Without arguments, the device uses the most recently saved Wi-Fi configuration.

**Go API:**

```go
// Connect to a specified Wi-Fi network.
out, err := helper.Join(sfport, &mkpgo.JoinOption{
    SSID:     "ssid",
    Password: "password1234",
})

// Use the most recently saved Wi-Fi configuration.
out, err = helper.Join(sfport, nil)
```

`Controller` proxy:

```go
out, err := ctrl.Join(&mkpgo.JoinOption{SSID: "ssid", Password: "password1234"})
out, err = ctrl.Join(nil)
```

**Parser:** `RawDirective_join`

- Output type: text, not JSON.
- EOF marker: `cli>`.
- Success rule: output contains `connect: Connected`.
- Failure rule: output containing `Command returned non-zero error code` / `error code` returns `ErrRawDirecitveExecutionFailed`.

Successful output example:

```text
join ssid password1234
I (29664) connect: Connecting to 'ssid'
W (29664) wifi:Password length matches WPA2 standards, authmode threshold changes from OPEN to WPA2
I (31648) esp_netif_handlers: sta ip: 192.168.71.79, mask: 255.255.255.0, gw: 192.168.71.1
I (31648) connect: Connected
cli>
```

Failure output example:

```text
join ssid password1234
I (8368) connect: Connecting to 'ssid'
W (18376) connect: Connection timed out
Command returned non-zero error code: 0x1 (ERROR)
cli>
```

No-argument successful output example:

```text
join
I (736848) connect: Connecting to ''
ssid ChinaNet-9Wfg pass password1234
W (736856) wifi:Password length matches WPA2 standards, authmode threshold changes from OPEN to WPA2
W (736856) wifi:sta is connected, disconnect before connecting to new ap
I (736872) connect: Connected
cli>
```

### `alive`: heartbeat check

**CLI:**

```text
alive
```

**Go API:**

```go
hb, err := helper.Alive(sfport)
fmt.Println(hb.Timetamp)
```

**Output:** JSON -> `mkpgo.Heartbeat`

```json
{"timetamp": 1234567890}
```

> The field name is `timetamp` in the current code/firmware payload.

### `atime`: get log duration

**CLI:**

```text
atime <path>
```

**Go API:**

```go
length, err := helper.Atime(sfport, "demo/record1")
length, err = helper.Atime(sfport, "/eMMC/applog/demo/record1.log")
```

**Path:** code comments indicate that both relative paths (optionally without `.log`, e.g. `mkpdemo/1129f40`) and absolute paths (e.g. `/eMMC/applog/mkpdemo/1129f40.log`) are supported.

**Output:** JSON -> `mkpgo.LogLength`

```json
{"seconds": 1, "milsec": 230}
```

### `aversion`: get version information

**CLI:**

```text
aversion
```

**Go API:**

```go
version, err := helper.Aversion(sfport)
fmt.Println(version.UVersion, version.AVersion)
```

**Output:** JSON -> `mkpgo.MKPVersion`

```json
{"uver":"...", "aver":"..."}
```

### `ainsp`: inspect basic log information

**CLI:**

```text
ainsp <path>
```

**Go API:**

```go
info, err := helper.AInspect(sfport, "demo/record1")
info, err = helper.AInspect(sfport, "/eMMC/applog/demo/record1.log")
```

**Output:** JSON -> `mkpgo.LogInfo`, a combination of `LogOption` and `LogLength`:

```json
{
  "width": 1920,
  "heigh": 1080,
  "stpos": {"x": 0, "y": 0},
  "seconds": 1,
  "milsec": 230
}
```

## Mouse Command: `m10`

### CLI

```text
m10 --port <port> [--b button] [--x dx] [--y dy] [--w wheel]
```

`SFSerialPort.Mouse10` always prefixes:

```text
m10 --port <sp.MousePortFlag> ...
```

Default `MousePortFlag = "1"`.

### Options

| Option | Source | Range / Description |
|---|---|---|
| `--port <port>` | `sp.MousePortFlag` | Mouse port flag, default `1`. |
| `--b <button>` | `M10Option.Button` | Mouse button bitmask, lower 5 bits. |
| `--x <dx>` | `M10Option.X` | Relative X movement. Code comment range: `-2048~2047`. |
| `--y <dy>` | `M10Option.Y` | Relative Y movement. Code comment range: `-2048~2047`. |
| `--w <wheel>` | `M10Option.Wheel` | Wheel delta. Code comment range: `-128~127`. |

### Button bitmask

| Name | Value | Description |
|---|---:|---|
| `ReleaseMouseButton` | `0` | Release / no button |
| `LeftMouseButton` | `1` | Left button |
| `RightMouseButton` | `2` | Right button |
| `BothLeftRightMouseButton` | `3` | Left + right buttons |
| `MiddleMouseButton` | `4` | Middle button; code comment range `[4-7]` |
| `BackMouseButton` | `8` | Back button; code comment range `[8-9]` |
| `FowardMouseButton` | `16` | Forward button; code comment range `[16-31]` |

### Go API

```go
// Default: async send.
err := sfport.Mouse10(mkpgo.NewM10Option().SetX(100).SetY(0))

// Press left button.
err = sfport.Mouse10(mkpgo.NewM10Option().WithLeftButton())

// Release all mouse buttons.
err = sfport.MouseReleaseAll()

// Synchronous send, wait for completion but ignore output.
err = sfport.Mouse10(mkpgo.NewM10Option().SetX(10).WithAsync(false))
// or:
err = sfport.Mouse10(mkpgo.NewM10Option().SetX(10).WithSyncIgnoreOutput(true))
```

### `M10Option` builder methods

| Method | Description |
|---|---|
| `NewM10Option()` | Create an option object; default `Async=true`. |
| `SetButton(v)` / `WithButton(v)` | Set the button bitmask directly. |
| `WithoutButton()` / `NoButton()` | Do not emit `--b`. |
| `WithLeftButton()` | Set left button. |
| `WithRightButton()` | Set right button. |
| `WithBothLeftRightButton()` | Set left + right buttons. |
| `WithMiddleButton()` | Set middle button. |
| `WithBackButton()` | Set back button. |
| `WithFowardButton()` | Set forward button. The code name uses `Foward`. |
| `SetX(v)` | Set `--x`. |
| `SetY(v)` | Set `--y`. |
| `SetWheel(v)` | Set `--w`. |
| `WithAsync(async)` | Control sync/async behavior; `false` is equivalent to sync-ignore-output mode. |
| `WithSyncIgnoreOutput(syncIgnoreOutput)` | Control synchronous sending while ignoring output. |
| `Reset()` | Clear Button/X/Y/Wheel fields. |
| `ToString()` | Build the CLI argument fragment. |

## Keyboard Command: `kpad`

### CLI

```text
kpad --port <port> [--s status] [--x1 key] [--x2 key] [--x3 key] [--x4 key] [--x5 key] [--x6 key] [--rel n] [--d n] [--v 1]
```

`SFSerialPort.Keypad` always prefixes:

```text
kpad --port <sp.KeyboardPortFlag> ...
```

Default `KeyboardPortFlag = "2"`.

### Options

| Option | Source | Description |
|---|---|---|
| `--port <port>` | `sp.KeyboardPortFlag` | Keyboard port flag, default `2`. |
| `--s <hex>` | `KpadOption.ModKeys.ToStatus()` | Modifier-key status byte, for example `0x01` for left Ctrl. |
| `--x1` ~ `--x6` | `KpadOption.Keys[0..5]` | Up to 6 normal key scan codes. |
| `--rel <n>` | `KpadOption.Release` | Release mode: `0` hold; `1` auto-release; `>1` duration in ms. |
| `--d <n>` | `KpadOption.Delay` | Command delay value. Code comment says seconds; normal default is `0`. |
| `--v 1` | `KpadOption.Verbose=true` | Enable verbose output. |

### Modifier keys

| Name | Status bit |
|---|---:|
| `MOD_LCTRL` | `0x01` |
| `MOD_LSHIFT` | `0x02` |
| `MOD_LALT` | `0x04` |
| `MOD_LMETA` | `0x08` |
| `MOD_RCTRL` | `0x10` |
| `MOD_RSHIFT` | `0x20` |
| `MOD_RALT` | `0x40` |
| `MOD_RMETA` | `0x80` |

Supported short aliases: `LCTRL`, `LSHIFT`, `LALT`, `LMETA`, `RCTRL`, `RSHIFT`, `RALT`, `RMETA`.  
Common generic aliases are normalized to left-side modifiers: `CTRL -> MOD_LCTRL`, `SHIFT -> MOD_LSHIFT`, `ALT -> MOD_LALT`, `META -> MOD_LMETA`.

### Normal key names

Normal key names are converted to HID scan codes through `KeyNameToHexCode`. The full mapping is in `types_keycode_maps.go`; categories include:

- Basic keys: `NONE`, `ERR_OVF`, `A`~`Z`
- Number keys: `1`~`0`
- Control keys: `ENTER`, `ESC`, `BACKSPACE`, `TAB`, `SPACE`, etc.
- Function keys: `F1`~`F24`
- System/navigation keys: arrows, Insert/Delete/Home/End/PageUp/PageDown, etc.
- Keypad keys: `KP*` series
- Special function keys, application control keys, international keyboard keys, independent left/right control keys, and media control keys

Printable symbols are expanded by `KeyDown` / `KeyUp` helper logic. Examples:

| Input | Expansion |
|---|---|
| `!` | `MOD_LSHIFT` + `1` |
| `@` | `MOD_LSHIFT` + `2` |
| `_` | `MOD_LSHIFT` + `MINUS` |
| `+` | `MOD_LSHIFT` + `EQUAL` |
| space | `SPACE` |
| `[` / `]` | `LEFTBRACE` / `RIGHTBRACE` |

### Go API

```go
// Tap A once with auto release.
err := sfport.Keypad(mkpgo.NewKpadOption().WithKeys([]string{"A"}).WithAutoRelease())

// Hold W.
err = sfport.Keypad(mkpgo.NewKpadOption().WithKeys([]string{"W"}).WithHold())

// Ctrl + C.
err = sfport.Keypad(mkpgo.NewKpadOption().WithKeys([]string{"CTRL", "C"}).WithAutoRelease())

// Release current slots.
err = sfport.Keypad(mkpgo.HidKpadRelease)

// Full release.
err = sfport.Keypad(mkpgo.HidKpadReleaseAll)
```

### `KpadOption` builder methods

| Method | Description |
|---|---|
| `NewKpadOption()` | Create an option object; defaults: `Release=0`, `Delay=0`, `Verbose=false`, `Async=true`. |
| `WithKeys(keys)` | Set up to 6 keys; automatically splits modifier keys and normal keys. |
| `WithKey(key)` | Put one key into the first key slot. |
| `WithModKeys(modKeys)` | Set modifier-key collection directly. |
| `WithDelay(delay)` | Set `--d`. |
| `WithRelease(release)` | Set `--rel` directly. |
| `WithHold()` | Set `Release=0`, meaning hold. |
| `WithAutoRelease()` | Set `Release=1`, meaning auto-release. |
| `WithDuration(duration)` | If `duration > 1`, set `Release=duration`. |
| `WithVerbose(verbose)` | Control `--v 1`. |
| `WithAsync(async)` | Control sync/async behavior; `false` is equivalent to sync-ignore-output mode. |
| `WithSyncIgnoreOutput(syncIgnoreOutput)` | Control synchronous sending while ignoring output. |
| `KeyDown(key)` | Build a hold packet from the current local cache plus the newly pressed key; commit cache after successful send. |
| `KeyUp(key)` | Build two packets: one to release the target key and one to keep remaining keys held. |
| `ToString()` | Build the CLI argument fragment. |

### Preset release options

| Variable | Equivalent configuration | Description |
|---|---|---|
| `mkpgo.HidKpadRelease` | `NewKpadOption().WithDelay(0).WithKey("NONE")` | Release current key slots to `NONE`. |
| `mkpgo.HidKpadReleaseAll` | `NewKpadOption().WithDelay(0).WithRelease(0)` | Send a full keyboard release packet. |

## Output Parser Index

| Command | Parser | JSON | EOF | Parsed result |
|---|---|---:|---|---|
| `alog` | `RawDirective_alog` | No | `cli>` | Text; may be empty. |
| `astop` | `RawDirective_astop` | No | `cli>` | Text; empty content is treated as parse failure. |
| `acancel` | `RawDirective_acancel` | Yes | `<EOF>` | JSON text; empty content is treated as parse failure. |
| `sn` | `RawDirective_sn` | Yes | `<EOF>` | JSON text -> `SN`. |
| `list_dir` | `RawDirective_list_dir` | Yes | `<EOF>` | JSON text -> `FileSystem`; empty content returns an empty string. |
| `clean_dir` | `RawDirective_clean_dir` | No | `cli>` | Empty on success; output containing `Failed to` returns an error. |
| `delete_file` | `RawDirective_delete_file` | No | `cli>` | Empty on success; output containing `Failed to remove` returns an error. |
| `join` | `RawDirective_join` | No | `cli>` | Text; success must contain `connect: Connected`; error-code output returns execution failure. |
| `alive` | `RawDirective_alive` | Yes | `<EOF>` | JSON text -> `Heartbeat`. |
| `atime` | `RawDirective_atime` | Yes | `<EOF>` | Finds a JSON line containing `"seconds"` -> `LogLength`. |
| `aversion` | `RawDirective_aversion` | Yes | `<EOF>` | JSON text -> `MKPVersion`. |
| `ainsp` | `RawDirective_ainsp` | Yes | `<EOF>` | Finds a JSON line containing both `"seconds"` and `"width"` -> `LogInfo`. |
| `aplay` | None | - | default `cli>` | API sends async by default. |
| `m10` | None | - | default `cli>` | Sync mode usually ignores output. |
| `kpad` | None | - | default `cli>` | Sync mode usually ignores output. |

Common error rule: all parsers run `PreFlight` first. If output contains `command returned non-zero error code` case-insensitively, the parser returns `ErrRawDirecitveExecutionFailed`.

## Response Model Index

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

type LogInfo struct {
    LogOption
    LogLength
}
```

## Go API Index

### `SFSerialPort` command methods

| Method | Command | Description |
|---|---|---|
| `SendDirective` / `SendDirectiveContext` / `SendSyncDirective` / `SendSyncDirectiveContext` | Any | Synchronously send and return raw output; optional `DirectiveOption` values override this wait. |
| `SendDirectiveIgnoreOutput` / `SendDirectiveIgnoreOutputContext` | Any | Synchronously send, wait for completion, and ignore output; optional `DirectiveOption` values override this wait. |
| `SendDirectiveAsync` / `SendDirectiveAsyncContext` | Any | Send asynchronously. |
| `StartRecording` / `StartRecordingContext` | `alog` | Start recording asynchronously. |
| `StartReplaying` / `StartReplayingContext` | `aplay` | Start replay asynchronously. |
| `StopRecording` / `StopRecordingContext` | `astop` | Stop recording asynchronously. |
| `Stop` / `StopContext` | `astop` | Alias of `StopRecording`. |
| `CancelReplay` / `CancelReplayContext` | `acancel` | Cancel replay asynchronously. |
| `Mouse10` / `Mouse10Context` | `m10` | Send a mouse command; `Async=false` waits synchronously and ignores output. |
| `MouseReleaseAll` / `MouseReleaseAllContext` | `m10 --b 0` | Release all mouse buttons. |
| `Keypad` / `KeypadContext` | `kpad` | Send a keyboard command; `Async=false` waits synchronously and ignores output. |

### `helper` package methods

Sync-output helpers such as `Alog`, `Astop`, `Join`, `DeviceSN`, `ListDir`, `CleanDir`, `DeleteFile`, `Alive`, `Atime`, `Aversion`, and `AInspect` all accept optional `mkpgo.DirectiveOption` values; use `mkpgo.WithSyncOutputTimeout(...)` to override the timeout for one wait.

| Method | Command | Return |
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

### Common `controller.Controller` proxies

Matching sync-output proxy methods on `Controller` also accept optional `mkpgo.DirectiveOption` values.

`controller.Controller` wraps helper functions and higher-level keyboard/mouse operations:

- Recording/replay: `StartRecord`, `StopRecord`, `Alog`, `Astop`, `Cancel`
- Device/files/network: `DeviceSN`, `ListDir`, `CleanDir`, `DeleteFile`, `Join`, `Alive`, `Atime`, `Aversion`, `AInspect`
- Keyboard: `KeyDown`, `KeyUp`, `KeyTap`, `KeyPresses`, `KeypadRelease`, `KeypadReleaseAll`
- Mouse: `MouseClick`, `MouseClickWithOption`, `MouseScroll`, `MouseScrollWithOption`, `MouseScrollWithButton`, `MouseDown`, `MouseReleaseAll`, `MouseUp`, `M10Move`, `MouseMove`

## Draft / Unwrapped Commands

`mmm/mkp_directive260417.md` records two proposed commands. They currently have no registered parser and no formal Go wrapper in the main code.

### `apause`: pause replay

Suggested behavior: pause the log currently being replayed.

Suggested parameters:

1. `logName`: optional. Empty means the currently replaying log.
2. `paused_at`: optional. Pause at a specified timestamp, for example `1.22 s`. If firmware-side implementation is inconvenient, pause immediately when the command is received.

Note: save keyboard/mouse states at pause time so that resume can restore them.

### `aresume`: resume replay

Suggested behavior: resume log replay.

Suggested parameters:

1. `logName`: optional. Empty means the currently replaying log.
2. `useStates`: `1/0`; default `1`. `1` means use states saved by `apause` before resuming; `0` means resume directly.

Note: states saved by `apause` should be cleared after `alog`, `astop`, `acancel`, or `aresume`.
