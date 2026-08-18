package centreon

import "testing"

func TestNilToEmpty(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"nil input", nil},
		{"empty non-nil", []int{}},
		{"non-empty passthrough", []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nilToEmpty(tt.in)
			if got == nil {
				t.Fatalf("nilToEmpty(%v) returned nil; want non-nil", tt.in)
			}
			if len(got) != len(tt.in) {
				t.Fatalf("len = %d; want %d", len(got), len(tt.in))
			}
			for i := range tt.in {
				if got[i] != tt.in[i] {
					t.Fatalf("element %d = %d; want %d", i, got[i], tt.in[i])
				}
			}
			// A non-nil input is returned as-is (same backing array), never copied.
			if len(tt.in) > 0 && &got[0] != &tt.in[0] {
				t.Fatal("non-nil input was copied; want the same slice returned")
			}
		})
	}
}
