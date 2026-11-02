package common

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// sensitiveQueryParams are redacted from access logs. The WebSocket upgrade carries the JWT as
// ?token= (browsers can't set headers on a WebSocket), so it must never reach the log output.
var sensitiveQueryParams = []string{"token"}

// AccessLogger is gin's request logger with the same line format as gin.Logger(), but with
// sensitive query parameters replaced by REDACTED.
func AccessLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(p gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %#v\n%s",
				p.TimeStamp.Format("2006/01/02 - 15:04:05"),
				p.StatusCode,
				p.Latency.Truncate(time.Microsecond),
				p.ClientIP,
				p.Method,
				RedactPath(p.Path),
				p.ErrorMessage,
			)
		},
	})
}

// RedactPath masks sensitive query parameter values in a request path ("/x?token=abc" →
// "/x?token=REDACTED"). Paths without a query are returned unchanged.
func RedactPath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}
	if u.RawQuery == "" {
		return path
	}
	q := u.Query()
	changed := false
	for _, k := range sensitiveQueryParams {
		if _, ok := q[k]; ok {
			q.Set(k, "REDACTED")
			changed = true
		}
	}
	if !changed {
		return path
	}
	u.RawQuery = q.Encode()
	return u.String()
}
