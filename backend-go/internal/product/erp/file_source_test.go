package erp

import "testing"

func TestFileSourceReadsSampleCSV(t *testing.T) {
	products, err := FileSource{Path: "../../../samples/erp/products.csv"}.FetchProducts()
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 5 {
		t.Fatalf("got %d products, want 5", len(products))
	}
	if products[0].SKU != "SKU-1001" || len(products[0].Barcodes) != 2 {
		t.Errorf("first product = %+v", products[0])
	}
}
