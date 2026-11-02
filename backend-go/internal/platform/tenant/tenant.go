// Package tenant holds the store every POS row belongs to until real Bhoomy stores are linked.
package tenant

import "github.com/google/uuid"

// DefaultStoreID matches the store created by migration 000004.
var DefaultStoreID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
