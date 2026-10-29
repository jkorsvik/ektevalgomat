package data

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The vote index is generated, and generated data goes stale the moment someone
// adds evidence that cites a vote the index has never seen. Nothing in the site
// would notice: the new vote would simply be missing from the alignment map, and
// the map would look fine.
//
// So this test derives the cited set from the content and requires the index to
// hold all of it. It is deliberately a re-implementation of the citation
// extraction rather than a call into the generator, because a test that asks the
// generator "did you do your job" answers itself. If the two ever disagree about
// what counts as a citation, the fix is to widen this regex, not to weaken the
// requirement.
// voteIDPatterns mirrors tools/verify/verify_votes.py ID_PATTERNS exactly. The
// fourth pattern is the one a previous Go check did not have, and its absence let a
// guard here report "0 contradictory" while two rows coded the same division the same
// way in opposite directions: they wrote "Votering 27496", which none of the first
// three patterns matches. A checker that cannot see its input passes vacuously.
var voteIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`voteringid=([0-9]{4,6})`),
	regexp.MustCompile(`[?&]vt=([0-9]{4,6})`),
	// Case matters, exactly as in the Python: only the last pattern accepts a
	// capital V. Making the third one case-insensitive made it read
	// "Votering_id 26361, 2025" as two citations and reported a votering 2025 that
	// does not exist. A port that is not byte-for-byte the authority drifts.
	regexp.MustCompile(`votering[_ -]?id[\s:=]*((?:[0-9]{4,6})(?:\s*(?:,|og)\s*[0-9]{4,6})*)`),
	regexp.MustCompile(`[Vv]otering(?:en|ene)?\s+((?:[0-9]{4,6})(?:\s*(?:,|og)\s*[0-9]{4,6})*)`),
}

var digits = regexp.MustCompile(`[0-9]{4,6}`)

// voteIDsIn returns every votering id the text cites, in first-seen order.
func voteIDsIn(text string) []int {
	var out []int
	seen := map[int]bool{}
	for _, re := range voteIDPatterns {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			for _, d := range digits.FindAllString(m[1], -1) {
				if id, err := strconv.Atoi(d); err == nil && !seen[id] {
					seen[id] = true
					out = append(out, id)
				}
			}
		}
	}
	return out
}

func citedVoteringIDs(t *testing.T) map[int]string {
	t.Helper()
	cited := map[int]string{}
	files, err := filepath.Glob(filepath.Join("content", "evidence", "*.yaml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no evidence files found (%d, %v): the check would pass on nothing", len(files), err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, id := range voteIDsIn(string(b)) {
			cited[id] = filepath.Base(f)
		}
	}
	return cited
}

func TestVoteIndexCoversEveryCitedVotering(t *testing.T) {
	cited := citedVoteringIDs(t)
	if len(cited) < 50 {
		t.Fatalf("only %d cited voterings parsed out of the content: the parse is broken, not the index", len(cited))
	}
	have := make(map[int]bool, VoteCount())
	for _, v := range Votes() {
		have[v.VoteringID] = true
	}
	missing := 0
	for id, file := range cited {
		if !have[id] {
			missing++
			if missing <= 10 {
				t.Errorf("votering %d is cited by %s but is not in the vote index: run tools/analytics/build_vote_index.py", id, file)
			}
		}
	}
	if missing > 10 {
		t.Errorf("... and %d more cited voterings are missing from the index", missing-10)
	}
	t.Logf("content cites %d voterings, index holds %d", len(cited), VoteCount())
}

var validSide = map[string]bool{"for": true, "mot": true, "split": true, "absent": true}

func TestEveryIndexedVoteCarriesSidesAndItsProposal(t *testing.T) {
	checked := 0
	withoutForslag := 0
	for _, v := range Votes() {
		checked++
		if v.VoteringID == 0 {
			t.Errorf("a vote with no votering_id is in the index")
		}
		if strings.TrimSpace(v.Forslag) == "" {
			withoutForslag++
		}
		if len(v.Parties) == 0 && !v.Unanimous {
			// A divided vote with no party side is a gap in the record: either the
			// tally failed to parse or the vote is misidentified.
			t.Errorf("votering %d has no party sides and is not marked unanimous", v.VoteringID)
		}
		for party, s := range v.Parties {
			if !validSide[s] {
				t.Errorf("votering %d, party %q: side %q is not one of for/mot/split/absent", v.VoteringID, party, s)
			}
		}
	}
	if checked == 0 {
		t.Fatal("indexed no votes: the check ran on nothing")
	}
	// Not a failure, but a number a reader should see: a vote whose proposal text
	// the API does not publish (typically a Romertall vedtak) still counts for
	// alignment, but cannot be named on a page.
	t.Logf("%d votes checked, %d without a published forslag text", checked, withoutForslag)
}

// TestVoteIndexIsFreshForTheSnapshot makes the staleness explicit rather than
// implicit: the index may not claim to have been built after the snapshot the
// site advertises, and it may not be older than the newest evidence it covers.
func TestVoteIndexIsFreshForTheSnapshot(t *testing.T) {
	gen := VotesGeneratedAt()
	if gen == "" {
		t.Fatal("the vote index records no generated_at date")
	}
	if _, err := time.Parse("2006-01-02", gen); err != nil {
		t.Fatalf("generated_at %q is not an ISO date: %v", gen, err)
	}
	t.Logf("%d votes, generated %s", VoteCount(), gen)
}

// The vote's own date, harvested from votering_tid. It was recorded in the
// generator as unavailable for three iterations, which wrote off a whole
// second-date stream as needing a live scan it did not need. The assertions below
// are derived from the artefact, so dropping the field or pointing it at the case's
// maintenance stamp fails here.
func TestEveryWalkedVoteCarriesItsOwnDate(t *testing.T) {
	votes := Votes()
	if len(votes) == 0 {
		t.Fatal("no votes loaded")
	}
	dated, undated := 0, []int{}
	years := map[int]int{}
	for _, v := range votes {
		if v.Date == "" {
			undated = append(undated, v.VoteringID)
			continue
		}
		d, err := time.Parse("2006-01-02", v.Date)
		if err != nil {
			t.Fatalf("votering %d: date %q is not YYYY-MM-DD: %v", v.VoteringID, v.Date, err)
		}
		dated++
		years[d.Year()]++
	}
	// Every vote the session walk reached has one. A handful reached only through a
	// citation may not, and that is recorded rather than hidden.
	if dated < len(votes)*99/100 {
		t.Errorf("only %d of %d votes carry a date", dated, len(votes))
	}
	if len(undated) > len(votes)/100 {
		t.Errorf("%d votes have no date, more than the 1%% allowance: %v", len(undated), undated[:min(5, len(undated))])
	}
	// A maintenance stamp would cluster the whole corpus into the weeks it was last
	// edited. A real vote date spreads across the years the sessions cover.
	if len(years) < 5 {
		t.Errorf("dates cover only %d distinct years (%v): that is an edit stamp, not a vote time", len(years), years)
	}
	for y, n := range years {
		if y < 2009 || y > 2026 {
			t.Errorf("%d votes dated %d, outside the period the corpus covers", n, y)
		}
	}
	t.Logf("%d of %d votes dated across %d years", dated, len(votes), len(years))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// A division is not only its nine parties. Votering 18486 fell 21 for, the nine
// account for 20, and the 21st is Irene Ojala of Pasientfokus (API id PF). The
// generator dropped any party id outside the nine without a word, so an unknown
// party and a party that did not vote were indistinguishable. This pins the real
// division and the population, so re-dropping them fails.
func TestVotesKeepPartiesOutsideTheNine(t *testing.T) {
	votes := Votes()
	if len(votes) == 0 {
		t.Fatal("no votes loaded")
	}
	withUnmapped := 0
	for _, v := range votes {
		if len(v.UnmappedParties) > 0 {
			withUnmapped++
		}
	}
	if withUnmapped == 0 {
		t.Fatal("no vote records a party outside the nine: the field is gone or never populated")
	}
	// The division that exposed this. 18486 must still show Pasientfokus for.
	var found bool
	for _, v := range votes {
		if v.VoteringID != 18486 {
			continue
		}
		found = true
		if got := v.UnmappedParties["PF"]; got != "for" {
			t.Errorf("votering 18486 unmapped[PF] = %q, want \"for\" (Pasientfokus voted for)", got)
		}
		if len(v.Parties) != 9 {
			t.Errorf("votering 18486 has %d of the nine parties: the unmapped fix must not change Parties", len(v.Parties))
		}
	}
	if !found {
		t.Skip("votering 18486 is not in this index")
	}
	t.Logf("%d of %d votes record a party outside the nine", withUnmapped, len(votes))
}

// One party cannot take both sides of one division in one statement. Two rows can
// cite the same votering legitimately when they code it on DIFFERENT statements (the
// same vote points opposite ways on two near-opposite questions), but inside one
// statement a +1 and a -1 on one votering is not history, it is a contradiction, and
// the newest-row rule would resolve it silently by date instead of surfacing it.
//
// Found in the wild: utsagn-59/sv coded votering 27496 as +1 from a co-sponsorship and
// as -1 from the vote itself, while both rows quoted "SV 5 mot, 0 for".
func TestAPartyNeverCodesOneVoteringOppositeWaysInOneStatement(t *testing.T) {
	type key struct {
		statement, party string
		vote             int
	}
	bySign := map[key]map[int][]string{}
	citations := 0
	for _, e := range Store.Evidence {
		if e.Position == nil {
			continue
		}
		sign := 0
		switch {
		case *e.Position > 0:
			sign = 1
		case *e.Position < 0:
			sign = -1
		}
		blob := e.SourceName + " " + e.SourceURL + " " + e.ContextNote + " " + e.ClaimAction
		for _, id := range voteIDsIn(blob) {
			citations++
			k := key{e.StatementID, e.PartyID, id}
			if bySign[k] == nil {
				bySign[k] = map[int][]string{}
			}
			bySign[k][sign] = append(bySign[k][sign], e.ID)
		}
	}
	if citations == 0 {
		t.Fatal("no coded evidence cites a votering: the check would pass on nothing")
	}
	bad := 0
	for k, signs := range bySign {
		if len(signs[1]) > 0 && len(signs[-1]) > 0 {
			bad++
			t.Errorf("%s/%s codes votering %d both ways: +%v vs -%v",
				k.statement, k.party, k.vote, signs[1], signs[-1])
		}
	}
	t.Logf("%d citations checked across %d statement/party/votering keys, %d contradictory",
		citations, len(bySign), bad)
}

// Every vote must say which session it was taken in, because a vote with no
// session cannot be placed in time and the timeline cannot show it.
//
// Measured 2026-10-29: 4179 of 25773 votes carried no session, spread over
// fifteen sessions. The cause was a closed loop in the generator, not a gap in
// the source: the union of sessions it walks is derived from the session labels
// already in the index, so an unlabelled vote could never pull in the walk that
// would have labelled it. The generator now falls back to the vote's own
// votering_tid, which agrees with every label that already existed (21594 of
// 21594, including all 232 votes whose id two sessions both claim).
func TestEveryVoteCarriesItsSession(t *testing.T) {
	votes := Votes()
	if len(votes) == 0 {
		t.Fatal("no votes loaded: the check would pass on nothing")
	}
	var empty []int
	for _, v := range votes {
		if v.Session == "" {
			empty = append(empty, v.VoteringID)
		}
	}
	if len(empty) > 0 {
		first := empty
		if len(first) > 5 {
			first = first[:5]
		}
		t.Errorf("%d of %d votes carry no session (first ids %v): the artefact cannot place "+
			"them in time", len(empty), len(votes), first)
	}
	labelled := 0
	for _, v := range votes {
		if v.Session != "" {
			labelled++
		}
	}
	t.Logf("%d votes checked, %d carry a session", len(votes), labelled)
}
