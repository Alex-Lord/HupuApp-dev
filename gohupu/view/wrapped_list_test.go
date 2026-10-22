package view

import (
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
