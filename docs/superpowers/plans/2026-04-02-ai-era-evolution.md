# dev CLI AI 时代演进 实施计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 精简 dev CLI 工具集中 AI 时代低价值功能，提升代码质量，增加 `--json` 输出模式，并创建 Skill 文件让 AI 助手直接调用 CLI 能力。

**Architecture:** 分四个阶段实施。阶段 1 移除 7 个命令文件和 4 个 internal 包；阶段 2 统一错误处理、清理代码、补充测试；阶段 3 为搜索命令增加结构化 JSON 输出；阶段 4 创建 Claude Code Skill 文件。

**Tech Stack:** Go 1.23, Cobra, Viper, resty, goquery, pterm

**Spec:** `docs/superpowers/specs/2026-04-02-ai-era-evolution-design.md`

---

## 文件结构映射

### 阶段 1 要删除的文件
- `cmd/encode.go` (127 行) - 加密命令
- `cmd/decode.go` (92 行) - 解密命令
- `cmd/json.go` (83 行) - JSON 转换命令（含独立 init()）
- `cmd/sql.go` (106 行) - SQL 转换命令
- `cmd/word.go` (63 行) - 单词格式转换命令
- `cmd/time.go` (45 行) - 时间转换命令
- `cmd/pypi.go` (124 行) - PyPI 镜像命令（含独立 init()）
- `internal/json2struct/parser.go` (152 行) - JSON 转 struct 逻辑
- `internal/json2struct/fields.go` (54 行) - JSON 字段处理
- `internal/sql2struct/mysql.go` (108 行) - MySQL struct 转换
- `internal/sql2struct/template.go` (69 行) - SQL 模板
- `internal/word/word.go` (40 行) - 单词格式转换逻辑
- `internal/timer/time.go` (12 行) - 时间工具
- `internal/tools/tools.go` (28 行) - 无引用死代码包
- `cmd/console_bench_test.go` (29 行) - 字符串拼接 bench（编解码相关）

### 阶段 1 要修改的文件
- `cmd/root.go` - 移除 7 个命令注册、移除 SQL 配置、替换 homedir
- `go.mod` / `go.sum` - 移除 3 个依赖

### 阶段 2 要修改的文件
- `cmd/console.go` - 移除编解码 case 和提示项、删除无用 import
- `cmd/search.go` - 删除死代码 `concurrentSearch()`、`Run` → `RunE`
- `cmd/http.go` - 安全修复、`Run` → `RunE`
- `cmd/install.go` - `Run` → `RunE`
- `cmd/go.go` - `Run` → `RunE`
- `cmd/url.go` - `Run` → `RunE`
- `cmd/root.go` - 错误处理统一

### 阶段 2 要创建的文件
- `internal/search/search_test.go` - 搜索引擎参数测试
- `internal/search/client_test.go` - HTTP 客户端测试

### 阶段 3 要修改的文件
- `cmd/search.go` - 增加 `--json` flag
- `internal/search/search_engine.go` - 增加 JSON 输出结构体和转换函数

### 阶段 3 要创建的文件
- `internal/search/json_output.go` - JSON 输出结构体和转换逻辑

### 阶段 4 要创建的文件
- `.claude/skills/dev-cli.md` - AI Skill 文件

---

## 阶段 1：功能精简

### Task 1: 删除低价值命令文件

**Files:**
- Delete: `cmd/encode.go`
- Delete: `cmd/decode.go`
- Delete: `cmd/json.go`
- Delete: `cmd/sql.go`
- Delete: `cmd/word.go`
- Delete: `cmd/time.go`
- Delete: `cmd/pypi.go`

- [ ] **Step 1: 删除命令文件**

```bash
rm cmd/encode.go cmd/decode.go cmd/json.go cmd/sql.go cmd/word.go cmd/time.go cmd/pypi.go
```

- [ ] **Step 2: 验证编译失败**

Run: `go build ./...`
Expected: 编译失败（root.go 中引用了已删除的命令）

### Task 2: 删除低价值 internal 包和测试文件

**Files:**
- Delete: `internal/json2struct/parser.go`
- Delete: `internal/json2struct/fields.go`
- Delete: `internal/sql2struct/mysql.go`
- Delete: `internal/sql2struct/template.go`
- Delete: `internal/word/word.go`
- Delete: `internal/timer/time.go`
- Delete: `internal/tools/tools.go`
- Delete: `cmd/console_bench_test.go`

- [ ] **Step 1: 删除 internal 包目录和死代码包**

```bash
rm -rf internal/json2struct internal/sql2struct internal/word internal/timer internal/tools
rm cmd/console_bench_test.go
```

### Task 3: 清理 root.go 命令注册和配置

**Files:**
- Modify: `cmd/root.go`

`cmd/root.go` 当前内容（关键部分）：
- 第 9 行: `"github.com/mitchellh/go-homedir"` import
- 第 47-58 行: 命令注册（需移除 decode、encode、time、word、sql、json 的 AddCommand）
- 第 73 行: `homedir.Dir()` → 改为 `os.UserHomeDir()`
- 第 103-118 行: `ViperInitSet()` 中 SQL 配置（6 行 viper.Set）

- [ ] **Step 1: 移除已删除命令的注册行**

在 `root.go` 的 `init()` 函数中，删除以下行：
```go
rootCmd.AddCommand(NewCmdDecode())  // 字符串解密
rootCmd.AddCommand(NewCmdEncode())  // 字符串加密
rootCmd.AddCommand(NewCmdTime())    // 时间转换
rootCmd.AddCommand(NewCmdWord())    // 单词格式转换
rootCmd.AddCommand(NewCmdSql())     // sql相关
```

保留的命令注册：
```go
rootCmd.AddCommand(NewCmdConsole()) // 类似ipython的交互式命令行
rootCmd.AddCommand(NewCmdInstall()) // 安装命令到PATH
rootCmd.AddCommand(NewCmdSearch())  // 搜索
rootCmd.AddCommand(NewCmdOpen())    // 打开文件或目录
rootCmd.AddCommand(NewCmdHttp())    // http服务
rootCmd.AddCommand(NewCmdUrl())     // 打开网址，文件夹或文件
rootCmd.AddCommand(NewCmdGo())      // go相关
```

- [ ] **Step 2: 替换 homedir 为标准库**

将 import 中的 `"github.com/mitchellh/go-homedir"` 替换为无（删除该 import）。

将 `initConfig()` 中第 73-76 行：
```go
home, err := homedir.Dir()
if err != nil {
    log.Fatalf("get home dir fail: %s", err.Error())
}
```
替换为：
```go
home, err := os.UserHomeDir()
if err != nil {
    log.Fatalf("get home dir fail: %s", err.Error())
}
```

- [ ] **Step 3: 清理 ViperInitSet 中 SQL 配置**

将 `ViperInitSet()` 中 SQL 相关的 6 行 viper.Set 删除：
```go
// 删除以下行
viper.Set(SqlConfigType, DefaultSqlType)
viper.Set(SqlConfigHost, DefaultSqlHost)
viper.Set(SqlConfigUserName, "")
viper.Set(SqlConfigPassword, "")
viper.Set(SqlConfigDb, "")
viper.Set(SqlConfigCharset, DefaultSqlCharset)
```

清理后 `ViperInitSet()` 应为：
```go
func ViperInitSet() {
	// http
	viper.Set(HttpConfigPort, DefaultHttpPort)

	// search
	viper.Set(SearchConfigEngine, DefaultSearchEngine)
	viper.Set(SearchConfigType, DefaultSearchType)
	viper.Set(SearchConfigCliIsDesc, DefaultCliIsDesc)
}
```

- [ ] **Step 4: 验证编译通过**

Run: `go build ./...`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "refactor: 移除 AI 时代低价值功能（encode/decode/json/sql/word/time/pypi）

移除 7 个命令和 5 个 internal 包，精简约 1000 行代码：
- 编解码命令（AI 直接处理）
- 代码转换命令（json2struct/sql2struct/word/time）
- PyPI 镜像管理（功能单一）
- internal/tools 死代码包

同时清理 SQL 配置和 go-homedir 依赖。"
```

### Task 4: 清理依赖

**Files:**
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: 运行 go mod tidy 清理依赖**

```bash
go mod tidy
```

验证以下依赖已被移除：
- `github.com/go-sql-driver/mysql`
- `github.com/mitchellh/go-homedir`
- `github.com/tidwall/gjson` 可能仍被 search 包使用，保留

- [ ] **Step 2: 验证编译和测试**

Run: `go build ./... && go test ./...`
Expected: 编译成功，所有测试通过

- [ ] **Step 3: 提交**

```bash
git add go.mod go.sum
git commit -m "chore: 清理依赖（移除 mysql driver 和 go-homedir）"
```

---

## 阶段 2：代码质量提升

> 依赖阶段 1 完成。

### Task 5: 清理 console.go 编解码逻辑

**Files:**
- Modify: `cmd/console.go`

`cmd/console.go` 当前编解码相关内容：
- import: `crypto/md5`, `crypto/sha1`, `encoding/base64`, `encoding/hex`, `net/url`, `strconv`（unicode_decode 用）
- `ConsolePromptSuggestList` 中 8 个编解码提示项（第 27-34 行）
- switch 中 8 个编解码 case（第 81-127 行）

- [ ] **Step 1: 移除编解码提示项**

将 `ConsolePromptSuggestList` 中的 8 个编解码相关项删除，保留：
```go
var ConsolePromptSuggestList = []prompt.Suggest{
	{Text: "help", Description: "查看帮助，列出所有命令"},
	{Text: "//", Description: "使用默认浏览器打开网址"},
	{Text: "??", Description: "使用搜索引擎搜索关键字"},
	{Text: ">", Description: "执行命令行命令"},
	{Text: "open", Description: "使用默认程序打开文件"},
}
```

- [ ] **Step 2: 移除编解码 case 分支**

从 switch 中删除 8 个 case：`md5`, `sha1`, `base64_encode`, `base64_decode`, `url_encode`, `url_decode`, `unicode_encode`, `unicode_decode`。

保留的 case：`open`, `//`, `??`, `>`, `default`。

- [ ] **Step 3: 清理无用 import**

删除不再使用的 import：`crypto/md5`, `crypto/sha1`, `encoding/base64`, `encoding/hex`, `net/url`, `strconv`。

- [ ] **Step 4: 验证编译通过**

Run: `go build ./...`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add cmd/console.go
git commit -m "refactor: 移除 console 中编解码逻辑（8 个命令已移除）"
```

### Task 6: 删除 search.go 死代码

**Files:**
- Modify: `cmd/search.go`

`cmd/search.go` 中第 40-73 行的 `concurrentSearch()` 函数是死代码（未被调用，实际使用的是 `search.ConcurrentSearch()`）。

- [ ] **Step 1: 删除死代码**

删除 `cmd/search.go` 中第 39-73 行的 `concurrentSearch()` 函数。

同时清理对应的无用 import：`sync`（如果只被该函数使用）。检查 `sync` 是否在其他地方使用——如果仅 `concurrentSearch` 使用则删除。

- [ ] **Step 2: 验证编译通过**

Run: `go build ./...`

- [ ] **Step 3: 提交**

```bash
git add cmd/search.go
git commit -m "refactor: 删除 search.go 中未使用的 concurrentSearch 死代码"
```

### Task 7: 统一错误处理 - search.go

**Files:**
- Modify: `cmd/search.go`

- [ ] **Step 1: 将 Run 改为 RunE**

将 `NewCmdSearch()` 中的 `Run` 改为 `RunE`，函数签名从 `func(cmd *cobra.Command, args []string)` 改为 `func(cmd *cobra.Command, args []string) error`。

将 `log.Fatalf(...)` 改为 `return fmt.Errorf(...)`。涉及两处：
- 第 119 行: `log.Fatalf("request search engine fail: %s", err.Error())` → `return fmt.Errorf("request search engine fail: %w", err)`
- 第 159 行: `log.Fatalf("检索失败: %s", err.Error())` → `return fmt.Errorf("检索失败: %w", err)`

确保 import 中添加 `"fmt"`（如果还没有），移除 `"log"`（如果不再使用）。

- [ ] **Step 2: 验证编译通过**

Run: `go build ./...`

- [ ] **Step 3: 提交**

```bash
git add cmd/search.go
git commit -m "refactor: search 命令错误处理改为 RunE + return error"
```

### Task 8: 统一错误处理 - http.go + 安全修复

**Files:**
- Modify: `cmd/http.go`

- [ ] **Step 1: 修复文件上传路径穿越漏洞**

将 `/upload` handler 中的 `os.Create(fh.Filename)` 改为使用 `filepath.Base()`：
```go
safeFilename := filepath.Base(fh.Filename)
dst, err := os.Create(safeFilename)
```

同步修改下面的路径拼接：
```go
fileFullSavePath := filepath.Join(pwd, safeFilename)
```

- [ ] **Step 2: 增加上传大小限制**

在 `/upload` handler 开头增加：
```go
r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB 限制
```

- [ ] **Step 3: 验证编译通过**

Run: `go build ./...`

- [ ] **Step 4: 提交**

```bash
git add cmd/http.go
git commit -m "fix: http 上传安全修复（路径穿越防护 + 100MB 大小限制）"
```

### Task 9: 统一错误处理 - install.go, go.go, url.go

**Files:**
- Modify: `cmd/install.go`
- Modify: `cmd/go.go`
- Modify: `cmd/url.go`

- [ ] **Step 1: install.go 改为 RunE**

将 `Run` 改为 `RunE`，将 `log.Fatalln(...)` 改为 `return fmt.Errorf(...)`。
移除 `"log"` import。

- [ ] **Step 2: go.go 改为 RunE**

将两个子命令的 `Run` 改为 `RunE`，将 `log.Fatalln(...)` / `log.Fatal(...)` 改为 `return fmt.Errorf(...)`。

- [ ] **Step 3: url.go 改为 RunE**

将 `Run` 改为 `RunE`。检查 `search.Open()` 返回的 error 并返回。

- [ ] **Step 4: 验证编译通过**

Run: `go build ./...`

- [ ] **Step 5: 提交**

```bash
git add cmd/install.go cmd/go.go cmd/url.go
git commit -m "refactor: install/go/url 命令错误处理改为 RunE + return error"
```

### Task 10: 补充搜索包单元测试

**Files:**
- Create: `internal/search/search_test.go`
- Create: `internal/search/client_test.go`

- [ ] **Step 1: 编写搜索引擎参数测试**

创建 `internal/search/search_test.go`：

```go
package search

import (
	"testing"
)

func TestEngineParamMap_HasAllEngines(t *testing.T) {
	engines := []string{
		EngineBing, EngineBaidu, EngineGoogle,
		EngineZhiHu, EngineWeiXin, EngineGithub,
		EngineKaiFa, EngineDouBan, EngineMovie,
		EngineBook, Engine360, EngineSoGou,
	}
	for _, eng := range engines {
		param, ok := EngineParamMap[eng]
		if !ok {
			t.Errorf("engine %q not found in EngineParamMap", eng)
			continue
		}
		if param.Domain == "" {
			t.Errorf("engine %q has empty Domain", eng)
		}
	}
}

func TestFormatSearchUrl_Bing(t *testing.T) {
	url := FormatSearchUrl(EngineBing, "golang")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if !contains(url, "bing.com") {
		t.Errorf("expected bing.com in URL, got %s", url)
	}
}

func TestFormatSearchUrl_UnknownEngine(t *testing.T) {
	url := FormatSearchUrl("nonexistent", "test")
	if url == "" {
		t.Fatal("expected fallback to bing URL")
	}
	if !contains(url, "bing.com") {
		t.Errorf("expected bing.com fallback URL, got %s", url)
	}
}

func TestGetEngineParamCached(t *testing.T) {
	param := getEngineParamCached(EngineBing)
	if param.Domain == "" {
		t.Error("expected non-empty Domain")
	}

	// 第二次调用应命中缓存
	param2 := getEngineParamCached(EngineBing)
	if param2.Domain != param.Domain {
		t.Error("cache should return same param")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: 编写 HTTP 客户端测试**

创建 `internal/search/client_test.go`：

```go
package search

import (
	"testing"
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
```

- [ ] **Step 3: 运行测试**

Run: `go test ./internal/search/ -v -count=1`
Expected: 所有测试通过

- [ ] **Step 4: 提交**

```bash
git add internal/search/search_test.go internal/search/client_test.go
git commit -m "test: 补充 search 包单元测试（引擎参数、HTTP 客户端）"
```

---

## 阶段 3：输出优化

> 依赖阶段 2 完成。

### Task 11: 定义 JSON 输出结构体

**Files:**
- Create: `internal/search/json_output.go`

- [ ] **Step 1: 创建 JSON 输出结构体和转换函数**

创建 `internal/search/json_output.go`：

```go
package search

import "encoding/json"

// SearchResult JSON 模式的搜索结果
type SearchResult struct {
	Engine  string       `json:"engine"`
	Query   string       `json:"query"`
	Results []ResultItem `json:"results"`
}

// ResultItem 单条搜索结果
type ResultItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// keyToFieldMapping 中文 key 到英文字段名的映射
var keyToFieldMapping = map[string]string{
	"标题": "title",
	"名称": "title",
	"链接": "url",
	"描述": "description",
	"评分": "",
	"评价人数": "",
	"出版信息": "",
}

// KeyValToResultItems 将 [][]KeyVal 转换为 []ResultItem
func KeyValToResultItems(engine string, query string, kvResults [][]KeyVal) SearchResult {
	items := make([]ResultItem, 0, len(kvResults))
	for _, kvRow := range kvResults {
		item := ResultItem{}
		for _, kv := range kvRow {
			field := keyToFieldMapping[kv.Key]
			switch field {
			case "title":
				item.Title = kv.Val
			case "url":
				item.URL = kv.Val
			case "description":
				item.Description = kv.Val
			}
		}
		items = append(items, item)
	}
	return SearchResult{
		Engine:  engine,
		Query:   query,
		Results: items,
	}
}

// ToJSON 将搜索结果序列化为 JSON 字符串
func (sr SearchResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(sr, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
```

- [ ] **Step 2: 编写 JSON 输出测试**

创建 `internal/search/json_output_test.go`：

```go
package search

import (
	"encoding/json"
	"testing"
)

func TestKeyValToResultItems(t *testing.T) {
	input := [][]KeyVal{
		{
			{Key: "标题", Val: "Test Title"},
			{Key: "链接", Val: "https://example.com"},
			{Key: "描述", Val: "Test Description"},
		},
	}

	result := KeyValToResultItems("bing", "test query", input)

	if result.Engine != "bing" {
		t.Errorf("expected engine 'bing', got %q", result.Engine)
	}
	if result.Query != "test query" {
		t.Errorf("expected query 'test query', got %q", result.Query)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", result.Results[0].Title)
	}
	if result.Results[0].URL != "https://example.com" {
		t.Errorf("expected url 'https://example.com', got %q", result.Results[0].URL)
	}
}

func TestKeyValToResultItems_DoubanKeyName(t *testing.T) {
	// 豆瓣用 "名称" 而非 "标题"
	input := [][]KeyVal{
		{
			{Key: "名称", Val: "Douban Title"},
			{Key: "链接", Val: "https://douban.com/xxx"},
			{Key: "描述", Val: "Douban Desc"},
		},
	}

	result := KeyValToResultItems("douban", "test", input)

	if result.Results[0].Title != "Douban Title" {
		t.Errorf("expected '名称' mapped to title, got %q", result.Results[0].Title)
	}
}

func TestSearchResult_ToJSON(t *testing.T) {
	sr := SearchResult{
		Engine: "bing",
		Query:  "test",
		Results: []ResultItem{
			{Title: "Title", URL: "https://example.com", Description: "Desc"},
		},
	}

	jsonStr, err := sr.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 验证是有效 JSON
	var parsed SearchResult
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed.Engine != "bing" {
		t.Error("JSON roundtrip failed")
	}
}
```

- [ ] **Step 3: 运行测试**

Run: `go test ./internal/search/ -v -count=1`
Expected: 所有测试通过

- [ ] **Step 4: 提交**

```bash
git add internal/search/json_output.go internal/search/json_output_test.go
git commit -m "feat: 添加搜索结果 JSON 输出结构体和转换逻辑"
```

### Task 12: search 命令增加 --json flag

**Files:**
- Modify: `cmd/search.go`

- [ ] **Step 1: 添加 --json flag 和输出逻辑**

在 `NewCmdSearch()` 中：
1. 添加 `var jsonOutput bool` 变量
2. 添加 flag: `cmd.Flags().BoolVar(&jsonOutput, "json", false, "以 JSON 格式输出搜索结果")`
3. 在 `SearchTypeCli` 分支中，当 `jsonOutput` 为 true 时，使用 `KeyValToResultItems` 转换并输出 JSON

关键逻辑：
```go
if jsonOutput {
	result := search.KeyValToResultItems(searchMode, searchStr, searchRes)
	jsonStr, err := result.ToJSON()
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}
	fmt.Println(jsonStr)
	return nil
}
```

- [ ] **Step 2: 更新 Example**

在 `getSearchExample()` 中添加 JSON 输出示例行：
```go
{`JSON 格式输出搜索结果`, `dev search -m bing -t cli --json "golang orm"`},
```

- [ ] **Step 3: 验证编译和测试**

Run: `go build ./... && go test ./...`

- [ ] **Step 4: 提交**

```bash
git add cmd/search.go
git commit -m "feat: search 命令增加 --json flag，支持结构化输出"
```

---

## 阶段 4：Skill 集成

> 无硬依赖，可与阶段 3 并行。

### Task 13: 创建 Claude Code Skill 文件

**Files:**
- Create: `.claude/skills/dev-cli.md`

- [ ] **Step 1: 创建 Skill 文件**

创建 `.claude/skills/dev-cli.md`：

```markdown
---
name: dev-cli
description: dev CLI 工具集 - 终端搜索、文件服务和 Go 工具
---

# dev CLI 工具集

开发者工具集 CLI，提供终端搜索、文件服务和 Go 开发辅助功能。

## 网页搜索

搜索互联网内容，支持多个搜索引擎，结果直接显示在终端。

- 网页搜索：`dev search "查询内容" --engine bing --cli`
- GitHub 搜索：`dev search "query" --engine github --cli`
- 并发搜索（Bing+Baidu+Google 同时搜索）：`dev search "query" --concurrent --cli`
- JSON 输出（推荐用于程序化解析）：`dev search "query" --cli --json`
- 浏览器搜索：`dev search "query" --engine baidu -t browser`

支持的搜索引擎：bing, baidu, google, github, zhihu, weixin, kaifa, douban, movie, book, 360, sogou

## 文件服务

快速启动 HTTP 文件服务器，支持上传下载。

- 启动服务：`dev http --port 8899`
- 启动并启用上传：`dev http --port 8899 --use_upload`
- 下载：访问 `http://localhost:8899/static/`
- 上传：访问 `http://localhost:8899` 使用拖拽上传

## Go 工具

- 模糊搜索常用 Go 包：`dev go search -p "关键词"`
- 安装 Go 包：`dev go search -p "gin" -i`
- 交叉编译：`dev go build main.go -n myapp -p build_output`

## 其他工具

- 安装 CLI 到 PATH：`dev install`
- 打开网址：`dev url https://example.com` 或 `dev // example.com`
- 打开文件：`dev open /path/to/file`
```

- [ ] **Step 2: 提交**

```bash
git add .claude/skills/dev-cli.md
git commit -m "feat: 添加 Claude Code Skill 文件，让 AI 助手直接调用 CLI"
```

### Task 14: 最终验证

- [ ] **Step 1: 完整编译和测试**

Run: `go build ./... && go test ./... -v`
Expected: 编译成功，所有测试通过

- [ ] **Step 2: 构建 binary 并验证基本功能**

```bash
go build -o dev ./main.go
./dev --version
./dev --help
```

验证帮助信息中不再出现 encode、decode、json、sql、word、time、pypi 命令。

- [ ] **Step 3: 最终提交（如有遗漏修复）**

检查 `git status`，确保无遗漏文件。
