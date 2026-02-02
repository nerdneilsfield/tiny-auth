package policy

import (
	"strings"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// MatchPolicy 匹配路由策略
// 返回第一个匹配的策略，如果没有匹配则返回 nil
func MatchPolicy(policies []config.RoutePolicy, host, uri, method string) *config.RoutePolicy {
	for i := range policies {
		p := &policies[i]

		// 匹配 host
		if !matchHost(p.Host, host) {
			continue
		}

		// 匹配 path prefix
		if p.PathPrefix != "" && !strings.HasPrefix(uri, p.PathPrefix) {
			continue
		}

		// 匹配 method
		if p.Method != "" && !strings.EqualFold(p.Method, method) {
			continue
		}

		// 所有条件都匹配，返回这个策略
		return p
	}

	// 没有策略匹配
	return nil
}

// matchHost 匹配 host 模式
// 支持精确匹配和通配符（*.example.com）
func matchHost(pattern, host string) bool {
	if pattern == "" {
		return true // 空模式匹配所有 host
	}

	// 通配符匹配：*.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // .example.com
		return strings.HasSuffix(host, suffix)
	}

	// 精确匹配（不区分大小写）
	return strings.EqualFold(pattern, host)
}
