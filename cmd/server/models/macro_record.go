package models

import "time"

type MacroRecord struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name         string `gorm:"size:255;not null" json:"name"`
	UniqueID     string `gorm:"size:64;uniqueIndex;not null" json:"unique_id"`
	MKPPath      string `gorm:"size:512;not null" json:"mkp_path"`
	StartPointX  int    `json:"start_point_x"`
	StartPointY  int    `json:"start_point_y"`
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
	OS           string `gorm:"size:64;not null" json:"os"`

	Seconds      int `json:"seconds"`
	Milliseconds int `json:"milliseconds"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
