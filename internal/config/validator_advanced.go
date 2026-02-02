package config

import (
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"strings"
)

// validatePolicyDependencies 验证策略依赖（检测循环依赖）
// 确保策略引用的认证方式名称都存在于配置中
func validatePolicyDependencies(cfg *Config) error {
	// 构建所有可用的认证方式名称集合
	availableNames := make(map[string]bool)

	for _, auth := range cfg.BasicAuths {
		availableNames[auth.Name] = true
	}
	for _, auth := range cfg.BearerTokens {
		availableNames[auth.Name] = true
	}
	for _, auth := range cfg.APIKeys {
		availableNames[auth.Name] = true
	}

	// 检查每个策略引用的名称是否存在
	for _, policy := range cfg.RoutePolicies {
		// 检查 allowed_basic_names
		for _, name := range policy.AllowedBasicNames {
			if !availableNames[name] {
				return fmt.Errorf("policy [%s] references non-existent basic_auth name: %q", policy.Name, name)
			}
		}

		// 检查 allowed_bearer_names
		for _, name := range policy.AllowedBearerNames {
			if !availableNames[name] {
				return fmt.Errorf("policy [%s] references non-existent bearer_token name: %q", policy.Name, name)
			}
		}

		// 检查 allowed_apikey_names
		for _, name := range policy.AllowedAPIKeyNames {
			if !availableNames[name] {
				return fmt.Errorf("policy [%s] references non-existent api_key name: %q", policy.Name, name)
			}
		}
	}

	return nil
}

// validatePolicyConflicts 验证策略冲突
// 检查是否有多个策略匹配相同的路由（host + path + method）
func validatePolicyConflicts(policies []RoutePolicy) error {
	if len(policies) == 0 {
		return nil
	}

	// 记录所有策略的匹配规则
	type policyKey struct {
		host   string
		path   string
		method string
	}

	conflicts := make(map[policyKey][]string)

	for _, policy := range policies {
		// 为每个匹配模式创建键
		hosts := []string{policy.Host}
		if policy.Host == "" {
			hosts = []string{"*"} // 空表示匹配所有
		}

		paths := []string{policy.PathPrefix}
		if policy.PathPrefix == "" {
			paths = []string{"/"} // 空表示根路径
		}

		methods := []string{policy.Method}
		if policy.Method == "" {
			methods = []string{"*"} // 空表示所有方法
		}

		// 检查每个组合
		for _, host := range hosts {
			for _, path := range paths {
				for _, method := range methods {
					key := policyKey{
						host:   host,
						path:   path,
						method: method,
					}
					conflicts[key] = append(conflicts[key], policy.Name)
				}
			}
		}
	}

	// 检查冲突
	hasConflict := false
	for key, policyNames := range conflicts {
		if len(policyNames) > 1 {
			// 发现冲突，但不是致命错误，只发出警告
			hasConflict = true
			fmt.Fprintf(os.Stderr, "⚠ Warning: Multiple policies match [host=%s, path=%s, method=%s]: %v\n",
				key.host, key.path, key.method, policyNames)
			fmt.Fprintf(os.Stderr, "  → First matching policy will be used (order matters)\n")
		}
	}

	// 如果有冲突，提示用户
	if hasConflict {
		fmt.Fprintf(os.Stderr, "💡 Tip: Use the 'priority' field to control policy match order\n")
	}

	return nil // 冲突不是错误，只是警告
}

// validateJWTSecretStrength 验证 JWT Secret 强度
// 检查密钥长度和熵值
func validateJWTSecretStrength(jwt *JWTConfig) error {
	if jwt.Secret == "" {
		return nil // JWT 是可选的
	}

	// 跳过环境变量（无法在配置阶段验证）
	if strings.HasPrefix(jwt.Secret, "env:") {
		fmt.Fprintf(os.Stderr, "ℹ JWT secret from environment variable - strength check skipped\n")
		return nil
	}

	secret := jwt.Secret

	// 1. 长度检查（已在 validateJWT 中完成，这里是双重保险）
	if len(secret) < 32 {
		return fmt.Errorf("jwt secret too short (< 32 chars): use a longer secret for HS256")
	}

	// 2. 熵值检查
	entropy := calculateEntropy(secret)
	minEntropy := 3.0 // 最小熵值（bits per character）

	if entropy < minEntropy {
		fmt.Fprintf(os.Stderr, "⚠ Warning: JWT secret has low entropy (%.2f bits/char, recommended: >%.1f)\n", entropy, minEntropy)
		fmt.Fprintf(os.Stderr, "  → Current secret may be predictable (e.g., repeated characters, simple patterns)\n")
		fmt.Fprintf(os.Stderr, "  → Recommendation: Use a cryptographically random string\n")
		fmt.Fprintf(os.Stderr, "     Example: openssl rand -base64 32\n")
	}

	// 3. 检查常见弱密钥模式
	weakPatterns := []string{
		"your-secret-key",
		"secret",
		"password",
		"12345",
		"qwerty",
		"admin",
		"test",
		"demo",
		"example",
	}

	lowerSecret := strings.ToLower(secret)
	for _, pattern := range weakPatterns {
		if strings.Contains(lowerSecret, pattern) {
			fmt.Fprintf(os.Stderr, "⚠ Warning: JWT secret contains common weak pattern: %q\n", pattern)
			fmt.Fprintf(os.Stderr, "  → Use a cryptographically random secret instead\n")
			break
		}
	}

	return nil
}

// calculateEntropy 计算字符串的香农熵（Shannon entropy）
// 返回每个字符的平均信息量（bits per character）
func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0.0
	}

	// 统计字符频率
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	// 计算熵值
	var entropy float64
	length := float64(len(s))

	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// calculateSecretComplexity 计算密钥复杂度评分（0-100）
// 综合考虑长度、字符种类、熵值等因素
func calculateSecretComplexity(secret string) int {
	if len(secret) == 0 {
		return 0
	}

	score := 0

	// 1. 长度评分（最高 40 分）
	length := len(secret)
	if length >= 32 {
		score += 40
	} else if length >= 24 {
		score += 30
	} else if length >= 16 {
		score += 20
	} else {
		score += length
	}

	// 2. 字符种类评分（最高 30 分）
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range secret {
		if c >= 'a' && c <= 'z' {
			hasLower = true
		} else if c >= 'A' && c <= 'Z' {
			hasUpper = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	if hasLower {
		score += 7
	}
	if hasUpper {
		score += 7
	}
	if hasDigit {
		score += 8
	}
	if hasSpecial {
		score += 8
	}

	// 3. 熵值评分（最高 30 分）
	entropy := calculateEntropy(secret)
	entropyScore := int(entropy * 6) // 5 bits/char = 30 分
	if entropyScore > 30 {
		entropyScore = 30
	}
	score += entropyScore

	if score > 100 {
		score = 100
	}

	return score
}

// hashSecret 对密钥进行哈希（用于安全日志记录）
// 只返回哈希的前 8 个字符
func hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", hash[:4]) // 前 4 字节 = 8 个十六进制字符
}
