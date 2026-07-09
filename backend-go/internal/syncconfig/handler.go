package syncconfig

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/common"
	membersync "github.com/krnskrll/event-mpos/internal/member/sync"
	"github.com/krnskrll/event-mpos/internal/product/erp"
)

// Handler exposes the admin sync administration under /api/admin/sync.
type Handler struct {
	cfg    *Service
	sched  *Scheduler
	erp    *erp.SyncService
	member *membersync.Service
}

func NewHandler(cfg *Service, sched *Scheduler, erpSync *erp.SyncService, memberSync *membersync.Service) *Handler {
	return &Handler{cfg: cfg, sched: sched, erp: erpSync, member: memberSync}
}

func (h *Handler) view(t SyncType, cfg *SyncConfig) SyncTypeView {
	v := SyncTypeView{Type: t, Enabled: cfg.Enabled, Cron: cfg.Cron, UpdatedAt: cfg.UpdatedAt}
	if t == TypeERP {
		s := h.erp.Status()
		v.LastSyncAt, v.LastResult, v.Count, v.Source, v.Message = s.LastSyncAt, s.LastResult, s.ProductCount, s.Source, s.Message
	} else {
		s := h.member.Status()
		v.LastSyncAt, v.LastResult, v.Count, v.Source, v.Message = s.LastSyncAt, s.LastResult, s.MemberCount, s.Source, s.Message
	}
	return v
}

func (h *Handler) Overview(c *gin.Context) {
	erpCfg, err := h.cfg.Get(TypeERP)
	if err != nil {
		common.Fail(c, err)
		return
	}
	memCfg, err := h.cfg.Get(TypeMember)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, SyncOverview{ERP: h.view(TypeERP, erpCfg), Member: h.view(TypeMember, memCfg)})
}

func (h *Handler) Update(c *gin.Context) {
	t, ok := ParseType(c.Param("type"))
	if !ok {
		common.AbortError(c, http.StatusBadRequest, "type: must be one of ERP, MEMBER")
		return
	}
	var req UpdateSyncConfigRequest
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
