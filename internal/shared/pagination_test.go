package shared

import "testing"

func TestNormalizePage(t *testing.T) {
	p := NormalizePage(0, 999)
	if p.Page != 1 || p.PageSize != MaxPageSize {
		t.Fatalf("unexpected page: %+v", p)
	}
}
