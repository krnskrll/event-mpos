package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/report/service"
)

// Handler exposes the admin reporting + audit surface under /api/admin.
type Handler struct {
	svc    *service.Service
	audit  auditiface.Audit
	events eventiface.Events
}

func NewHandler(svc *service.Service, aud auditiface.Audit, events eventiface.Events) *Handler {
	return &Handler{svc: svc, audit: aud, events: events}
}

func (h *Handler) Report(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
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
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
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
// ?page (0-based, default 0) &size (default 50, capped at 200) &actor (user id) &action &orderId &from &to.
func (h *Handler) Audit(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	q := auditiface.Query{EventID: eventID, Page: 0, Size: defaultAuditPageSize}
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
	if c.Query("from") != "" || c.Query("to") != "" {
		from, to, ok := period(c)
		if !ok {
			return
		}
		q.From, q.To = from, to.AddDate(0, 0, 1)
	}
	if raw := c.Query("action"); raw != "" {
		a, ok := auditiface.ParseAction(raw)
		if !ok {
			common.AbortError(c, http.StatusBadRequest, "Unknown action '"+raw+"'")
			return
		}
		q.Action = &a
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), eventID); err != nil {
		common.Fail(c, err)
		return
	}
	page, err := h.audit.Search(q)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// Sales is the back-office report of one location: ?from &to are WIB dates, default today.
func (h *Handler) Sales(c *gin.Context) {
	locationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid location id")
		return
	}
	from, to, ok := period(c)
	if !ok {
		return
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), locationID); err != nil {
		common.Fail(c, err)
		return
	}
	rep, err := h.svc.Sales(locationID, from, to, time.Now())
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

// period reads ?from &to (YYYY-MM-DD, WIB, both included, at most a year apart).
func period(c *gin.Context) (time.Time, time.Time, bool) {
	today := service.Today(time.Now())
	from, err1 := service.Day(c.DefaultQuery("from", today))
	to, err2 := service.Day(c.DefaultQuery("to", today))
	if err1 != nil || err2 != nil || to.Before(from) || to.Sub(from) > 366*24*time.Hour {
		common.AbortError(c, http.StatusBadRequest, "from and to must be dates (YYYY-MM-DD), from first, at most a year apart")
		return from, to, false
	}
	return from, to, true
}
