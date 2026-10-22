package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wudizhangzhi/HupuApp/gohupu/api"
	"github.com/wudizhangzhi/HupuApp/gohupu/auth"
	"github.com/wudizhangzhi/HupuApp/gohupu/spider"
	"github.com/wudizhangzhi/HupuApp/gohupu/teams"
)

type HomeCmd struct{}

func (*HomeCmd) Run() error {
	if !cli.Guest {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		cookie, err := auth.Ensure(ctx, os.Stderr, cli.LoginWait)
		stop()
		if err != nil {
			return err
		}
		spider.SetSession(cookie)
	}
	notice := ""
	for {
		team, err := teams.Preferred()
		if err != nil {
			return err
		}
		choice, err := choose("虎扑 · 功能菜单", homeItems(team), notice)
		if err != nil {
			return err
		}
		notice = ""
		switch choice {
		case -1, 5:
			return nil
		case 0:
			err = (&LiveCmd{GameType: api.NBA}).Run()
		case 1:
			err = (&NewsCmd{Region: spider.NBA}).Run()
		case 2:
			err = (&NewsCmd{Region: spider.Region("news:" + team)}).Run()
		case 3:
			err = (&NewsCmd{Region: spider.Region(team)}).Run()
		case 4:
			items := []string{}
			for _, t := range teams.Teams {
				items = append(items, t.Name)
			}
			var index int
			index, err = choose("选择主队", items, "主队偏好保存在本机，不改变默认功能菜单")
			if err == nil && index >= 0 {
				err = teams.Save(teams.Teams[index].Key)
			}
		}
		if err != nil {
			notice = "操作未完成：" + err.Error()
		}
	}
}
func homeItems(team string) []string {
	name := teams.Name(team)
	return []string{"Live · NBA 比赛文字直播", "NBA 新闻", name + "新闻 · NBA标题筛选", name + "专区 · 新闻与讨论", "设置主队（当前：" + name + "）", "退出"}
}
func choose(title string, items []string, notice string) (int, error) {
	ui := tview.NewApplication()
	index := -1
	list := tview.NewList().ShowSecondaryText(false)
	list.SetSelectedStyle(tcell.StyleDefault.Bold(true)).SetHighlightFullLine(false)
	list.SetBorder(true).SetTitle(title)
	for i, item := range items {
		n := i
		list.AddItem(item, "", 0, func() { index = n; ui.Stop() })
	}
	help := tview.NewTextView().SetText("↑↓ 选择 · Enter 进入 · Esc / Ctrl+C 返回或退出\n" + notice)
	panel := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(list, 0, 1, true).AddItem(help, 4, 0, false)
	panel.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEsc || e.Key() == tcell.KeyCtrlC {
			ui.Stop()
			return nil
		}
		return e
	})
	err := ui.SetRoot(panel, true).Run()
	return index, err
}

type LoginCmd struct{}

func (*LoginCmd) Run() error { return (&HomeCmd{}).Run() }

type AuthCmd struct{}

func (*AuthCmd) Run() error {
	cookie, err := auth.Load()
	if err != nil {
		return err
	}
	if strings.TrimSpace(cookie) == "" {
		fmt.Println("未保存登录会话，运行 hupu 自动登录。")
		return nil
	}
	ok, err := auth.Verify(context.Background(), cookie)
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("已登录，可复用本地会话。")
	} else {
		fmt.Println("会话已过期，运行 hupu 恢复登录。")
	}
	return nil
}
