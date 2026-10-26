// Package auth reuses the local login strategy developed for this user's CLI.
// Credentials remain outside the repository and are sent only to the BBS origin.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

func SessionPath() (string, error) {
	d, e := os.UserConfigDir()
	return filepath.Join(d, "hupu", "session.cookie"), e
}
func Load() (string, error) {
	p, e := SessionPath()
	if e != nil {
		return "", e
	}
	b, e := os.ReadFile(p)
	if os.IsNotExist(e) {
		return "", nil
	}
	return strings.TrimSpace(string(b)), e
}
func Save(cookie string) error {
	if cookie == "" || strings.ContainsAny(cookie, "\r\n") {
		return errors.New("无效 Cookie")
	}
	req := &http.Request{Header: http.Header{"Cookie": []string{cookie}}}
	if len(req.Cookies()) == 0 {
		return errors.New("无效 Cookie")
	}
	p, e := SessionPath()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), ".session-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.WriteString(cookie); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(f.Name(), p)
}

// Attach rejects cross-origin redirects instead of risking credential forwarding.
func Attach(client *http.Client) {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		if req.URL.Scheme != "https" || req.URL.Host != "bbs.hupu.com" {
			return errors.New("BBS redirected outside its origin")
		}
		return nil
	}
}
func Verify(ctx context.Context, cookie string) (bool, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://bbs.hupu.com/", nil)
	if e != nil {
		return false, e
	}
	req.Header.Set("User-Agent", UserAgent)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	Attach(client)
	res, e := client.Do(req)
	if e != nil {
		return false, e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return false, fmt.Errorf("登录验证 HTTP %d", res.StatusCode)
	}
	return loginState(io.LimitReader(res.Body, 8<<20))
}

var stateStart = regexp.MustCompile(`window\.\$\$data\s*=\s*`)

func loginState(r io.Reader) (bool, error) {
	doc, e := goquery.NewDocumentFromReader(r)
	if e != nil {
		return false, e
	}
	var value *bool
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		loc := stateStart.FindStringIndex(text)
		if loc == nil {
			return
		}
		var data struct {
			IsLogin *bool `json:"isLogin"`
		}
		if json.NewDecoder(strings.NewReader(text[loc[1]:])).Decode(&data) == nil {
			value = data.IsLogin
		}
	})
	if value == nil {
		return false, errors.New("无法识别登录状态，请稍后重试；不会清除已有会话")
	}
	return *value, nil
}
func Ensure(ctx context.Context, out io.Writer, wait time.Duration) (string, error) {
	return ensure(ctx, out, wait, Verify, BrowserLogin)
}
func ensure(ctx context.Context, out io.Writer, wait time.Duration, verify func(context.Context, string) (bool, error), browser func(context.Context, io.Writer, time.Duration) (string, error)) (string, error) {
	cookie, e := Load()
	if e != nil {
		return "", e
	}
	if cookie != "" {
		valid, e := verify(ctx, cookie)
		if e != nil {
			return "", e
		}
		if valid {
			fmt.Fprintln(out, "已复用本地登录会话。")
			return cookie, nil
		}
		fmt.Fprintln(out, "Cookie 已失效，尝试恢复上次的浏览器会话。")
	}
	return browser(ctx, out, wait)
}
