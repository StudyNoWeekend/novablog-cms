package githubartifact

import (
	"errors"
	"testing"
)

// TestParseRepoURL 覆盖仓库地址解析的常见形态与非法输入。
func TestParseRepoURL(t *testing.T) {
	cases := []struct {
		raw    string
		owner  string
		repo   string
		hasErr bool
	}{
		{raw: "https://github.com/owner/repo", owner: "owner", repo: "repo"},
		{raw: "https://github.com/owner/repo/", owner: "owner", repo: "repo"},
		{raw: "https://github.com/owner/repo.git", owner: "owner", repo: "repo"},
		{raw: "https://github.com/owner/repo/tree/main", owner: "owner", repo: "repo"},
		{raw: "https://github.com/owner/repo/tree/HEAD/sub/dir", owner: "owner", repo: "repo"},
		{raw: "https://github.com/owner/repo#readme", owner: "owner", repo: "repo"},
		{raw: "https://github.com/owner/repo?tab=readme", owner: "owner", repo: "repo"},
		{raw: "https://github.com/vercel/next.js", owner: "vercel", repo: "next.js"},
		{raw: "  https://github.com/owner/repo  ", owner: "owner", repo: "repo"},
		{raw: "http://github.com/owner/repo", hasErr: true},
		{raw: "https://gitlab.com/owner/repo", hasErr: true},
		{raw: "https://github.com/owner", hasErr: true},
		{raw: "https://github.com/", hasErr: true},
		{raw: "not a url", hasErr: true},
		{raw: "", hasErr: true},
	}

	for _, c := range cases {
		owner, repo, err := ParseRepoURL(c.raw)
		if c.hasErr {
			if err == nil {
				t.Errorf("ParseRepoURL(%q) 期望报错，实际 owner=%q repo=%q", c.raw, owner, repo)
			}
			if !errors.Is(err, ErrInvalidOpenRepoURL) && err != nil {
				t.Errorf("ParseRepoURL(%q) 错误应可 errors.Is 到 ErrInvalidOpenRepoURL，实际 %v", c.raw, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRepoURL(%q) 意外报错: %v", c.raw, err)
			continue
		}
		if owner != c.owner || repo != c.repo {
			t.Errorf("ParseRepoURL(%q) = (%q, %q)，期望 (%q, %q)", c.raw, owner, repo, c.owner, c.repo)
		}
	}
}
