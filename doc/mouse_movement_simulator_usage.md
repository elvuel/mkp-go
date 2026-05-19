# mouse_movement_simulator 使用说明 / FAQ

## 1. 作用概览

`mouse_movement_simulator` 用来生成一段“更像人手”的鼠标相对移动轨迹，并通过 `m10` 指令逐步发送到设备。

轨迹主要由以下几个部分组成：

- **冲刺阶段（Sprint）**：从起点快速接近目标。
- **过冲阶段（Overshoot，可选）**：先略微超过目标点。
- **修正阶段（Correction）**：从过冲点回拉到真实目标。
- **抖动（Jitter，可选）**：给轨迹加入少量随机扰动。
- **停顿（Pause，可选）**：冲刺后到修正前插入短暂停顿。

默认配置在“自然度、可控性、易用性”之间做了一个比较均衡的折中。

---

## 2. 快速开始

### 2.1 直接使用默认配置

```go
cfg := mkpgo.DefaultMouseMovementSimulatorConfig()

sim := mkpgo.NewMouseMovementSimulator(cfg, true, true, true)
sim.SetSFPort(sfport)

sim.MoveTo(
    int(mkpgo.LeftMouseButton),
    [2]float64{0, 0},
    [2]float64{120, -35},
    220*time.Millisecond,
)
```

### 2.2 通过 Controller 调整配置

```go
cfg := mkpgo.DefaultMouseMovementSimulatorConfig()
cfg.SpeedMultiplier = 0.9
cfg.JitterMag = 0.7
cfg.CorrectionMag = 0.3
cfg.OvershootMin = 4
cfg.OvershootMax = 8

err := ctrl.MouseMove(
    "left",
    120, -35,
    220*time.Millisecond,
    mkpgo.WithConfig(cfg),
    mkpgo.WithOvershoot(true),
    mkpgo.WithPause(true),
    mkpgo.WithJitter(true),
)
```

---

## 3. 参数怎么理解

### 3.1 时间与速度

#### `SpeedMultiplier`

- **作用**：统一缩放整段动作的时间。
- **效果**：
  - `< 1.0`：整体更快。
  - `> 1.0`：整体更慢。
- **注意**：它不仅影响总耗时，也会同步影响采样间隔和 Pause。

经验值：

- `0.6 ~ 0.85`：明显偏快。
- `0.9 ~ 1.1`：比较平衡，推荐先从这里调。
- `1.2 ~ 1.6`：偏稳、偏慢、微调感更强。

#### `SprintTimeRatio` / `CorrectionTimeRatio`

- **作用**：控制冲刺阶段和修正阶段如何分配 `baseTime`。
- **建议**：
  - 常规推荐：`0.72 ~ 0.78` / `0.22 ~ 0.28`
  - 默认值 `0.75 / 0.25` 就比较均衡
- **注意**：两者最好加起来约等于 `1.0`。

---

### 3.2 路径形状

#### `CtrlOffset`

- **作用**：冲刺阶段贝塞尔控制点偏移，决定主路径弯曲程度。
- **调大**：弯曲更明显，轨迹更“甩”。
- **调小**：更直、更干脆。

经验值：

- `15 ~ 25`：较直
- `25 ~ 40`：平衡推荐
- `40+`：弧线更明显，适合长距离移动

#### `CorrectionOffset`

- **作用**：修正阶段的贝塞尔控制点偏移。
- **建议**：通常明显小于 `CtrlOffset`，否则回拉会显得过于飘。

经验值：

- `4 ~ 6`：稳
- `6 ~ 10`：平衡推荐
- `10+`：修正段会更明显

---

### 3.3 采样密度

#### `SampleIntervalMin` / `SampleIntervalMax`

- **作用**：控制每个轨迹点之间的时间间隔。
- **调小**：点更密，运动更细。
- **调大**：点更稀，动作更“跳”。

默认值：

- `8ms ~ 14ms`

推荐区间：

- `7 ~ 12`：更细腻
- `8 ~ 14`：平衡推荐
- `10 ~ 16`：更稀疏

#### `WithAverageDurationPerStep`

- `true`：总耗时更稳定，每步时长围绕平均值轻微波动。
- `false`：每步时长直接在 `SampleIntervalMin/Max` 区间里随机，更强调报点波动感。

如果你想先把整体手感调顺，建议先保持 `true`。

---

### 3.4 抖动

#### `JitterMag`

- **作用**：冲刺阶段抖动幅度。
- **调大**：更飘、更散、更像大幅移动时的细碎误差。
- **调小**：更稳、更干净。

#### `CorrectionMag`

- **作用**：修正阶段抖动幅度。
- **建议**：通常应小于 `JitterMag`，否则回拉阶段会显得不够“收”。

经验值：

- `JitterMag = 0.4 ~ 0.8`：平衡推荐
- `CorrectionMag = 0.2 ~ 0.4`：平衡推荐
- 常见搭配：`CorrectionMag ≈ JitterMag 的 30% ~ 60%`

---

### 3.5 过冲与停顿

#### `OvershootMin` / `OvershootMax`

- **作用**：定义过冲像素区间。
- **调大**：更容易出现“冲过头再拉回”的感觉。
- **调小**：更直接命中目标。

默认值：

- `5 ~ 10 px`

建议：

- 短距离（`< 30px`）：建议关闭过冲，或设为 `0 ~ 2`
- 中距离（`30 ~ 150px`）：建议 `3 ~ 8`
- 长距离（`> 150px`）：建议 `5 ~ 12`

一个实用原则：

- **过冲不要明显大于总位移的 10% ~ 20%**
- 否则修正段会显得过长、过假

#### `PauseMinMs` / `PauseMaxMs`

- **作用**：冲刺结束后、修正开始前的停顿区间。
- **默认值**：`20 ~ 60ms`
- **建议**：
  - `10 ~ 25ms`：节奏更紧
  - `20 ~ 60ms`：平衡推荐
  - `40 ~ 90ms`：更犹豫、更明显

---

## 4. 几套实用预设

### 4.1 平衡移动（推荐先从这里开始）

适合大多数普通鼠标移动。

```go
cfg := mkpgo.DefaultMouseMovementSimulatorConfig()
cfg.SpeedMultiplier = 0.9
cfg.CtrlOffset = 30
cfg.CorrectionOffset = 7
cfg.SampleIntervalMin = 8
cfg.SampleIntervalMax = 14
cfg.JitterMag = 0.6
cfg.CorrectionMag = 0.3
cfg.OvershootMin = 3
cfg.OvershootMax = 8
cfg.PauseMinMs = 20
cfg.PauseMaxMs = 50
cfg.SprintTimeRatio = 0.75
cfg.CorrectionTimeRatio = 0.25
```

特点：

- 主路径不算太直，也不会太甩
- 末端修正存在，但不夸张
- 适合先建立“基线手感”

### 4.2 稳定微调

适合小距离、希望更稳时使用。

```go
cfg := mkpgo.DefaultMouseMovementSimulatorConfig()
cfg.SpeedMultiplier = 1.15
cfg.CtrlOffset = 18
cfg.CorrectionOffset = 4
cfg.JitterMag = 0.25
cfg.CorrectionMag = 0.1
cfg.OvershootMin = 0
cfg.OvershootMax = 2
cfg.PauseMinMs = 10
cfg.PauseMaxMs = 25
cfg.SprintTimeRatio = 0.82
cfg.CorrectionTimeRatio = 0.18
```

建议同时考虑：

- 小于 `20 ~ 30px` 的位移直接关闭 `UseOvershoot`

### 4.3 快速冲刺

适合长距离快速移动。

```go
cfg := mkpgo.DefaultMouseMovementSimulatorConfig()
cfg.SpeedMultiplier = 0.7
cfg.CtrlOffset = 38
cfg.CorrectionOffset = 8
cfg.SampleIntervalMin = 7
cfg.SampleIntervalMax = 12
cfg.JitterMag = 0.8
cfg.CorrectionMag = 0.25
cfg.OvershootMin = 6
cfg.OvershootMax = 12
cfg.PauseMinMs = 10
cfg.PauseMaxMs = 30
cfg.SprintTimeRatio = 0.78
cfg.CorrectionTimeRatio = 0.22
```

特点：

- 冲刺阶段更强
- 抖动略大
- 过冲更明显，但修正仍需保持收敛

---

## 5. FAQ

### Q1：平衡移动应该先调哪些参数？

建议顺序：

1. 先调 `SpeedMultiplier`
2. 再调 `CtrlOffset / CorrectionOffset`
3. 再调 `JitterMag / CorrectionMag`
4. 最后调 `OvershootMin / OvershootMax` 和 `Pause`

原因：

- 速度决定“节奏”
- Offset 决定“轨迹形状”
- Jitter 决定“细碎随机感”
- Overshoot/Pause 决定“收尾风格”

如果一开始所有参数一起改，通常很难判断到底是哪一项让手感变差。

### Q2：平衡移动推荐的起步值是什么？

直接从默认值开始，再做小幅调整即可：

- `SpeedMultiplier = 0.9 ~ 1.0`
- `CtrlOffset = 25 ~ 35`
- `CorrectionOffset = 6 ~ 8`
- `JitterMag = 0.5 ~ 0.7`
- `CorrectionMag = 0.2 ~ 0.35`
- `OvershootMin/Max = 3 ~ 8`

### Q3：冲刺抖动浮动怎么设置比较自然？

优先调 `JitterMag`，并让 `CorrectionMag` 保持更小。

推荐经验：

- 保守：`0.35 / 0.15`
- 平衡：`0.6 / 0.3`
- 激进：`0.8 / 0.25 ~ 0.4`

不建议一开始就把 `JitterMag` 拉得太高：

- `< 0.3`：容易显得太直太干净
- `0.4 ~ 0.8`：通常比较自然
- `> 1.0`：容易出现明显锯齿感，特别是在短距离移动上

### Q4：过冲区间怎么设置比较合适？

最核心的判断标准不是“固定多少像素”，而是**相对当前位移长度是否合理**。

推荐：

- **短距离**：优先关闭过冲，或 `0 ~ 2px`
- **中距离**：`3 ~ 8px`
- **长距离**：`5 ~ 12px`

如果你发现轨迹经常“先冲太远再慢慢回来”，一般说明：

- `OvershootMax` 偏大
- 或 `CorrectionTimeRatio` 偏高
- 或 `CorrectionOffset` 偏大

### Q5：什么时候应该关闭 `UseOvershoot`？

以下情况建议优先关闭：

- 小范围微调
- 目标点很近
- 你希望轨迹更直接、更稳

一个很实用的经验：

- **位移小于 20~30px 时，默认先不要过冲**

### Q6：`WithAverageDurationPerStep` 该开还是关？

建议：

- **先开 (`true`)**：更容易得到稳定、可控的整体动作时长
- **再尝试关 (`false`)**：当你想让每一步报点间隔更有随机感时再切换

如果你在调参阶段还没把速度和路径调顺，通常没必要先改这个。

### Q7：为什么我设置了 Pause 反而报错或行为异常？

当前实现里：

```go
pauseMs := mc.Cfg.PauseMinMs + rand.Intn(pRange)
```

这意味着在 `UsePause = true` 时，**`PauseMaxMs` 必须大于 `PauseMinMs`**。

也就是说：

- `20 / 60`：可以
- `20 / 21`：可以
- `20 / 20`：当前实现下不建议

如果你想固定停顿时间，当前更稳妥的做法是给它留一个很小区间，比如 `20 ~ 21ms`。

### Q8：为什么我的短距离移动看起来很奇怪？

短距离移动最容易出问题，常见原因是：

- `CtrlOffset` 太大
- `JitterMag` 太大
- `Overshoot` 太大

处理建议：

- 降低 `CtrlOffset`
- 降低 `JitterMag`
- 关闭过冲
- 让 `SpeedMultiplier` 略大于 `1.0`

### Q9：一套配置能覆盖所有距离吗？

可以先用一套通用配置，但更推荐按距离做轻微分档：

- 短距离：更稳、少抖、少过冲
- 中距离：默认平衡参数
- 长距离：更快、更弯、允许适度过冲

如果你后续要做自动调参，最值得优先按距离动态调整的参数是：

1. `SpeedMultiplier`
2. `OvershootMin / OvershootMax`
3. `CtrlOffset`
4. `JitterMag`

---

## 6. 一个简单的调参策略

如果你不知道从哪里开始，建议按下面顺序试：

1. 用默认配置跑一遍
2. 觉得太快/太慢，只改 `SpeedMultiplier`
3. 觉得太直/太飘，只改 `CtrlOffset` 和 `CorrectionOffset`
4. 觉得太机械，只改 `JitterMag` 和 `CorrectionMag`
5. 觉得收尾不自然，再改 `Overshoot`、`Pause`、`CorrectionTimeRatio`

建议每次只改 1~2 个参数，并且每次变化幅度不要太大。

这样最容易找到稳定、可复用的参数区间。
