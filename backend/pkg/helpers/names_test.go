package helpers

import "testing"

func TestNormalizateNames(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"single lowercase name", "john", "John"},
		{"single name with surrounding spaces", "  john  ", "John"},
		{"already capitalized", "John", "John"},
		{"full name lowercase", "john doe", "John Doe"},
		{"full name uppercase", "JOHN DOE", "John Doe"},
		{"mixed case", "jOhN dOe", "John Doe"},
		{"three names", "juan carlos pérez", "Juan Carlos Pérez"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizateNames(tc.input); got != tc.want {
				t.Errorf("NormalizateNames(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	cases := []struct {
		name  string
		count uint
		limit uint
		want  uint
	}{
		{"exact division", 100, 10, 10},
		{"partial last page", 101, 10, 11},
		{"fewer items than limit", 3, 10, 1},
		{"zero items", 0, 10, 0},
		{"single item", 1, 1, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CalculateTotalPages(tc.count, tc.limit); got != tc.want {
				t.Errorf("CalculateTotalPages(%d, %d) = %d, want %d", tc.count, tc.limit, got, tc.want)
			}
		})
	}
}
