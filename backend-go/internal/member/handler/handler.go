package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krnskrll/event-mpos/internal/member/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
	members, err := h.svc.List()
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}
