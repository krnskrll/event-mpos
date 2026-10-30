package report

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/common"
)

// Handler exposes the admin reporting + audit surface under /api/admin.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, aud *audit.Service) *Handler {
	return &Handler{svc: svc, audit: aud}
}

func (h *Handler) Report(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	rep, err := h.svc.EndOfEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

func (h *Handler) ReportCSV(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	rep, err := h.svc.EndOfEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\"report-"+rep.EventCode+".csv\"")
	c.Data(http.StatusOK, "text/csv", []byte(h.svc.ToCSV(rep)))
}

func (h *Handler) Audit(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if raw := c.Query("orderId"); raw != "" {
		orderID, err := uuid.Parse(raw)
		if err != nil {
			common.AbortError(c, http.StatusBadRequest, "Invalid order id")
			return
		}
		entries, err := h.audit.ForOrder(orderID)
		if err != nil {
			common.Fail(c, err)
			return
		}
		c.JSON(http.StatusOK, entries)
		return
	}
	entries, err := h.audit.ForEvent(eventID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, entries)
}
