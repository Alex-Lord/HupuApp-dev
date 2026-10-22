package spider

import "testing"

func TestTeamNewsFilter(t *testing.T) {
	r := Region("news:sixers")
	if r.String() != "502-hot" {
		t.Fatal(r.String())
	}
	news := BBS{Uid: "123", Title: "[流言板]76人比赛新闻"}
	discussion := BBS{Uid: "124", Title: "关于76人的讨论"}
	if !MatchesRegion(r, news) || MatchesRegion(r, discussion) {
		t.Fatal("team news includes discussions")
	}
	other := BBS{Uid: "125", Title: "[流言板]湖人比赛新闻"}
	if MatchesRegion(r, other) {
		t.Fatal("included other team")
	}
	if !MatchesRegion(Region("sixers"), discussion) {
		t.Fatal("team forum filtered discussion")
	}
}
