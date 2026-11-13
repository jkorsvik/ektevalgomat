package server

import (
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func TestStatementPutsPartyAnswersBeforeMethodAndPromises(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utsagn/25")
	ordered := []string{
		"Partienes dokumenterte posisjoner",
		"Hva har faktisk skjedd?",
		`id="party-ap"`,
		"Kjeden for dette utsagnet",
		"Partiene om dette utsagnet",
	}
	previous := -1
	for _, marker := range ordered {
		at := strings.Index(body, marker)
		if at < 0 {
			t.Fatalf("statement page is missing %q", marker)
		}
		if at <= previous {
			t.Errorf("%q is out of reader order", marker)
		}
		previous = at
	}
}

func TestStatementQuoteAndOriginalSourceAreVisibleBeforeTechnicalDetails(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utsagn/25")
	party := strings.Index(body, `id="party-ap"`)
	if party < 0 {
		t.Fatal("party card is missing")
	}
	segment := body[party:]
	quote := strings.Index(segment, "Komiteens medlemmer fra Arbeiderpartiet og Senterpartiet")
	source := strings.Index(segment, `class="source-open"`)
	technical := strings.Index(segment, "Teknisk kildegrunnlag og kontroll")
	if quote < 0 || source < 0 || technical < 0 {
		t.Fatalf("missing fixture: quote=%d source=%d technical=%d", quote, source, technical)
	}
	if !(quote < source && source < technical) {
		t.Error("quote and original source are not visible before the technical disclosure")
	}
	openDetails := strings.Index(segment[:quote], "<details")
	if openDetails >= 0 {
		t.Error("the lead quote is hidden inside a details element")
	}
}

func TestStatementReturnPreservesExactFinderQuery(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utsagn/25?q=tannhelse&tema=universell-grunnmur&external=https%3A%2F%2Fevil.example")
	want := `href="/utforsk?q=tannhelse&amp;tema=universell-grunnmur#saker"`
	if !strings.Contains(body, want) {
		t.Errorf("statement does not return to the exact finder query; missing %q", want)
	}
	if strings.Contains(body, "evil.example") {
		t.Error("an unrelated query parameter entered the return link")
	}
}

func TestStatementReturnPreservesRommetComparison(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utsagn/55?fra=rommet&tema=finansiering-av-fellesskapet&a=ap&b=h&alle=1&external=https%3A%2F%2Fevil.example")
	want := `href="/rommet?a=ap&amp;alle=1&amp;b=h&amp;tema=finansiering-av-fellesskapet#saker"`
	if !strings.Contains(body, want) {
		t.Errorf("statement does not return to the exact comparison; missing %q", want)
	}
	if !strings.Contains(body, "Til uenighetene") {
		t.Error("the return link does not name the comparison the reader came from")
	}
	if strings.Contains(body, "evil.example") {
		t.Error("an unrelated query parameter entered the comparison return link")
	}
}

func TestStatementNeighboursFollowTheRealNumbering(t *testing.T) {
	for i, statement := range data.Store.Statements {
		page := buildStatementPage(data.Store, statement)
		if i == 0 {
			if page.Previous != nil {
				t.Error("first statement has a previous link")
			}
		} else if page.Previous == nil || page.Previous.Number != data.Store.Statements[i-1].Number {
			t.Errorf("%s has the wrong previous link", statement.ID)
		}
		if i == len(data.Store.Statements)-1 {
			if page.Next != nil {
				t.Error("last statement has a next link")
			}
		} else if page.Next == nil || page.Next.Number != data.Store.Statements[i+1].Number {
			t.Errorf("%s has the wrong next link", statement.ID)
		}
		if !strings.Contains(page.ExploreTopicHref, "/utforsk?tema=") {
			t.Errorf("%s does not link back to its topic in the hub", statement.ID)
		}
	}

	body := bodyOf(t, newTestServer(), "/utsagn/25")
	for _, want := range []string{`href="/utsagn/24"`, `href="/utsagn/26"`, "Alle saker", "Til valgomaten"} {
		if !strings.Contains(body, want) {
			t.Errorf("statement navigation is missing %q", want)
		}
	}
}
