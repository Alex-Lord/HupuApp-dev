package spider

import (
	"fmt"
	"github.com/wudizhangzhi/HupuApp/gohupu/teams"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTeamNewsCursorAndCache(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Cookie") != "" {
			t.Error("sent credentials")
		}
		if r.URL.Query().Get("teamId") != "1901000000501281" || r.URL.Query().Get("cateGoryCode") != "basketball" {
			t.Error("wrong team query")
		}
		switch r.URL.Query().Get("newsId") {
		case "0":
			fmt.Fprint(w, `{"success":true,"result":[{"nid":"90","tid":"123","title":"恩比德：期待比赛","replies":8,"lights":3,"publishTime":"2026-10-22"}]}`)
		case "90":
			fmt.Fprint(w, `{"success":true,"result":[{"nid":"80","tid":"124","title":"马克西谈防守"}]}`)
		case "80":
			fmt.Fprint(w, `{"success":true,"result":[]}`)
		default:
			t.Error("invalid cursor")
		}
	}))
	defer server.Close()
	f := NewTeamNewsFeed("news:sixers")
	f.endpoint = server.URL
	first, err := f.Page(1)
	if err != nil || len(first) != 1 {
		t.Fatalf("%v %v", first, err)
	}
	if first[0].Uid != "123" || first[0].Href != "/123.html" || first[0].ReplyCnt != 8 {
		t.Fatal(first)
	}
	second, err := f.Page(2)
	if err != nil || len(second) != 1 || second[0].Uid != "124" {
		t.Fatalf("%v %v", second, err)
	}
	first[0].Title = "mutated"
	prev, err := f.Page(1)
	if err != nil || prev[0].Title == "mutated" || calls != 2 {
		t.Fatal("cache not preserved")
	}
	for i := 0; i < 2; i++ {
		end, err := f.Page(3)
		if err != nil || len(end) != 0 {
			t.Fatal(end, err)
		}
	}
	if calls != 3 {
		t.Fatal("refetched exhausted feed")
	}
}
func TestTeamNewsErrorsDoNotAdvance(t *testing.T) {
	for _, body := range []string{`{"success":false,"result":[]}`, `{}`, `not json`, `{"success":true,"result":[{"nid":"0","tid":"123","title":"news"}]}`} {
		t.Run(body, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
			defer s.Close()
			f := NewTeamNewsFeed("news:sixers")
			f.endpoint = s.URL
			if _, err := f.Page(1); err == nil {
				t.Fatal("expected error")
			}
			if f.cursor != "0" || len(f.pages) != 0 {
				t.Fatal("advanced on failure")
			}
		})
	}
}
func TestAllTeamsHaveUniqueAppIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, team := range teams.Teams {
		id, ok := teams.AppID(team.Key)
		if !ok || !newsNumericID.MatchString(id) || seen[id] {
			t.Fatal(team, id)
		}
		seen[id] = true
	}
	if _, ok := teams.AppID("nba"); ok {
		t.Fatal("NBA is not a team")
	}
}
