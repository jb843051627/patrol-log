package service

import (
	"context"
	"path/filepath"
	"sync"
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

// 并发执行巡检（产生告警）不应有 data race
func TestBug003_ConcurrentPatrolRace(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	plan, err := svc.CreatePlan(ctx, "race-plan", "dev-race", 24, []model.CheckItem{
		{Name: "temp", Unit: "C", WarnMin: 10, WarnMax: 60},
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	// 读数越界 → 每次巡检都生成告警 → 写共享统计 map
	readings := map[string]float64{"temp": 95}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				if _, err := svc.ExecutePatrol(ctx, plan.ID, "dev-race", readings); err != nil {
					t.Errorf("execute patrol: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}