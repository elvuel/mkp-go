# Controller.MouseMove 使用说明

本文说明 `github.com/elvuel/mkp-go/controller.Controller` 中与拟人化鼠标移动相关的 API，重点是 `Controller.MouseMove`。

## 1. 适用场景

`Controller.MouseMove` 用于通过 MKP 设备发送一段相对鼠标移动。它不是一次性发送一个 `m10 --x --y`，而是通过 `MouseMovementSimulator` 生成多段轨迹点，再逐步发送 `m10` 指令，使移动更平滑、更接近人工移动。

适合：

- 拖拽时保持某个鼠标按钮按下；
- 将一次较大的相对移动拆成平滑轨迹；
- 需要过冲、修正、抖动、停顿等拟人化轨迹特征；
- 移动开始时附带一次滚轮输入。

如果你只需要发送一个低层 `m10` 指令，可以直接使用：

```go
sfport.Mouse10(mkpgo.NewM10Option().SetX(100).SetY(50))
```

如果你需要多个动态 step、每段单独设置 button/wheel/pause，见本文末尾的 `MouseMoveOffsets` 相关说明。

## 2. 基本初始化

```go
package main

import (
    "time"

    mkpgo "github.com/elvuel/mkp-go"
    "github.com/elvuel/mkp-go/controller"
)

func main() {
    sfport := mkpgo.NewSFSerialPort()
    sfport.Name = "COM5" // 按实际设备端口修改

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

## 3. API 签名

```go
func (c *Controller) MouseMove(
    button string,
    relX, relY int,
    interval time.Duration,
    opts ...mkpgo.MouseMovementSimulatorOption,
) error
```

参数说明：

| 参数 | 说明 |
|---|---|
| `button` | 移动过程中按住的鼠标按钮名称。空字符串表示不按按钮。 |
| `relX` | 相对 X 位移，单位为 M10 相对移动单位。正数通常表示向右。 |
| `relY` | 相对 Y 位移，单位为 M10 相对移动单位。正数通常表示向下，具体方向取决于设备/系统坐标约定。 |
| `interval` | 轨迹基础耗时。实际每个采样点还会受模拟器配置影响，例如 `SpeedMultiplier`、采样间隔、过冲阶段分配等。 |
| `opts` | 本次调用临时覆盖的 `MouseMovementSimulatorOption`，不会永久修改 `Controller.MouseMovement` 的基础配置。 |

支持的常见按钮名称来自 `mkpgo.CheckMouseButton`：

| 名称 | 含义 |
|---|---|
| `""` 或未知字符串 | 不按按钮 / 释放状态 |
| `"left"` | 左键 |
| `"right"` | 右键 |
| `"both"` | 左键 + 右键 |
| `"middle"` | 中键 |
| `"backword"` | 后退键（保留当前拼写） |
| `"forword"` | 前进键（保留当前拼写） |

## 4. 常用示例

### 4.1 不按按钮平滑移动

```go
err := ctrl.MouseMove("", 100, 50, 200*time.Millisecond)
```

### 4.2 按住左键拖拽

```go
err := ctrl.MouseMove("left", 180, 0, 300*time.Millisecond)
```

`MouseMove` 会在轨迹结束后发送一次释放按钮的 m10 指令。

### 4.3 移动开始时附带一次滚轮

```go
err := ctrl.MouseMove(
    "",
    60,
    30,
    180*time.Millisecond,
    mkpgo.WithWheel(1),
)
```

`WithWheel(1)` 会在本次移动回放的第一条 `m10` 指令中发送一次 `--w 1`。它不会被每个轨迹采样点重复发送，因此不会因为轨迹拆分而放大滚轮次数。

清除已有默认 wheel：

```go
err := ctrl.MouseMove("", 60, 30, 180*time.Millisecond, mkpgo.WithoutWheel())
```

### 4.4 关闭过冲、抖动、停顿

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

适合需要更确定轨迹、减少随机性的场景。

### 4.5 调整轨迹曲线

```go
err := ctrl.MouseMove(
    "left",
    120,
    -35,
    220*time.Millisecond,
    mkpgo.WithBesselOffset(2.0, 1.0),
)
```

`WithBesselOffset(ctrlOffset, correctionOffset)` 会影响贝塞尔控制点偏移，进而改变轨迹弯曲程度。

### 4.6 使用自定义配置

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

注意：`WithConfig(cfg)` 会替换本次调用使用的配置对象。若后续还会修改 `cfg`，请自行管理共享引用带来的影响。

## 5. 可用的移动模拟器选项

常用 `MouseMovementSimulatorOption`：

| 选项 | 说明 |
|---|---|
| `mkpgo.WithWheel(wheel)` | 本次移动开始时附带一次滚轮位移。 |
| `mkpgo.WithoutWheel()` | 清除本次移动的滚轮位移。 |
| `mkpgo.WithBesselOffset(ctrl, correction)` | 设置冲刺/修正阶段贝塞尔控制点偏移。 |
| `mkpgo.WithUnitsPerSecond(v)` | 设置自动估算移动时长时使用的 M10 单位/秒。 |
| `mkpgo.WithPixelsPerUnit(v)` | 设置屏幕像素到 M10 单位比例。 |
| `mkpgo.WithOvershoot(true/false)` | 开关过冲阶段。 |
| `mkpgo.WithoutOvershoot()` | 关闭过冲阶段。 |
| `mkpgo.WithPause(true/false)` | 开关阶段间停顿。 |
| `mkpgo.WithoutPause()` | 关闭阶段间停顿。 |
| `mkpgo.WithJitter(true/false)` | 开关随机抖动。 |
| `mkpgo.WithoutJitter()` | 关闭随机抖动。 |
| `mkpgo.WithConfig(cfg)` | 替换本次调用的模拟器配置。 |

## 6. 与 `MouseMoveOffsets` 的关系

`MouseMove` 适合一次完整移动：

```go
ctrl.MouseMove("left", 100, 0, 200*time.Millisecond)
```

当你需要多个 step，并且每个 step 有独立的 `button`、`wheel`、`pause`，使用 `MouseMoveOffsets`。

当前签名：

```go
func (c *Controller) MouseMoveOffsets(
    ctx context.Context,
    button string,
    offsets interface{},
    opts ...mkpgo.MouseMovementSimulatorOption,
) error
```

`ctx` 是第一个参数：

- 可传 `context.Background()` 表示不主动取消；
- 可传 `context.WithCancel` / `context.WithTimeout` 创建的 context 来取消动态/流式移动；
- 如果传入 `nil`，内部会按 `context.Background()` 处理。

### 6.1 固定 offsets

```go
offsets := []mkpgo.MouseMoveOffset{
    mkpgo.NewMouseMoveOffset(50, 0).WithButton("left").WithPause(80),
    mkpgo.NewMouseMoveOffset(0, 20).WithWheel(1).WithPause(50),
    mkpgo.NewMouseMoveOffset(-10, 0).WithoutButton(),
}

err := ctrl.MouseMoveOffsets(context.Background(), "", offsets)
```

旧版 `[][2]int` 仍然可用，但也需要传入 `ctx`：

```go
err := ctrl.MouseMoveOffsets(context.Background(), "left", [][2]int{
    {50, 0},
    {0, 20},
})
```

### 6.2 动态/流式 offsets

当 offsets 是动态产生或流式产生时，可以把 channel 直接传给 `MouseMoveOffsets`：

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

也可以显式使用 `MouseMoveOffsetsStream`，语义相同，但类型更明确：

```go
err := ctrl.MouseMoveOffsetsStream(ctx, "", offsetCh)
```

> 注意：当前 `ctx` 会在每个 offset step 开始前检查；如果某个 step 已经开始执行轨迹或正在执行 step 后的 `Pause`，会等该 step 完成后再返回取消错误。

## 7. 注意事项

- `relX` / `relY` 是 M10 相对移动单位，不一定等同于屏幕像素。
- `MouseMove` 会将一次移动拆成多个 `m10` 指令，因此移动耗时和轨迹点数量会受配置影响。
- `WithWheel` 只在本次移动开始时发送一次，不会在每个轨迹点重复发送。
- 默认 `m10` 选项是异步发送，适合高频轨迹回放。
- 如果直接操作底层同步指令，确保按需启动 `go sfport.Read()`；普通异步 `m10` 移动通常不依赖同步输出读取。


