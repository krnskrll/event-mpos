package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/member/service"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// StoreList is the read-only back-office list; the QR code shows only its last 4 characters.
func (h *Handler) StoreList(c *gin.Context) {
	members, err := h.svc.List(authiface.StoreOf(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	for i := range members {
		members[i].QRCode = mask(members[i].QRCode)
	}
	c.JSON(http.StatusOK, members)
}

func mask(qr string) string {
	r := []rune(qr)
	if len(r) <= 4 {
		return "••••"
	}
	return "•••• " + string(r[len(r)-4:])
}
