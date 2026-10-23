package data

import (
	"strings"
	"testing"
	"unicode"
)

// TestStatementsAreAtomicWithStakes is the gate for the 2026-09-27
// atomicity pass (see ATOMIC-HANDOFF.md): every statement must be one
// proposition a person can agree or disagree with without splitting their
// answer, and every statement must carry practical_stakes so a reader
// unfamiliar with the case still knows what each extreme answer means.
//
// "Exactly one sentence" is checked as: after stripping the final period,
// no other ". " (period-space) sequence remains, UNLESS the period is
// immediately preceded by a digit — that is an ordinal marker ("1.
// konsesjonsrunde", "1. til 3. trinn"), not a sentence boundary. A naive
// strings.Count(text, ". ") == 0 check would misfire on exactly those two
// real statements (14 and 32), which is why this exists as a helper
// instead of an inline one-liner.
func TestStatementsAreAtomicWithStakes(t *testing.T) {
	for _, s := range Store.Statements {
		if s.PracticalStakes == nil {
			t.Errorf("utsagn-%d (%s): missing practical_stakes", s.Number, s.ShortName)
			continue
		}
		if strings.TrimSpace(s.PracticalStakes.Enig) == "" {
			t.Errorf("utsagn-%d (%s): practical_stakes.enig is empty", s.Number, s.ShortName)
		}
		if strings.TrimSpace(s.PracticalStakes.Uenig) == "" {
			t.Errorf("utsagn-%d (%s): practical_stakes.uenig is empty", s.Number, s.ShortName)
		}

		text := strings.TrimSpace(s.Text)
		if text == "" {
			t.Errorf("utsagn-%d (%s): text is empty", s.Number, s.ShortName)
			continue
		}

		if strings.Contains(text, "—") || strings.Contains(text, "–") {
			t.Errorf("utsagn-%d (%s): text contains an em or en dash: %q", s.Number, s.ShortName, text)
		}
		if strings.Contains(text, " selv om ") {
			t.Errorf("utsagn-%d (%s): text contains a rejectable condition clause (' selv om '): %q", s.Number, s.ShortName, text)
		}
		if strings.Contains(text, " fordi ") {
			t.Errorf("utsagn-%d (%s): text contains a justification clause (' fordi '): %q", s.Number, s.ShortName, text)
		}
		if n := len([]rune(text)); n > 200 {
			t.Errorf("utsagn-%d (%s): text is %d runes, want <= 200: %q", s.Number, s.ShortName, n, text)
		}
		if ok, near := isOneSentence(text); !ok {
			t.Errorf("utsagn-%d (%s): text looks like more than one sentence, near %q: %q", s.Number, s.ShortName, near, text)
		}
	}
}

// isOneSentence reports whether text (assumed already whitespace-normalized
// and trimmed) reads as a single sentence. It strips one trailing period,
// then scans for any remaining ". " that is not immediately preceded by a
// digit (an ordinal marker, not a sentence break).
func isOneSentence(text string) (ok bool, near string) {
	t := strings.TrimSuffix(text, ".")
	runes := []rune(t)
	for i := 0; i < len(runes)-1; i++ {
		if runes[i] == '.' && runes[i+1] == ' ' {
			if i > 0 && unicode.IsDigit(runes[i-1]) {
				continue
			}
			start := i - 20
			if start < 0 {
				start = 0
			}
			end := i + 20
			if end > len(runes) {
				end = len(runes)
			}
			return false, string(runes[start:end])
		}
	}
	return true, ""
}
