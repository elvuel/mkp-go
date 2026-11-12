# MKP Go

English | [中文](./README.md) | [Command Reference](./directives.en.md) | [Controller.MouseMove Guide](./doc/controller_mouse_move.en.md)

---

### What is MKP?

In this repository, **MKP** means a serial-controlled **mouse / keyboard recording, replay, and file-management device protocol**.

It usually runs on an external USB HID-capable board or firmware. A host program sends directives such as `m10`, `kpad`, `alog`, `aplay`, and `list_dir` to the MKP device over a serial port. The device then emits USB HID mouse / keyboard events to the target computer, or performs device-side recording, replay, inspection, and file management for input logs.

`mkp-go` is the Go SDK / client library for the MKP protocol. It mainly helps you:

- open and manage the MKP serial port;
- send raw MKP CLI directives;
- wrap common mouse, keyboard, recording, replay, and file-management operations;
- parse JSON or text output returned by the device;
- use higher-level helper / controller APIs for clicks, moves, key actions, recording, and replay.

> In short: **MKP is the hardware / firmware-side mouse-keyboard automation and replay protocol; MKP Go is the Go library for controlling it.**

### Project Scope

`mkp-go` is not just an OS-level mouse / keyboard simulation package. Its main data path is:

```text
Go program -> Serial port -> MKP device / firmware -> USB HID mouse / keyboard events -> Target system
```

This model is useful when you need:

- input events generated through an external HID device;
- device-side recording, storage, inspection, and replay of input logs;
- low-level MKP CLI directives wrapped as Go APIs, controller capabilities, or tooling.

### Features

| Feature | Description | Commands/APIs |
|---|---|---|
| Serial connection | Configure and read / write the device serial port | `NewSFSerialPort`, `Open`, `Read`, `Close` |
| Raw directives | Send arbitrary MKP CLI directives, sync or async | `SendDirective`, `SendDirectiveAsync` |
| Mouse control | Relative move, buttons, wheel, release | `m10`, `Mouse10`, `MouseReleaseAll` |
| Keyboard control | Down, up, tap, combos, release-all | `kpad`, `Keypad`, `helper.KeyTap` |
| Recording / replay | Record, stop, cancel, and replay device-side logs | `alog`, `astop`, `acancel`, `aplay` |
| File management | Browse directories, delete files, clean paths, import / export logs | `list_dir`, `delete_file`, `clean_dir`, `aget`, `aput`, `ajson2log` |
| Device info | SN, heartbeat, version, log duration, and log metadata | `sn`, `alive`, `aversion`, `atime`, `ainsp` |
| Network config | Device-side Wi‑Fi and HTTP base configuration | `join`, `wifi_auto`, `ahttpbase` |
| Higher-level wrappers | helper / controller mouse and keyboard actions | `helper`, `controller` |

## Installation

```bash
go get github.com/elvuel/mkp-go
```

The current `go.mod` declares:

```text
go 1.25.0
```

## Quick Start

```go
package main

import (
    "fmt"

    mkpgo "github.com/elvuel/mkp-go"
    "github.com/elvuel/mkp-go/helper"
)

func main() {
    sfport := mkpgo.NewSFSerialPort()
    sfport.Name = "COM5" // Change to your actual device port.

    if err := sfport.Open(); err != nil {
        panic(err)
    }
    defer sfport.Close()

    // Required when waiting for synchronous command output.
    go sfport.Read()

    sn, err := helper.DeviceSN(sfport)
    if err != nil {
        panic(err)
    }

    fmt.Println("device sn:", sn.SN)
}
```

## Examples

The table below uses the same format across languages: source path, run command, and observed behavior.

| Example | Source path | Run command | Behavior |
|---|---|---|---|
| `m10` | `examples/m10/main.go` | `go run .\examples\m10` | Demonstrates relative mouse movement plus button press/release |
| `kpad` | `examples/kpad/main.go` | `go run .\examples\kpad` | Demonstrates text input, key combos, and auto release |
| `keys` | `examples/keys/main.go` | `go run .\examples\keys` | Demonstrates `helper.KeyTap`, `KeyDown`, and `KeyUp` |
| `context` | `examples/context/main.go` | `go run .\examples\context` | Demonstrates timeout cancelation and manual cancelation with `context` |
| `commonpkg` | `examples/commonpkg/common.go` | Reused by other examples | Provides serial setup, stop-record, and logging helpers |

Common directive examples are summarized the same way:

| Directive examples | Source path | Run command | Purpose |
|---|---|---|---|
| `sn` / `aversion` / `alive` | `examples/directives/<name>/main.go` | `go run .\examples\directives\sn` | Query device serial number, version, and liveness |
| `list_dir` / `clean_dir` / `delete_file` | `examples/directives/<name>/main.go` | `go run .\examples\directives\list_dir /eMMC/applog` | Browse directories, clean directories, and delete files |
| `join` / `wifi_auto` / `ahttpbase` | `examples/directives/<name>/main.go` | `go run .\examples\directives\join <ssid> <password>` | Configure Wi‑Fi, auto-connect, and HTTP base address |
| `ainsp` / `atime` | `examples/directives/<name>/main.go` | `go run .\examples\directives\atime /eMMC/applog/spike.log` | Inspect log metadata and duration |
| `adumj` / `aget` / `aput` / `ajson2log` | `examples/directives/<name>/main.go` | `go run .\examples\directives\ajson2log D:\tmp\demo.json /eMMC/applog/demo.log` | Upload, download, and convert between JSON and MKP logs |
| `alog` / `astop` | `examples/directives/<name>/main.go` | `go run .\examples\directives\alog mkpdemo/demo` | Start and stop device-side recording |

## Recording and Replay

```go
// Start recording asynchronously.
err := sfport.StartRecording("demo")

// Stop recording.
err = sfport.StopRecording()

// Replay demo asynchronously with delay 0.
err = sfport.StartReplaying("demo", 0)

// Cancel replay.
err = sfport.CancelReplay()
```

To send `alog` synchronously and wait for output:

```go
out, err := helper.Alog(sfport, "demo", nil)
fmt.Println(out, err)
```

## File and Device Information

```go
fs, err := helper.ListDir(sfport, "/eMMC/applog")
version, err := helper.Aversion(sfport)
hb, err := helper.Alive(sfport)
length, err := helper.Atime(sfport, "demo/record1")
info, err := helper.AInspect(sfport, "demo/record1")
```

## Directive Modes

mkp-go supports three send modes:

| Mode | API | Description |
|---|---|---|
| Sync with output | `SendDirective` | Wait for device output and return raw output. |
| Sync ignoring output | `SendDirectiveIgnoreOutput` | Wait for completion but return empty output; useful for `m10` and `kpad`. |
| Async | `SendDirectiveAsync` | Write to the serial port without waiting; useful for high-frequency input events. |

## Repository Layout

| Path | Description |
|---|---|
| `serial_port_base.go` | Base serial read/write and synchronous output handling. |
| `sf_serial_port.go` | MKP/SF serial wrapper and core command methods. |
| `directive_parser.go` | Built-in command output parsers. |
| `types_cli.go` | Response models for device info, log info, and filesystem data. |
| `types_m10.go` | Mouse `m10` option model. |
| `types_keycode*.go` | Keyboard HID key codes and `kpad` option model. |
| `helper/` | Business-friendly helper functions. |
| `controller/` | Higher-level mouse/keyboard controller. |
| `directives.en.md` | Full English MKP command reference. |

## Notes

- For synchronous commands, make sure `SyncOuputEnabled=true` and `go sfport.Read()` is running.
- `m10` and `kpad` are asynchronous by default. Use `WithAsync(false)` or `WithSyncIgnoreOutput(true)` when you need to wait for completion while ignoring output.
- File deletion/cleaning helpers are intentionally limited to `/eMMC/applog` to reduce accidental damage.
- Defaults from `NewSFSerialPort()` may need to be adjusted for your actual port, VID/PID, and serial number.
- Firmware support may vary by device/firmware version.













