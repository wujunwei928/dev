package updater

import "testing"

// semver 大小比较: 判断「已是最新」时不重复下载 (normalizeVersion 只做等值归一)
func TestCompareVersions(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
		ok   bool
	}{
		{"相等", "v0.2.1", "v0.2.1", 0, true},
		{"v 前缀归一后相等", "0.2.1", "v0.2.1", 0, true},
		{"小于", "v0.2.0", "v0.2.1", -1, true},
		{"大于", "v0.2.1", "v0.2.0", 1, true},
		{"数值比较而非字典序", "v0.10.0", "v0.9.0", 1, true},
		{"正式版大于预发布", "v1.0.0", "v1.0.0-rc.1", 1, true},
		{"伪版本 (Go 1.24+ 本地构建推导) 小于正式版", "v0.0.0-20250101120000-abcdef123456", "v0.2.1", -1, true},
		{"预发布段数值比较", "v1.0.0-rc.10", "v1.0.0-rc.9", 1, true},
		{"预发布段数多者大", "v1.0.0-rc.1", "v1.0.0-rc", 1, true},
		{"devel 版本不可比较", "devel+0123456", "v0.2.1", 0, false},
		{"纯 devel 不可比较", "devel", "v0.2.1", 0, false},
		{"空串不可比较", "", "v0.2.1", 0, false},
		{"两段式不可解析", "v0.2", "v0.2.0", 0, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := compareVersions(c.a, c.b)
			if ok != c.ok {
				t.Fatalf("compareVersions(%q, %q) ok = %v, want %v", c.a, c.b, ok, c.ok)
			}
			if got != c.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
			}
		})
	}
}
