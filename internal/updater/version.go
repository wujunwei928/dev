// Package updater 实现安装方式感知的自更新支持。
package updater

import (
	"regexp"
	"runtime/debug"
	"strconv"
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
	if isSemVer(bi.MainVersion) && !isPseudoVersion(bi.MainVersion) {
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

// pseudoVersionRe 匹配 go 命令由 VCS 状态推导的伪版本, 三种形态共用同一 <14位UTC时间戳>-<commit hash> 后缀:
// vX.0.0-<ts>-<hash> / vX.Y.Z-<pre>.0.<ts>-<hash> / vX.Y[.Z]-N.0.<ts>-<hash>
var pseudoVersionRe = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+-([0-9A-Za-z]+\.)*[0-9]{14}-[0-9a-f]{12,}$`)

// isPseudoVersion 报告 s 是否为 go 命令生成的伪版本。
// 伪版本表示构建时由 VCS 状态推导 (含本地 go build 与 go install @branch),
// 不是 go install 安装的发布版本, 故不计入 go install 探测。
func isPseudoVersion(s string) bool {
	return pseudoVersionRe.MatchString(s)
}

// isSemVer 报告 s 是否为合法 semver 版本号 (v 前缀可选, 可带 -prerelease / +build 后缀)。
// 用于从主模块版本识别 go install 构建; 本地构建的主模块版本为 "(devel)" 或空串, 据此区分。
// go 模块版本恒带 v 前缀且不含 build metadata, parseSemver 的额外宽容不影响安装方式判定
func isSemVer(s string) bool {
	_, ok := parseSemver(s)
	return ok
}

// eqVersion 报告两个版本号去 v/V 前缀后是否字面等价 ("v0.2.1" 与 "0.2.1" 等价)
func eqVersion(a, b string) bool {
	return normalizeVersion(a) == normalizeVersion(b)
}

// compareVersions 按 semver 次序比较 a 与 b: a<b 返回 -1, a>b 返回 1, 相等返回 0。
// 接受 vX.Y.Z 与可选 -prerelease 后缀 (v 前缀可选); 任一方不可解析 (如 devel 形态) 时 ok=false,
// 调用方按「无法判定新旧」处理。
// 预发布段按 semver 规则比较: 正式版 > 预发布, 段内数字按数值、标识符按字典序。
func compareVersions(a, b string) (cmp int, ok bool) {
	va, oka := parseSemver(a)
	vb, okb := parseSemver(b)
	if !oka || !okb {
		return 0, false
	}
	if va.major != vb.major {
		return sign(va.major - vb.major), true
	}
	if va.minor != vb.minor {
		return sign(va.minor - vb.minor), true
	}
	if va.patch != vb.patch {
		return sign(va.patch - vb.patch), true
	}
	switch {
	case va.prerelease == vb.prerelease:
		return 0, true
	case va.prerelease == "":
		return 1, true
	case vb.prerelease == "":
		return -1, true
	default:
		return comparePrerelease(va.prerelease, vb.prerelease), true
	}
}

// semver compareVersions 所需的解析结果
type semver struct {
	major, minor, patch int
	prerelease          string // 不含前导 -, 空串表示正式版
}

// parseSemver 解析 vX.Y.Z(-prerelease) 形态版本号, 容忍 build metadata (+...) 后缀;
// 伪版本 (vX.0.0-<ts>-<hash>) 与 tag 后接的预发布一并落入 prerelease 处理
func parseSemver(s string) (semver, bool) {
	s = normalizeVersion(s)
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	var v semver
	if i := strings.IndexByte(s, '-'); i >= 0 {
		v.prerelease = s[i+1:]
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 || !isNumeric(parts[0]) || !isNumeric(parts[1]) || !isNumeric(parts[2]) {
		return semver{}, false
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return semver{}, false
		}
		nums[i] = n
	}
	v.major, v.minor, v.patch = nums[0], nums[1], nums[2]
	return v, true
}

// comparePrerelease 按 semver 规则比较预发布段: 以 . 分段, 段数少者小;
// 同位置均为纯数字按数值比较, 数字段小于字母段, 其余按字典序
func comparePrerelease(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		xNum, yNum := isNumeric(as[i]), isNumeric(bs[i])
		switch {
		case xNum && yNum:
			x, _ := strconv.Atoi(as[i])
			y, _ := strconv.Atoi(bs[i])
			if x != y {
				return sign(x - y)
			}
		case xNum != yNum:
			if xNum {
				return -1
			}
			return 1
		default:
			if c := strings.Compare(as[i], bs[i]); c != 0 {
				return c
			}
		}
	}
	return sign(len(as) - len(bs))
}

// sign 返回 n 的符号
func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
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
