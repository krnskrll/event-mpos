package iface

import "github.com/google/uuid"

type Receipts interface {
	LogPrint(orderID, baUserID uuid.UUID) error
}
