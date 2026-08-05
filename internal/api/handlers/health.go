package handlers

import (
	"context"

	"github.com/76Parker/metrico/internal/applications/health"
	"github.com/76Parker/metrico/pkg/logger"
	"github.com/gin-gonic/gin"
)

type healthService interface {
	CheckAvailability(ctx context.Context) health.AvailabilityResult
}
type HealthHandler struct {
	log     logger.Logger
	service healthService
}

func NewHealthHandler(log logger.Logger, service healthService) *HealthHandler {
	return &HealthHandler{log: log, service: service}
}

func (h *HealthHandler) CheckAvailability(c *gin.Context) {
	ctx := c.Request.Context()
	result := h.service.CheckAvailability(ctx)
	if result.Ready {
		c.Status(200)
		return
	}
	for _, check := range result.AvailabilityResults {
		if !check.IsAvailable {
			h.log.Error("System is not available", "system", check.System, "error", check.Error)
		}
	}
	c.Status(500)
}
