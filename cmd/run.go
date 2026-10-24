package main

import (
	"time"

	"github.com/alecthomas/kong"
	"github.com/wudizhangzhi/HupuApp/gohupu/api"
	"github.com/wudizhangzhi/HupuApp/gohupu/api_utils"
	"github.com/wudizhangzhi/HupuApp/gohupu/live"
	"github.com/wudizhangzhi/HupuApp/gohupu/menu"
	"github.com/wudizhangzhi/HupuApp/gohupu/spider"
	"github.com/wudizhangzhi/HupuApp/gohupu/view"
	"github.com/wudizhangzhi/HupuApp/gohupu/watch"
)

type LiveCmd struct {
	GameType api.GameType `arg:"" name:"gameType" help:"比赛类型(nba/cba)."`
}

// 直播
func (r *LiveCmd) Run() error {
	matches, _ := api_utils.GetMatchesFromDate(r.GameType)
	if len(matches) == 0 {
		matches, _ = api_utils.GetAnyMatches(api.NBA, 10, false)
	}
	interfaceItems := make([]interface{}, 0)
	for _, item := range matches {
		interfaceItems = append(interfaceItems, item)
	}
	m := menu.Menu{
		Label:     menu.LabelLive,
		Items:     interfaceItems,
		Templates: menu.LiveTemplate,
		Size:      len(interfaceItems),
	}

	idx, err := m.Start()
	if err != nil {
		return err
	}
	match := matches[idx]
	client := live.New(match)
	// // 退出过快可能导致print打印不显示
	time.Sleep(1 * time.Second)
	client.Start()
	return nil
}

type NewsCmd struct {
	Region spider.Region `arg:"" name:"region" help:"领域(nba/cba/湿乎乎/步行街)."`
}

func (r *NewsCmd) Run() error {
	spider.New()
	view.NewApp()
	bbsView := view.NewBBSView(r.Region)
	bbsView.Display()
	return nil
}

type WatchCmd struct {
	GameType api.GameType `arg:"" name:"gameType" help:"比赛类型(nba/cba)."`
	Team     string       `arg:"" name:"team" help:"球队名称(如 开拓者)."`
	Player   string       `arg:"" name:"player" help:"球员名称(如 杨瀚森)."`
	Notify   string       `name:"notify" help:"通知命令(如 dx-push), 消息内容将作为参数传入."`
}

// 监控球员上场
func (r *WatchCmd) Run() error {
	watcher := watch.New(r.GameType, r.Team, r.Player, r.Notify)
	return watcher.Start()
}

var cli struct {
	Debug bool `help:"Enable debug mode."`

	Live  LiveCmd  `cmd:"" help:"比赛直播."`
	News  NewsCmd  `cmd:"" help:"新闻、帖子."`
	Watch WatchCmd `cmd:"" help:"监控球员上场."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("Hupu"),
		kong.Description("A command line tool for Hupu."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree:      true,
			Compact:   true,
			Summary:   true,
			FlagsLast: true,
			// NoExpandSubcommands: true,
		}))
	// TODO 读取配置, 注册signal
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
