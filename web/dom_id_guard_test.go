package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestStatementCardIdHasOneSource makes "a statement card has one id, built
// in one place" a tested property of the shipped JS, not a convention.
//
// On 2026-10-13 the live site carried two schemes for the same id:
// buildStatementCard set card.id = "utsagn-" + st.number, while markWeight
// looked up document.getElementById("statement-" + st.id). getElementById
// returned null, markWeight returned early on the very next line, and the
// importance control kept showing "Vanlig" after the visitor pressed
// "Viktig". The weight WAS saved, so the result moved while the control
// denied it — a control misreporting the visitor's own input, silently, on
// the one axis the trade-off step reads.
//
// Nothing caught it: both expressions were legal, the page rendered, no
// console error appeared, and the answer buttons (which used the right
// prefix) looked correct. The id was the only thing wrong, and an id has no
// type. So the rule is enforced over the source text the way
// token_guard_test.go enforces the colour rule: one scheme, one builder,
// and a failing test if a second appears.
//
// Comments are stripped first: the comment in quiz.js quotes BOTH schemes
// to explain the bug, and a gate that reads its own explanation as code
// would fail on the fix it documents. Block and line comments are removed;
// a `//` inside a string literal would also be removed, which is harmless
// here because the only literals this test looks for are id prefixes.
func TestStatementCardIdHasOneSource(t *testing.T) {
	b, err := fs.ReadFile(Static, "static/js/quiz.js")
	if err != nil {
		t.Fatalf("read quiz.js: %v", err)
	}
	src := stripJSComments(string(b))

	// 1. Exactly one prefix scheme may exist. A second scheme is the bug.
	prefix := regexp.MustCompile(`"(utsagn|statement)-"`)
	found := map[string]int{}
	for _, m := range prefix.FindAllString(src, -1) {
		found[m]++
	}
	if len(found) != 1 {
		t.Errorf("quiz.js uses %d statement-card id prefixes, want 1: %v", len(found), found)
	}
	if _, ok := found[`"utsagn-"`]; !ok {
		t.Errorf("quiz.js does not use \"utsagn-\" as the card id prefix; found %v", found)
	}

	// 2. And it is built in exactly one place, so two sites cannot drift.
	if n := strings.Count(src, `"utsagn-"`); n != 1 {
		t.Errorf("the card id prefix is built in %d places, want exactly 1 (statementCardId)", n)
	}

	// 3. No lookup may assemble an id of its own. This is the line that
	//    would have failed on the bug even with the right prefix, a wrong
	//    field ("utsagn-" + st.id) being the same mistake one step over.
	call := regexp.MustCompile(`getElementById\(([^)]*)\)`)
	lookups, selfBuilt := 0, 0
	for _, m := range call.FindAllStringSubmatch(src, -1) {
		lookups++
		if prefix.MatchString(m[1]) {
			selfBuilt++
			t.Errorf("getElementById builds a statement-card id itself instead of using statementCardId: %s", m[1])
		}
	}

	// 4. Fail loudly if the check ran on nothing, and require the single
	//    builder to actually be used. 1 definition plus 4 call sites:
	//    markAnswered, buildStatementCard, scrollToCard, markWeight.
	const wantUses = 5
	if got := strings.Count(src, "statementCardId("); got < wantUses {
		t.Fatalf("statementCardId( appears %d times, want at least %d; "+
			"the lookup sites have been removed and this test would pass on an empty file", got, wantUses)
	}

	// 5. The bug's exact field mistake: the suffix must be the statement's
	//    NUMBER, never its id. "utsagn-" + st.id yields "utsagn-utsagn-01".
	if strings.Contains(src, `statementCardId(st.id)`) {
		t.Errorf("statementCardId is called with st.id; it needs st.number, or the id becomes utsagn-utsagn-NN")
	}

	t.Logf("checked %d getElementById lookups, %d prefix scheme(s), %d builder, %d statementCardId uses, %d self-built",
		lookups, len(found), strings.Count(src, `"utsagn-"`), strings.Count(src, "statementCardId("), selfBuilt)
}

// stripJSComments removes block and line comments so a test that asserts on
// source text cannot be satisfied or broken by prose.
func stripJSComments(s string) string {
	s = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(s, "")
	return regexp.MustCompile(`(?m)//[^\n]*`).ReplaceAllString(s, "")
}
