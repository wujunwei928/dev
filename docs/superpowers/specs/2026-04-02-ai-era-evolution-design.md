# dev CLI 工具集 - AI 时代演进设计

## 背景与动机

dev 是一个 Go 语言的开发者工具集 CLI，基于 Cobra 框架，提供搜索、编解码、代码转换、文件服务等功能。在 AI 编程时代，许多"转换类"工具（如 JSON→struct、编解码、时间转换）的价值已被 AI 助手取代。

本次演进的目标是：
1. **精简** - 移除 AI 时代低价值功能，聚焦核心竞争力
2. **质量** - 提升代码质量、统一错误处理、补充测试
3. **集成** - 通过 Skill 文件让 AI 助手直接调用 CLI 能力
4. **演进** - 保持架构灵活性，未来可扩展为 MCP Server

---

## 阶段一：功能精简

### 移除的功能

| 命令 | 包 | 移除原因 |
|------|-----|---------|
| `encode` | `cmd/encode.go` | AI 直接做编解码 |
| `decode` | `cmd/decode.go` | AI 直接做编解码 |
| `json` | `cmd/json.go` + `internal/json2struct/` | AI 直接转换 |
| `sql` | `cmd/sql.go` + `internal/sql2struct/` | AI 直接转换 |
| `word` | `cmd/word.go` + `internal/word/` | AI 直接做格式转换 |
| `time` | `cmd/time.go` + `internal/timer/` | AI 直接做时间转换 |
| `pypi` | `cmd/pypi.go` | 功能单一，使用场景少 |
| console 中的编解码部分 | `cmd/console.go` | 重复逻辑，AI 可替代 |

### 保留并增强的功能

| 命令 | 定位 | 后续增强 |
|------|------|---------|
| **search** | 核心功能 | 增加 `--json` 输出，封装为 Skill 工具 |
| **http** | 文件服务 | 安全修复，封装为 Skill 工具 |
| **console** | 交互入口 | 精简后保留核心交互能力 |
| **go** | Go 包搜索 | 与搜索能力整合 |
| **install** | 工具安装 | 保持不变 |
| **open/url** | 系统操作 | 保持不变 |

### 依赖清理

- 移除 `go-homedir` → 用 `os.UserHomeDir()` 替代
- 移除 `go-sql-driver/mysql`（sql 命令移除后）
- 修复 `strings.Title()` 弃用 → 用 `golang.org/x/text/cases`
- 统一 CI Go 版本为 1.23（当前 CI 仍为 1.17）

**预计影响：** 移除约 800-900 行代码，项目从 ~3000 行精简到 ~2100 行。

---

## 阶段二：代码质量提升

### 2.1 错误处理统一

**当前问题：** `log.Fatalf`（直接退出）、`log.Printf`（打印继续）、`return err` 三种模式混用。

**改进方案：**
- 所有命令的 `Run` 改为 `RunE`（返回 error）
- 在 `root.go` 的 `PersistentPreRunE` 或 `cmd.SetErrorFunc` 中统一处理错误输出
- 移除所有 `log.Fatalf`，改为返回错误
- 错误输出格式：`Error: <message>` 统一前缀

### 2.2 代码清理

- 删除 `cmd/search.go` 中未使用的 `concurrentSearch()` 死代码
- 统一命令注册模式：全部使用 `rootCmd.AddCommand(NewCmdXxx())` 模式
- 清理 console.go 中内嵌的编解码逻辑

### 2.3 测试补充

为保留的核心包增加单元测试：
- `internal/search/` - 搜索引擎参数映射、HTML/JSON 解析规则
- `cmd/` - 关键命令的基本输入输出测试

### 2.4 安全修复

- `cmd/http.go` 文件上传：增加路径安全检查，防止路径穿越攻击
  - 使用 `filepath.Base()` 或 `filepath.Clean()` 处理上传文件名
  - 验证文件路径在允许的目录范围内

---

## 阶段三：输出优化

### 3.1 搜索命令增加 `--json` 输出

为 `dev search` 命令增加 `--json` flag，输出结构化 JSON 结果：

```json
{
  "engine": "bing",
  "query": "golang orm",
  "results": [
    {
      "title": "GORM - The fantastic ORM library for Golang",
      "url": "https://gorm.io/",
      "description": "..."
    }
  ]
}
```

这让 AI 助手通过 Skill 调用时更容易解析结果。

### 3.2 统一退出码

- 成功：退出码 0
- 一般错误：退出码 1
- 让 AI 助手通过退出码判断执行结果

### 3.3 错误输出标准化

错误输出统一为 JSON 格式（当指定 `--json` 时）：

```json
{ "error": "search engine 'xxx' not found" }
```

---

## 阶段四：Skill 集成

### 4.1 创建 AI Skill 文件

创建一个 Claude Code Skill 文件，教 AI 助手如何使用 dev CLI：

```markdown
# dev CLI 工具集

## 搜索
- 网页搜索：`dev search "查询内容" --engine bing --cli`
- GitHub 搜索：`dev search "query" --engine github --cli`
- 并发搜索：`dev search "query" --concurrent --cli`
- JSON 输出：`dev search "query" --cli --json`

## 文件服务
- 启动：`dev http --port 8899`
- 上传：访问 http://localhost:8899

## Go 工具
- 包搜索：`dev go search "关键词"`
```

### 4.2 使用方式

用户将 Skill 文件安装到 Claude Code，之后 AI 助手即可：
- 直接调用搜索功能获取实时网络信息
- 启动文件服务进行文件分享
- 搜索 Go 包信息

**零额外代码**，完全复用现有 CLI 能力。

---

## 未来扩展：MCP Server（可选）

当前设计不阻碍未来向 MCP Server 演进：

- 如果 Skill 方案不够用（需要状态管理、结构化交互等），可在 `internal/mcp/` 中实现 MCP Server
- 新增 `dev mcp` 子命令启动 stdio/SSE 模式的 MCP 服务
- 复用现有 `internal/search/` 的搜索能力

MCP 作为**可选增强**，不是当前优先级。

---

## 实施计划

| 阶段 | 内容 | 预计变更 |
|------|------|---------|
| 阶段 1 | 功能精简 | 删除 ~800 行，移除 3 个依赖 |
| 阶段 2 | 质量提升 | 修改 ~200 行，新增 ~300 行测试 |
| 阶段 3 | 输出优化 | 修改 ~100 行，新增 --json flag |
| 阶段 4 | Skill 集成 | 新增 1 个 Markdown 文件 |

每个阶段独立可交付，风险可控。
