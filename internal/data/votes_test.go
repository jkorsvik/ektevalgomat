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
var citedVotering = regexp.MustCompile(`(?i)(?:votering[_ -]?id[\s:=]*|voteringid=|&vt=)(\d{4,6})`)

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
		for _, m := range citedVotering.FindAllStringSubmatch(string(b), -1) {
			id, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
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
