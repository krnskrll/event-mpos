package bhoomy

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
)

const pageSize = 500

// CatalogSync pulls each linked store's Produk Toko changes into POS's copy, page by page; a page
// and its cursor are saved together, so a failed pull resumes where it stopped.
type CatalogSync struct {
	client    Client
	catalog   catalogiface.Catalog
	catalogue productiface.Catalogue
}

// NewCatalogSync pulls into catalog; Katalog POS items linked to changed products follow them.
func NewCatalogSync(client Client, catalog catalogiface.Catalog, catalogue productiface.Catalogue) *CatalogSync {
	return &CatalogSync{client: client, catalog: catalog, catalogue: catalogue}
}

func (s *CatalogSync) RunAll(ctx context.Context) {
	links, err := s.catalog.Links()
	if err != nil {
		log.Printf("bhoomy catalog sync: %v", err)
		return
	}
	for _, l := range links {
		if err := s.run(ctx, l); err != nil {
			log.Printf("bhoomy catalog sync %s: %v", l.StoreID, err)
		}
	}
}

func (s *CatalogSync) RunStore(ctx context.Context, storeID uuid.UUID) error {
	links, err := s.catalog.Links()
	if err != nil {
		return err
	}
	for _, l := range links {
		if l.StoreID == storeID {
			return s.run(ctx, l)
		}
	}
	return common.NewAppError(http.StatusConflict, "This store is not linked to Bhoomy yet")
}

func (s *CatalogSync) run(ctx context.Context, l catalogiface.Link) error {
	cur, err := s.catalog.Cursor(l.StoreID)
	if err != nil {
		return err
	}
	for {
		page, err := s.client.VariantChanges(ctx, l.BhoomyStoreID, cur, pageSize)
		if err != nil {
			if saveErr := s.catalog.SaveError(l.StoreID, err); saveErr != nil {
				log.Printf("bhoomy catalog sync %s: %v", l.StoreID, saveErr)
			}
			return common.NewAppError(http.StatusBadGateway, "Bhoomy did not answer: "+err.Error())
		}
		at := cur
		if n := len(page.Items); n > 0 {
			at = &catalogiface.Cursor{Since: page.Items[n-1].ChangedAt, After: page.Items[n-1].BhoomyVariantID}
		}
		if at != nil {
			if err := s.catalog.SavePage(l.StoreID, *page, *at); err != nil {
				return err
			}
			if err := s.catalogue.Reconcile(l.StoreID, productIDs(page.Items)); err != nil {
				return err
			}
		}
		if page.Next == nil {
			return nil
		}
		cur = page.Next
	}
}

func productIDs(items []catalogiface.Variant) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var out []uuid.UUID
	for _, v := range items {
		if !seen[v.BhoomyProductID] {
			seen[v.BhoomyProductID] = true
			out = append(out, v.BhoomyProductID)
		}
	}
	return out
}
