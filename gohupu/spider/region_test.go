package spider

import "testing"

func TestForumKeepsDiscussion(t *testing.T) {
	if !MatchesRegion(Region("sixers"), BBS{Uid: "124", Title: "关于76人的讨论"}) {
		t.Fatal("filtered discussion")
	}
	if MatchesRegion(Region("sixers"), BBS{Title: "missing id"}) {
		t.Fatal("accepted missing id")
	}
}
