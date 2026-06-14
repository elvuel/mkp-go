package helper

import (
	"testing"
	"time"
)

func TestPixelsToM10Units(t *testing.T) {
	tests := []struct {
		name             string
		distancePixels   float64
		pixelsPerM10Unit float64
		want             float64
	}{
		{name: "positive", distancePixels: 150, pixelsPerM10Unit: 1.5, want: 100},
		{name: "preserve sign", distancePixels: -150, pixelsPerM10Unit: 1.5, want: -100},
		{name: "invalid scale", distancePixels: 150, pixelsPerM10Unit: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PixelsToM10Units(tt.distancePixels, tt.pixelsPerM10Unit); got != tt.want {
				t.Fatalf("PixelsToM10Units() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoM10Duration(t *testing.T) {
	tests := []struct {
		name              string
		distanceM10Units  float64
		m10UnitsPerSecond float64
		want              time.Duration
	}{
		{name: "normal", distanceM10Units: 520, m10UnitsPerSecond: 2600, want: 200 * time.Millisecond},
		{name: "min clamp", distanceM10Units: 140, m10UnitsPerSecond: 4600, want: 60 * time.Millisecond},
		{name: "max clamp", distanceM10Units: 10000, m10UnitsPerSecond: 1000, want: 360 * time.Millisecond},
		{name: "negative distance", distanceM10Units: -520, m10UnitsPerSecond: 2600, want: 200 * time.Millisecond},
		{name: "invalid speed", distanceM10Units: 520, m10UnitsPerSecond: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AutoM10Duration(tt.distanceM10Units, tt.m10UnitsPerSecond); got != tt.want {
				t.Fatalf("AutoM10Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}
