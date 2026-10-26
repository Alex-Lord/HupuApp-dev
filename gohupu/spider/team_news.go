package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/wudizhangzhi/HupuApp/gohupu/teams"
)

const teamNewsEndpoint = "https://games.mobileapi.hupu.com/3/7.5.60/basketballapi/news/v2/teamNewsById"

type teamNewsItem struct {
	NID         string `json:"nid"`
	TID         string `json:"tid"`
	Title       string `json:"title"`
	Replies     int    `json:"replies"`
	Lights      int    `json:"lights"`
	PublishTime string `json:"publishTime"`
}
type teamNewsResponse struct {
	Success bool           `json:"success"`
	Result  []teamNewsItem `json:"result"`
}

// TeamNewsFeed owns a browsing session's pages and API cursors. It never shares
// the authenticated BBS client or cookie jar with the mobile API.
type TeamNewsFeed struct {
	region   Region
	client   *http.Client
	endpoint string
	pages    [][]BBS
	cursor   string
	done     bool
}

func NewTeamNewsFeed(region Region) *TeamNewsFeed {
	return &TeamNewsFeed{region: region, endpoint: teamNewsEndpoint, cursor: "0",
		client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}
}

var newsNumericID = regexp.MustCompile(`^[0-9]+$`)

func (f *TeamNewsFeed) Page(page int) ([]BBS, error) {
	id, ok := teams.AppID(strings.TrimPrefix(string(f.region), "news:"))
	if !ok {
		return nil, fmt.Errorf("未知球队：%s", f.region)
	}
	if page < 1 {
		page = 1
	}
	for len(f.pages) < page && !f.done {
		items, err := f.fetch(id)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			f.done = true
			break
		}
		next := items[len(items)-1].NID
		if !newsNumericID.MatchString(next) || next == f.cursor {
			return nil, fmt.Errorf("球队新闻接口返回无效的翻页游标")
		}
		posts := make([]BBS, 0, len(items))
		seen := map[string]bool{}
		for _, cached := range f.pages {
			for _, p := range cached {
				seen[p.Uid] = true
			}
		}
		for _, item := range items {
			if !newsNumericID.MatchString(item.TID) || strings.TrimSpace(item.Title) == "" {
				continue
			}
			if seen[item.TID] {
				continue
			}
			seen[item.TID] = true
			posts = append(posts, BBS{Region: f.region, Uid: item.TID, Href: "/" + item.TID + ".html",
				Title: item.Title, ReplyCnt: item.Replies, LightCnt: item.Lights, PostTime: item.PublishTime, Label: "球队新闻"})
		}
		if len(posts) == 0 {
			return nil, fmt.Errorf("球队新闻接口未返回新的可阅读帖子")
		}
		f.pages = append(f.pages, posts)
		f.cursor = next
	}
	if page > len(f.pages) {
		return []BBS{}, nil
	}
	return append([]BBS(nil), f.pages[page-1]...), nil
}

func (f *TeamNewsFeed) fetch(id string) ([]teamNewsItem, error) {
	u, err := url.Parse(f.endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("cateGoryCode", "basketball")
	q.Set("clientId", "93977196")
	q.Set("newsId", f.cursor)
	q.Set("teamId", id)
	u.RawQuery = q.Encode()
	resp, err := f.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("读取球队新闻：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("球队新闻 HTTP %d", resp.StatusCode)
	}
	var data teamNewsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&data); err != nil {
		return nil, fmt.Errorf("球队新闻响应格式错误：%w", err)
	}
	if !data.Success || data.Result == nil {
		return nil, fmt.Errorf("球队新闻接口未返回有效结果")
	}
	return data.Result, nil
}
