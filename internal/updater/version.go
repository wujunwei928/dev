// Package updater 实现安装方式感知的自更新支持。
package updater

import (
	"runtime/debug"
	"strings"
)

// Method 二进制安装方式, 由版本解析结果直接得出
type Method int

const (
	MethodRelease   Method = iota // release 二进制 (ldflags 注入版本)
	MethodGoInstall               // go install 构建
	MethodLocal                   // 本地构建
)

// BuildInfo 版本解析所需的构建信息最小集
type BuildInfo struct {
	MainVersion string // 主模块版本, go install 构建为 "vX.Y.Z", 本地构建为 "(devel)" 或 ""
	VCSRevision string // vcs.revision, 完整 commit hash; 非 VCS 构建为空
}

// VersionInfo 版本解析结果, 同时表达安装方式探测结论
type VersionInfo struct {
	Method  Method // 安装方式
	Version string // 展示用版本号: "0.2.0" / "v0.2.0" / "devel+0123456"
}

// ResolveVersion 解析版本号与安装方式。
// 三级优先: ldflags 注入 > buildinfo 主模块版本 > devel+短commit。
func ResolveVersion(ldflagsVersion string, bi BuildInfo) VersionInfo {
	if ldflagsVersion != "" {
		return VersionInfo{Method: MethodRelease, Version: ldflagsVersion}
	}
	if isSemVer(bi.MainVersion) {
		return VersionInfo{Method: MethodGoInstall, Version: bi.MainVersion}
	}
	if rev := shortCommit(bi.VCSRevision); rev != "" {
		return VersionInfo{Method: MethodLocal, Version: "devel+" + rev}
	}
	return VersionInfo{Method: MethodLocal, Version: "devel"}
}

// shortCommit 取 commit hash 前 7 位短形式; 空串原样返回。
func shortCommit(rev string) string {
	if len(rev) > 7 {
		return rev[:7]
	}
	return rev
}

// isSemVer 报告 s 是否为合法模块版本号 vX.Y.Z (可带 -prerelease)。
// 本地构建的主模块版本为 "(devel)" 或空串, 须据此与 go install 构建区分。
func isSemVer(s string) bool {
	if !strings.HasPrefix(s, "v") {
		return false
	}
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s = s[:i] // 去掉可选的 -prerelease
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 || !isNumeric(parts[0][1:]) {
		return false
	}
	return isNumeric(parts[1]) && isNumeric(parts[2])
}

// isNumeric 报告 s 是否为非空纯十进制数字串。
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// CurrentVersion 读取当前二进制的运行时构建信息, 与注入的 ldflags 版本一起解析。
// injected 为 release 构建经 -ldflags "-X github.com/wujunwei928/dev/cmd.Version=..." 注入的值,
// 未注入时传空串, 解析退回 buildinfo / devel 分支。
func CurrentVersion(injected string) VersionInfo {
	var bi BuildInfo
	if info, ok := debug.ReadBuildInfo(); ok {
		bi.MainVersion = info.Main.Version
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				bi.VCSRevision = s.Value
			}
		}
	}
	return ResolveVersion(injected, bi)
}
