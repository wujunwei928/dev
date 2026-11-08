# 安装方式感知的 update 自更新机制

dev 经 `go install`、GitHub Releases、Docker 三条渠道分发，单一更新路径无法同时服务；决定用 `debug.ReadBuildInfo()` 探测安装来源——`go install` 构建的二进制委托 `go install <module>@<version>` 更新，release/本地构建走 GitHub Releases 下载替换（sha256 校验 + 冒烟测试 + rename 原子交换，Windows 借「运行中的 exe 可 rename 不可删」完成替换与 `.old` 延迟清理）。发布流水线选用 goreleaser 本地手动运行，产物为 6 平台归档 + checksums.txt。校验强度比同类（pi agent 无 checksum 校验）更严，延续 27a891a 出站加固的安全基线。

## Considered Options

- 单一路径 Releases 下载替换：实现最简，但 `go install` 用户的 `GOPATH/bin` 与 module cache 版本记录脱节。
- 只委托 `go install`：零发布产物依赖，但无 Go 工具链的用户（Releases/Docker 渠道存在的理由）被排除。
- 参考系 pi agent（badlogic/pi-mono）：其自更新全部建立在包管理器/受管安装布局上，独立二进制形态明确不支持自更新；Go 单二进制 + rename 交换可补齐该缺口，故不照搬其 npm/受管布局。

## Consequences

- 版本号必须停止硬编码：解析顺序为 ldflags 注入 > buildinfo > `devel+<commit>`，这是安装方式探测的前提。
- Docker Hub 自动发布不纳入新流水线（docker.yml 已于 adb8485 移除），容器内 update 的结果随容器重建失效，README 需标注。
