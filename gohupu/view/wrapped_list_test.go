package view

import (
	"github.com/go-resty/resty/v2"
	"github.com/wudizhangzhi/HupuApp/gohupu/spider"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func screenText(s tcell.SimulationScreen, w, h int) string {
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, _, width := s.GetContent(x, y)
			if r != 0 {
				b.WriteRune(r)
			}
			if width > 1 {
				x += width - 1
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestWrappedListLongTextNavigationAndResize(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(30, 12)
	list := newWrappedList()
	list.SetRect(0, 0, 30, 12)
	list.SetItems([]string{"恩比德谈新赛季：保持健康，全力以赴争取胜利\nheat 123", "马克西谈防守\nheat 42"})
	list.Draw(screen)
	rendered := screenText(screen, 30, 12)
	if !strings.Contains(rendered, "heat 123") || !strings.Contains(rendered, "胜利") {
		t.Fatal(rendered)
	}
	_, _, style, _ := screen.GetContent(0, 0)
	_, _, attrs := style.Decompose()
	if attrs&tcell.AttrBold == 0 || attrs&tcell.AttrReverse != 0 {
		t.Fatal("selection must be bold only")
	}
	selected := -1
	list.selected = func(i int) { selected = i }
	handler := list.InputHandler()
	focus := func(tview.Primitive) {}
	handler(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), focus)
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), focus)
	if selected != 1 {
		t.Fatal(selected)
	}
	list.SetRect(0, 0, 16, 6)
	list.Draw(screen)
	if !strings.Contains(screenText(screen, 16, 6), "heat 42") {
		t.Fatal("resize hid selection")
	}
}

func TestWrappedListLongCommentScrollAndEscape(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 5)
	list := newWrappedList()
	list.SetRect(0, 0, 20, 5)
	list.SetItems([]string{strings.Repeat("很长的评论内容", 20) + "\n[red]昵称 · 亮 575"})
	list.Draw(screen)
	handler := list.InputHandler()
	for i := 0; i < 30; i++ {
		handler(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), func(tview.Primitive) {})
		list.Draw(screen)
	}
	rendered := screenText(screen, 20, 5)
	if !strings.Contains(rendered, "[red]昵称") || !strings.Contains(rendered, "575") {
		t.Fatal(rendered)
	}
	list.SetItems(nil)
	list.Draw(screen)
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(tview.Primitive) {})
}

func TestMetadataUsesSilverWithoutBold(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(30, 8)
	list := newWrappedList()
	list.SetRect(0, 0, 30, 8)
	list.SetEntries([]wrappedItem{{text: "新闻正文", metadata: "heat 123"}})
	list.Draw(screen)
	_, _, style, _ := screen.GetContent(0, 1)
	fg, _, attrs := style.Decompose()
	if fg != tcell.ColorSilver || attrs&tcell.AttrBold != 0 {
		t.Fatalf("metadata style: %v %v", fg, attrs)
	}
}

func TestRightCyclesPanelsWithoutMovingSelection(t *testing.T) {
	oldApp := app
	app = tview.NewApplication()
	defer func() { app = oldApp }()
	b := NewBBSView("nba")
	b.BBSListView.SetItems([]string{"one", "two"})
	b.BBSListView.current = 1
	app.SetFocus(b.BBSListView)
	_, primitive := b.Main.GetFrontPage()
	handler := primitive.InputHandler()
	for _, want := range []tview.Primitive{b.PostView, b.CommentsView, b.BBSListView} {
		handler(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
		if app.GetFocus() != want {
			t.Fatal("wrong focused panel")
		}
	}
	if b.BBSListView.current != 1 {
		t.Fatal("right changed selected news")
	}
}

type commentTransport func(*http.Request) (*http.Response, error)

func (f commentTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestCommentArrowsChangePagesAndKeepFocus(t *testing.T) {
	oldApp := app
	app = tview.NewApplication()
	defer func() { app = oldApp }()
	old := spider.SpiderClient
	defer func() { spider.SpiderClient = old }()
	var paths []string
	spider.SpiderClient = resty.New()
	spider.SpiderClient.GetClient().Transport = commentTransport(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		html := "<div class='post-reply-list-wrapper'><div class='thread-content-detail'><p>评论</p></div></div>"
		if r.URL.Path == "/123-3.html" {
			html = ""
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(html)), Request: r}, nil
	})
	b := NewBBSView("nba")
	b.SelectedBBS = spider.BBS{Uid: "123"}
	b.CommentPage = 1
	app.SetFocus(b.CommentsView)
	handler := b.CommentsView.InputHandler()
	focus := func(p tview.Primitive) { app.SetFocus(p) }
	handler(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), focus)
	if b.CommentPage != 2 || app.GetFocus() != b.CommentsView {
		t.Fatal("down did not keep focus and advance")
	}
	handler(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), focus)
	if b.CommentPage != 2 || len(b.Comments) != 1 {
		t.Fatal("empty next page erased current comments")
	}
	handler(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), focus)
	handler(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), focus)
	if b.CommentPage != 1 || len(paths) != 3 {
		t.Fatal(b.CommentPage, paths)
	}
}
