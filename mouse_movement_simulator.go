package mkpgo

import (
	"math"
	"math/rand"
	"time"
)

// MouseMovementPoint 结构体
type MouseMovementPoint struct {
	RelX     float64
	RelY     float64
	AbsX     float64
	AbsY     float64
	Duration time.Duration
}

/*
对抗机器学习检测：建议大幅增加 JitterMag 并减小 SampleInterval。

模拟日常办公：可以降低 OvershootMax 并增加 PauseMaxMs。

模拟竞技游戏：减小 SpeedMultiplier（即提速）并开启 UseOvershoot，因为高手在拉枪时通常会有明显的过冲修正动作。
*/

// MouseMovementSimulatorConfig 包含所有可配置参数
type MouseMovementSimulatorConfig struct {
	// 时间的统一缩放：SpeedMultiplier 不仅改变了总路程的耗时，还自动调整了采样点之间的 Interval 以及 Pause 的长度。这保证了轨迹在变快或变慢时，其运动特征（如加速度曲线）保持比例一致，不会因为变快就显得闪烁。
	// 动态响应：你可以根据目标距离动态调整倍率。例如：
	// 距离很远：SpeedMultiplier = 0.6（快速划过）。
	// 距离很近：SpeedMultiplier = 1.5（小心微调）。
	// SpeedMultiplier = 0.5：动作变快一倍（耗时缩短）。
	// SpeedMultiplier = 2.0：动作变慢一倍（更像是在犹豫或仔细查找）。
	SpeedMultiplier float64 // 总体速度倍率

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

// DefaultMouseMovementSimulatorConfig 提供默认参数
func DefaultMouseMovementSimulatorConfig() *MouseMovementSimulatorConfig {
	return &MouseMovementSimulatorConfig{
		SpeedMultiplier:  1.0,
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

type MouseMovementSimulator struct {
	Cfg          *MouseMovementSimulatorConfig
	UseOvershoot bool
	UsePause     bool
	UseJitter    bool

	SFPort *SFSerialPort
}

func NewMouseController(cfg *MouseMovementSimulatorConfig, overshoot, pause, jitter bool) *MouseMovementSimulator {
	return &MouseMovementSimulator{
		Cfg:          cfg,
		UseOvershoot: overshoot,
		UsePause:     pause,
		UseJitter:    jitter,
	}
}

// generatePath 生成路径
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

// GenerateTrajectory 主生成逻辑
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

// MoveTo 执行移动
func (mc *MouseMovementSimulator) MoveTo(start, end [2]float64, baseTime time.Duration) {
	trajectory := mc.GenerateTrajectory(start, end, baseTime)
	for _, p := range trajectory {
		m10Opt := NewM10Option()
		// m10Opt.SetX(int(p.RelX)).SetY(int(p.RelY)).SetButton(0)
		m10Opt.SetX(int(p.RelX)).SetY(int(p.RelY)).SetButton(2)
		mc.SFPort.Mouse10(m10Opt)
		time.Sleep(p.Duration)
	}
}
