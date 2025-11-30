// Package transliterate provides IAST to Devanagari transliteration.
package transliterate

import (
	"strings"
	"unicode"
)

// IAST to SLP1 mapping
var iastToSLP = map[string]string{
	// Vowels
	"a": "a", "ā": "A", "i": "i", "ī": "I", "u": "u", "ū": "U",
	"ṛ": "f", "ṝ": "F", "ḷ": "x", "ḹ": "X",
	"e": "e", "ai": "E", "o": "o", "au": "O",
	// Anusvara and Visarga
	"ṃ": "M", "ḥ": "H", "~": "~",
	// Velars
	"k": "k", "kh": "K", "g": "g", "gh": "G", "ṅ": "N",
	// Palatals
	"c": "c", "ch": "C", "j": "j", "jh": "J", "ñ": "Y",
	// Retroflexes
	"ṭ": "w", "ṭh": "W", "ḍ": "q", "ḍh": "Q", "ṇ": "R",
	// Dentals
	"t": "t", "th": "T", "d": "d", "dh": "D", "n": "n",
	// Labials
	"p": "p", "ph": "P", "b": "b", "bh": "B", "m": "m",
	// Semivowels
	"y": "y", "r": "r", "l": "l", "v": "v",
	// Sibilants
	"ś": "S", "ṣ": "z", "s": "s", "h": "h",
	// Avagraha
	"'": "'",
}

// SLP1 to Devanagari mapping
var slpToDeva = map[rune]string{
	// Vowels (independent)
	'a': "अ", 'A': "आ", 'i': "इ", 'I': "ई", 'u': "उ", 'U': "ऊ",
	'f': "ऋ", 'F': "ॠ", 'x': "ऌ", 'X': "ॡ",
	'e': "ए", 'E': "ऐ", 'o': "ओ", 'O': "औ",
	// Anusvara and Visarga
	'M': "ं", 'H': "ः", '~': "ँ",
	// Velars
	'k': "क", 'K': "ख", 'g': "ग", 'G': "घ", 'N': "ङ",
	// Palatals
	'c': "च", 'C': "छ", 'j': "ज", 'J': "झ", 'Y': "ञ",
	// Retroflexes
	'w': "ट", 'W': "ठ", 'q': "ड", 'Q': "ढ", 'R': "ण",
	// Dentals
	't': "त", 'T': "थ", 'd': "द", 'D': "ध", 'n': "न",
	// Labials
	'p': "प", 'P': "फ", 'b': "ब", 'B': "भ", 'm': "म",
	// Semivowels
	'y': "य", 'r': "र", 'l': "ल", 'v': "व",
	// Sibilants
	'S': "श", 'z': "ष", 's': "स", 'h': "ह",
	// Avagraha
	'\'': "ऽ",
}

// SLP1 vowel mātrās (dependent vowel signs)
var slpToMatra = map[rune]string{
	'a': "", // Inherent 'a' - no mātrā
	'A': "ा", 'i': "ि", 'I': "ी", 'u': "ु", 'U': "ू",
	'f': "ृ", 'F': "ॄ", 'x': "ॢ", 'X': "ॣ",
	'e': "े", 'E': "ै", 'o': "ो", 'O': "ौ",
}

// Consonants in SLP1
var consonants = map[rune]bool{
	'k': true, 'K': true, 'g': true, 'G': true, 'N': true,
	'c': true, 'C': true, 'j': true, 'J': true, 'Y': true,
	'w': true, 'W': true, 'q': true, 'Q': true, 'R': true,
	't': true, 'T': true, 'd': true, 'D': true, 'n': true,
	'p': true, 'P': true, 'b': true, 'B': true, 'm': true,
	'y': true, 'r': true, 'l': true, 'v': true,
	'S': true, 'z': true, 's': true, 'h': true,
}

// Vowels in SLP1
var vowels = map[rune]bool{
	'a': true, 'A': true, 'i': true, 'I': true, 'u': true, 'U': true,
	'f': true, 'F': true, 'x': true, 'X': true,
	'e': true, 'E': true, 'o': true, 'O': true,
}

// IASTToSLP converts IAST to SLP1 transliteration.
func IASTToSLP(iast string) string {
	iast = strings.ToLower(iast)
	var result strings.Builder
	runes := []rune(iast)

	for i := 0; i < len(runes); i++ {
		// Try two-character sequences first
		if i+1 < len(runes) {
			twoChar := string(runes[i : i+2])
			if slp, ok := iastToSLP[twoChar]; ok {
				result.WriteString(slp)
				i++
				continue
			}
		}

		// Single character
		oneChar := string(runes[i])
		if slp, ok := iastToSLP[oneChar]; ok {
			result.WriteString(slp)
		} else {
			// Pass through unchanged (spaces, punctuation, etc.)
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

// SLPToDevanagari converts SLP1 to Devanagari script.
func SLPToDevanagari(slp string) string {
	var result strings.Builder
	runes := []rune(slp)
	prevWasConsonant := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if consonants[r] {
			// Write the consonant
			if deva, ok := slpToDeva[r]; ok {
				result.WriteString(deva)
			}
			prevWasConsonant = true

			// Check if next char is a vowel (add mātrā) or consonant (add virāma)
			if i+1 < len(runes) {
				next := runes[i+1]
				if vowels[next] {
					// Next is vowel - mātrā will be added in next iteration
					continue
				}
				if consonants[next] || next == 'M' || next == 'H' {
					// Consonant cluster or anusvara/visarga - add virāma
					result.WriteString("्")
					prevWasConsonant = false
				}
			} else {
				// End of string after consonant - add virāma
				result.WriteString("्")
				prevWasConsonant = false
			}
		} else if vowels[r] {
			if prevWasConsonant {
				// Add mātrā (dependent vowel sign)
				if matra, ok := slpToMatra[r]; ok {
					result.WriteString(matra)
				}
			} else {
				// Independent vowel
				if deva, ok := slpToDeva[r]; ok {
					result.WriteString(deva)
				}
			}
			prevWasConsonant = false
		} else if r == 'M' || r == 'H' || r == '~' {
			// Anusvara, visarga, candrabindu
			if deva, ok := slpToDeva[r]; ok {
				result.WriteString(deva)
			}
			prevWasConsonant = false
		} else {
			// Pass through (spaces, punctuation, numbers)
			result.WriteRune(r)
			prevWasConsonant = false
		}
	}

	return result.String()
}

// IASTToDevanagari converts IAST directly to Devanagari.
func IASTToDevanagari(iast string) string {
	slp := IASTToSLP(iast)
	return SLPToDevanagari(slp)
}

// IsDevanagari returns true if the string contains Devanagari characters.
func IsDevanagari(s string) bool {
	for _, r := range s {
		if r >= 0x0900 && r <= 0x097F {
			return true
		}
	}
	return false
}

// NormalizeQuery normalizes a search query, returning both IAST and Devanagari forms.
func NormalizeQuery(query string) (iast, deva string) {
	query = strings.TrimSpace(query)

	if IsDevanagari(query) {
		// Already Devanagari, return as-is
		// TODO: Add Devanagari to IAST conversion if needed
		return query, query
	}

	// Assume IAST input
	iast = strings.ToLower(query)
	deva = IASTToDevanagari(iast)
	return iast, deva
}

// ToSearchTerms returns the query in forms suitable for searching.
func ToSearchTerms(query string) []string {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	terms := []string{query}

	if IsDevanagari(query) {
		return terms
	}

	// Add Devanagari form
	deva := IASTToDevanagari(query)
	if deva != "" && deva != query {
		terms = append(terms, deva)
	}

	// Also try lowercase
	lower := strings.ToLower(query)
	if lower != query {
		terms = append(terms, lower)
		devaLower := IASTToDevanagari(lower)
		if devaLower != "" && devaLower != deva {
			terms = append(terms, devaLower)
		}
	}

	return unique(terms)
}

func unique(strs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// ContainsDevanagari checks if any character in s is Devanagari.
func ContainsDevanagari(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Devanagari, r) {
			return true
		}
	}
	return false
}
