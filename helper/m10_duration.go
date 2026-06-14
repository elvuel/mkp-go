package helper

import (
	"math"
	"time"
)

const (
	autoM10DurationMin = 60 * time.Millisecond
	autoM10DurationMax = 360 * time.Millisecond
)

// PixelsToM10Units converts a screen-pixel distance to M10 units.
// PixelsToM10Units 根据屏幕像素距离和 pixelsPerM10Unit 返回对应的 M10 unit 数值。
func PixelsToM10Units(distancePixels, pixelsPerM10Unit float64) float64 {
	if pixelsPerM10Unit <= 0 {
		return 0
	}
	return distancePixels / pixelsPerM10Unit
}

// AutoM10Duration estimates an M10 movement duration from an M10-unit distance.
//
// The estimated duration is clamped to the same 60ms-360ms range used by the
// simulator's automatic M10 duration helpers. Invalid speed parameters return 0.
//
// AutoM10Duration 根据 M10 unit 距离和 m10UnitsPerSecond 自动估算 M10 移动时长。
// 返回值会限制在 60ms-360ms；速度参数无效时返回 0。
func AutoM10Duration(distanceM10Units, m10UnitsPerSecond float64) time.Duration {
	distanceM10Units = math.Abs(distanceM10Units)
	if distanceM10Units <= 0 || m10UnitsPerSecond <= 0 {
		return 0
	}

	duration := time.Duration(distanceM10Units / m10UnitsPerSecond * float64(time.Second))
	if duration < autoM10DurationMin {
		return autoM10DurationMin
	}
	if duration > autoM10DurationMax {
		return autoM10DurationMax
	}
	return duration
}
