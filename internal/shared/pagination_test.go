package shared

import "testing"

func TestNormalizePage(t *testing.T) {
	page := NormalizePage(0, 500)
	if page.Page != 1 || page.PageSize != 100 {
		t.Fatalf("unexpected page: %+v", page)
	}
}
