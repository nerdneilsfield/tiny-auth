package server

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// parseTrustedProxies 解析可信代理配置为 CIDR 网段
func parseTrustedProxies(proxies []string) []*net.IPNet {
	var cidrs []*net.IPNet

	for _, proxy := range proxies {
		// 如果是单个 IP，添加 /32 (IPv4) 或 /128 (IPv6) 后缀
		if !strings.Contains(proxy, "/") {
			if strings.Contains(proxy, ":") {
				// IPv6
				proxy = proxy + "/128"
			} else {
				// IPv4
				proxy = proxy + "/32"
			}
		}

		_, cidr, err := net.ParseCIDR(proxy)
		if err != nil {
			// 忽略无效配置（已在 validator 中检查）
			continue
		}
		cidrs = append(cidrs, cidr)
	}

	return cidrs
}

// isTrustedProxy 检查 IP 是否在可信代理列表中
func isTrustedProxy(ip string, trustedCIDRs []*net.IPNet) bool {
	// 如果没有配置可信代理，默认信任所有（向后兼容）
	if len(trustedCIDRs) == 0 {
		return true
	}

	// 解析 IP 地址
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	// 检查是否在任何 CIDR 范围内
	for _, cidr := range trustedCIDRs {
		if cidr.Contains(clientIP) {
			return true
		}
	}

	return false
}

// getClientIP 安全地获取客户端 IP
// 只有来自可信代理的请求才信任 X-Forwarded-For
func getClientIP(c *fiber.Ctx, cfg *config.Config, trustedCIDRs []*net.IPNet) string {
	// 获取直接连接的 IP
	directIP := c.IP()

	// 检查直接 IP 是否是可信代理
	if !isTrustedProxy(directIP, trustedCIDRs) {
		// 不是可信代理，直接返回连接 IP（忽略 X-Forwarded-For）
		return directIP
	}

	// 来自可信代理，检查 X-Forwarded-For
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor == "" {
		return directIP
	}

	// X-Forwarded-For 可能包含多个 IP（逗号分隔）
	// 格式：client, proxy1, proxy2
	// 我们取第一个（最左边）作为真实客户端 IP
	ips := strings.Split(forwardedFor, ",")
	if len(ips) > 0 {
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	return directIP
}

// getForwardedHeaders 安全地获取转发的 headers
// 只有来自可信代理的请求才信任这些 headers
func getForwardedHeaders(c *fiber.Ctx, trustedCIDRs []*net.IPNet) (host, uri, method string, trusted bool) {
	directIP := c.IP()

	// 检查是否来自可信代理
	if !isTrustedProxy(directIP, trustedCIDRs) {
		// 不可信，使用直接请求的值
		return c.Hostname(), c.OriginalURL(), c.Method(), false
	}

	// 可信代理，使用转发的 headers
	host = c.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Get("X-Forwarded-Server")
	}
	if host == "" {
		host = c.Hostname()
	}

	uri = c.Get("X-Forwarded-Uri")
	if uri == "" {
		uri = c.OriginalURL()
	}

	method = c.Get("X-Forwarded-Method")
	if method == "" {
		method = c.Method()
	}

	return host, uri, method, true
}
