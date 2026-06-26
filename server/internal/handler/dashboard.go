package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	stats, err := h.dashboardService.GetStats(userID)
	if err != nil {
		response.InternalError(c, "Failed to get stats")
		return
	}

	response.Success(c, stats)
}

func (h *DashboardHandler) GetStudyTimeChart(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	period := c.DefaultQuery("period", "week")

	data, err := h.dashboardService.GetStudyTimeChart(userID, period)
	if err != nil {
		response.InternalError(c, "Failed to get chart data")
		return
	}

	response.Success(c, data)
}

func (h *DashboardHandler) GetCategoryStats(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	data, err := h.dashboardService.GetCategoryStats(userID)
	if err != nil {
		response.InternalError(c, "Failed to get category stats")
		return
	}

	response.Success(c, data)
}

func (h *DashboardHandler) GetCompletionRateChart(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	period := c.DefaultQuery("period", "week")

	data, err := h.dashboardService.GetCompletionRateChart(userID, period)
	if err != nil {
		response.InternalError(c, "Failed to get completion rate data")
		return
	}

	response.Success(c, data)
}
