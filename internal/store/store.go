package store

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"patrol-log/internal/model"
)

// 业务错误
var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicateName = errors.New("duplicate plan name")
)

// Store 数据访问层
type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// CreatePlan 创建巡检计划；同名计划返回 ErrDuplicateName
func (s *Store) CreatePlan(p *model.Plan) error {
	var count int64
	if err := s.db.Model(&model.Plan{}).Where("name = ?", p.Name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateName
	}
	if err := s.db.Create(p).Error; err != nil {
		return fmt.Errorf("create plan: %w", err)
	}
	return nil
}

// GetPlanByID 查询计划（含巡检项）；不存在时返回 ErrNotFound
func (s *Store) GetPlanByID(id uint) (*model.Plan, error) {
	var p model.Plan
	err := s.db.Preload("Items").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateResult 写入一条巡检结果
func (s *Store) CreateResult(r *model.Result) error {
	return s.db.Create(r).Error
}

// CreateAlert 写入一条告警记录
func (s *Store) CreateAlert(a *model.Alert) error {
	return s.db.Create(a).Error
}

// ListResults 按设备与起始时间查询巡检结果（倒序）
func (s *Store) ListResults(deviceID string, since time.Time) ([]model.Result, error) {
	var rs []model.Result
	err := s.db.Where("device_id = ? AND created_at >= ?", deviceID, since).
		Order("created_at DESC").Find(&rs).Error
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// ListAlerts 按设备查询告警记录（倒序）
func (s *Store) ListAlerts(deviceID string) ([]model.Alert, error) {
	var as []model.Alert
	err := s.db.Where("device_id = ?", deviceID).
		Order("created_at DESC").Find(&as).Error
	if err != nil {
		return nil, err
	}
	return as, nil
}
