package updater

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// MaxResponseBodySize 单个响应体大小上限, 防止异常响应耗尽内存。
	// 版本检查只读小体积元数据; 下载路径 (#8) 复用该客户端时再放宽
	MaxResponseBodySize = 2 << 20 // 2MB

	// MaxRedirects 单次请求允许的最大跳转次数
	MaxRedirects = 5
)

var (
	httpClient     *resty.Client
	httpClientOnce sync.Once
)

// GetHTTPClient 获取 updater 独立的 HTTP 客户端。
// 不复用 search 的 2MB 搜索单例: updater 的响应体上限按更新流程独立设定
func GetHTTPClient() *resty.Client {
	httpClientOnce.Do(func() {
		httpClient = resty.New().
			SetTimeout(10*time.Second).
			SetHeader("User-Agent", "dev-updater").
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

// redirectPolicy 限制跳转次数, 防止恶意跳转循环
func redirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) >= MaxRedirects {
		return fmt.Errorf("stopped after %d redirects", MaxRedirects)
	}
	return nil
}
