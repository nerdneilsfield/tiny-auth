package ratelimit

import (
	"testing"
	"time"
)

// TestNewLimiter 测试创建限制器
func TestNewLimiter(t *testing.T) {
	limiter := NewLimiter(5, time.Minute, time.Minute*5)
	defer limiter.Stop()

	if limiter == nil {
		t.Fatal("Expected non-nil limiter")
	}

	if limiter.maxAttempts != 5 {
		t.Errorf("Expected maxAttempts=5, got %d", limiter.maxAttempts)
	}

	if limiter.window != time.Minute {
		t.Errorf("Expected window=1m, got %v", limiter.window)
	}

	if limiter.banDuration != time.Minute*5 {
		t.Errorf("Expected banDuration=5m, got %v", limiter.banDuration)
	}
}

// TestAllow_BasicFlow 测试基本流程
func TestAllow_BasicFlow(t *testing.T) {
	limiter := NewLimiter(3, time.Minute, time.Minute*5)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// 前 3 次应该允许
	for i := 0; i < 3; i++ {
		allowed, retryAfter := limiter.Allow(ip)
		if !allowed {
			t.Errorf("Attempt %d should be allowed", i+1)
		}
		if retryAfter != 0 {
			t.Errorf("RetryAfter should be 0, got %v", retryAfter)
		}
	}

	// 第 4 次应该被拒绝（触发封禁）
	allowed, retryAfter := limiter.Allow(ip)
	if allowed {
		t.Error("Attempt 4 should be denied")
	}
	if retryAfter == 0 {
		t.Error("RetryAfter should be non-zero")
	}
}

// TestAllow_DifferentIPs 测试不同 IP 独立限制
func TestAllow_DifferentIPs(t *testing.T) {
	limiter := NewLimiter(2, time.Minute, time.Minute)
	defer limiter.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// IP1 尝试 2 次
	limiter.Allow(ip1)
	limiter.Allow(ip1)

	// IP1 第 3 次应该被拒绝
	allowed, _ := limiter.Allow(ip1)
	if allowed {
		t.Error("IP1 third attempt should be denied")
	}

	// IP2 应该仍然可以尝试
	allowed, _ = limiter.Allow(ip2)
	if !allowed {
		t.Error("IP2 first attempt should be allowed")
	}
}

// TestAllow_SlidingWindow 测试滑动窗口
func TestAllow_SlidingWindow(t *testing.T) {
	// 使用很短的时间窗口和封禁时长方便测试
	limiter := NewLimiter(2, time.Millisecond*100, time.Millisecond*100)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// 第 1 次尝试
	allowed, _ := limiter.Allow(ip)
	if !allowed {
		t.Error("First attempt should be allowed")
	}

	// 第 2 次尝试
	allowed, _ = limiter.Allow(ip)
	if !allowed {
		t.Error("Second attempt should be allowed")
	}

	// 第 3 次尝试（应该被拒绝，触发封禁）
	allowed, _ = limiter.Allow(ip)
	if allowed {
		t.Error("Third attempt should be denied")
	}

	// 等待窗口和封禁都过期
	time.Sleep(time.Millisecond * 150)

	// 封禁过期后应该可以再次尝试（记录已清空）
	allowed, _ = limiter.Allow(ip)
	if !allowed {
		t.Error("Attempt after ban expiry should be allowed")
	}
}

// TestAllow_BanDuration 测试封禁时长
func TestAllow_BanDuration(t *testing.T) {
	limiter := NewLimiter(2, time.Minute, time.Millisecond*100)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// 触发封禁
	limiter.Allow(ip)
	limiter.Allow(ip)
	allowed, retryAfter := limiter.Allow(ip)

	if allowed {
		t.Error("Should be denied after exceeding limit")
	}

	if retryAfter <= 0 || retryAfter > time.Millisecond*100 {
		t.Errorf("RetryAfter should be ~100ms, got %v", retryAfter)
	}

	// 封禁期内应该继续拒绝
	allowed, _ = limiter.Allow(ip)
	if allowed {
		t.Error("Should still be banned")
	}

	// 等待封禁过期
	time.Sleep(time.Millisecond * 150)

	// 封禁过期后应该可以尝试
	allowed, _ = limiter.Allow(ip)
	if !allowed {
		t.Error("Should be allowed after ban expires")
	}
}

// TestReset 测试重置
func TestReset(t *testing.T) {
	limiter := NewLimiter(2, time.Minute, time.Minute)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// 尝试 2 次
	limiter.Allow(ip)
	limiter.Allow(ip)

	// 第 3 次应该被拒绝
	allowed, _ := limiter.Allow(ip)
	if allowed {
		t.Error("Third attempt should be denied")
	}

	// 重置后应该可以再次尝试
	limiter.Reset(ip)
	allowed, _ = limiter.Allow(ip)
	if !allowed {
		t.Error("After reset, first attempt should be allowed")
	}
}

// TestGetStats 测试统计信息
func TestGetStats(t *testing.T) {
	limiter := NewLimiter(3, time.Minute, time.Minute*5)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// 初始状态
	attempts, isBanned, retryAfter := limiter.GetStats(ip)
	if attempts != 0 || isBanned || retryAfter != 0 {
		t.Error("Initial stats should be zero")
	}

	// 尝试 2 次
	limiter.Allow(ip)
	limiter.Allow(ip)

	attempts, isBanned, _ = limiter.GetStats(ip)
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
	if isBanned {
		t.Error("Should not be banned yet")
	}

	// 触发封禁
	limiter.Allow(ip)
	limiter.Allow(ip)

	attempts, isBanned, retryAfter = limiter.GetStats(ip)
	if !isBanned {
		t.Error("Should be banned")
	}
	if retryAfter == 0 {
		t.Error("RetryAfter should be non-zero when banned")
	}
}

// TestCleanup 测试清理功能
func TestCleanup(t *testing.T) {
	// 使用短窗口和清理间隔
	limiter := NewLimiter(5, time.Millisecond*50, time.Millisecond*50)
	limiter.cleanupInterval = time.Millisecond * 100 // 快速清理
	defer limiter.Stop()

	// 添加一些记录
	limiter.Allow("192.168.1.1")
	limiter.Allow("192.168.1.2")
	limiter.Allow("192.168.1.3")

	initialCount := limiter.GetTotalRecords()
	if initialCount != 3 {
		t.Errorf("Expected 3 records, got %d", initialCount)
	}

	// 等待记录过期 + 清理运行
	time.Sleep(time.Millisecond * 200)

	// 手动触发清理
	limiter.cleanup()

	finalCount := limiter.GetTotalRecords()
	if finalCount != 0 {
		t.Errorf("Expected 0 records after cleanup, got %d", finalCount)
	}
}

// TestConcurrency 测试并发安全
func TestConcurrency(t *testing.T) {
	limiter := NewLimiter(100, time.Minute, time.Minute)
	defer limiter.Stop()

	done := make(chan bool)
	goroutines := 10
	iterations := 100

	// 多个 goroutine 并发访问
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			ip := "192.168.1.1"
			for j := 0; j < iterations; j++ {
				limiter.Allow(ip)
				limiter.GetStats(ip)
				if j%10 == 0 {
					limiter.Reset(ip)
				}
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// 如果没有 panic 或死锁，测试通过
}

// TestZeroMaxAttempts 测试边界情况：maxAttempts = 0
func TestZeroMaxAttempts(t *testing.T) {
	limiter := NewLimiter(0, time.Minute, time.Minute)
	defer limiter.Stop()

	// maxAttempts = 0 意味着任何请求都会被拒绝
	allowed, _ := limiter.Allow("192.168.1.1")
	if allowed {
		t.Error("With maxAttempts=0, all requests should be denied")
	}
}

// TestMultipleIPsConcurrently 测试多个 IP 并发访问
func TestMultipleIPsConcurrently(t *testing.T) {
	limiter := NewLimiter(5, time.Minute, time.Minute)
	defer limiter.Stop()

	done := make(chan bool)
	numIPs := 10

	for i := 0; i < numIPs; i++ {
		go func(id int) {
			ip := "192.168.1." + string(rune('0'+id))
			for j := 0; j < 3; j++ {
				limiter.Allow(ip)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numIPs; i++ {
		<-done
	}

	totalRecords := limiter.GetTotalRecords()
	if totalRecords != numIPs {
		t.Errorf("Expected %d IP records, got %d", numIPs, totalRecords)
	}
}

// BenchmarkAllow 基准测试：Allow 操作
func BenchmarkAllow(b *testing.B) {
	limiter := NewLimiter(1000, time.Minute, time.Minute)
	defer limiter.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("192.168.1.1")
	}
}

// BenchmarkAllowParallel 基准测试：并发 Allow 操作
func BenchmarkAllowParallel(b *testing.B) {
	limiter := NewLimiter(1000, time.Minute, time.Minute)
	defer limiter.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow("192.168.1.1")
		}
	})
}
