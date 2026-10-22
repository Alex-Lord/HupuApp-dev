package auth

import (
	"context"
	"errors"
	"github.com/chromedp/cdproto/network"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoginState(t *testing.T) {
	for _, tc := range []struct {
		html        string
		valid, fail bool
	}{
		{`<script>window.$$data = {"isLogin":true};</script>`, true, false},
		{`<script>window.$$data={"isLogin":false,"nested":{"isLogin":true}};</script>`, false, false},
		{`<h1>访问验证</h1>`, false, true},
	} {
		got, e := loginState(strings.NewReader(tc.html))
		if got != tc.valid || (e != nil) != tc.fail {
			t.Fatalf("%v %v", got, e)
		}
	}
}
func TestPersistence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if e := Save("token=test"); e != nil {
		t.Fatal(e)
	}
	got, e := Load()
	if got != "token=test" || e != nil {
		t.Fatal("save/load failed")
	}
	p, _ := SessionPath()
	s, _ := os.Stat(p)
	if s.Mode().Perm() != 0600 {
		t.Fatal("wrong permissions")
	}
	if e = Save("bad\r\nheader"); e == nil {
		t.Fatal("accepted header injection")
	}
	got, _ = Load()
	if got != "token=test" {
		t.Fatal("invalid login overwrote valid session")
	}
}
func TestRedirectBoundary(t *testing.T) {
	client := &http.Client{}
	Attach(client)
	for _, target := range []string{"https://evil.example/", "https://sub.bbs.hupu.com/", "http://bbs.hupu.com/"} {
		u, _ := url.Parse(target)
		if client.CheckRedirect(&http.Request{URL: u}, nil) == nil {
			t.Fatal("allowed redirect", target)
		}
	}
}
func TestBrowserCookieFilter(t *testing.T) {
	got := bbsCookieHeader([]*network.Cookie{{Name: "token", Value: "ok", Domain: ".hupu.com"}, {Name: "other", Value: "bad", Domain: "example.com"}, {Name: "old", Value: "bad", Domain: "bbs.hupu.com", Expires: 1}}, time.Now())
	if got != "token=ok" {
		t.Fatal("unexpected cookies")
	}
	if _, err := BrowserLogin(context.Background(), io.Discard, 0); err == nil {
		t.Fatal("accepted invalid wait")
	}
}

func TestEnsureReusesSessionAndPreservesItOnNetworkError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if e := Save("token=test"); e != nil {
		t.Fatal(e)
	}
	calls := 0
	browser := func(context.Context, io.Writer, time.Duration) (string, error) { calls++; return "new", nil }
	got, e := ensure(context.Background(), io.Discard, time.Minute, func(context.Context, string) (bool, error) { return true, nil }, browser)
	if got != "token=test" || e != nil || calls != 0 {
		t.Fatal("valid session did not bypass browser")
	}
	_, e = ensure(context.Background(), io.Discard, time.Minute, func(context.Context, string) (bool, error) { return false, errors.New("offline") }, browser)
	if e == nil || calls != 0 {
		t.Fatal("network failure triggered login")
	}
	got, _ = Load()
	if got != "token=test" {
		t.Fatal("lost saved session")
	}
	_, e = ensure(context.Background(), io.Discard, time.Minute, func(context.Context, string) (bool, error) { return false, nil }, browser)
	if e != nil || calls != 1 {
		t.Fatal("expired cookie did not restore browser")
	}
}
