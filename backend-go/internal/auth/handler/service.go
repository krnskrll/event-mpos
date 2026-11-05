package handler

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// ServiceMiddleware admits the Bhoomy backend, which has already checked the owner or manager's
// login and POS access: the shared service token plus the Bhoomy store they work in. storeFor
// finds the POS store, creating it on the store's first visit.
func ServiceMiddleware(token string, storeFor func(bhoomyStoreID uuid.UUID, name string) (uuid.UUID, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		got := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			common.AbortUnauthorized(c)
			return
		}
		bhoomyStore, err := uuid.Parse(c.GetHeader("X-Bhoomy-Store-Id"))
		if err != nil {
			common.AbortError(c, http.StatusBadRequest, "X-Bhoomy-Store-Id: required")
			return
		}
		name, _ := url.QueryUnescape(c.GetHeader("X-Bhoomy-Store-Name"))
		store, err := storeFor(bhoomyStore, strings.TrimSpace(name))
		if err != nil {
			common.Fail(c, err)
			return
		}
		iface.SetPrincipal(c, iface.Principal{StoreID: store, Role: useriface.RoleAdmin})
		c.Next()
	}
}
