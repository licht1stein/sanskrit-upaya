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

func TestDevanagariToIAST(t *testing.T) {
	tests := []struct {
		deva string
		want string
	}{
		{"अ", "a"},
		{"क", "ka"},
		{"कि", "ki"},
		{"कु", "ku"},
		{"का", "kā"},
		{"कृ", "kṛ"},
		{"के", "ke"},
		{"को", "ko"},
		{"कै", "kai"},
		{"कौ", "kau"},
		{"ं", "ṃ"},
		{"ः", "ḥ"},
		{"संस्कृत", "saṃskṛta"},
		{"नमस्ते", "namaste"},
		{"योग", "yoga"},
		{"धर्म", "dharma"},
		{"कर्म", "karma"},
		{"जनक", "janaka"},
		{"राम", "rāma"},
		{"कृष्ण", "kṛṣṇa"},
	}

	for _, tt := range tests {
		t.Run(tt.deva, func(t *testing.T) {
			got := DevanagariToIAST(tt.deva)
			if got != tt.want {
				t.Errorf("DevanagariToIAST(%q) = %q, want %q", tt.deva, got, tt.want)
			}
		})
	}
}

func TestRoundtrip(t *testing.T) {
	// IAST -> Devanagari -> IAST should return original
	iastTests := []string{
		"a", "ā", "i", "ī", "u", "ū",
		"ṛ", "ṝ", "ḷ",
		"e", "ai", "o", "au",
		"ka", "kha", "ga", "gha", "ṅa",
		"ca", "cha", "ja", "jha", "ña",
		"ṭa", "ṭha", "ḍa", "ḍha", "ṇa",
		"ta", "tha", "da", "dha", "na",
		"pa", "pha", "ba", "bha", "ma",
		"ya", "ra", "la", "va",
		"śa", "ṣa", "sa", "ha",
		"saṃskṛta", "namaste", "yoga", "dharma", "karma",
		"janaka", "rāma", "kṛṣṇa", "āśrama", "upaniṣad",
	}

	for _, iast := range iastTests {
		t.Run("IAST_"+iast, func(t *testing.T) {
			deva := IASTToDevanagari(iast)
			back := DevanagariToIAST(deva)
			if back != iast {
				t.Errorf("Roundtrip failed: %q -> %q -> %q", iast, deva, back)
			}
		})
	}

	// Devanagari -> IAST -> Devanagari should return original
	devaTests := []string{
		"अ", "आ", "इ", "ई", "उ", "ऊ",
		"ऋ", "ॠ", "ऌ",
		"ए", "ऐ", "ओ", "औ",
		"क", "ख", "ग", "घ", "ङ",
		"च", "छ", "ज", "झ", "ञ",
		"ट", "ठ", "ड", "ढ", "ण",
		"त", "थ", "द", "ध", "न",
		"प", "फ", "ब", "भ", "म",
		"य", "र", "ल", "व",
		"श", "ष", "स", "ह",
		"संस्कृत", "नमस्ते", "योग", "धर्म", "कर्म",
		"जनक", "राम", "कृष्ण", "आश्रम", "उपनिषद्",
	}

	for _, deva := range devaTests {
		t.Run("Deva_"+deva, func(t *testing.T) {
			iast := DevanagariToIAST(deva)
			back := IASTToDevanagari(iast)
			if back != deva {
				t.Errorf("Roundtrip failed: %q -> %q -> %q", deva, iast, back)
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
