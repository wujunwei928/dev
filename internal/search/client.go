package search

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	httpClient     *resty.Client
	httpClientOnce sync.Once
)

// GetHTTPClient 获取单例 HTTP 客户端
func GetHTTPClient() *resty.Client {
	httpClientOnce.Do(func() {
		httpClient = resty.New().
			SetTimeout(10 * time.Second).
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.81 Safari/537.36 Edg/104.0.1293.54").
			SetTransport(&http.Transport{
				MaxIdleConnsPerHost: 10,
			})
	})
	return httpClient
}
