package auth

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func BrowserLogin(ctx context.Context, out io.Writer, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		return "", fmt.Errorf("登录等待时间必须大于零")
	}
	ctx, stop := context.WithTimeout(ctx, timeout)
	defer stop()
	profile, err := browserProfilePath()
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(profile, 0700); err != nil {
		return "", err
	}
	if err = os.Chmod(profile, 0700); err != nil {
		return "", err
	}
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts, chromedp.UserDataDir(profile), chromedp.Flag("restore-last-session", true), chromedp.Flag("headless", false), chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false), chromedp.UserAgent(UserAgent),
		chromedp.WindowSize(1100, 800))
	alloc, closeAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer closeAlloc()
	browser, closeBrowser := chromedp.NewContext(alloc)
	defer closeBrowser()
	target := "https://bbs.hupu.com/"
	loginURL := "https://passport.hupu.com/login?project=www&from=pc&jumpurl=" + url.QueryEscape(target)
	prompted := false
	fmt.Fprintln(out, "正在恢复上次的虎扑浏览器会话。如需验证，请在网页完成；之后会自动返回 功能菜单。")
	if err := chromedp.Run(browser, chromedp.Navigate(target)); err != nil {
		return "", fmt.Errorf("无法打开登录窗口（需要 Chrome / Chromium）：%w", err)
	}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("登录等待已结束：%w", ctx.Err())
		case <-chromedp.FromContext(browser).Browser.LostConnection:
			return "", fmt.Errorf("登录窗口已关闭，请重新运行 hupu login")
		case <-browser.Done():
			return "", fmt.Errorf("登录窗口已关闭，请重新运行 hupu login")
		case <-ticker.C:
			var cookies []*network.Cookie
			err := chromedp.Run(browser, chromedp.ActionFunc(func(ctx context.Context) error {
				var e error
				cookies, e = network.GetCookies().WithURLs([]string{"https://bbs.hupu.com/"}).Do(ctx)
				return e
			}))
			if err != nil {
				return "", fmt.Errorf("读取登录会话失败：%w", err)
			}
			cookie := bbsCookieHeader(cookies, time.Now())

			valid, err := Verify(ctx, cookie)
			if err != nil {
				// Login may still be propagating or a navigation may be in flight.
				// Keep waiting without saving an unverified session.
				continue
			}
			if !valid {
				if !prompted {
					prompted = true
					if err := chromedp.Run(browser, chromedp.Navigate(loginURL)); err != nil {
						return "", fmt.Errorf("打开登录页面失败：%w", err)
					}
				}
				continue
			}
			if err = Save(cookie); err != nil {
				return "", err
			}
			// Close and wait for our isolated browser before returning to the CLI.
			// Graceful shutdown flushes Chrome cookies and site storage to disk.
			if err := chromedp.Cancel(browser); err != nil {
				fmt.Fprintln(out, "Cookie 已保存；登录窗口关闭异常，请手动关闭该窗口。")
			}
			closeBrowser()
			closeAlloc()
			fmt.Fprintln(out, "登录成功，Cookie 已保存，登录窗口已关闭。正在进入 功能菜单…")
			return cookie, nil
		}
	}
}
func bbsCookieHeader(cookies []*network.Cookie, now time.Time) string {
	parts := []string{}
	for _, c := range cookies {
		domain := strings.TrimPrefix(c.Domain, ".")
		if domain != "hupu.com" && domain != "bbs.hupu.com" {
			continue
		}
		if c.Expires > 0 && c.Expires <= float64(now.Unix()) {
			continue
		}
		if c.Name == "" || strings.ContainsAny(c.Name+c.Value, "\r\n;") {
			continue
		}
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}

func browserProfilePath() (string, error) {
	path, err := SessionPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "browser-profile"), nil
}
