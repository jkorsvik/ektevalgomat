package server

import (
	"strings"
	"testing"
)

func TestNavigationGroupsAreExplicitAndMarkExploreCurrent(t *testing.T) {
	if got := len(coreNavLinks()); got != 3 {
		t.Errorf("core navigation has %d links, want 3", got)
	}
	// The hub plus nine analysis pages. The hub is the reader-facing entry;
	// the nine existing pages stay directly reachable from the grouped menu.
	if got := len(exploreNavLinks()); got != 10 {
		t.Errorf("explore navigation has %d links, want 10", got)
	}
	if got := len(metaNavLinks()); got != 4 {
		t.Errorf("meta navigation has %d links, want 4", got)
	}
	for _, link := range navLinks {
		switch link.Group {
		case "core", "explore", "meta":
		default:
			t.Errorf("%s has no valid navigation group", link.Href)
		}
	}
	body := bodyOf(t, newTestServer(), "/rommet")
	if !strings.Contains(body, `<summary aria-current="page">Utforsk</summary>`) {
		t.Error("the collapsed Utforsk menu does not mark its own section as current")
	}
}
