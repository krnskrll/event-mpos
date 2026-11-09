package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/syncconfig/repository"
	"github.com/krnskrll/event-mpos/internal/syncconfig/service"
)

// Handler exposes the member sync schedule under /api/admin/sync.
type Handler struct {
	cfg    *service.Service
	sched  *service.Scheduler
	member iface.StatusReporter
}

func NewHandler(cfg *service.Service, sched *service.Scheduler, memberSync iface.StatusReporter) *Handler {
	return &Handler{cfg: cfg, sched: sched, member: memberSync}
}

func (h *Handler) view(cfg *repository.SyncConfig) iface.SyncTypeView {
	v := iface.SyncTypeView{Type: iface.TypeMember, Enabled: cfg.Enabled, Cron: cfg.Cron, UpdatedAt: cfg.UpdatedAt}
	r := h.member.LastRun()
	v.LastSyncAt, v.LastResult, v.Count, v.Source, v.Message = r.At, r.Result, r.Count, r.Source, r.Message
	return v
}

func (h *Handler) Overview(c *gin.Context) {
	cfg, err := h.cfg.Get(iface.TypeMember)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, iface.SyncOverview{Member: h.view(cfg)})
}

func (h *Handler) Update(c *gin.Context) {
	t := iface.SyncType(c.Param("type"))
	switch {
	case t == "ERP":
		common.AbortError(c, http.StatusBadRequest, "ERP product sync is retired")
		return
	case t != iface.TypeMember:
		common.AbortError(c, http.StatusBadRequest, "type: must be MEMBER")
		return
	}
	var req iface.UpdateSyncConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return
	}
	cfg, err := h.cfg.Update(t, req.Enabled, req.Cron)
	if err != nil {
		common.Fail(c, err)
		return
	}
	h.sched.Reschedule(t, cfg.Cron)
	c.JSON(http.StatusOK, h.view(cfg))
}
