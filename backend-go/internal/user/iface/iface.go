package iface

import (
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

// Role is one of the three system roles, serialized as e.g. "ADMIN".
type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleBA        Role = "BA"
	RoleWarehouse Role = "WAREHOUSE"
)

func ParseRole(s string) (Role, bool) {
	switch Role(s) {
	case RoleAdmin, RoleBA, RoleWarehouse:
		return Role(s), true
	}
	return "", false
}

// Account is a user as other modules see it; PasswordHash is only for auth.
type Account struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Role         Role
	Active       bool
}

// Users is the user module's contract for other modules. Finders return nil when not found.
type Users interface {
	FindByID(id uuid.UUID) (*Account, error)
	FindByEmail(email string) (*Account, error)
	NameOf(tx database.Tx, id uuid.UUID) (string, error)
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"fullName" binding:"required"`
	Role     Role   `json:"role" binding:"required"`
}

// UpdateUserRequest is a partial update: only non-nil fields are applied.
type UpdateUserRequest struct {
	FullName *string `json:"fullName"`
	Role     *Role   `json:"role"`
	Active   *bool   `json:"active"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	Role      Role      `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}
