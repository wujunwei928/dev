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
