package iface

import "github.com/google/uuid"

type Receipts interface {
	LogPrint(orderID, locationID, staffID uuid.UUID) error
}
