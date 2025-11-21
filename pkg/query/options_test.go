package query

import "testing"

func TestOptionsNormalize(t *testing.T) {
	opts := Options{}.Normalize(50)
	if opts.Limit != 50 || opts.Offset != 0 {
		t.Fatalf("expected defaults applied, got %+v", opts)
	}

	opts = Options{Limit: 10, Offset: -5}.Normalize(20)
	if opts.Limit != 10 || opts.Offset != 0 {
		t.Fatalf("expected offset clamped to 0, got %+v", opts)
	}
}
