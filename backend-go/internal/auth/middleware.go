package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/user"
)

const principalKey = "authPrincipal"

// Principal is the authenticated user attached to the request context.
type Principal struct {
	UserID   uuid.UUID
	Email    string
	FullName string
	Role     user.Role
}

// JWTMiddleware verifies the Bearer access token and attaches the Principal. A missing or invalid
// token yields 401 in the terse security-layer shape (matches Kotlin's authenticationEntryPoint).
func JWTMiddleware(jwt *JwtService) gin.HandlerFunc {
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
		role, valid := user.ParseRole(roleStr)
		if !valid {
			common.AbortUnauthorized(c)
			return
		}
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)
		c.Set(principalKey, Principal{UserID: uid, Email: email, FullName: name, Role: role})
		c.Next()
	}
}

// RequireRole allows the request only if the caller holds one of the given roles; otherwise 403.
func RequireRole(roles ...user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := PrincipalFrom(c)
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

// PrincipalFrom retrieves the authenticated Principal from the context.
func PrincipalFrom(c *gin.Context) (Principal, bool) {
	v, exists := c.Get(principalKey)
	if !exists {
		return Principal{}, false
	}
	p, ok := v.(Principal)
	return p, ok
}
