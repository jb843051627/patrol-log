package service

import (
	"context"
	"time"

	"patrol-log/internal/model"
)

// ExportLatest 导出某设备最近 n 条巡检结果，并打上导出标记
func (s *Service) ExportLatest(ctx context.Context, deviceID string, n int) ([]model.Result, error) {
	rs, err := s.store.ListResults(deviceID, time.Time{})
	if err != nil {
		return nil, err
	}
	if n <= 0 || n > len(rs) {
		n = len(rs)
	}
	out := make([]model.Result, n)
	copy(out, rs[:n])
	for i := range out {
		out[i].Note = "exported"
	}
	return out, nil
}
