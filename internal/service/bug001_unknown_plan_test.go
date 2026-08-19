package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"patrol-log/internal/model"
	"patrol-log/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	// 每次测试用独立临时文件库，避免共享内存库在 -count=N 多次迭代间残留数据
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Windows 上文件句柄不关会导致 TempDir 清理失败，测试结束前显式关闭
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&model.Plan{}, &model.CheckItem{}, &model.Result{}, &model.Alert{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(store.New(db))
}

// 对不存在的计划执行巡检：不应 panic，应返回 ErrNotFound
func TestBug001_ExecutePatrolUnknownPlan(t *testing.T) {
	svc := newTestService(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, err := svc.ExecutePatrol(context.Background(), 9999, "dev-1", map[string]float64{"pressure": 1.0})
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}