package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

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

// 请求取消后：巡检应中止（返回 context 错误），不能继续写完全部结果
func TestBug005_PatrolCancellation(t *testing.T) {
	svc := newTestService(t)

	plan, err := svc.CreatePlan(context.Background(), "cancel-plan", "dev-01", 24, []model.CheckItem{
		{Name: "p1", Unit: "u", WarnMin: 0, WarnMax: 100},
		{Name: "p2", Unit: "u", WarnMin: 0, WarnMax: 100},
		{Name: "p3", Unit: "u", WarnMin: 0, WarnMax: 100},
		{Name: "p4", Unit: "u", WarnMin: 0, WarnMax: 100},
		{Name: "p5", Unit: "u", WarnMin: 0, WarnMax: 100},
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	readings := map[string]float64{"p1": 1, "p2": 1, "p3": 1, "p4": 1, "p5": 1}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = svc.ExecutePatrol(ctx, plan.ID, "dev-01", readings)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want DeadlineExceeded, got %v", err)
	}

	rs, err := svc.ListResults(context.Background(), "dev-01", time.Time{})
	if err != nil {
		t.Fatalf("list results: %v", err)
	}
	if len(rs) >= 5 {
		t.Fatalf("cancelled patrol still wrote all results: got %d", len(rs))
	}
}