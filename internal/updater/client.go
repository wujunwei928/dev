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
	// 版本检查只读小体积元数据; Release 归档下载走独立的 GetDownloadClient
	MaxResponseBodySize = 2 << 20 // 2MB

	// MaxDownloadBodySize 单个下载产物大小上限 (Release 归档), 流式读取时强制限额
	MaxDownloadBodySize = 100 << 20 // 100MB

	// MaxRedirects 单次请求允许的最大跳转次数
	MaxRedirects = 5
)

var (
	httpClient     *resty.Client
	httpClientOnce sync.Once

	downloadClient     *resty.Client
	downloadClientOnce sync.Once
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

// GetDownloadClient 获取 Release 归档下载专用客户端。
// 与元数据客户端共享代理与跳转策略, 但限额独立: 归档上限 100MB, 流式落盘不驻留内存;
// 超时放宽以容纳大归档下载
func GetDownloadClient() *resty.Client {
	downloadClientOnce.Do(func() {
		downloadClient = resty.New().
			SetTimeout(10*time.Minute).
			SetHeader("User-Agent", "dev-updater").
			SetRedirectPolicy(resty.RedirectPolicyFunc(redirectPolicy)).
			SetTransport(newTransport())
	})
	return downloadClient
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
