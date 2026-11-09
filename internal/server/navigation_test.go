package server

import (
	"strings"
	"testing"
)

func TestNavigationGroupsAreExplicitAndMarkExploreCurrent(t *testing.T) {
	if got := len(coreNavLinks()); got != 3 {
		t.Errorf("core navigation has %d links, want 3", got)
	}
	// 9 since /personer joined Utforsk on 2026-11-09: the public-office records
	// are an overview page, and a page nobody can reach is a page for nobody.
	if got := len(exploreNavLinks()); got != 9 {
		t.Errorf("explore navigation has %d links, want 9", got)
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
