package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/media"
	"github.com/krnskrll/event-mpos/internal/product/service"
)

type Handler struct {
	svc    *service.Service
	events eventiface.Events
}

func NewHandler(svc *service.Service, events eventiface.Events) *Handler {
	return &Handler{svc: svc, events: events}
}

func (h *Handler) UploadPhoto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid product id")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, media.MaxBytes+64<<10)
	fh, err := c.FormFile("file")
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "file: required, 2 MB or less")
		return
	}
	f, err := fh.Open()
	if err != nil {
		common.Fail(c, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, media.MaxBytes+1))
	if err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.svc.SetPhoto(authiface.StoreOf(c), id, data)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
