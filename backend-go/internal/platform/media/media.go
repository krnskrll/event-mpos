// Package media keeps uploaded images (POS photos) in a local folder and serves them; an S3 store
// can replace it later behind the same methods.
package media

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/common"
)

const MaxBytes = 2 << 20

var extensions = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

var fileName = regexp.MustCompile(`^[0-9a-f-]{36}\.(jpg|png|webp)$`)

type Store struct{ dir, baseURL string }

func NewStore(dir, baseURL string) *Store {
	return &Store{dir: dir, baseURL: strings.TrimRight(baseURL, "/")}
}

// Save checks the image type from its content, stores it under a random name and returns its URL.
func (s *Store) Save(data []byte) (string, error) {
	if len(data) > MaxBytes {
		return "", common.NewAppError(http.StatusBadRequest, "file: must be 2 MB or less")
	}
	ext, ok := extensions[http.DetectContentType(data)]
	if !ok {
		return "", common.NewAppError(http.StatusBadRequest, "file: must be a JPEG, PNG or WebP image")
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", err
	}
	name := uuid.NewString() + ext
	if err := os.WriteFile(filepath.Join(s.dir, name), data, 0o644); err != nil {
		return "", err
	}
	return s.baseURL + "/" + name, nil
}

// Serve answers GET /media/:file; names other than the ones Save creates are 404.
func (s *Store) Serve(c *gin.Context) {
	name := c.Param("file")
	if !fileName.MatchString(name) {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(filepath.Join(s.dir, name))
}
