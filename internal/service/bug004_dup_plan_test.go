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

// 同名计划第二次创建：错误必须能被 errors.Is 识别为 ErrDuplicateName
func TestBug004_DuplicatePlanErrorChain(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, err := svc.CreatePlan(ctx, "dup-plan", "dev-1", 24, nil); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := svc.CreatePlan(ctx, "dup-plan", "dev-2", 24, nil)
	if !errors.Is(err, store.ErrDuplicateName) {
		t.Fatalf("want ErrDuplicateName via errors.Is, got %v", err)
	}
}