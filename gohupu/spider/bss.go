package spider

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/wudizhangzhi/HupuApp"
	"github.com/wudizhangzhi/HupuApp/gohupu/logger"
	"github.com/wudizhangzhi/HupuApp/gohupu/teams"
)

type Region string

const (
	NBA        Region = "nba"
	CBA        Region = "cba"
	Vote       Region = "湿乎乎"
	TopicDaily Region = "步行街" // 步行街主干道
	// Gambia     Region = "步行街"
)

// 返回链接中对应的region部分, https://bbs.hupu.com/<region>
func (r Region) String() string {
	if strings.HasPrefix(string(r), "news:") {
		return "502-hot"
	}
	key := string(r)
	if team, ok := teams.ResolveTeam(key); ok && team != "nba" {
		return team
	}
	switch r {
	case NBA:
		return "502-hot"
	case CBA:
		return "cba"
	case Vote:
		return "vote"
	// case Gambia:
	// 	return "all-gambia"
	case TopicDaily:
		return "topic-daily"
	default:
		return "502-hot"
	}
}

// 对应帖子列表的selector
func (r Region) GetBBSListSelector() string {
	switch r {
	case NBA:
		return "div.bbs-sl-web-post > ul > li > div.bbs-sl-web-post-layout"
	case CBA:
		return "div.bbs-sl-web-post > ul > li > div.bbs-sl-web-post-layout"
	case Vote:
		return "div.bbs-sl-web-post > ul > li > div.bbs-sl-web-post-layout"
	default:
		return "div.bbs-sl-web-post > ul > li > div.bbs-sl-web-post-layout"
	}
}

type GetBBSInfoFunc func(i int, s *goquery.Selection, region Region, page int) BBS

// 湿乎乎、步行街等获取帖子信息的方法
func VoteGetBBS(i int, s *goquery.Selection, region Region, page int) BBS {
	selection := s.Find("div.post-title > a")
	title := selection.Text()
	href, _ := selection.Attr("href")
	nickname := s.Find("div.post-auth > a").Text()
	postTime := s.Find("div.post-time").Text()
	viewReplyCntS := s.Find("div.post-datum").Text()
	viewReplyCntList := strings.Split(viewReplyCntS, "/")
	replyCnt := 0
	viewCnt := 0
	if len(viewReplyCntList) == 2 {
		replyCnt, _ = strconv.Atoi(regexp.MustCompile(`\d+`).FindString(viewReplyCntList[0]))
		viewCnt, _ = strconv.Atoi(regexp.MustCompile(`\d+`).FindString(viewReplyCntList[1]))
	}
	label := s.Find("div.t-label > a").Text()
	uid := regexp.MustCompile(`\d+`).FindString(href)
	logger.Debug.Printf("UID.%d: UID:%s 标题:%s 作者:%s 回复/浏览:%s\n", i*page+1, uid, title, nickname, viewReplyCntS)

	return BBS{
		Uid:      uid,
		Title:    title,
		Nickname: nickname,
		ReplyCnt: replyCnt,
		ViewCnt:  viewCnt,
		Href:     href,
		Label:    label,
		PostTime: postTime,
	}
}

// NBA/CBA获取帖子信息的方法
// 暂时不需要，使用类似步行街的页面了
func NBAGetBBS(i int, s *goquery.Selection, region Region, page int) BBS {
	selection := s.Find("div > a")
	title := selection.Text()
	href, _ := selection.Attr("href")
	lightCntS := s.Find("div.t-info > span.t-lights").Text()
	replyCntS := s.Find("div.t-info > span.t-replies").Text()
	label := s.Find("div.t-label > a").Text()
	// fmt.Printf("No.%d: 标题:%s 亮:%s 回复:%s\n", i*page+1, title, lightCntS, replyCntS)
	logger.Debug.Printf("No.%d: 标题:%s 亮:%s 回复:%s\n", i*page+1, title, lightCntS, replyCntS)

	uid := regexp.MustCompile(`\d+`).FindString(href)
	lightCnt, _ := strconv.Atoi(regexp.MustCompile(`\d+`).FindString(lightCntS))
	replyCnt, _ := strconv.Atoi(regexp.MustCompile(`\d+`).FindString(replyCntS))
	return BBS{
		Uid:      uid,
		Title:    title,
		LightCnt: lightCnt,
		ReplyCnt: replyCnt,
		Href:     href,
		Label:    label,
	}
}

// 获取帖子信息的方法
func (r Region) GetBBS(i int, s *goquery.Selection, page int) BBS {
	switch r {
	case NBA:
		return VoteGetBBS(i, s, r, page)
	case CBA:
		return VoteGetBBS(i, s, r, page)
	case Vote:
		return VoteGetBBS(i, s, r, page)
	case TopicDaily:
		return VoteGetBBS(i, s, r, page)
	default:
		return VoteGetBBS(i, s, r, page)
	}
}

type BBS struct {
	Region   Region `comment:"领域"`
	Uid      string `comment:"帖子id"`
	Title    string `comment:"标题"`
	Href     string `comment:"链接"`
	Label    string `comment:"标签"`
	ReplyCnt int    `comment:"回复"`
	ViewCnt  int    `comment:"浏览"`
	LightCnt int    `comment:"亮了"`
	Content  string `comment:"内容"`
	// Author   User   `comment:"作者"`
	Nickname string `comment:"作者"`
	PostTime string `comment:"发帖时间"`
}

type Comment struct {
	Uid       string
	Content   string `comment:"内容"`
	Location  string
	ReplyTime time.Time
	LightCnt  int
	Nickname  string
}

func GetBBSList(region Region, page int) ([]BBS, error) {
	if page < 1 {
		page = 1
	}
	if !strings.HasPrefix(string(region), "news:") {
		return getBBSPage(region, page)
	}
	// One team-news page examines three NBA source pages, without duplicate posts.
	result := []BBS{}
	seen := map[string]bool{}
	for n := (page-1)*3 + 1; n <= page*3; n++ {
		posts, err := getBBSPage(region, n)
		if err != nil {
			return nil, err
		}
		for _, p := range posts {
			if !seen[p.Uid] {
				seen[p.Uid] = true
				result = append(result, p)
			}
		}
	}
	return result, nil
}
func getBBSPage(region Region, page int) ([]BBS, error) {
	url := "https://bbs.hupu.com/" + region.String()
	if page > 1 {
		url += fmt.Sprintf("-%d", page)
	}

	bbsList := make([]BBS, 0)
	resp, err := SpiderClient.R().
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("新闻列表 HTTP %d", resp.StatusCode())
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return bbsList, err
	}
	// Find the bbs items
	doc.Find(region.GetBBSListSelector()).
		Each(func(i int, s *goquery.Selection) {
			item := region.GetBBS(i, s, page)
			if MatchesRegion(region, item) {
				bbsList = append(bbsList, item)
			}
		})
	return bbsList, nil
}

func (bbs *BBS) GetDetail() (string, error) {
	url := "https://bbs.hupu.com/" + bbs.Uid + ".html"
	resp, err := SpiderClient.R().
		Get(url)
	if err != nil {
		return bbs.Content, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return bbs.Content, err
	}
	content := ""
	// 选择第一个
	doc.Find("div[class^=post-content_main-post-info] > div >div.thread-content-detail > p").
		Each(func(i int, s *goquery.Selection) {
			content += s.Text() + "\n"
		})
	bbs.Content = content
	logger.Debug.Printf("帖子内容: %s\n", content)
	return bbs.Content, nil
}

func (bbs *BBS) GetComments(page int) ([]Comment, error) {
	url := "https://bbs.hupu.com/" + bbs.Uid + fmt.Sprintf("-%d", page) + ".html"
	comments := make([]Comment, 0)
	resp, err := SpiderClient.R().
		Get(url)
	if err != nil {
		return comments, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return comments, nil
	}
	doc.Find("div.post-reply-list-wrapper").Each(func(i int, s *goquery.Selection) {
		content := s.Find("div.thread-content-detail > p").Text()
		nickname := s.Find("div.user-base-info > a").Text()
		replyTimeS := s.Find("div.user-base-info > span.post-reply-list-user-info-top-time").Text()
		replyTime, _ := HupuApp.GetTimeFromString(replyTimeS)
		location := s.Find("div.user-base-info > span.post-reply-list-user-info-user-location").Text()
		location = strings.Trim(location, "发布于")
		lightCntS := s.Find("div.post-reply-list-operate > div.light > span").Text()
		lightCntS = regexp.MustCompile(`\d+`).FindString(lightCntS)
		lightCnt, _ := strconv.Atoi(lightCntS)
		// fmt.Printf("评论: %s\n", content)
		comments = append(comments, Comment{
			Content:   content,
			Location:  location,
			ReplyTime: replyTime,
			Nickname:  nickname,
			LightCnt:  lightCnt,
		})
	})
	for _, comment := range comments {
		// fmt.Printf("评论: %+v\n", comment)
		logger.Debug.Printf("评论: %+v\n", comment)
	}
	return comments, nil
}

// Team news selects explicit news titles mentioning the team from NBA news pages.
func MatchesRegion(region Region, item BBS) bool {
	if item.Uid == "" || strings.TrimSpace(item.Title) == "" {
		return false
	}
	if strings.HasPrefix(string(region), "news:") {
		team := strings.TrimPrefix(string(region), "news:")
		key, ok := teams.ResolveTeam(team)
		if !ok || key == "nba" {
			return false
		}
		news := strings.Contains(item.Title, "[流言板]") || strings.Contains(item.Title, "【流言板】")
		return news && strings.Contains(item.Title, teams.Name(key))
	}
	return true
}
func (r Region) Label() string {
	if strings.HasPrefix(string(r), "news:") {
		return teams.Name(strings.TrimPrefix(string(r), "news:")) + "新闻（NBA标题筛选）"
	}
	if key, ok := teams.ResolveTeam(string(r)); ok && key != "nba" {
		return teams.Name(key) + "专区（含讨论）"
	}
	return string(r)
}
