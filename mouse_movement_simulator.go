package mkpgo

import (
	"context"
	"math"
	"math/rand"
	"time"
)

const (
	// DefaultMouseMovementUnitsPerSecond is the default target speed used for automatic duration calculation.
	// DefaultMouseMovementUnitsPerSecond 是自动估算移动时长时使用的默认目标速度，单位为 M10 units/s。
	DefaultMouseMovementUnitsPerSecond = 2600.0
	// DefaultMouseMovementPixelsPerUnit is the default screen-pixel to M10-unit scale.
	// DefaultMouseMovementPixelsPerUnit 是默认的屏幕像素到 M10 单位比例，表示 1 个 M10 unit 对应多少像素。
	DefaultMouseMovementPixelsPerUnit = 1.0
	// DefaultMouseMovementMinDuration avoids splitting very small movements too finely.
	// DefaultMouseMovementMinDuration 避免小位移拆得过碎；自动时长低于该值时会被抬高。
	DefaultMouseMovementMinDuration = 60 * time.Millisecond
	// DefaultMouseMovementMaxDuration caps automatic duration for large movements.
	// DefaultMouseMovementMaxDuration 避免大位移拖得过久；自动时长高于该值时会被压低。
	DefaultMouseMovementMaxDuration = 360 * time.Millisecond
)

// MouseMovementPoint represents one sampled movement step.
// MouseMovementPoint 表示一次采样得到的鼠标移动步点。
type MouseMovementPoint struct {
	RelX     float64
	RelY     float64
	AbsX     float64
	AbsY     float64
	Duration time.Duration
}

// MouseMoveOffset represents one relative movement segment with optional wheel, button, and post-step pause overrides.
// MouseMoveOffset 表示一段相对移动，并可为该段单独覆盖滚轮、鼠标按键与段后停顿。
type MouseMoveOffset struct {
	X      int  `json:"x"`
	Y      int  `json:"y"`
	Wheel  *int `json:"wheel,omitempty"`
	Button *int `json:"button,omitempty"`
	// Pause is an optional sleep duration after this movement segment, in milliseconds.
	// Pause 是当前移动段结束后的可选停顿时长，单位为毫秒。
	Pause *int `json:"pause,omitempty"`
}

// NewMouseMoveOffset creates a relative movement segment.
// NewMouseMoveOffset 创建一段相对移动。
func NewMouseMoveOffset(x, y int) MouseMoveOffset {
	return MouseMoveOffset{X: x, Y: y}
}

// MouseMoveOffsetsFromPairs converts legacy [][2]int offsets into MouseMoveOffset values.
// MouseMoveOffsetsFromPairs 将旧版 [][2]int offsets 转换为 MouseMoveOffset。
func MouseMoveOffsetsFromPairs(offsets [][2]int) []MouseMoveOffset {
	items := make([]MouseMoveOffset, 0, len(offsets))
	for _, offset := range offsets {
		items = append(items, NewMouseMoveOffset(offset[0], offset[1]))
	}
	return items
}

// WithWheel sets this movement segment's optional wheel delta.
// WithWheel 设置当前移动段的可选滚轮位移值。
func (o MouseMoveOffset) WithWheel(wheel int) MouseMoveOffset {
	o.Wheel = &wheel
	return o
}

// WithoutWheel clears this movement segment's optional wheel delta.
// WithoutWheel 清除当前移动段的可选滚轮位移值。
func (o MouseMoveOffset) WithoutWheel() MouseMoveOffset {
	o.Wheel = nil
	return o
}

// WithPause sets an optional sleep duration after this movement segment, in milliseconds.
// WithPause 设置当前移动段结束后的可选停顿时长，单位为毫秒。
func (o MouseMoveOffset) WithPause(pauseMs int) MouseMoveOffset {
	o.Pause = &pauseMs
	return o
}

// WithoutPause clears this movement segment's optional pause.
// WithoutPause 清除当前移动段结束后的可选停顿。
func (o MouseMoveOffset) WithoutPause() MouseMoveOffset {
	o.Pause = nil
	return o
}

// WithButton sets this movement segment's button override by button name.
// WithButton 按按钮名称设置当前移动段的鼠标按键覆盖。
func (o MouseMoveOffset) WithButton(button string) MouseMoveOffset {
	return o.WithButtonMask(int(CheckMouseButton(button)))
}

// WithButtonMask sets this movement segment's button override by M10 bitmask.
// WithButtonMask 按 M10 按键位掩码设置当前移动段的鼠标按键覆盖。
func (o MouseMoveOffset) WithButtonMask(button int) MouseMoveOffset {
	o.Button = &button
	return o
}

// WithoutButton explicitly releases mouse buttons for this movement segment.
// WithoutButton 显式让当前移动段释放鼠标按键。
func (o MouseMoveOffset) WithoutButton() MouseMoveOffset {
	return o.WithButtonMask(int(ReleaseMouseButton))
}

/*
对抗机器学习检测：建议大幅增加 JitterMag 并减小 SampleInterval。

模拟日常办公：可以降低 OvershootMax 并增加 PauseMaxMs。

模拟竞技游戏：减小 SpeedMultiplier（即提速）并开启 UseOvershoot，因为高手在拉枪时通常会有明显的过冲修正动作。
*/

// MouseMovementSimulatorConfig contains tunable trajectory parameters.
// MouseMovementSimulatorConfig 包含轨迹生成的可调参数。
type MouseMovementSimulatorConfig struct {
	// 时间的统一缩放：SpeedMultiplier 不仅改变了总路程的耗时，还自动调整了采样点之间的 Interval 以及 Pause 的长度。这保证了轨迹在变快或变慢时，其运动特征（如加速度曲线）保持比例一致，不会因为变快就显得闪烁。
	// 动态响应：你可以根据目标距离动态调整倍率。例如：
	// 距离很远：SpeedMultiplier = 0.6（快速划过）。
	// 距离很近：SpeedMultiplier = 1.5（小心微调）。
	// SpeedMultiplier = 0.5：动作变快一倍（耗时缩短）。
	// SpeedMultiplier = 2.0：动作变慢一倍（更像是在犹豫或仔细查找）。
	SpeedMultiplier float64 // 总体速度倍率

	// UnitsPerSecond is the target M10 speed used by AutoM10Duration; <=0 uses the default.
	// UnitsPerSecond 是自动估算移动时长时使用的目标 M10 单位/秒；<=0 时使用默认值。
	UnitsPerSecond float64
	// PixelsPerUnit describes how many screen pixels correspond to one M10 unit; <=0 uses the default.
	// PixelsPerUnit 表示 1 个 M10 单位对应多少屏幕像素；<=0 时使用默认值。
	PixelsPerUnit float64

	// 路径形状
	CtrlOffset       float64 // 冲刺阶段贝塞尔控制点偏移
	CorrectionOffset float64 // 修正阶段贝塞尔控制点偏移

	WithAverageDurationPerStep bool // 是否启用平均步长，影响轨迹点密度， 如果启用 则在 generatePath 中 不使用采样 范围内随机生成当前点的持续时间

	// 采样 可以轻松控制轨迹的“细腻程度”。数值越小，轨迹点越密集，对高精度检测的抗性越高。
	// SampleInterval float64 // 采样间隔(ms)，影响轨迹点密度
	// 采样间隔随机范围 (ms)
	SampleIntervalMin float64 // 最小采样间隔 (默认 8.0)
	SampleIntervalMax float64 // 最大采样间隔 (默认 14.0)

	// 抖动
	JitterMag     float64 // 冲刺阶段抖动幅度
	CorrectionMag float64 // 修正阶段抖动幅度

	// 过冲 通过 Min 和 Max 定义一个区间，让每次移动都产生不可预测的随机性。
	OvershootMin float64 // 最小过冲像素
	OvershootMax float64 // 最大过冲像素

	// 停顿 通过 Min 和 Max 定义一个区间，让每次移动都产生不可预测的随机性。
	PauseMinMs int // 最小停顿时间
	PauseMaxMs int // 最大停顿时间

	// 时间分配比例 (两者之和应为 1.0)
	SprintTimeRatio     float64 // 冲刺阶段耗时比例 (默认 0.75)
	CorrectionTimeRatio float64 // 修正阶段耗时比例 (默认 0.25)
}

// DefaultMouseMovementSimulatorConfig returns sensible default simulator config.
// DefaultMouseMovementSimulatorConfig 返回默认模拟参数。
func DefaultMouseMovementSimulatorConfig() *MouseMovementSimulatorConfig {
	return &MouseMovementSimulatorConfig{
		SpeedMultiplier: 1.0,
		UnitsPerSecond:  DefaultMouseMovementUnitsPerSecond,
		PixelsPerUnit:   DefaultMouseMovementPixelsPerUnit,

		CtrlOffset:       35.0,
		CorrectionOffset: 8.0,

		WithAverageDurationPerStep: true,
		// SampleInterval:      11.0,
		SampleIntervalMin: 8.0, // 模拟大约 125Hz - 80Hz 的动态报点率
		SampleIntervalMax: 14.0,

		JitterMag:           0.6,
		CorrectionMag:       0.3,
		OvershootMin:        5.0,
		OvershootMax:        10.0,
		PauseMinMs:          20,
		PauseMaxMs:          60,
		SprintTimeRatio:     0.75, // 75% 的时间在移动
		CorrectionTimeRatio: 0.25, // 25% 的时间在修正
	}
}

// MouseMovementSimulator generates and replays human-like mouse trajectories.
// MouseMovementSimulator 用于生成并执行拟人化鼠标轨迹。
type MouseMovementSimulator struct {
	Cfg          *MouseMovementSimulatorConfig
	UseOvershoot bool
	UsePause     bool
	UseJitter    bool

	// Wheel is an optional wheel delta sent once at the start of MoveTo replay.
	// Wheel 是可选滚轮位移值；非 nil 时会在 MoveTo 回放的第一条 m10 指令中发送一次。
	Wheel *int

	SFPort *SFSerialPort
}

// NewMouseMovementSimulator constructs a simulator with explicit feature toggles.
// NewMouseMovementSimulator 使用指定开关创建模拟器实例。
func NewMouseMovementSimulator(cfg *MouseMovementSimulatorConfig, overshoot, pause, jitter bool) *MouseMovementSimulator {
	return &MouseMovementSimulator{
		Cfg:          cfg,
		UseOvershoot: overshoot,
		UsePause:     pause,
		UseJitter:    jitter,
	}
}

// MouseMovementSimulatorOption applies functional options to simulator.
// MouseMovementSimulatorOption 是模拟器函数式配置项。
type MouseMovementSimulatorOption func(*MouseMovementSimulator)

// WithBesselOffset sets Bezier control-point offsets.
// WithBesselOffset 设置贝塞尔控制点偏移参数。
func WithBesselOffset(ctrlOffset, correctionOffset float64) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.Cfg.CtrlOffset = ctrlOffset
		mms.Cfg.CorrectionOffset = correctionOffset
	}
}

// WithUnitsPerSecond sets the target M10 units/s used by automatic duration calculation.
// WithUnitsPerSecond 设置自动估算移动时长时使用的目标 M10 单位/秒。
func WithUnitsPerSecond(unitsPerSecond float64) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.Cfg.UnitsPerSecond = unitsPerSecond
	}
}

// WithPixelsPerUnit sets the screen-pixel to M10-unit scale.
// WithPixelsPerUnit 设置屏幕像素到 M10 单位的比例，表示 1 个 M10 unit 对应多少像素。
func WithPixelsPerUnit(pixelsPerUnit float64) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.Cfg.PixelsPerUnit = pixelsPerUnit
	}
}

// WithWheel sets an optional wheel delta for MoveTo/Controller.MouseMove.
// The wheel delta is sent once on the first replayed m10 directive so it is not multiplied by trajectory samples.
// WithWheel 为 MoveTo/Controller.MouseMove 设置可选滚轮位移值；该值只会在第一条回放 m10 指令中发送一次，避免被轨迹采样点重复放大。
func WithWheel(wheel int) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.SetWheel(wheel)
	}
}

// WithoutWheel clears the optional wheel delta for MoveTo/Controller.MouseMove.
// WithoutWheel 清除 MoveTo/Controller.MouseMove 的可选滚轮位移值。
func WithoutWheel() MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.WithoutWheel()
	}
}

// WithOvershoot enables/disables overshoot phase.
// WithOvershoot 启用或关闭过冲阶段。
func WithOvershoot(use bool) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.UseOvershoot = use
	}
}

// WithoutOvershoot disables overshoot phase.
// WithoutOvershoot 关闭过冲阶段。
func WithoutOvershoot() MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.UseOvershoot = false
	}
}

// WithPause enables/disables pause between phases.
// WithPause 启用或关闭阶段间停顿。
func WithPause(use bool) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.UsePause = use
	}
}

// WithoutPause disables pause between phases.
// WithoutPause 关闭阶段间停顿。
func WithoutPause() MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.UsePause = false
	}
}

// WithJitter enables/disables random jitter.
// WithJitter 启用或关闭抖动扰动。
func WithJitter(use bool) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.UseJitter = use
	}
}

// WithoutJitter disables random jitter.
// WithoutJitter 关闭抖动扰动。
func WithoutJitter() MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.UseJitter = false
	}
}

// WithSFPort binds target serial port for replay.
// WithSFPort 绑定用于执行轨迹的串口对象。
func WithSFPort(port *SFSerialPort) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.SFPort = port
	}
}

// WithConfig replaces simulator config.
// WithConfig 替换模拟器配置。
func WithConfig(cfg *MouseMovementSimulatorConfig) MouseMovementSimulatorOption {
	return func(mms *MouseMovementSimulator) {
		mms.Cfg = cfg
	}
}

// ApplyOptions applies functional options in order.
// ApplyOptions 按顺序应用函数式配置项。
func (mms *MouseMovementSimulator) ApplyOptions(opts ...MouseMovementSimulatorOption) *MouseMovementSimulator {
	for _, opt := range opts {
		opt(mms)
	}
	return mms
}

// SetSFPort sets target serial port.
// SetSFPort 设置目标串口。
func (mms *MouseMovementSimulator) SetSFPort(port *SFSerialPort) {
	mms.SFPort = port
}

// SetConfig sets simulator config directly.
// SetConfig 直接设置模拟参数。
func (mms *MouseMovementSimulator) SetConfig(cfg *MouseMovementSimulatorConfig) {
	mms.Cfg = cfg
}

// WithOvershoot toggles overshoot behavior.
// WithOvershoot 切换过冲行为开关。
func (mms *MouseMovementSimulator) WithOvershoot(use bool) {
	mms.UseOvershoot = use
}

// WithoutOvershoot disables overshoot behavior.
// WithoutOvershoot 关闭过冲行为。
func (mms *MouseMovementSimulator) WithoutOvershoot() {
	mms.UseOvershoot = false
}

// WithPause toggles phase pause behavior.
// WithPause 切换阶段停顿开关。
func (mms *MouseMovementSimulator) WithPause(use bool) {
	mms.UsePause = use
}

// WithoutPause disables phase pause behavior.
// WithoutPause 关闭阶段停顿。
func (mms *MouseMovementSimulator) WithoutPause() {
	mms.UsePause = false
}

// WithJitter toggles jitter behavior.
// WithJitter 切换抖动开关。
func (mms *MouseMovementSimulator) WithJitter(use bool) {
	mms.UseJitter = use
}

// WithoutJitter disables jitter behavior.
// WithoutJitter 关闭抖动行为。
func (mms *MouseMovementSimulator) WithoutJitter() {
	mms.UseJitter = false
}

// SetWheel sets an optional wheel delta for MoveTo replay.
// SetWheel 设置 MoveTo 回放时附带的一次性滚轮位移值。
func (mms *MouseMovementSimulator) SetWheel(wheel int) {
	mms.Wheel = &wheel
}

// WithoutWheel clears the optional wheel delta for MoveTo replay.
// WithoutWheel 清除 MoveTo 回放时附带的一次性滚轮位移值。
func (mms *MouseMovementSimulator) WithoutWheel() {
	mms.Wheel = nil
}

// PixelsToUnits converts a screen-pixel distance to M10 units using PixelsPerUnit.
// PixelsToUnits 根据传入的屏幕像素距离和 PixelsPerUnit 返回对应的 M10 unit 数值。
func (mms *MouseMovementSimulator) PixelsToUnits(distancePixels float64) float64 {
	return distancePixels / mms.effectivePixelsPerUnit()
}

// AutoM10Duration estimates a movement duration from M10 unit deltas.
// AutoM10Duration 根据 M10 单位位移大小和 UnitsPerSecond 自动估算平滑移动总时长。
func (mms *MouseMovementSimulator) AutoM10Duration(x, y int) time.Duration {
	return mms.AutoM10DurationForUnits(math.Hypot(float64(x), float64(y)))
}

// AutoM10DurationForPixels estimates duration from a screen-pixel distance.
// AutoM10DurationForPixels 根据屏幕像素距离先换算为 M10 units，再自动估算平滑移动总时长。
func (mms *MouseMovementSimulator) AutoM10DurationForPixels(distancePixels float64) time.Duration {
	return mms.AutoM10DurationForUnits(math.Abs(mms.PixelsToUnits(distancePixels)))
}

// AutoM10DurationForUnits estimates duration from an M10-unit distance.
// AutoM10DurationForUnits 根据 M10 unit 距离和 UnitsPerSecond 自动估算平滑移动总时长。
func (mms *MouseMovementSimulator) AutoM10DurationForUnits(distanceUnits float64) time.Duration {
	distanceUnits = math.Abs(distanceUnits)
	if distanceUnits <= 0 {
		return 0
	}

	duration := time.Duration(distanceUnits / mms.effectiveUnitsPerSecond() * float64(time.Second))
	if duration < DefaultMouseMovementMinDuration {
		return DefaultMouseMovementMinDuration
	}
	if duration > DefaultMouseMovementMaxDuration {
		return DefaultMouseMovementMaxDuration
	}
	return duration
}

func (mms *MouseMovementSimulator) effectiveUnitsPerSecond() float64 {
	if mms != nil && mms.Cfg != nil && mms.Cfg.UnitsPerSecond > 0 {
		return mms.Cfg.UnitsPerSecond
	}
	return DefaultMouseMovementUnitsPerSecond
}

func (mms *MouseMovementSimulator) effectivePixelsPerUnit() float64 {
	if mms != nil && mms.Cfg != nil && mms.Cfg.PixelsPerUnit > 0 {
		return mms.Cfg.PixelsPerUnit
	}
	return DefaultMouseMovementPixelsPerUnit
}

// generatePath builds one movement segment path.
// generatePath 生成单段路径（冲刺段或修正段）。
func (mc *MouseMovementSimulator) generatePath(start, end [2]float64, totalTime time.Duration, isCorrection bool, lastAbs [2]float64) ([]MouseMovementPoint, [2]float64) {
	// 应用速度倍率
	adjustedTime := time.Duration(float64(totalTime) * mc.Cfg.SpeedMultiplier)

	// 原: 随机采样率 SampleInterval 当固定为 11.0(固定值时)时<可能会出现 固定报点频率的统计检测, 在数据分析中会呈现出极高的人工痕迹, 模拟更进一步，可以让 SampleInterval 也在一个微小的范围内随机（例如 11ms +/- 2ms）>
	// 因此配置中有了 SampleIntervalMin/SampleIntervalMax
	// steps := int(adjustedTime.Milliseconds() / int64(mc.Cfg.SampleInterval))

	// 新: 使用平均采样间隔预估步数
	avgInterval := (mc.Cfg.SampleIntervalMin + mc.Cfg.SampleIntervalMax) / 2
	steps := int(adjustedTime.Milliseconds() / int64(avgInterval))

	if steps < 5 {
		steps = 5
	}

	midX := (start[0] + end[0]) / 2
	midY := (start[1] + end[1]) / 2

	offset := mc.Cfg.CtrlOffset
	if isCorrection {
		offset = mc.Cfg.CorrectionOffset
	}

	control := [2]float64{
		midX + rand.NormFloat64()*offset,
		midY + rand.NormFloat64()*offset,
	}

	points := make([]MouseMovementPoint, 0, steps)
	currentAbs := lastAbs

	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		var easedT float64
		if isCorrection {
			easedT = 1 - math.Pow(1-t, 2)
		} else {
			easedT = -(math.Cos(math.Pi*t) - 1) / 2
		}

		mt := 1 - easedT
		targetAbsX := mt*mt*start[0] + 2*easedT*mt*control[0] + easedT*easedT*end[0]
		targetAbsY := mt*mt*start[1] + 2*easedT*mt*control[1] + easedT*easedT*end[1]

		if mc.UseJitter {
			mag := mc.Cfg.JitterMag
			if isCorrection {
				mag = mc.Cfg.CorrectionMag
			}
			targetAbsX += rand.NormFloat64() * mag
			targetAbsY += rand.NormFloat64() * mag
		}

		roundedAbsX := math.Round(targetAbsX)
		roundedAbsY := math.Round(targetAbsY)

		relX := roundedAbsX - currentAbs[0]
		relY := roundedAbsY - currentAbs[1]

		var interval time.Duration

		// 原: 总时长除以步数的平均值
		// interval := time.Duration((float64(adjustedTime.Milliseconds())/float64(steps))+(rand.Float64()*2-1)) * time.Millisecond

		// // 新：--- 核心改动：在配置范围内随机生成当前点的持续时间 ---
		// intervalRange := mc.Cfg.SampleIntervalMax - mc.Cfg.SampleIntervalMin
		// randomInterval := mc.Cfg.SampleIntervalMin + rand.Float64()*intervalRange
		// // 基础间隔受速度倍率影响
		// interval := time.Duration(randomInterval*mc.Cfg.SpeedMultiplier) * time.Millisecond

		if mc.Cfg.WithAverageDurationPerStep {
			interval = time.Duration((float64(adjustedTime.Milliseconds())/float64(steps))+(rand.Float64()*2-1)) * time.Millisecond
		} else {
			intervalRange := mc.Cfg.SampleIntervalMax - mc.Cfg.SampleIntervalMin
			randomInterval := mc.Cfg.SampleIntervalMin + rand.Float64()*intervalRange
			// 基础间隔受速度倍率影响
			interval = time.Duration(randomInterval*mc.Cfg.SpeedMultiplier) * time.Millisecond
		}

		points = append(points, MouseMovementPoint{
			RelX: relX, RelY: relY, AbsX: roundedAbsX, AbsY: roundedAbsY, Duration: interval,
		})

		currentAbs[0] = roundedAbsX
		currentAbs[1] = roundedAbsY
	}
	return points, currentAbs
}

// GenerateTrajectory creates complete trajectory from start to end.
// GenerateTrajectory 生成从起点到终点的完整轨迹。
func (mc *MouseMovementSimulator) GenerateTrajectory(start, end [2]float64, baseTime time.Duration) []MouseMovementPoint {
	lastAbs := start

	if !mc.UseOvershoot {
		path, _ := mc.generatePath(start, end, baseTime, false, lastAbs)
		return path
	}

	// 1. 计算过冲假终点
	angle := math.Atan2(end[1]-start[1], end[0]-start[0])
	dist := mc.Cfg.OvershootMin + rand.Float64()*(mc.Cfg.OvershootMax-mc.Cfg.OvershootMin)
	fakeEnd := [2]float64{
		end[0] + math.Cos(angle)*dist,
		end[1] + math.Sin(angle)*dist,
	}

	// 2. 根据配置中的 Ratio 分配时间
	sprintTime := time.Duration(float64(baseTime) * mc.Cfg.SprintTimeRatio)
	correctionTime := time.Duration(float64(baseTime) * mc.Cfg.CorrectionTimeRatio)

	// 第一阶段：冲刺
	path1, lastAbs := mc.generatePath(start, fakeEnd, sprintTime, false, lastAbs)

	// 停顿处理
	if mc.UsePause && len(path1) > 0 {

		// 原:
		// pause := time.Duration(float64(20+rand.Intn(40))*mc.SpeedMultiplier) * time.Millisecond

		// 新: 将 Pause 的计算应用了 SpeedMultiplier。这意味着如果你设置整体速度变慢，人类反应时间（停顿）也会相应地模拟得更久。
		pRange := mc.Cfg.PauseMaxMs - mc.Cfg.PauseMinMs
		// pauseMs := mc.Cfg.PauseMinMs + rand.Intn(pRange+1)
		pauseMs := mc.Cfg.PauseMinMs + rand.Intn(pRange)
		// 停顿也受速度倍率影响
		pause := time.Duration(float64(pauseMs)*mc.Cfg.SpeedMultiplier) * time.Millisecond
		path1[len(path1)-1].Duration += pause
	}

	// 第二阶段：修正
	path2, _ := mc.generatePath(fakeEnd, end, correctionTime, true, lastAbs)

	return append(path1, path2...)
}

// MoveOffsets moves through multiple relative M10 offsets with automatically calculated duration.
//
// The button name follows CheckMouseButton, for example "left", "right", "middle" or "both".
// When button is not empty and resolves to a pressed button, it is used as the default button for every offset.
// Use MoveOffsetSteps for per-offset button/wheel overrides.
//
// MoveOffsets 按多个相对 M10 offset 自动计算每段耗时并依次移动。
// button 使用 CheckMouseButton 支持的名称，例如 "left"、"right"、"middle"、"both"。
// 当 button 非空且能解析为按下状态时，会作为每段 offset 的默认按键。
// 如需为每段单独指定 button/wheel，请使用 MoveOffsetSteps。
func (mc *MouseMovementSimulator) MoveOffsets(button string, offsets [][2]int) {
	mc.MoveOffsetsWithButton(int(CheckMouseButton(button)), offsets)
}

// MoveOffsetsWithButton moves through multiple relative M10 offsets with automatically calculated duration.
// MoveOffsetsWithButton 使用按钮位掩码按多个相对 M10 offset 自动计算每段耗时并依次移动。
func (mc *MouseMovementSimulator) MoveOffsetsWithButton(button int, offsets [][2]int) {
	mc.MoveOffsetStepsWithButton(button, MouseMoveOffsetsFromPairs(offsets))
}

// MoveOffsetSteps moves through relative M10 offsets with per-offset optional button/wheel/pause overrides.
// MoveOffsetSteps 按多段相对 M10 offset 移动，并允许每段单独覆盖 button/wheel/pause。
func (mc *MouseMovementSimulator) MoveOffsetSteps(button string, offsets []MouseMoveOffset) {
	mc.MoveOffsetStepsContext(context.Background(), button, offsets)
}

// MoveOffsetStepsContext moves through relative M10 offsets and checks ctx before each offset.
// MoveOffsetStepsContext 按多段相对 M10 offset 移动，并在每段开始前检查 ctx。
func (mc *MouseMovementSimulator) MoveOffsetStepsContext(ctx context.Context, button string, offsets []MouseMoveOffset) error {
	return mc.MoveOffsetStepsWithButtonContext(ctx, int(CheckMouseButton(button)), offsets)
}

// MoveOffsetStepsWithButton moves through relative M10 offsets with a default button bitmask.
// Per-offset Button overrides the default; Button==0 explicitly releases buttons for that segment.
// Per-offset Wheel is sent once at the beginning of that segment; mc.Wheel is used as a default wheel when set.
// Per-offset Pause sleeps after that segment, in milliseconds.
// MoveOffsetStepsWithButton 使用默认按钮位掩码按多段相对 M10 offset 移动。
// 每段的 Button 会覆盖默认按钮；Button==0 表示该段显式释放按钮。
// 每段的 Wheel 会在该段开始时发送一次；如果设置了 mc.Wheel，则作为默认滚轮值使用。
// 每段的 Pause 会在该段结束后停顿，单位为毫秒。
func (mc *MouseMovementSimulator) MoveOffsetStepsWithButton(defaultButton int, offsets []MouseMoveOffset) {
	_ = mc.MoveOffsetStepsWithButtonContext(context.Background(), defaultButton, offsets)
}

// MoveOffsetStepsWithButtonContext moves through relative M10 offsets with a default button bitmask and checks ctx before each offset.
// MoveOffsetStepsWithButtonContext 使用默认按钮位掩码按多段相对 M10 offset 移动，并在每段开始前检查 ctx。
func (mc *MouseMovementSimulator) MoveOffsetStepsWithButtonContext(ctx context.Context, defaultButton int, offsets []MouseMoveOffset) error {
	if ctx == nil {
		ctx = context.Background()
	}
	m10Opt := NewM10Option()
	if mc.shouldReleaseAfterOffsets(defaultButton, offsets) {
		defer mc.releaseMouseAfterOffsets(m10Opt)()
	}

	current := [2]float64{0, 0}
	defaultButtonField := legacyM10ButtonField(defaultButton)
	for _, offset := range offsets {
		if err := ctx.Err(); err != nil {
			return err
		}
		mc.moveOffsetStep(m10Opt, &current, defaultButtonField, offset)
	}
	return nil
}

// MoveOffsetStream consumes relative M10 offsets from a channel until it is closed or ctx is canceled.
// MoveOffsetStream 从 channel 持续读取相对 M10 offset，直到 channel 关闭或 ctx 取消。
func (mc *MouseMovementSimulator) MoveOffsetStream(ctx context.Context, button string, offsets <-chan MouseMoveOffset) error {
	return mc.MoveOffsetStreamWithButton(ctx, int(CheckMouseButton(button)), offsets)
}

// MoveOffsetStreamWithButton consumes relative M10 offsets from a channel with a default button bitmask.
// MoveOffsetStreamWithButton 使用默认按钮位掩码从 channel 持续读取相对 M10 offset。
func (mc *MouseMovementSimulator) MoveOffsetStreamWithButton(ctx context.Context, defaultButton int, offsets <-chan MouseMoveOffset) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if offsets == nil {
		return nil
	}

	m10Opt := NewM10Option()
	needRelease := defaultButton != int(ReleaseMouseButton)
	defer func() {
		if needRelease {
			mc.releaseMouseAfterOffsets(m10Opt)()
		}
	}()

	current := [2]float64{0, 0}
	defaultButtonField := legacyM10ButtonField(defaultButton)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case offset, ok := <-offsets:
			if !ok {
				return nil
			}
			if offset.Button != nil && *offset.Button != int(ReleaseMouseButton) {
				needRelease = true
			}
			mc.moveOffsetStep(m10Opt, &current, defaultButtonField, offset)
		}
	}
}

// MoveTo sends generated trajectory to device as m10 commands.
// MoveTo 按生成轨迹发送 m10 指令执行移动。
func (mc *MouseMovementSimulator) MoveTo(button int, start, end [2]float64, baseTime time.Duration) {
	m10Opt := NewM10Option()

	defer func() {
		m10Opt.Reset()
		m10Opt.WithoutButton().SetX(0).SetY(0)
		mc.SFPort.Mouse10(m10Opt)
	}()

	trajectory := mc.GenerateTrajectory(start, end, baseTime)
	mc.replayTrajectoryWithWheel(m10Opt, button, trajectory, mc.Wheel)
}

func (mc *MouseMovementSimulator) replayTrajectory(m10Opt *M10Option, button int, trajectory []MouseMovementPoint) {
	mc.replayTrajectoryWithWheel(m10Opt, button, trajectory, nil)
}

func (mc *MouseMovementSimulator) replayTrajectoryWithWheel(m10Opt *M10Option, button int, trajectory []MouseMovementPoint, wheel *int) {
	mc.replayTrajectoryWithButtonAndWheel(m10Opt, legacyM10ButtonField(button), trajectory, wheel)
}

func (mc *MouseMovementSimulator) replayTrajectoryWithButtonAndWheel(m10Opt *M10Option, button *int, trajectory []MouseMovementPoint, wheel *int) {
	wheelSent := false
	for _, p := range trajectory {
		mc.prepareM10Point(m10Opt, button, int(p.RelX), int(p.RelY), wheel, !wheelSent)
		if wheel != nil && !wheelSent {
			wheelSent = true
		}
		mc.SFPort.Mouse10(m10Opt)
		time.Sleep(p.Duration)
	}
}

func (mc *MouseMovementSimulator) moveOffsetStep(m10Opt *M10Option, current *[2]float64, defaultButtonField *int, offset MouseMoveOffset) {
	buttonField := defaultButtonField
	if offset.Button != nil {
		buttonField = cloneIntPtr(offset.Button)
	}
	wheel := offset.Wheel
	if wheel == nil {
		wheel = mc.Wheel
	}

	if offset.X == 0 && offset.Y == 0 {
		if buttonField != nil || wheel != nil {
			mc.sendM10Point(m10Opt, buttonField, 0, 0, wheel)
		}
		sleepMouseMoveOffsetPause(offset.Pause)
		return
	}

	next := [2]float64{
		current[0] + float64(offset.X),
		current[1] + float64(offset.Y),
	}
	duration := mc.AutoM10Duration(offset.X, offset.Y)
	trajectory := mc.GenerateTrajectory(*current, next, duration)
	mc.replayTrajectoryWithButtonAndWheel(m10Opt, buttonField, trajectory, wheel)
	*current = next
	sleepMouseMoveOffsetPause(offset.Pause)
}

func (mc *MouseMovementSimulator) sendM10Point(m10Opt *M10Option, button *int, x, y int, wheel *int) {
	mc.prepareM10Point(m10Opt, button, x, y, wheel, true)
	mc.SFPort.Mouse10(m10Opt)
}

func (mc *MouseMovementSimulator) prepareM10Point(m10Opt *M10Option, button *int, x, y int, wheel *int, includeWheel bool) {
	m10Opt.Reset()
	if button == nil {
		m10Opt = m10Opt.WithoutButton()
	} else {
		m10Opt = m10Opt.WithButton(*button)
	}
	m10Opt.SetX(x).SetY(y)
	if wheel != nil && includeWheel {
		m10Opt.SetWheel(*wheel)
	}
}

func (mc *MouseMovementSimulator) shouldReleaseAfterOffsets(defaultButton int, offsets []MouseMoveOffset) bool {
	if defaultButton != int(ReleaseMouseButton) {
		return true
	}
	for _, offset := range offsets {
		if offset.Button != nil && *offset.Button != int(ReleaseMouseButton) {
			return true
		}
	}
	return false
}

func (mc *MouseMovementSimulator) releaseMouseAfterOffsets(m10Opt *M10Option) func() {
	return func() {
		m10Opt.Reset()
		m10Opt.SetButton(int(ReleaseMouseButton)).SetX(0).SetY(0)
		mc.SFPort.Mouse10(m10Opt)
	}
}

func legacyM10ButtonField(button int) *int {
	if button == int(ReleaseMouseButton) {
		return nil
	}
	v := button
	return &v
}

func cloneIntPtr(v *int) *int {
	if v == nil {
		return nil
	}
	clone := *v
	return &clone
}

func sleepMouseMoveOffsetPause(pauseMs *int) {
	if pauseMs == nil || *pauseMs <= 0 {
		return
	}
	time.Sleep(time.Duration(*pauseMs) * time.Millisecond)
}
