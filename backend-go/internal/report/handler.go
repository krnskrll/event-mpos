package report

import (
	"net/http"
	"strconv"

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

const (
	defaultAuditPageSize = 50
	maxAuditPageSize     = 200
)

// Audit returns one page of an event's audit log, newest first:
// ?page (0-based, default 0) &size (default 50, capped at 200) &actor (user id) &action &orderId.
func (h *Handler) Audit(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	q := audit.Query{EventID: eventID, Page: 0, Size: defaultAuditPageSize}
	if raw := c.Query("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			common.AbortError(c, http.StatusBadRequest, "page must be a non-negative integer")
			return
		}
		q.Page = n
	}
	if raw := c.Query("size"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			common.AbortError(c, http.StatusBadRequest, "size must be a positive integer")
			return
		}
		q.Size = min(n, maxAuditPageSize)
	}
	for _, f := range []struct {
		param string
		dst   **uuid.UUID
	}{{"orderId", &q.OrderID}, {"actor", &q.ActorID}} {
		if raw := c.Query(f.param); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				common.AbortError(c, http.StatusBadRequest, "Invalid "+f.param)
				return
			}
			*f.dst = &id
		}
	}
	if raw := c.Query("action"); raw != "" {
		a, ok := audit.ParseAction(raw)
		if !ok {
			common.AbortError(c, http.StatusBadRequest, "Unknown action '"+raw+"'")
			return
		}
		q.Action = &a
	}
	page, err := h.audit.Search(q)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}
