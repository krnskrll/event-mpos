package iface

import (
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

// Role of a user account. Only ADMIN signs in (the old dashboard); cashiers are staff on paired
// devices, and older BA/WAREHOUSE rows can no longer log in.
type Role string

const RoleAdmin Role = "ADMIN"

func ParseRole(s string) (Role, bool) {
	if Role(s) == RoleAdmin {
		return RoleAdmin, true
	}
	return "", false
}

// Account is a user as other modules see it; PasswordHash is only for auth.
type Account struct {
	ID           uuid.UUID
	StoreID      uuid.UUID
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

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	StoreID   uuid.UUID `json:"storeId"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	Role      Role      `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}
