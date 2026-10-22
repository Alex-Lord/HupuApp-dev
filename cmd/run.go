package main

import (
	"fmt"
	"github.com/wudizhangzhi/HupuApp/gohupu/auth"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/wudizhangzhi/HupuApp/gohupu/api"
	"github.com/wudizhangzhi/HupuApp/gohupu/api_utils"
	"github.com/wudizhangzhi/HupuApp/gohupu/live"
	"github.com/wudizhangzhi/HupuApp/gohupu/menu"
	"github.com/wudizhangzhi/HupuApp/gohupu/spider"
	"github.com/wudizhangzhi/HupuApp/gohupu/view"
)

type LiveCmd struct {
	GameType api.GameType `arg:"" name:"gameType" help:"比赛类型(nba/cba)."`
}

// 直播
func (r *LiveCmd) Run() error {
	matches, err := api_utils.GetMatchesFromDate(r.GameType)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		matches, err = api_utils.GetAnyMatches(r.GameType, 10, false)
		if err != nil {
			return err
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("暂未获取到比赛，可能暂无赛程或上游直播接口已变化")
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
	if idx < 0 || idx >= len(matches) {
		return nil
	}
	match := matches[idx]
	client := live.New(match)
	// // 退出过快可能导致print打印不显示
	time.Sleep(1 * time.Second)
	return client.Start()
}

type NewsCmd struct {
	Region spider.Region `arg:"" name:"region" help:"领域(nba/cba/湿乎乎/步行街)."`
}

func (r *NewsCmd) Run() error {
	if !cli.Guest {
		cookie, err := auth.Load()
		if err != nil {
			return err
		}
		spider.SetSession(cookie)
	}
	spider.New()
	view.NewApp()
	bbsView := view.NewBBSView(r.Region)
	bbsView.Display()
	return nil
}

var cli struct {
	Debug     bool          `help:"Enable debug mode."`
	Guest     bool          `help:"跳过登录，使用公开内容。"`
	LoginWait time.Duration `name:"login-wait" default:"5m" help:"等待网页登录的时间。"`
	Home      HomeCmd       `cmd:"" default:"1" help:"自动登录并进入功能菜单。"`
	Login     LoginCmd      `cmd:"" help:"恢复登录并进入功能菜单。"`
	Auth      AuthCmd       `cmd:"" help:"检查本地登录状态。"`

	Live LiveCmd `cmd:"" help:"比赛直播."`
	News NewsCmd `cmd:"" help:"新闻、帖子."`
}

func main() {
	if os.Getenv("TERM") == "" || os.Getenv("TERM") == "dumb" {
		os.Setenv("TERM", "xterm-256color")
	}
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
