package pagination

import "testing"

func TestNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Page
		want Page
	}{
		{name: "defaults", in: Page{}, want: Page{Limit: DefaultLimit, Offset: 0}},
		{name: "max limit", in: Page{Limit: 999, Offset: 5}, want: Page{Limit: MaxLimit, Offset: 5}},
		{name: "negative offset", in: Page{Limit: 10, Offset: -1}, want: Page{Limit: 10, Offset: 0}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Normalize(tt.in); got != tt.want {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}
