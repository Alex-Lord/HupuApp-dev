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
