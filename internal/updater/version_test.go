package updater

import "testing"

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name    string
		ldflags string
		bi      BuildInfo
		want    VersionInfo
	}{
		{
			name:    "ldflags 注入时直接采用, 优先于 buildinfo",
			ldflags: "0.2.0",
			bi:      BuildInfo{MainVersion: "v0.2.0", VCSRevision: "0123456789abcdef"},
			want:    VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		},
		{
			name:    "仅 ldflags 注入, 无 buildinfo",
			ldflags: "1.2.3",
			want:    VersionInfo{Method: MethodRelease, Version: "1.2.3"},
		},
		{
			name: "go install 构建取主模块版本",
			bi:   BuildInfo{MainVersion: "v0.2.0", VCSRevision: "0123456789abcdef"},
			want: VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		},
		{
			name: "带 prerelease 的主模块版本同样视作 go install 构建",
			bi:   BuildInfo{MainVersion: "v1.0.0-rc.1"},
			want: VersionInfo{Method: MethodGoInstall, Version: "v1.0.0-rc.1"},
		},
		{
			name:    "主模块版本为 (devel) 时取 commit 前 7 位",
			ldflags: "",
			bi:      BuildInfo{MainVersion: "(devel)", VCSRevision: "0123456789abcdef"},
			want:    VersionInfo{Method: MethodLocal, Version: "devel+0123456"},
		},
		{
			name: "主模块版本为空串时同样视作本地构建",
			bi:   BuildInfo{VCSRevision: "abc1234"},
			want: VersionInfo{Method: MethodLocal, Version: "devel+abc1234"},
		},
		{
			name: "非 VCS 构建 (如源码压缩包) 退化为 devel",
			bi:   BuildInfo{MainVersion: "(devel)"},
			want: VersionInfo{Method: MethodLocal, Version: "devel"},
		},
		{
			name: "VCS 伪版本主模块版本视作本地构建, 不误判为 go install",
			bi:   BuildInfo{MainVersion: "v0.0.0-20261108091214-4e79244c5459", VCSRevision: "4e79244c5459abcdef0123456789abcdef"},
			want: VersionInfo{Method: MethodLocal, Version: "devel+4e79244"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveVersion(tt.ldflags, tt.bi)
			if got != tt.want {
				t.Errorf("ResolveVersion(%q, %+v) = %+v, want %+v", tt.ldflags, tt.bi, got, tt.want)
			}
		})
	}
}

func TestIsSemVer(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"v0.2.0", true},
		{"v1.0.0-rc.1", true},
		{"v0.0.1", true},
		{"", false},        // 本地构建可能为空
		{"(devel)", false}, // 本地构建占位
		{"latest", false},
		{"v1.2", false},     // 段数不足
		{"v1.2.3.4", false}, // 段数超出
		{"v1.x.3", false},   // 非数字段
		// parseSemver 带来的宽容度: 无 v 前缀 / build metadata 也接受。
		// go 模块版本恒带 v 前缀且不含 build metadata, 不影响 ResolveVersion 的安装方式判定
		{"0.2.0", true},
		{"v0.2.0+build.1", true},
	}
	for _, tt := range tests {
		if got := isSemVer(tt.in); got != tt.want {
			t.Errorf("isSemVer(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestIsPseudoVersion(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		// go 命令生成的三种伪版本形态
		{"v0.0.0-20261108091214-4e79244c5459", true},        // 无任何 tag
		{"v1.0.0-rc.1.0.20261108091214-4e79244c5459", true}, // 早于 vX.Y.Z-pre
		{"v0.2.0-2.0.20261108091214-4e79244c5459", true},    // 晚于 vX.Y.Z 两个 commit
		{"v0.2.0", false},                             // go install 安装的发布版本
		{"v1.0.0-rc.1", false},                        // go install 安装的 prerelease
		{"v0.0.0-2026110809121-4e79244c5459", false},  // 时间戳不足 14 位
		{"v0.0.0-20261108091214-4e79244c545", false},  // commit hash 不足 12 位
		{"v0.0.0-20261108091214-4e79244c545g", false}, // commit hash 含非 hex 字符
	}
	for _, tt := range tests {
		if got := isPseudoVersion(tt.in); got != tt.want {
			t.Errorf("isPseudoVersion(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
