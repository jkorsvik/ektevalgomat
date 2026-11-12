package server

import (
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The spine is the only thing standing between nine separate doors and a walk, so
// these tests check the walk itself: nobody missing, no dead ends except the last
// step, and every page actually rendering its contract.

func TestSpineCoversEveryExplorePage(t *testing.T) {
	inSpine := map[string]bool{}
	for _, s := range exploreSpine {
		if inSpine[s.Href] {
			t.Errorf("%s appears twice in the spine", s.Href)
		}
		inSpine[s.Href] = true
	}
	for _, link := range navLinks {
		if link.Group != "explore" {
			continue
		}
		if !inSpine[link.Href] {
			t.Errorf("%s is in the Utforsk menu but not in the spine", link.Href)
		}
	}
	if len(inSpine) != len(exploreSpine) {
		t.Fatalf("%d entries but %d distinct hrefs", len(exploreSpine), len(inSpine))
	}
}

// A step with an empty contract line would render a heading over nothing.
func TestEveryStepStatesQuestionSourceAndGap(t *testing.T) {
	for _, s := range exploreSpine {
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

// Every step but the last must hand off to the next question, and the handoff must
// point somewhere in the walk. This is what stops the spine drifting into the dead
// ends it replaced.
func TestSpineIsOneWalkWithoutDeadEnds(t *testing.T) {
	for i, s := range exploreSpine {
		info := exploreInfo(s.Href)
		last := i == len(exploreSpine)-1
		if last {
			// The walk closes back at the valgomat rather than stopping dead.
			if info.NextHref != "/" || info.NextLabel == "" || info.NextQuestion == "" {
				t.Errorf("the last step %s does not close the loop: %q %q %q",
					s.Href, info.NextHref, info.NextLabel, info.NextQuestion)
			}
			continue
		}
		want := exploreSpine[i+1]
		if info.NextHref != want.Href {
			t.Errorf("%s hands off to %q, want %q", s.Href, info.NextHref, want.Href)
		}
		if info.NextLabel == "" || info.NextQuestion == "" {
			t.Errorf("%s hands off without naming the next question", s.Href)
		}
		if exploreInfo(info.NextHref).Href == "" {
			t.Errorf("%s points at %s, which is not in the spine", s.Href, info.NextHref)
		}
	}
}

// The pages must show what the spine says. A contract defined in Go and never
// rendered is a comment, not a mechanism.
func TestEveryExplorePageRendersItsContract(t *testing.T) {
	srv := newTestServer()
	for i, s := range exploreSpine {
		body := bodyOf(t, srv, s.Href)
		if !strings.Contains(body, "Spørsmålet:") {
			t.Errorf("%s does not state its question", s.Href)
		}
		if !strings.Contains(body, "Det som mangler:") {
			t.Errorf("%s does not state what it cannot show", s.Href)
		}
		next := exploreSpine[(i+1)%len(exploreSpine)]
		nextHref := next.Href
		if i == len(exploreSpine)-1 {
			nextHref = "/"
		}
		if !strings.Contains(body, `href="`+nextHref+`"`) {
			t.Errorf("%s does not link onward to %s", s.Href, nextHref)
		}
		if !strings.Contains(body, "Neste spørsmål:") {
			t.Errorf("%s has no handoff line", s.Href)
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
