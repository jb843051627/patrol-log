package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"patrol-log/internal/model"
	"patrol-log/internal/store"
)

// Service 业务层
type Service struct {
	store *store.Store
}

func New(s *store.Store) *Service {
	return &Service{store: s}
}

// Summary 一次巡检的执行汇总
type Summary struct {
	PlanID   uint   `json:"plan_id"`
	DeviceID string `json:"device_id"`
	Total    int    `json:"total"`
	Abnormal int    `json:"abnormal"`
	Alerts   int    `json:"alerts"`
}

// 设备告警次数统计（内存态，仅用于快速查看）
var (
	alertMu     sync.Mutex
	alertCounts = map[string]int{}
)

func recordAlert(deviceID string) {
	alertMu.Lock()
	alertCounts[deviceID]++
	alertMu.Unlock()
}

// AlertCount 查询某设备的累计告警次数
func AlertCount(deviceID string) int {
	alertMu.Lock()
	defer alertMu.Unlock()
	return alertCounts[deviceID]
}

// CreatePlan 创建巡检计划；同名计划返回 ErrDuplicateName
func (s *Service) CreatePlan(ctx context.Context, name, deviceID string, interval int, items []model.CheckItem) (*model.Plan, error) {
	plan := &model.Plan{
		Name: name, DeviceID: deviceID, Interval: interval,
		Enabled: true, Items: items,
	}
	if err := s.store.CreatePlan(plan); err != nil {
		return nil, fmt.Errorf("create plan failed: %w", err)
	}
	return plan, nil
}

// GetPlan 查询计划（含巡检项）
func (s *Service) GetPlan(ctx context.Context, id uint) (*model.Plan, error) {
	return s.store.GetPlanByID(id)
}

// ExecutePatrol 执行一次巡检：逐项记录读数，越界生成告警
func (s *Service) ExecutePatrol(ctx context.Context, planID uint, deviceID string, readings map[string]float64) (*Summary, error) {
	plan, err := s.store.GetPlanByID(planID)
	if err != nil {
		return nil, err
	}
	var total, abnormal int
	for _, it := range plan.Items {
		// 模拟单条写入耗时
		time.Sleep(10 * time.Millisecond)
		reading, ok := readings[it.Name]
		if !ok {
			continue
		}
		total++
		isAbnormal := reading < it.WarnMin || reading > it.WarnMax
		r := &model.Result{
			PlanID: plan.ID, DeviceID: deviceID, ItemName: it.Name,
			Reading: reading, Abnormal: isAbnormal,
		}
		if err := s.store.CreateResult(r); err != nil {
			return nil, err
		}
		if isAbnormal {
			abnormal++
			a := &model.Alert{
				PlanID: plan.ID, DeviceID: deviceID, ItemName: it.Name,
				Reading: reading, Note: "reading out of range",
			}
			if err := s.store.CreateAlert(a); err != nil {
				return nil, err
			}
			recordAlert(deviceID)
		}
	}
	return &Summary{PlanID: plan.ID, DeviceID: deviceID, Total: total, Abnormal: abnormal, Alerts: abnormal}, nil
}

// ListResults 按设备与起始时间查询巡检结果
func (s *Service) ListResults(ctx context.Context, deviceID string, since time.Time) ([]model.Result, error) {
	return s.store.ListResults(deviceID, since)
}

// ListAlerts 查询某设备的告警记录
func (s *Service) ListAlerts(ctx context.Context, deviceID string) ([]model.Alert, error) {
	return s.store.ListAlerts(deviceID)
}
