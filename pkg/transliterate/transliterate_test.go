package transliterate

import "testing"

func TestIASTToDevanagari(t *testing.T) {
	tests := []struct {
		iast string
		want string
	}{
		{"a", "अ"},
		{"ka", "क"},
		{"ki", "कि"},
		{"ku", "कु"},
		{"kā", "का"},
		{"kṛ", "कृ"},
		{"ke", "के"},
		{"ko", "को"},
		{"kai", "कै"},
		{"kau", "कौ"},
		{"ṃ", "ं"},
		{"ḥ", "ः"},
		{"saṃskṛta", "संस्कृत"},
		{"namaste", "नमस्ते"},
		{"yoga", "योग"},
		{"dharma", "धर्म"},
		{"karma", "कर्म"},
		{"janaka", "जनक"},
		{"rāma", "राम"},
		{"kṛṣṇa", "कृष्ण"},
	}

	for _, tt := range tests {
		t.Run(tt.iast, func(t *testing.T) {
			got := IASTToDevanagari(tt.iast)
			if got != tt.want {
				t.Errorf("IASTToDevanagari(%q) = %q, want %q", tt.iast, got, tt.want)
			}
		})
	}
}

func TestIsDevanagari(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"संस्कृत", true},
		{"sanskrit", false},
		{"saṃskṛta", false},
		{"नमस्ते", true},
		{"hello", false},
		{"राम", true},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := IsDevanagari(tt.s)
			if got != tt.want {
				t.Errorf("IsDevanagari(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToSearchTerms(t *testing.T) {
	tests := []struct {
		query    string
		wantLen  int
		contains string
	}{
		{"janaka", 2, "जनक"},
		{"जनक", 1, "जनक"},
		{"rāma", 2, "राम"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			terms := ToSearchTerms(tt.query)
			if len(terms) < tt.wantLen {
				t.Errorf("ToSearchTerms(%q) returned %d terms, want at least %d", tt.query, len(terms), tt.wantLen)
			}
			found := false
			for _, term := range terms {
				if term == tt.contains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ToSearchTerms(%q) = %v, should contain %q", tt.query, terms, tt.contains)
			}
		})
	}
}
