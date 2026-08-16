package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"patrol-log/internal/model"
	"patrol-log/internal/service"
	"patrol-log/internal/store"
)

// Handler HTTP 处理器
type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

type createPlanReq struct {
	Name     string          `json:"name" binding:"required"`
	DeviceID string          `json:"device_id" binding:"required"`
	Interval int             `json:"interval"`
	Items    []model.CheckItem `json:"items"`
}

func (h *Handler) CreatePlan(c *gin.Context) {
	var req createPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.svc.CreatePlan(c.Request.Context(), req.Name, req.DeviceID, req.Interval, req.Items)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateName) {
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate plan name"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) GetPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	plan, err := h.svc.GetPlan(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

type executeReq struct {
	PlanID   uint               `json:"plan_id" binding:"required"`
	DeviceID string             `json:"device_id" binding:"required"`
	Readings map[string]float64 `json:"readings"`
}

func (h *Handler) ExecutePatrol(c *gin.Context) {
	var req executeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sum, err := h.svc.ExecutePatrol(c.Request.Context(), req.PlanID, req.DeviceID, req.Readings)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sum)
}

func (h *Handler) ListResults(c *gin.Context) {
	deviceID := c.Query("device_id")
	since := time.Time{}
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = t
		}
	}
	rs, err := h.svc.ListResults(c.Request.Context(), deviceID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rs)
}

func (h *Handler) ExportLatest(c *gin.Context) {
	deviceID := c.Query("device_id")
	n, _ := strconv.Atoi(c.DefaultQuery("n", "10"))
	rs, err := h.svc.ExportLatest(c.Request.Context(), deviceID, n)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rs)
}

func (h *Handler) ListAlerts(c *gin.Context) {
	deviceID := c.Query("device_id")
	as, err := h.svc.ListAlerts(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, as)
}

func (h *Handler) AlertCount(c *gin.Context) {
	deviceID := c.Query("device_id")
	c.JSON(http.StatusOK, gin.H{"device_id": deviceID, "alerts": service.AlertCount(deviceID)})
}
