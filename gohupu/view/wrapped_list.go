package view

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var metadataColor = tcell.NewRGBColor(135, 175, 175) // Muted cyan, distinct from terminal white/gray.

// wrappedList keeps selection at the item level while scrolling wrapped lines.
type wrappedList struct {
	*tview.Box
	items           []wrappedItem
	current, offset int
	width           int
	reveal          bool
	selected        func(int)
	layoutWidth     int
	layoutLines     []wrappedLine
	layoutStarts    []int
}

func newWrappedList() *wrappedList {
	return &wrappedList{Box: tview.NewBox(), reveal: true}
}

type wrappedItem struct{ text, metadata string }
type wrappedLine struct {
	text     string
	metadata bool
}

func (l *wrappedList) SetItems(items []string) {
	entries := make([]wrappedItem, len(items))
	for i, text := range items {
		entries[i] = wrappedItem{text: text}
	}
	l.SetEntries(entries)
}

func (l *wrappedList) SetEntries(items []wrappedItem) {
	l.items = append([]wrappedItem(nil), items...)
	l.layoutStarts = nil
	l.layoutLines = nil
	l.current = 0
	l.offset = 0
	l.reveal = true
}

func (l *wrappedList) lines(width int) ([]wrappedLine, []int) {
	if l.layoutStarts != nil && l.layoutWidth == width {
		return l.layoutLines, l.layoutStarts
	}
	var lines []wrappedLine
	starts := make([]int, len(l.items)+1)
	for i, item := range l.items {
		starts[i] = len(lines)
		for part, text := range []string{item.text, item.metadata} {
			if part == 1 && text == "" {
				continue
			}
			for _, paragraph := range strings.Split(text, "\n") {
				wrapped := tview.WordWrap(tview.Escape(paragraph), width)
				if len(wrapped) == 0 {
					wrapped = []string{""}
				}
				for _, line := range wrapped {
					lines = append(lines, wrappedLine{text: line, metadata: part == 1})
				}
			}
		}
	}
	starts[len(l.items)] = len(lines)
	l.layoutWidth = width
	l.layoutLines = lines
	l.layoutStarts = starts
	return lines, starts
}

func (l *wrappedList) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, w, h := l.GetInnerRect()
	if w <= 0 || h <= 0 {
		return
	}
	lines, starts := l.lines(w)
	if l.width != w {
		l.width = w
		l.reveal = true
	}
	if len(l.items) > 0 && l.reveal {
		start, end := starts[l.current], starts[l.current+1]
		if start < l.offset || end-start > h {
			l.offset = start
		} else if end > l.offset+h {
			l.offset = end - h
		}
	}
	l.reveal = false
	maxOffset := len(lines) - h
	if maxOffset < 0 {
		maxOffset = 0
	}
	if l.offset > maxOffset {
		l.offset = maxOffset
	}
	if l.offset < 0 {
		l.offset = 0
	}
	for row := 0; row < h && l.offset+row < len(lines); row++ {
		n := l.offset + row
		line := lines[n].text
		color := tcell.ColorDefault
		if lines[n].metadata {
			color = metadataColor
		}
		if !lines[n].metadata && len(l.items) > 0 && n >= starts[l.current] && n < starts[l.current+1] {
			line = "[::b]" + line + "[::-]"
		}
		tview.Print(screen, line, x, y+row, w, tview.AlignLeft, color)
	}
}

func (l *wrappedList) move(delta int) {
	next := l.current + delta
	if next < 0 || next >= len(l.items) {
		return
	}
	l.current = next
	l.reveal = true
}

func (l *wrappedList) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		_, _, _, h := l.GetInnerRect()
		switch event.Key() {
		case tcell.KeyUp:
			l.move(-1)
		case tcell.KeyDown:
			l.move(1)
		case tcell.KeyHome:
			l.current = 0
			l.reveal = true
		case tcell.KeyEnd:
			if len(l.items) > 0 {
				l.current = len(l.items) - 1
				l.reveal = true
			}
		case tcell.KeyPgDn:
			l.scroll(max(1, h-1))
		case tcell.KeyPgUp:
			l.scroll(-max(1, h-1))
		case tcell.KeyEnter:
			if len(l.items) > 0 && l.selected != nil {
				l.selected(l.current)
			}
		}
	})
}

func (l *wrappedList) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return l.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		if !l.InRect(mx, my) {
			return false, nil
		}
		switch action {
		case tview.MouseScrollUp:
			l.scroll(-1)
		case tview.MouseScrollDown:
			l.scroll(1)
		case tview.MouseLeftClick, tview.MouseLeftDoubleClick:
			setFocus(l)
			_, y, w, _ := l.GetInnerRect()
			if w <= 0 {
				return true, nil
			}
			_, starts := l.lines(w)
			row := my - y + l.offset
			for i := range l.items {
				if row >= starts[i] && row < starts[i+1] {
					l.current = i
					if action == tview.MouseLeftDoubleClick && l.selected != nil {
						l.selected(i)
					}
					break
				}
			}
		default:
			return false, nil
		}
		return true, nil
	})
}

// Clamp each input event, so queued wheel events at a boundary do not create
// overscroll debt that delays the first event in the opposite direction.
func (l *wrappedList) scroll(delta int) {
	_, _, w, h := l.GetInnerRect()
	if w <= 0 || h <= 0 {
		return
	}
	lines, _ := l.lines(w)
	l.offset = max(0, min(l.offset+delta, max(0, len(lines)-h)))
	l.reveal = false
}
