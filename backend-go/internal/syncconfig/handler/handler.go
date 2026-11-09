package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/syncconfig/repository"
	"github.com/krnskrll/event-mpos/internal/syncconfig/service"
)

// Handler serves the Member POS sync bar under /api/backoffice/members/sync.
type Handler struct {
	cfg    *service.Service
	sched  *service.Scheduler
	member iface.StatusReporter
	run    func()
}

func NewHandler(cfg *service.Service, sched *service.Scheduler, memberSync iface.StatusReporter, run func()) *Handler {
	return &Handler{cfg: cfg, sched: sched, member: memberSync, run: run}
}

func (h *Handler) view(cfg *repository.SyncConfig) iface.MemberSync {
	r := h.member.LastRun()
	return iface.MemberSync{Enabled: cfg.Enabled, Time: service.DailyTime(cfg.Cron),
		LastSyncAt: r.At, LastResult: r.Result, Count: r.Count, Source: r.Source, Message: r.Message}
}

func (h *Handler) status(c *gin.Context) {
	cfg, err := h.cfg.Get(iface.TypeMember)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, h.view(cfg))
}

func (h *Handler) Status(c *gin.Context) { h.status(c) }

// SyncNow runs the member sync and answers with the new status.
func (h *Handler) SyncNow(c *gin.Context) {
	h.run()
	h.status(c)
}

// Schedule turns the daily sync on or off at a WIB time.
func (h *Handler) Schedule(c *gin.Context) {
	var req iface.MemberScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	cron, err := service.DailyCron(req.Time)
	if err != nil {
		common.Fail(c, err)
		return
	}
	cfg, err := h.cfg.Update(iface.TypeMember, req.Enabled, cron)
	if err != nil {
		common.Fail(c, err)
		return
	}
	h.sched.Reschedule(iface.TypeMember, cfg.Cron)
	c.JSON(http.StatusOK, h.view(cfg))
}
