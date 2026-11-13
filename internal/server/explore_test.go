package server

import (
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// These tests cover the shared contract, the hub and the related-question links.

func TestPageSetCoversEveryExploreMenuPage(t *testing.T) {
	inPages := map[string]bool{}
	for _, s := range explorePages {
		if inPages[s.Href] {
			t.Errorf("%s appears twice in the page set", s.Href)
		}
		inPages[s.Href] = true
	}
	for _, link := range navLinks {
		if link.Group != "explore" || link.Href == "/utforsk" {
			continue
		}
		if !inPages[link.Href] {
			t.Errorf("%s is in the Utforsk menu but not in the page set", link.Href)
		}
	}
	if len(inPages) != len(explorePages) {
		t.Fatalf("%d entries but %d distinct hrefs", len(explorePages), len(inPages))
	}
}

// A step with an empty contract line would render a heading over nothing.
func TestEveryStepStatesQuestionSourceAndGap(t *testing.T) {
	for _, s := range explorePages {
		info := exploreInfo(s.Href)
		if strings.TrimSpace(info.Question) == "" {
			t.Errorf("%s has no question", s.Href)
		}
		if strings.TrimSpace(info.Source) == "" {
			t.Errorf("%s has no source line", s.Href)
		}
		if strings.TrimSpace(info.Missing) == "" {
			t.Errorf("%s says nothing about what it cannot show", s.Href)
		}
	}
}

// Every page offers two related questions in its own reader group. There is no
// prescribed next chapter and no last page that becomes a dead end.
func TestEveryExplorePageHasRelatedQuestions(t *testing.T) {
	for _, s := range explorePages {
		info := exploreInfo(s.Href)
		if len(info.Related) != 2 {
			t.Errorf("%s has %d related pages, want 2", s.Href, len(info.Related))
		}
		for _, related := range info.Related {
			if related.Href == s.Href {
				t.Errorf("%s points to itself", s.Href)
			}
			if exploreInfo(related.Href).Href == "" {
				t.Errorf("%s points at unknown related page %s", s.Href, related.Href)
			}
		}
	}
}

// The pages must show the contract and the way back to the hub. A contract
// defined in Go and never rendered is a comment, not a mechanism.
func TestEveryExplorePageRendersItsContractAndRelatedQuestions(t *testing.T) {
	srv := newTestServer()
	for _, s := range explorePages {
		body := bodyOf(t, srv, s.Href)
		if !strings.Contains(body, "Spørsmålet:") {
			t.Errorf("%s does not state its question", s.Href)
		}
		if !strings.Contains(body, "Det som mangler:") {
			t.Errorf("%s does not state what it cannot show", s.Href)
		}
		if !strings.Contains(body, `href="/utforsk"`) || !strings.Contains(body, "Alle spørsmål") {
			t.Errorf("%s does not link back to the question hub", s.Href)
		}
		if !strings.Contains(body, "Relaterte spørsmål:") {
			t.Errorf("%s has no related-question navigation", s.Href)
		}
		for _, related := range exploreInfo(s.Href).Related {
			if !strings.Contains(body, `href="`+related.Href+`"`) {
				t.Errorf("%s does not link to related page %s", s.Href, related.Href)
			}
		}
	}
}

// quoteCounts returns (total, withQuote) and the source line uses both. Swapping
// them shipped "730 av 730 har et ordrett sitat" next to "406 mangler sitat" on
// the same page. This ties each number to the function that produces it.
func TestContractQuoteNumbersMatchTheData(t *testing.T) {
	total, withQuote := quoteCounts(data.Store)
	src := exploreSource("/kart")
	gap := exploreGap("/kart")
	if !strings.Contains(src, thousands(total)+" påstander") {
		t.Errorf("source line does not state %d påstander: %s", total, src)
	}
	if !strings.Contains(src, thousands(withQuote)+" av dem har et ordrett sitat") {
		t.Errorf("source line does not state %d with a quote: %s", withQuote, src)
	}
	if total == withQuote {
		t.Fatal("the fixture has as many quotes as rows; this test cannot tell the two numbers apart")
	}
	if !strings.Contains(gap, thousands(total-withQuote)+" påstander har ingen lagret sitat") {
		t.Errorf("gap line does not state %d without a quote: %s", total-withQuote, gap)
	}
}
