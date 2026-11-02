package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/auth/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// JWTMiddleware verifies the Bearer access token and attaches the Principal. A missing or invalid
// token yields 401 in the terse security-layer shape (matches Kotlin's authenticationEntryPoint).
func JWTMiddleware(jwt iface.TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			common.AbortUnauthorized(c)
			return
		}
		claims, ok := jwt.Parse(strings.TrimSpace(header[len("Bearer "):]))
		if !ok {
			common.AbortUnauthorized(c)
			return
		}
		sub, _ := claims["sub"].(string)
		uid, err := uuid.Parse(sub)
		if err != nil {
			common.AbortUnauthorized(c)
			return
		}
		roleStr, _ := claims["role"].(string)
		role, valid := useriface.ParseRole(roleStr)
		if !valid {
			common.AbortUnauthorized(c)
			return
		}
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)
		iface.SetPrincipal(c, iface.Principal{UserID: uid, StoreID: iface.StoreFromClaims(claims), Email: email, FullName: name, Role: role})
		c.Next()
	}
}

// RequireRole allows the request only if the caller holds one of the given roles; otherwise 403.
func RequireRole(roles ...useriface.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := iface.PrincipalFrom(c)
		if !ok {
			common.AbortUnauthorized(c)
			return
		}
		for _, r := range roles {
			if p.Role == r {
				c.Next()
				return
			}
		}
		common.AbortForbidden(c)
	}
}
