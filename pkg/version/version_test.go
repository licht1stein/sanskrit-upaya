package version

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		{"v1.0.1", "v1.0.0", true},
		{"v1.1.0", "v1.0.0", true},
		{"v2.0.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.0.1", false},
		{"v1.0.0", "v1.1.0", false},
		{"v1.0.0", "v2.0.0", false},
		{"1.0.1", "1.0.0", true},   // without v prefix
		{"v1.0.1", "1.0.0", true},  // mixed prefixes
		{"1.0.1", "v1.0.0", true},  // mixed prefixes
		{"v1.0.10", "v1.0.9", true},
		{"v1.10.0", "v1.9.0", true},
		{"v10.0.0", "v9.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			got := isNewer(tt.latest, tt.current)
			if got != tt.expected {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"1.0.0", []int{1, 0, 0}},
		{"1.2.3", []int{1, 2, 3}},
		{"10.20.30", []int{10, 20, 30}},
		{"1", []int{1}},
		{"1.0", []int{1, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseVersion(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.expected)
					return
				}
			}
		})
	}
}
