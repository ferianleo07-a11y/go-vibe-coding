package report

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler menangani request HTTP untuk laporan metrik KPI & keuangan.
type Handler struct {
	service Service
	dbReady func() bool
}

// NewHandler membuat instance baru dari Report Handler.
func NewHandler(service Service, dbReady func() bool) *Handler {
	return &Handler{
		service: service,
		dbReady: dbReady,
	}
}

// GetDashboardMetrics menghitung metrik KPI hotel (Occupancy, ADR, RevPAR) dan rekapitulasi Laba/Rugi standar USALI.
func (h *Handler) GetDashboardMetrics(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	metrics, err := h.service.GetDashboardMetrics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung metrik dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": metrics,
	})
}
