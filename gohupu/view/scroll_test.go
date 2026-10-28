package view

import (
	"github.com/gdamore/tcell/v2"
	"strings"
	"testing"
)

func BenchmarkScrollLayout(b *testing.B) {
	l := newWrappedList()
	items := make([]wrappedItem, 50)
	for i := range items {
		items[i] = wrappedItem{text: strings.Repeat("这是一条包含中文的较长评论，用于检查滚动排版性能。", 8), metadata: "昵称 · 亮 123"}
	}
	l.SetEntries(items)
	l.lines(50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.lines(50)
	}
}

func TestSelectionScrollsOnlyNecessaryLines(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	defer s.Fini()
	s.SetSize(20, 6)
	l := newWrappedList()
	l.SetRect(0, 0, 20, 6)
	l.SetEntries([]wrappedItem{{text: "a", metadata: "1"}, {text: "b", metadata: "2"}, {text: "c", metadata: "3"}, {text: "d", metadata: "4"}, {text: "e", metadata: "5"}})
	l.Draw(s)
	l.move(1)
	l.move(1)
	l.move(1)
	l.Draw(s)
	if l.offset != 2 {
		t.Fatalf("expected minimal scroll of 2 rows, got %d", l.offset)
	}
}

func TestLayoutCacheInvalidatesOnResizeAndContent(t *testing.T) {
	l := newWrappedList()
	l.SetEntries([]wrappedItem{{text: strings.Repeat("中文测试", 10), metadata: "heat 1"}})
	first, _ := l.lines(40)
	cached, _ := l.lines(40)
	if &first[0] != &cached[0] {
		t.Fatal("layout not reused")
	}
	narrow, _ := l.lines(10)
	if len(narrow) <= len(first) {
		t.Fatal("resize did not reflow")
	}
	l.SetEntries([]wrappedItem{{text: "new"}})
	changed, _ := l.lines(10)
	if len(changed) != 1 || changed[0].text != "new" {
		t.Fatal("stale layout")
	}
}

func TestWheelBoundaryHasNoOverscrollDebt(t *testing.T) {
	l := newWrappedList()
	l.SetRect(0, 0, 20, 3)
	l.SetEntries([]wrappedItem{{text: "one\ntwo\nthree\nfour\nfive"}})
	for i := 0; i < 100; i++ {
		l.scroll(1)
	}
	if l.offset != 2 {
		t.Fatal(l.offset)
	}
	l.scroll(-1)
	if l.offset != 1 {
		t.Fatal("reverse wheel delayed", l.offset)
	}
	for i := 0; i < 100; i++ {
		l.scroll(-1)
	}
	l.scroll(1)
	if l.offset != 1 {
		t.Fatal("reverse at start delayed")
	}
}
