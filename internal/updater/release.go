package updater

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

// releaseAsset Release 已上传产物 (平台归档 / checksums.txt)
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// releaseInfo Releases API 响应中更新流程所需的字段
type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// fetchRelease 获取 Release 信息: tag 为空串查 latest, 否则查指定 tag。
// tag 查询 404 时以 v 前缀互换形态 (历史 Release tag 无 v 前缀) 重试一次
func fetchRelease(apiBase, tag string) (*releaseInfo, error) {
	if apiBase == "" {
		apiBase = defaultAPIBase
	}
	base := strings.TrimRight(apiBase, "/") + "/repos/" + repoSlug() + "/releases/"
	path := "latest"
	if tag != "" {
		path = "tags/" + tag
	}

	rel, resp, err := requestRelease(base + path)
	if err != nil {
		return nil, fmt.Errorf("查询版本 %s 失败: %w", tagLabel(tag), err)
	}
	if resp.StatusCode() == 404 && tag != "" {
		if alt := alternateTag(tag); alt != tag {
			rel, resp, err = requestRelease(base + "tags/" + alt)
			if err != nil {
				return nil, fmt.Errorf("查询版本 %s 失败: %w", tagLabel(tag), err)
			}
		}
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("查询版本 %s 失败: GitHub API 返回 %s", tagLabel(tag), resp.Status())
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("查询版本 %s 失败: 响应缺少 tag_name", tagLabel(tag))
	}
	return rel, nil
}

// requestRelease 请求单个 release API URL, 返回解析结果与原始响应供调用方判断状态码
func requestRelease(url string) (*releaseInfo, *resty.Response, error) {
	rel := &releaseInfo{}
	resp, err := GetHTTPClient().R().SetResult(rel).Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("请检查网络: %w", err)
	}
	return rel, resp, nil
}

// tagLabel 错误信息中使用的版本描述
func tagLabel(tag string) string {
	if tag == "" {
		return "latest"
	}
	return tag
}

// alternateTag 返回 v 前缀互换形态 ("v0.1.3" <-> "0.1.3"); 无需互换时返回原值
func alternateTag(tag string) string {
	n := normalizeVersion(tag)
	if n == tag {
		return "v" + tag
	}
	return n
}

// pickAsset 按平台选择下载产物。
// 兼容两代命名: goreleaser (dev_<ver>_<os>_<arch>.zip|.tar.gz, ver 无 v) 与
// 历史手工发布 (dev-<tag>-<os>-<arch>.tar.gz); 同名候选按列表顺序优先
func pickAsset(rel *releaseInfo, goos, arch string) (*releaseAsset, error) {
	ver := normalizeVersion(rel.TagName)
	candidates := []string{
		"dev_" + ver + "_" + goos + "_" + arch + ".zip", // goreleaser windows 首选
		"dev_" + ver + "_" + goos + "_" + arch + ".tar.gz", // goreleaser 其余平台 / 历史 windows
		"dev-" + rel.TagName + "-" + goos + "-" + arch + ".tar.gz", // 历史 tag 原形
		"dev-" + ver + "-" + goos + "-" + arch + ".tar.gz",
	}
	for _, want := range candidates {
		for i := range rel.Assets {
			if rel.Assets[i].Name == want {
				return &rel.Assets[i], nil
			}
		}
	}
	return nil, fmt.Errorf("版本 %s 未提供适配 %s/%s 的产物", rel.TagName, goos, arch)
}

// findChecksumsAsset 返回 Release 中的 checksums.txt 产物。
// 历史 Release 未上传该文件时 ok=false, 由调用方决定降级行为
func findChecksumsAsset(rel *releaseInfo) (*releaseAsset, bool) {
	for i := range rel.Assets {
		if rel.Assets[i].Name == "checksums.txt" {
			return &rel.Assets[i], true
		}
	}
	return nil, false
}
