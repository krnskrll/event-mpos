package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/syncconfig/repository"
	"github.com/krnskrll/event-mpos/internal/syncconfig/service"
)

// Handler exposes the admin sync administration under /api/admin/sync.
type Handler struct {
	cfg    *service.Service
	sched  *service.Scheduler
	erp    iface.StatusReporter
	member iface.StatusReporter
}

func NewHandler(cfg *service.Service, sched *service.Scheduler, erpSync iface.StatusReporter, memberSync iface.StatusReporter) *Handler {
	return &Handler{cfg: cfg, sched: sched, erp: erpSync, member: memberSync}
}

func (h *Handler) view(t iface.SyncType, cfg *repository.SyncConfig) iface.SyncTypeView {
	v := iface.SyncTypeView{Type: t, Enabled: cfg.Enabled, Cron: cfg.Cron, UpdatedAt: cfg.UpdatedAt}
	var r iface.LastRun
	if t == iface.TypeERP {
		r = h.erp.LastRun()
	} else {
		r = h.member.LastRun()
	}
	v.LastSyncAt, v.LastResult, v.Count, v.Source, v.Message = r.At, r.Result, r.Count, r.Source, r.Message
	return v
}

func (h *Handler) Overview(c *gin.Context) {
	erpCfg, err := h.cfg.Get(iface.TypeERP)
	if err != nil {
		common.Fail(c, err)
		return
	}
	memCfg, err := h.cfg.Get(iface.TypeMember)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, iface.SyncOverview{ERP: h.view(iface.TypeERP, erpCfg), Member: h.view(iface.TypeMember, memCfg)})
}

func (h *Handler) Update(c *gin.Context) {
	t, ok := iface.ParseType(c.Param("type"))
	if !ok {
		common.AbortError(c, http.StatusBadRequest, "type: must be one of ERP, MEMBER")
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
	c.JSON(http.StatusOK, h.view(t, cfg))
}
