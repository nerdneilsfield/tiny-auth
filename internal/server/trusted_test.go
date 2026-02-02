package server

import (
	"net"
	"testing"
)

func TestParseTrustedProxies(t *testing.T) {
	tests := []struct {
		name      string
		proxies   []string
		wantCount int
		testIPs   map[string]bool // IP -> should be trusted
	}{
		{
			name:      "单个 IPv4 地址",
			proxies:   []string{"192.168.1.1"},
			wantCount: 1,
			testIPs: map[string]bool{
				"192.168.1.1": true,
				"192.168.1.2": false,
			},
		},
		{
			name:      "IPv4 CIDR",
			proxies:   []string{"192.168.1.0/24"},
			wantCount: 1,
			testIPs: map[string]bool{
				"192.168.1.1":   true,
				"192.168.1.100": true,
				"192.168.1.255": true,
				"192.168.2.1":   false,
			},
		},
		{
			name:      "多个 CIDR",
			proxies:   []string{"10.0.0.0/8", "172.16.0.0/12"},
			wantCount: 2,
			testIPs: map[string]bool{
				"10.1.2.3":    true,
				"172.16.1.1":  true,
				"172.31.1.1":  true,
				"192.168.1.1": false,
			},
		},
		{
			name:      "IPv6 地址",
			proxies:   []string{"::1"},
			wantCount: 1,
			testIPs: map[string]bool{
				"::1":    true,
				"::2":    false,
				"fe80::": false,
			},
		},
		{
			name:      "混合 IPv4 和 IPv6",
			proxies:   []string{"192.168.1.1", "::1", "10.0.0.0/8"},
			wantCount: 3,
			testIPs: map[string]bool{
				"192.168.1.1": true,
				"::1":         true,
				"10.20.30.40": true,
				"8.8.8.8":     false,
			},
		},
		{
			name:      "空列表",
			proxies:   []string{},
			wantCount: 0,
			testIPs: map[string]bool{
				"192.168.1.1": true, // 空列表应该信任所有
			},
		},
		{
			name:      "无效 CIDR（应该被忽略）",
			proxies:   []string{"192.168.1.1", "invalid", "10.0.0.0/8"},
			wantCount: 2, // 只有 2 个有效
			testIPs: map[string]bool{
				"192.168.1.1": true,
				"10.1.1.1":    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cidrs := parseTrustedProxies(tt.proxies)

			if len(cidrs) != tt.wantCount {
				t.Errorf("parseTrustedProxies() count = %d, want %d", len(cidrs), tt.wantCount)
			}

			// 测试每个 IP 是否正确判断
			for ip, shouldTrust := range tt.testIPs {
				result := isTrustedProxy(ip, cidrs)
				if result != shouldTrust {
					t.Errorf("isTrustedProxy(%q) = %v, want %v", ip, result, shouldTrust)
				}
			}
		})
	}
}

func TestIsTrustedProxy(t *testing.T) {
	// 准备测试用的 CIDR
	_, cidr1, _ := net.ParseCIDR("192.168.0.0/16")
	_, cidr2, _ := net.ParseCIDR("10.0.0.0/8")
	trustedCIDRs := []*net.IPNet{cidr1, cidr2}

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"在第一个 CIDR 内", "192.168.1.1", true},
		{"在第一个 CIDR 内（边界）", "192.168.255.255", true},
		{"在第二个 CIDR 内", "10.1.2.3", true},
		{"不在任何 CIDR 内", "8.8.8.8", false},
		{"不在任何 CIDR 内（邻近）", "192.169.1.1", false},
		{"无效 IP", "invalid-ip", false},
		{"空 IP", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTrustedProxy(tt.ip, trustedCIDRs)
			if result != tt.expected {
				t.Errorf("isTrustedProxy(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestIsTrustedProxy_EmptyList(t *testing.T) {
	// 空列表应该信任所有（向后兼容）
	emptyCIDRs := []*net.IPNet{}

	ips := []string{
		"192.168.1.1",
		"10.0.0.1",
		"8.8.8.8",
		"::1",
	}

	for _, ip := range ips {
		if !isTrustedProxy(ip, emptyCIDRs) {
			t.Errorf("empty trusted list should trust all IPs, but %q was not trusted", ip)
		}
	}
}

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"hostname", "example.com", "example.com"},
		{"hostname with port", "example.com:443", "example.com"},
		{"hostname uppercase", "EXAMPLE.COM", "example.com"},
		{"multiple values", "example.com, proxy.local", "example.com"},
		{"ipv6 bracketed", "[::1]:443", "::1"},
		{"ipv6 raw", "::1", "::1"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeHost(tt.host); got != tt.want {
				t.Errorf("normalizeHost(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

// 性能基准测试
func BenchmarkIsTrustedProxy_Match(b *testing.B) {
	_, cidr, _ := net.ParseCIDR("192.168.0.0/16")
	trustedCIDRs := []*net.IPNet{cidr}
	ip := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTrustedProxy(ip, trustedCIDRs)
	}
}

func BenchmarkIsTrustedProxy_NoMatch(b *testing.B) {
	_, cidr, _ := net.ParseCIDR("192.168.0.0/16")
	trustedCIDRs := []*net.IPNet{cidr}
	ip := "8.8.8.8"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTrustedProxy(ip, trustedCIDRs)
	}
}

func BenchmarkParseTrustedProxies(b *testing.B) {
	proxies := []string{
		"192.168.0.0/16",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"::1",
		"fe80::/10",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseTrustedProxies(proxies)
	}
}
