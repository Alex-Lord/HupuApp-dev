package spider

import (
	"github.com/go-resty/resty/v2"
	"github.com/wudizhangzhi/HupuApp/gohupu/auth"
	"github.com/wudizhangzhi/HupuApp/gohupu/logger"
	"net/http"
	"time"
)

var SpiderClient *resty.Client

var sessionCookie string

func SetSession(cookie string) { sessionCookie = cookie }

func New() *resty.Client {
	SpiderClient = resty.New().SetTimeout(15*time.Second).SetHeader("User-Agent", auth.UserAgent).
		SetPreRequestHook(func(c *resty.Client, r *http.Request) error {
			if r.URL.Scheme == "https" && r.URL.Host == "bbs.hupu.com" && sessionCookie != "" {
				r.Header.Set("Cookie", sessionCookie)
			}
			logger.Debug.Printf("Request: %+v\n", r.URL)
			return nil
		})
	auth.Attach(SpiderClient.GetClient())
	return SpiderClient
}
