package search

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetHTTPClient_NotNil(t *testing.T) {
	client := GetHTTPClient()
	if client == nil {
		t.Fatal("expected non-nil HTTP client")
	}
}

func TestGetHTTPClient_Singleton(t *testing.T) {
	client1 := GetHTTPClient()
	client2 := GetHTTPClient()
	if client1 != client2 {
		t.Error("expected same client instance (singleton)")
	}
}

// 传输层必须遵循 HTTP(S)_PROXY 环境变量, 不能静默直连
func TestNewTransportUsesEnvironmentProxy(t *testing.T) {
	transport := newTransport()
	if transport.Proxy == nil {
		t.Error("transport should honor HTTP(S)_PROXY environment proxy")
	}
}

// 跨主机跳转时, 请求上的 Referer 与 net/http 自动生成的 Referer 都必须被清理
func TestRedirectPolicyDropsRefererCrossHost(t *testing.T) {
	var receivedReferer atomic.Value
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReferer.Store(r.Header.Get("Referer"))
	}))
	defer target.Close()

	// 不同端口即不同主机, 模拟跨主机跳转
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/landing", http.StatusFound)
	}))
	defer source.Close()

	_, err := GetHTTPClient().R().
		SetHeader("Referer", "https://www.bing.com/search?q=secret").
		Get(source.URL + "/redirect")
	if err != nil {
		t.Fatalf("cross-host redirect request failed: %v", err)
	}
	if referer := receivedReferer.Load(); referer != nil && referer.(string) != "" {
		t.Errorf("Referer leaked to cross-host redirect target: %q", referer.(string))
	}
}

// 同主机跳转保留 Referer, 满足引擎侧的来源校验
func TestRedirectPolicyKeepsRefererSameHost(t *testing.T) {
	var receivedReferer atomic.Value
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReferer.Store(r.Header.Get("Referer"))
	})
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		final(w, r)
	}))
	defer source.Close()

	_, err := GetHTTPClient().R().
		SetHeader("Referer", "https://www.bing.com/search?q=secret").
		Get(source.URL + "/redirect")
	if err != nil {
		t.Fatalf("same-host redirect request failed: %v", err)
	}
	referer, _ := receivedReferer.Load().(string)
	if referer == "" {
		t.Error("Referer should be kept on same-host redirect")
	}
}

func TestRedirectPolicyLimitsHops(t *testing.T) {
	var hops int32
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hops, 1)
		http.Redirect(w, r, fmt.Sprintf("/hop-%d", n), http.StatusFound)
	}))
	defer source.Close()

	_, err := GetHTTPClient().R().Get(source.URL + "/start")
	if err == nil {
		t.Fatal("expected error for redirect loop")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("stopped after %d redirects", MaxRedirects)) {
		t.Errorf("unexpected redirect limit error: %v", err)
	}
}

func TestResponseBodyLimit(t *testing.T) {
	// 超过 MaxResponseBodySize 的响应应报错, 而非整体读入内存
	big := strings.Repeat("a", MaxResponseBodySize+(1<<20))
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprint(len(big)))
		fmt.Fprint(w, big)
	}))
	defer source.Close()

	_, err := GetHTTPClient().R().Get(source.URL)
	if err == nil {
		t.Error("expected error when response body exceeds limit")
	}
}

// 慢速服务器应在客户端超时内中断, 不无限等待
func TestClientTimeoutStillApplied(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 客户端超时断开后请求上下文取消, handler 随之退出, 避免阻塞 Close
		<-r.Context().Done()
	}))
	defer source.Close()

	start := time.Now()
	_, err := GetHTTPClient().R().Get(source.URL)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Errorf("request took too long before timeout: %v", elapsed)
	}
}
