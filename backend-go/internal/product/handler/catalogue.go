package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authiface "github.com/krnskrll/event-mpos/internal/auth/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/product/iface"
)

// Katalog POS and the products of a location, for the Bhoomy back office.

func (h *Handler) CatalogueSource(c *gin.Context) {
	reply(c, http.StatusOK)(h.svc.CatalogueSource(authiface.StoreOf(c), c.Query("q")))
}

func (h *Handler) CatalogueList(c *gin.Context) {
	reply(c, http.StatusOK)(h.svc.CatalogueItems(authiface.StoreOf(c)))
}

func (h *Handler) CatalogueAdd(c *gin.Context) {
	var req iface.AddItemsRequest
	if !bind(c, &req) {
		return
	}
	ids, err := h.svc.AddFromProdukToko(authiface.StoreOf(c), req.ProductIDs)
	reply(c, http.StatusCreated)(iface.AddItemsResponse{IDs: ids}, err)
}

func (h *Handler) CatalogueGet(c *gin.Context) {
	id, ok := param(c, "id")
	if !ok {
		return
	}
	store := authiface.StoreOf(c)
	events, err := h.events.ListEvents(store)
	if err != nil {
		common.Fail(c, err)
		return
	}
	locs := make([]iface.ItemLocation, 0, len(events))
	for _, e := range events {
		if e.Status != eventiface.StatusClosed {
			locs = append(locs, iface.ItemLocation{ID: e.ID, Name: e.Name, Type: string(e.Type), Status: string(e.Status)})
		}
	}
	reply(c, http.StatusOK)(h.svc.CatalogueItem(store, id, locs))
}

func (h *Handler) CatalogueUpdate(c *gin.Context) {
	id, ok := param(c, "id")
	var req iface.UpdateItemRequest
	if !ok || !bind(c, &req) {
		return
	}
	store := authiface.StoreOf(c)
	for _, l := range req.Locations {
		if err := h.events.RequireInStore(store, l.ID); err != nil {
			common.Fail(c, err)
			return
		}
	}
	if err := h.svc.UpdateCatalogueItem(store, id, req); err != nil {
		common.Fail(c, err)
		return
	}
	h.CatalogueGet(c)
}

func (h *Handler) Locations(c *gin.Context) {
	store := authiface.StoreOf(c)
	events, err := h.events.ListEvents(store)
	if err != nil {
		common.Fail(c, err)
		return
	}
	counts, err := h.svc.LocationItemCounts(store)
	out := make([]eventiface.LocationResponse, 0, len(events))
	for _, e := range events {
		out = append(out, eventiface.LocationResponse{EventResponse: e, Items: counts[e.ID]})
	}
	reply(c, http.StatusOK)(out, err)
}

func (h *Handler) LocationRows(c *gin.Context) {
	id, ok := h.location(c)
	if ok {
		reply(c, http.StatusOK)(h.svc.LocationRows(authiface.StoreOf(c), id))
	}
}

func (h *Handler) SaveLocationRows(c *gin.Context) {
	id, ok := h.location(c)
	var req iface.SaveLocationRowsRequest
	if !ok || !bind(c, &req) {
		return
	}
	if err := h.svc.SaveLocationRows(authiface.StoreOf(c), id, req.Rows); err != nil {
		common.Fail(c, err)
		return
	}
	h.LocationRows(c)
}

func (h *Handler) AddLocationItems(c *gin.Context) {
	id, ok := h.location(c)
	var req iface.AddLocationItemsRequest
	if !ok || !bind(c, &req) {
		return
	}
	if err := h.svc.AddToLocation(authiface.StoreOf(c), id, req.ItemIDs); err != nil {
		common.Fail(c, err)
		return
	}
	h.LocationRows(c)
}

// location reads the :id location and checks it belongs to the caller's store.
func (h *Handler) location(c *gin.Context) (uuid.UUID, bool) {
	id, ok := param(c, "id")
	if !ok {
		return id, false
	}
	if err := h.events.RequireInStore(authiface.StoreOf(c), id); err != nil {
		common.Fail(c, err)
		return id, false
	}
	return id, true
}

func param(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		common.AbortError(c, http.StatusBadRequest, "Invalid "+name)
	}
	return id, err == nil
}

func bind(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		common.AbortError(c, http.StatusBadRequest, common.BindingMessage(err))
		return false
	}
	return true
}

// reply writes a service result: reply(c, 200)(svc.Get(...)).
func reply(c *gin.Context, status int) func(any, error) {
	return func(v any, err error) {
		if err != nil {
			common.Fail(c, err)
			return
		}
		c.JSON(status, v)
	}
}
