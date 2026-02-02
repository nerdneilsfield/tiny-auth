package policy

import (
	"testing"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// TestMatchPolicy_WithPriority 测试带优先级的策略匹配
func TestMatchPolicy_WithPriority(t *testing.T) {
	tests := []struct {
		name           string
		policies       []config.RoutePolicy
		host           string
		uri            string
		method         string
		expectedPolicy string // 期望匹配的策略名称
	}{
		{
			name: "Higher priority wins",
			policies: []config.RoutePolicy{
				{Name: "low-priority", Priority: 1, Host: "api.example.com"},
				{Name: "high-priority", Priority: 10, Host: "api.example.com"},
				{Name: "medium-priority", Priority: 5, Host: "api.example.com"},
			},
			host:           "api.example.com",
			uri:            "/api",
			method:         "GET",
			expectedPolicy: "high-priority",
		},
		{
			name: "Same priority - first in config wins (stable sort)",
			policies: []config.RoutePolicy{
				{Name: "first", Priority: 5, Host: "api.example.com"},
				{Name: "second", Priority: 5, Host: "api.example.com"},
				{Name: "third", Priority: 5, Host: "api.example.com"},
			},
			host:           "api.example.com",
			uri:            "/api",
			method:         "GET",
			expectedPolicy: "first",
		},
		{
			name: "Zero priority (default) vs explicit priority",
			policies: []config.RoutePolicy{
				{Name: "default-priority", Priority: 0, Host: "api.example.com"},
				{Name: "explicit-priority", Priority: 1, Host: "api.example.com"},
			},
			host:           "api.example.com",
			uri:            "/api",
			method:         "GET",
			expectedPolicy: "explicit-priority",
		},
		{
			name: "Negative priority still works",
			policies: []config.RoutePolicy{
				{Name: "negative", Priority: -10, Host: "api.example.com"},
				{Name: "zero", Priority: 0, Host: "api.example.com"},
				{Name: "positive", Priority: 10, Host: "api.example.com"},
			},
			host:           "api.example.com",
			uri:            "/api",
			method:         "GET",
			expectedPolicy: "positive",
		},
		{
			name: "Priority with path specificity",
			policies: []config.RoutePolicy{
				{Name: "generic-high", Priority: 100, PathPrefix: "/"},
				{Name: "specific-low", Priority: 1, PathPrefix: "/api/v2"},
				{Name: "specific-medium", Priority: 50, PathPrefix: "/api"},
			},
			host:           "api.example.com",
			uri:            "/api/v2/users",
			method:         "GET",
			expectedPolicy: "generic-high", // 100 优先级最高，虽然路径不够具体
		},
		{
			name: "Only matching policy has lower priority",
			policies: []config.RoutePolicy{
				{Name: "wrong-host", Priority: 100, Host: "wrong.example.com"},
				{Name: "correct-host", Priority: 1, Host: "api.example.com"},
				{Name: "wrong-path", Priority: 50, PathPrefix: "/wrong"},
			},
			host:           "api.example.com",
			uri:            "/api",
			method:         "GET",
			expectedPolicy: "correct-host",
		},
		{
			name: "All priorities equal, order matters",
			policies: []config.RoutePolicy{
				{Name: "first", Priority: 0, Host: "api.example.com"},
				{Name: "second", Priority: 0, Host: "api.example.com"},
				{Name: "third", Priority: 0, Host: "api.example.com"},
			},
			host:           "api.example.com",
			uri:            "/api",
			method:         "GET",
			expectedPolicy: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := MatchPolicy(tt.policies, tt.host, tt.uri, tt.method)

			if tt.expectedPolicy == "" {
				if policy != nil {
					t.Errorf("Expected no policy match, got %q", policy.Name)
				}
				return
			}

			if policy == nil {
				t.Fatalf("Expected policy %q, got nil", tt.expectedPolicy)
			}

			if policy.Name != tt.expectedPolicy {
				t.Errorf("Expected policy %q, got %q", tt.expectedPolicy, policy.Name)
			}
		})
	}
}

// TestMatchPolicy_PriorityPreservesOriginal 测试优先级排序不修改原始切片
func TestMatchPolicy_PriorityPreservesOriginal(t *testing.T) {
	originalPolicies := []config.RoutePolicy{
		{Name: "first", Priority: 1},
		{Name: "second", Priority: 2},
		{Name: "third", Priority: 3},
	}

	// 复制一份用于比较
	expected := make([]config.RoutePolicy, len(originalPolicies))
	copy(expected, originalPolicies)

	// 调用 MatchPolicy
	MatchPolicy(originalPolicies, "api.example.com", "/api", "GET")

	// 验证原始切片未被修改
	if len(originalPolicies) != len(expected) {
		t.Fatal("Original policies slice length changed")
	}

	for i := range originalPolicies {
		if originalPolicies[i].Name != expected[i].Name {
			t.Errorf("Original policies order changed at index %d: expected %q, got %q",
				i, expected[i].Name, originalPolicies[i].Name)
		}
		if originalPolicies[i].Priority != expected[i].Priority {
			t.Errorf("Original policies priority changed at index %d: expected %d, got %d",
				i, expected[i].Priority, originalPolicies[i].Priority)
		}
	}
}

// TestMatchPolicy_EmptyPolicies 测试空策略列表
func TestMatchPolicy_EmptyPolicies(t *testing.T) {
	policy := MatchPolicy([]config.RoutePolicy{}, "api.example.com", "/api", "GET")
	if policy != nil {
		t.Errorf("Expected nil for empty policies, got %+v", policy)
	}
}

// TestMatchPolicy_LargePriorities 测试大优先级数值
func TestMatchPolicy_LargePriorities(t *testing.T) {
	policies := []config.RoutePolicy{
		{Name: "low", Priority: -999999},
		{Name: "medium", Priority: 0},
		{Name: "high", Priority: 999999},
	}

	policy := MatchPolicy(policies, "example.com", "/", "GET")
	if policy == nil || policy.Name != "high" {
		t.Errorf("Expected 'high' priority policy, got %v", policy)
	}
}

// BenchmarkMatchPolicy_WithPriority 基准测试：带优先级的策略匹配
func BenchmarkMatchPolicy_WithPriority(b *testing.B) {
	policies := []config.RoutePolicy{
		{Name: "p1", Priority: 10, Host: "api.example.com", PathPrefix: "/v1"},
		{Name: "p2", Priority: 20, Host: "api.example.com", PathPrefix: "/v2"},
		{Name: "p3", Priority: 5, Host: "api.example.com", PathPrefix: "/v3"},
		{Name: "p4", Priority: 15, Host: "api.example.com", PathPrefix: "/v4"},
		{Name: "p5", Priority: 1, Host: "api.example.com", PathPrefix: "/v5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchPolicy(policies, "api.example.com", "/v3/users", "GET")
	}
}

// BenchmarkMatchPolicy_ManyPolicies 基准测试：大量策略
func BenchmarkMatchPolicy_ManyPolicies(b *testing.B) {
	policies := make([]config.RoutePolicy, 100)
	for i := 0; i < 100; i++ {
		policies[i] = config.RoutePolicy{
			Name:     string(rune('a' + i)),
			Priority: i,
			Host:     "api.example.com",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchPolicy(policies, "api.example.com", "/api", "GET")
	}
}
