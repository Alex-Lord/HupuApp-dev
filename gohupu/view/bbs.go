package view

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wudizhangzhi/HupuApp/gohupu/logger"
	"github.com/wudizhangzhi/HupuApp/gohupu/spider"
	"strings"
)

const help_content = `
Esc: 返回列表/菜单 Ctrl+C: 返回菜单 Enter: 选择
Ctrl+N/P: 翻页 评论区 ↑/↓: 上/下一页
→ / Ctrl+Q: 列表 → 正文 → 评论 PgUp/PgDn: 滚动长内容
`

var BBSView *BBSViewT

type BBSViewT struct {
	Region       spider.Region
	newsFeed     *spider.TeamNewsFeed
	Main         *tview.Pages
	BBSListView  *wrappedList
	PostView     *tview.TextView
	CommentsView *wrappedList
	// 页数
	BBSPage     int
	CommentPage int
	// 显示
	HelpContent string
	BBSList     []spider.BBS
	Comments    []spider.Comment
	PostContent string
	SelectedBBS spider.BBS
}

func NewBBSView(region spider.Region) *BBSViewT {
	pages := tview.NewPages()
	// 左侧
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	leftListView := newWrappedList()
	leftListView.SetBorder(true)
	leftListView.selected = onSelected

	help := tview.NewTextView().SetTextColor(tcell.ColorAqua)
	fmt.Fprint(help, help_content)
	leftPanel.AddItem(leftListView, 0, 5, true).AddItem(help, 0, 1, false)

	// 右侧
	postContentView := tview.NewTextView().SetDynamicColors(true).SetChangedFunc(func() {
		app.Draw()
	})
	postContentView.SetBorder(true).SetTitle("Post")
	commentsView := newWrappedList()
	commentsView.SetBorder(true).SetTitle("Comments-第1页")
	for _, box := range []*tview.Box{leftListView.Box, postContentView.Box, commentsView.Box} {
		markFocusedBorder(box)
	}
	commentsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown:
			BBSView.NextPageComments()
			return nil
		case tcell.KeyUp:
			BBSView.PrevPageComments()
			return nil
		}
		return event
	})
	postPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	postPanel.AddItem(postContentView, 0, 1, false)
	postPanel.AddItem(commentsView, 0, 1, false)
	// 主界面
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftPanel, 0, 1, true).
		AddItem(postPanel, 0, 2, false)
	// 全局快捷键
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			if app.GetFocus() != leftListView {
				app.SetFocus(leftListView)
			} else {
				app.Stop()
			}
			return nil
		} else if event.Key() == tcell.KeyCtrlC {
			app.Stop()
		} else if event.Key() == tcell.KeyCtrlQ || event.Key() == tcell.KeyRight {
			// logger.Debug.Printf("Focus: %+v\n", app.GetFocus())
			// 切换焦点
			switch app.GetFocus() {
			case leftListView:
				// Opening the currently browsed headline should not show an older article.
				index := leftListView.current
				if index < len(BBSView.BBSList) && BBSView.SelectedBBS.Uid != BBSView.BBSList[index].Uid {
					onSelected(index)
				} else {
					app.SetFocus(postContentView)
				}
			case postContentView:
				app.SetFocus(commentsView)
			case commentsView:
				app.SetFocus(leftListView)
			default:
				app.SetFocus(leftListView)
			}
			return nil
		}
		return event
	})
	// 左侧快捷键
	leftPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			BBSView.NextPageBBS()
		} else if event.Key() == tcell.KeyCtrlP {
			BBSView.PrevPageBBS()
		}
		return event
	})
	// 右侧快捷键
	postPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			BBSView.NextPageComments()
		} else if event.Key() == tcell.KeyCtrlP {
			BBSView.PrevPageComments()
		}
		return event
	})

	pages.AddPage("BBS", mainFlex, true, true)

	BBSView = &BBSViewT{
		Region:       region,
		Main:         pages,
		BBSListView:  leftListView,
		PostView:     postContentView,
		CommentsView: commentsView,
		HelpContent:  help_content,
	}
	if strings.HasPrefix(string(region), "news:") {
		BBSView.newsFeed = spider.NewTeamNewsFeed(region)
	}
	return BBSView
}

func onSelected(index int) {
	// 选择之后加载帖子内容和评论
	if index < 0 || index >= len(BBSView.BBSList) {
		return
	}
	bbs := BBSView.BBSList[index]
	var err error
	BBSView.PostContent, err = bbs.GetDetail()
	if err != nil {
		BBSView.PostView.SetText("读取失败：" + err.Error())
		return
	}
	BBSView.CommentPage = 0 // 重置评论页数
	BBSView.SelectedBBS = bbs
	BBSView.RefreshPost()
	BBSView.NextPageComments()
	app.SetFocus(BBSView.PostView)
}

func (b *BBSViewT) Display() {
	b.NextPageBBS()
	if err := app.SetRoot(b.Main, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// 刷新左侧标题
func (b *BBSViewT) RefreshTitle() {
	b.BBSListView.SetTitle(b.Region.Label() + "-" + fmt.Sprintf("第%d页", b.BBSPage))
}

// 刷新评论标题
func (b *BBSViewT) RefreshCommentsTitle() {
	b.CommentsView.SetTitle("Comments-" + fmt.Sprintf("第%d页", b.CommentPage))
}

// 帖子下一页
func (b *BBSViewT) NextPageBBS() { b.loadPage(b.BBSPage + 1) }
func (b *BBSViewT) PrevPageBBS() {
	page := b.BBSPage - 1
	if page < 1 {
		page = 1
	}
	b.loadPage(page)
}
func (b *BBSViewT) loadPage(page int) {
	var posts []spider.BBS
	var err error
	if b.newsFeed != nil {
		posts, err = b.newsFeed.Page(page)
	} else {
		posts, err = spider.GetBBSList(b.Region, page)
	}
	if err != nil {
		b.PostView.SetText("读取失败：" + err.Error())
		return
	}
	if len(posts) == 0 {
		b.PostView.SetText("没有更多内容。Esc 返回菜单。")
		return
	}
	b.BBSPage = page
	b.BBSList = posts
	b.RefreshBBSList()
	b.RefreshTitle()
}

// 评论下一页
func (b *BBSViewT) NextPageComments() { b.loadCommentPage(b.CommentPage + 1) }

func (b *BBSViewT) PrevPageComments() {
	if b.CommentPage <= 1 {
		return
	}
	b.loadCommentPage(b.CommentPage - 1)
}

func (b *BBSViewT) loadCommentPage(page int) {
	if b.SelectedBBS.Uid == "" {
		return
	}
	comments, err := b.SelectedBBS.GetComments(page)
	if err != nil {
		b.CommentsView.SetTitle(fmt.Sprintf("评论 · 第%d页 · 加载失败，请重试", b.CommentPage))
		return
	}
	if len(comments) == 0 {
		if b.CommentPage == 0 {
			b.Comments = nil
			b.RefreshComments()
		}
		b.CommentsView.SetTitle(fmt.Sprintf("评论 · 第%d页 · 没有更多评论", b.CommentPage))
		return
	}
	b.CommentPage = page
	b.Comments = comments
	b.RefreshComments()
	b.RefreshCommentsTitle()
}

// 刷新帖子列表
func (b *BBSViewT) RefreshBBSList() {
	items := make([]wrappedItem, 0, len(b.BBSList))
	for _, item := range b.BBSList {
		items = append(items, wrappedItem{text: item.Title, metadata: fmt.Sprintf("heat %d", item.ReplyCnt+item.LightCnt)})
	}
	b.BBSListView.SetEntries(items)
}

// 刷新帖子内容
func (b *BBSViewT) RefreshPost() {
	b.PostView.Clear()
	b.PostView.SetText(tview.Escape(strings.TrimSpace(b.PostContent)))
	logger.Debug.Printf("更新post content: %s", b.PostContent)
}

// 刷新评论
func (b *BBSViewT) RefreshComments() {
	items := make([]wrappedItem, 0, len(b.Comments))
	for _, item := range b.Comments {
		items = append(items, wrappedItem{text: item.Content, metadata: fmt.Sprintf("%s · 亮 %d", item.Nickname, item.LightCnt)})
	}
	b.CommentsView.SetEntries(items)
}

// Focus callbacks cover keyboard navigation, mouse clicks, and Esc equally.
func markFocusedBorder(box *tview.Box) {
	normal := box.GetBorderColor()
	box.SetFocusFunc(func() { box.SetBorderColor(tcell.ColorRed) })
	box.SetBlurFunc(func() { box.SetBorderColor(normal) })
}
