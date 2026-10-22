package teams

import "strings"

type Team struct {
	Key  string
	Name string
}

var Teams = []Team{
	{"lakers", "湖人"}, {"warriors", "勇士"}, {"rockets", "火箭"}, {"spurs", "马刺"}, {"celtics", "凯尔特人"},
	{"sixers", "76人"}, {"knicks", "尼克斯"}, {"thunder", "雷霆"}, {"blazers", "开拓者"}, {"cavaliers", "骑士"},
	{"mavericks", "独行侠"}, {"heat", "热火"}, {"nuggets", "掘金"}, {"pistons", "活塞"}, {"raptors", "猛龙"},
	{"timberwolves", "森林狼"}, {"clippers", "快船"}, {"suns", "太阳"}, {"bucks", "雄鹿"}, {"nets", "篮网"},
	{"grizzlies", "灰熊"}, {"pelicans", "鹈鹕"}, {"wizards", "奇才"}, {"magic", "魔术"}, {"kings", "国王"},
	{"jazz", "爵士"}, {"hornets", "黄蜂"}, {"pacers", "步行者"}, {"bulls", "公牛"}, {"hawks", "老鹰"},
}

func ResolveTeam(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "nba" || value == "全部" {
		return "nba", true
	}
	for _, t := range Teams {
		if value == t.Key || value == t.Name {
			return t.Key, true
		}
	}
	return "", false
}
func FeedLabel(key string) string {
	for _, t := range Teams {
		if t.Key == key {
			return t.Name + "专区 · 新闻与讨论"
		}
	}
	return "NBA 资讯"
}

// AppID returns the team ID used by Hupu's NBA standings and App team webview.
// Source: https://m.hupu.com/nba/team-rank (2026-10-22).
func AppID(value string) (string, bool) {
	key, ok := ResolveTeam(value)
	if !ok {
		return "", false
	}
	id, ok := appIDs[key]
	return id, ok
}

var appIDs = map[string]string{
	"pistons":      "1901000000501290",
	"hawks":        "1901000000501284",
	"sixers":       "1901000000501281",
	"raptors":      "1901000000501298",
	"wizards":      "1901000000501283",
	"hornets":      "1901000000501334",
	"pacers":       "1901000000501292",
	"knicks":       "1901000000501273",
	"nets":         "1901000000501336",
	"magic":        "1901000000501279",
	"celtics":      "1901000000501263",
	"bucks":        "1901000000501294",
	"bulls":        "1901000000501286",
	"heat":         "1901000000501267",
	"cavaliers":    "1901000000501288",
	"blazers":      "1901000000501325",
	"clippers":     "1901000000501333",
	"rockets":      "1901000000501311",
	"lakers":       "1901000000501323",
	"grizzlies":    "1901000000501313",
	"thunder":      "1901000000501329",
	"timberwolves": "1901000000501315",
	"spurs":        "1901000000501317",
	"mavericks":    "1901000000501300",
	"jazz":         "1901000000501320",
	"warriors":     "1901000000501331",
	"kings":        "1901000000501321",
	"suns":         "1901000000501327",
	"pelicans":     "1901000000501296",
	"nuggets":      "1901000000501301",
}
