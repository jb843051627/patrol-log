package model

import "time"

// Plan 巡检计划：绑定一台设备，含若干巡检项
type Plan struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `json:"name"`
	DeviceID  string      `json:"device_id"`
	Interval  int         `json:"interval"` // 巡检间隔（小时）
	Enabled   bool        `json:"enabled"`
	CreatedAt time.Time   `json:"created_at"`
	Items     []CheckItem `gorm:"foreignKey:PlanID" json:"items"`
}

// CheckItem 巡检项：一项一个告警阈值区间 [WarnMin, WarnMax]
type CheckItem struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	PlanID  uint    `json:"plan_id"`
	Name    string  `json:"name"`
	Unit    string  `json:"unit"`
	WarnMin float64 `json:"warn_min"`
	WarnMax float64 `json:"warn_max"`
}

// Result 巡检结果：一次巡检中单个巡检项的读数与判定
type Result struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PlanID    uint      `json:"plan_id"`
	DeviceID  string    `json:"device_id"`
	ItemName  string    `json:"item_name"`
	Reading   float64   `json:"reading"`
	Abnormal  bool      `json:"abnormal"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// Alert 告警记录：读数越界时生成
type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PlanID    uint      `json:"plan_id"`
	DeviceID  string    `json:"device_id"`
	ItemName  string    `json:"item_name"`
	Reading   float64   `json:"reading"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}
