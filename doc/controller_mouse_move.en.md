# Controller.MouseMove Usage Guide

This document explains the human-like mouse movement APIs on `github.com/elvuel/mkp-go/controller.Controller`, with a focus on `Controller.MouseMove`.

## 1. When to use it

`Controller.MouseMove` sends one relative mouse movement through an MKP device. Instead of sending a single low-level `m10 --x --y` command, it uses `MouseMovementSimulator` to generate multiple trajectory points and replays them as `m10` commands, producing smoother and more human-like movement.

Use it when you need to:

- hold a mouse button while dragging;
- split a larger relative movement into a smooth trajectory;
- use human-like movement features such as overshoot, correction, jitter, and pauses;
- send one wheel delta at the beginning of a movement.

If you only need a single low-level `m10` command, use `Mouse10` directly:

```go
sfport.Mouse10(mkpgo.NewM10Option().SetX(100).SetY(50))
```

If you need multiple dynamic steps with per-step `button`, `wheel`, or `pause`, see the `MouseMoveOffsets` section near the end of this document.

## 2. Basic setup

```go
package main

import (
    "time"

    mkpgo "github.com/elvuel/mkp-go"
    "github.com/elvuel/mkp-go/controller"
)

func main() {
    sfport := mkpgo.NewSFSerialPort()
    sfport.Name = "COM5" // Change to your actual device port.

    if err := sfport.Open(); err != nil {
        panic(err)
    }
    defer sfport.Close()

    ctrl := controller.NewController(sfport)

    if err := ctrl.MouseMove("", 120, -40, 220*time.Millisecond); err != nil {
        panic(err)
    }
}
```

## 3. API signature

```go
func (c *Controller) MouseMove(
    button string,
    relX, relY int,
    interval time.Duration,
    opts ...mkpgo.MouseMovementSimulatorOption,
) error
```

Parameters:

| Parameter | Description |
|---|---|
| `button` | Mouse button name to hold during movement. An empty string means no button. |
| `relX` | Relative X movement in M10 relative movement units. Positive values usually move right. |
| `relY` | Relative Y movement in M10 relative movement units. Positive values usually move down, depending on the device/system coordinate convention. |
| `interval` | Base trajectory duration. Actual per-sample timing is also affected by simulator configuration such as `SpeedMultiplier`, sampling interval, and overshoot phase ratios. |
| `opts` | Per-call `MouseMovementSimulatorOption` overrides. They do not permanently mutate the controller's base `MouseMovement` configuration. |

Common button names supported by `mkpgo.CheckMouseButton`:

| Name | Meaning |
|---|---|
| `""` or unknown string | No button / released state |
| `"left"` | Left button |
| `"right"` | Right button |
| `"both"` | Left + right buttons |
| `"middle"` | Middle button |
| `"backword"` | Back button, preserving the current spelling |
| `"forword"` | Forward button, preserving the current spelling |

## 4. Common examples

### 4.1 Smooth movement without holding a button

```go
err := ctrl.MouseMove("", 100, 50, 200*time.Millisecond)
```

### 4.2 Drag while holding the left button

```go
err := ctrl.MouseMove("left", 180, 0, 300*time.Millisecond)
```

`MouseMove` sends a final release `m10` command after the trajectory finishes.

### 4.3 Send one wheel delta at movement start

```go
err := ctrl.MouseMove(
    "",
    60,
    30,
    180*time.Millisecond,
    mkpgo.WithWheel(1),
)
```

`WithWheel(1)` sends `--w 1` once on the first replayed `m10` command. It is not repeated for every trajectory sample, so the wheel event is not multiplied by the number of generated points.

To clear an existing default wheel value:

```go
err := ctrl.MouseMove("", 60, 30, 180*time.Millisecond, mkpgo.WithoutWheel())
```

### 4.4 Disable overshoot, jitter, and pause

```go
err := ctrl.MouseMove(
    "",
    120,
    -35,
    220*time.Millisecond,
    mkpgo.WithoutOvershoot(),
    mkpgo.WithoutJitter(),
    mkpgo.WithoutPause(),
)
```

This is useful when you want more deterministic movement with less randomness.

### 4.5 Adjust the trajectory curve

```go
err := ctrl.MouseMove(
    "left",
    120,
    -35,
    220*time.Millisecond,
    mkpgo.WithBesselOffset(2.0, 1.0),
)
```

`WithBesselOffset(ctrlOffset, correctionOffset)` changes Bezier control-point offsets and therefore changes the trajectory curvature.

### 4.6 Use a custom simulator config

```go
cfg := mkpgo.DefaultMouseMovementSimulatorConfig()
cfg.SpeedMultiplier = 0.9
cfg.JitterMag = 0.7
cfg.CorrectionMag = 0.3
cfg.OvershootMin = 4
cfg.OvershootMax = 8

err := ctrl.MouseMove(
    "",
    160,
    20,
    260*time.Millisecond,
    mkpgo.WithConfig(cfg),
)
```

Note: `WithConfig(cfg)` replaces the config object used by this call. If you mutate `cfg` later, manage shared references accordingly.

## 5. Movement simulator options

Common `MouseMovementSimulatorOption` values:

| Option | Description |
|---|---|
| `mkpgo.WithWheel(wheel)` | Send one wheel delta at the start of this movement. |
| `mkpgo.WithoutWheel()` | Clear the wheel delta for this movement. |
| `mkpgo.WithBesselOffset(ctrl, correction)` | Set Bezier control-point offsets for sprint/correction phases. |
| `mkpgo.WithUnitsPerSecond(v)` | Set the M10 units/second used for automatic duration estimation. |
| `mkpgo.WithPixelsPerUnit(v)` | Set the screen-pixel to M10-unit scale. |
| `mkpgo.WithOvershoot(true/false)` | Enable or disable the overshoot phase. |
| `mkpgo.WithoutOvershoot()` | Disable the overshoot phase. |
| `mkpgo.WithPause(true/false)` | Enable or disable phase pause behavior. |
| `mkpgo.WithoutPause()` | Disable phase pause behavior. |
| `mkpgo.WithJitter(true/false)` | Enable or disable random jitter. |
| `mkpgo.WithoutJitter()` | Disable random jitter. |
| `mkpgo.WithConfig(cfg)` | Replace the simulator config for this call. |

## 6. Relationship to `MouseMoveOffsets`

Use `MouseMove` for one complete movement:

```go
ctrl.MouseMove("left", 100, 0, 200*time.Millisecond)
```

Use `MouseMoveOffsets` when you need multiple steps and each step may have its own `button`, `wheel`, or `pause`.

Current signature:

```go
func (c *Controller) MouseMoveOffsets(
    ctx context.Context,
    button string,
    offsets interface{},
    opts ...mkpgo.MouseMovementSimulatorOption,
) error
```

`ctx` is the first parameter:

- pass `context.Background()` when you do not need active cancellation;
- pass a context created with `context.WithCancel` / `context.WithTimeout` to cancel dynamic or streaming movement;
- if `nil` is passed, it is treated as `context.Background()` internally.

### 6.1 Fixed offsets

```go
offsets := []mkpgo.MouseMoveOffset{
    mkpgo.NewMouseMoveOffset(50, 0).WithButton("left").WithPause(80),
    mkpgo.NewMouseMoveOffset(0, 20).WithWheel(1).WithPause(50),
    mkpgo.NewMouseMoveOffset(-10, 0).WithoutButton(),
}

err := ctrl.MouseMoveOffsets(context.Background(), "", offsets)
```

The legacy `[][2]int` form is still supported, but it also requires `ctx`:

```go
err := ctrl.MouseMoveOffsets(context.Background(), "left", [][2]int{
    {50, 0},
    {0, 20},
})
```

### 6.2 Dynamic / streaming offsets

For dynamically generated or streaming offsets, pass a channel directly to `MouseMoveOffsets`:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

offsetCh := make(chan mkpgo.MouseMoveOffset)

go func() {
    defer close(offsetCh)
    offsetCh <- mkpgo.NewMouseMoveOffset(10, 0).WithPause(10)
    offsetCh <- mkpgo.NewMouseMoveOffset(10, 5).WithWheel(1)
}()

err := ctrl.MouseMoveOffsets(ctx, "", offsetCh)
```

You can also call `MouseMoveOffsetsStream` explicitly; it has the same behavior but a more specific type:

```go
err := ctrl.MouseMoveOffsetsStream(ctx, "", offsetCh)
```

> Note: `ctx` is currently checked before each offset step starts. If a step is already replaying its trajectory or sleeping in its post-step `Pause`, cancellation returns after that step finishes.

## 7. Notes

- `relX` / `relY` are M10 relative movement units, not necessarily screen pixels.
- `MouseMove` splits one movement into multiple `m10` commands, so duration and sample count depend on simulator configuration.
- `WithWheel` is sent once at movement start and is not repeated for every trajectory point.
- `m10` options are asynchronous by default, which is suitable for high-frequency trajectory replay.
- If you use low-level synchronous commands directly, make sure `go sfport.Read()` is running when needed. Normal asynchronous `m10` movement usually does not depend on synchronous output reading.


