package service

import (
	"context"
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

// 导出不应污染后续查询到的原始数据（note 保持为空）
func TestBug002_ExportNoPollution(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	plan, err := svc.CreatePlan(ctx, "export-plan", "dev-01", 24, []model.CheckItem{
		{Name: "pressure", Unit: "MPa", WarnMin: 0.5, WarnMax: 1.5},
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := svc.ExecutePatrol(ctx, plan.ID, "dev-01",
			map[string]float64{"pressure": 1.0}); err != nil {
			t.Fatalf("execute patrol: %v", err)
		}
	}

	// 1) 首次查询（触发缓存）
	if _, err := svc.ListResults(ctx, "dev-01", time.Time{}); err != nil {
		t.Fatalf("list results: %v", err)
	}
	// 2) 导出最近 1 条（内部打上 exported 标记）
	if _, err := svc.ExportLatest(ctx, "dev-01", 1); err != nil {
		t.Fatalf("export latest: %v", err)
	}
	// 3) 再查历史，note 必须全部为空（未被导出污染）
	rs, err := svc.ListResults(ctx, "dev-01", time.Time{})
	if err != nil {
		t.Fatalf("list results again: %v", err)
	}
	if len(rs) == 0 {
		t.Fatalf("no results")
	}
	for _, r := range rs {
		if r.Note != "" {
			t.Fatalf("result %d note polluted: %q", r.ID, r.Note)
		}
	}
}