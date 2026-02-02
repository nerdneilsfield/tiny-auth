package policy

import (
	"sort"
	"strings"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// MatchPolicy 匹配路由策略
// 策略按优先级排序（priority 越大越优先），优先级相同时按配置顺序
// 返回第一个匹配的策略，如果没有匹配则返回 nil
func MatchPolicy(policies []config.RoutePolicy, host, uri, method string) *config.RoutePolicy {
	if len(policies) == 0 {
		return nil
	}

	// 创建策略副本并按优先级排序
	sortedPolicies := make([]config.RoutePolicy, len(policies))
	copy(sortedPolicies, policies)

	sort.SliceStable(sortedPolicies, func(i, j int) bool {
		// 按 priority 降序排序（数字越大越优先）
		// 优先级相同时保持原有顺序（StableSort）
		return sortedPolicies[i].Priority > sortedPolicies[j].Priority
	})

	// 遍历排序后的策略
	for i := range sortedPolicies {
		p := &sortedPolicies[i]

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
