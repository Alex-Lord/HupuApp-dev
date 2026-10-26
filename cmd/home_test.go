package main

import (
	"github.com/alecthomas/kong"
	"strings"
	"testing"
)

func TestDefaultHomeCommand(t *testing.T) {
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Command() != "home" {
		t.Fatal(ctx.Command())
	}
}
func TestHomeSeparatesNewsAndForum(t *testing.T) {
	items := homeItems("sixers")
	if !strings.Contains(items[2], "76人新闻") || !strings.Contains(items[3], "76人专区") {
		t.Fatal(items)
	}
}
