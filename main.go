package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"patrol-log/internal/api"
	"patrol-log/internal/model"
	"patrol-log/internal/service"
	"patrol-log/internal/store"
)

func main() {
	db, err := gorm.Open(sqlite.Open("patrol.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Plan{}, &model.CheckItem{}, &model.Result{}, &model.Alert{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	st := store.New(db)
	svc := service.New(st)
	h := api.New(svc)

	r := gin.Default()
	r.POST("/api/plans", h.CreatePlan)
	r.GET("/api/plans/:id", h.GetPlan)
	r.POST("/api/patrols", h.ExecutePatrol)
	r.GET("/api/results", h.ListResults)
	r.GET("/api/export", h.ExportLatest)
	r.GET("/api/alerts", h.ListAlerts)
	r.GET("/api/alerts/count", h.AlertCount)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
