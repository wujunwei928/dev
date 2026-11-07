package search

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// MaxResponseBodySize 单个响应体大小上限, 防止异常响应耗尽内存
	MaxResponseBodySize = 2 << 20 // 2MB

	// MaxRedirects 单次请求允许的最大跳转次数
	MaxRedirects = 5
)

var (
	httpClient     *resty.Client
	httpClientOnce sync.Once
)

// GetHTTPClient 获取单例 HTTP 客户端
func GetHTTPClient() *resty.Client {
	httpClientOnce.Do(func() {
		httpClient = resty.New().
			SetTimeout(10*time.Second).
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.81 Safari/537.36 Edg/104.0.1293.54").
			SetResponseBodyLimit(MaxResponseBodySize).
			SetRedirectPolicy(resty.RedirectPolicyFunc(redirectPolicy)).
			SetTransport(newTransport())
	})
	return httpClient
}

// newTransport 构造出站传输层, 遵循 HTTP(S)_PROXY/NO_PROXY 环境变量代理配置
func newTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConnsPerHost: 10,
		Proxy:               http.ProxyFromEnvironment,
	}
}

// redirectPolicy 限制跳转次数; 跨主机跳转时清理 Referer,
// 防止携带检索词的引擎URL泄漏到跳转目标主机
// (net/http 跳转时还会自动把上一跳URL设为Referer, 同样被此处的清理覆盖)
func redirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) >= MaxRedirects {
		return fmt.Errorf("stopped after %d redirects", MaxRedirects)
	}
	if len(via) > 0 && !strings.EqualFold(req.URL.Host, via[0].URL.Host) {
		req.Header.Del("Referer")
	}
	return nil
}
