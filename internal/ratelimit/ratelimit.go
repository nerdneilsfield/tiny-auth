package ratelimit

import (
	"sync"
	"time"
)

// Limiter 速率限制器（基于滑动窗口算法）
type Limiter struct {
	// 每个 IP 的请求记录
	records map[string]*record
	mu      sync.RWMutex

	// 配置
	maxAttempts int           // 时间窗口内的最大尝试次数
	window      time.Duration // 时间窗口
	banDuration time.Duration // 封禁时长

	// 清理器
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// record 记录单个 IP 的请求历史
type record struct {
	attempts  []time.Time // 请求时间戳列表
	bannedUntil time.Time // 封禁截止时间
}

// NewLimiter 创建新的速率限制器
func NewLimiter(maxAttempts int, window, banDuration time.Duration) *Limiter {
	l := &Limiter{
		records:         make(map[string]*record),
		maxAttempts:     maxAttempts,
		window:          window,
		banDuration:     banDuration,
		cleanupInterval: time.Minute * 5, // 每 5 分钟清理一次过期记录
		stopCleanup:     make(chan struct{}),
	}

	// 启动后台清理任务
	go l.startCleanup()

	return l
}

// Allow 检查 IP 是否允许继续尝试
// 返回 (allowed bool, retryAfter time.Duration)
func (l *Limiter) Allow(ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// 获取或创建记录
	rec, exists := l.records[ip]
	if !exists {
		rec = &record{
			attempts: make([]time.Time, 0),
		}
		l.records[ip] = rec
	}

	// 检查是否在封禁期内
	if now.Before(rec.bannedUntil) {
		return false, time.Until(rec.bannedUntil)
	}

	// 封禁已过期，清空尝试记录重新开始
	if !rec.bannedUntil.IsZero() && now.After(rec.bannedUntil) {
		rec.attempts = make([]time.Time, 0)
		rec.bannedUntil = time.Time{} // 重置封禁时间
	}

	// 移除时间窗口外的旧记录
	cutoff := now.Add(-l.window)
	validAttempts := make([]time.Time, 0, len(rec.attempts))
	for _, t := range rec.attempts {
		if t.After(cutoff) {
			validAttempts = append(validAttempts, t)
		}
	}
	rec.attempts = validAttempts

	// 检查是否超过限制
	if len(rec.attempts) >= l.maxAttempts {
		// 触发封禁
		rec.bannedUntil = now.Add(l.banDuration)
		return false, l.banDuration
	}

	// 记录本次尝试
	rec.attempts = append(rec.attempts, now)

	return true, 0
}

// Reset 重置指定 IP 的限制（用于成功认证后）
func (l *Limiter) Reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.records, ip)
}

// GetStats 获取指定 IP 的统计信息
func (l *Limiter) GetStats(ip string) (attempts int, isBanned bool, retryAfter time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	rec, exists := l.records[ip]
	if !exists {
		return 0, false, 0
	}

	now := time.Now()

	// 检查封禁状态
	if now.Before(rec.bannedUntil) {
		return len(rec.attempts), true, time.Until(rec.bannedUntil)
	}

	// 统计有效尝试次数
	cutoff := now.Add(-l.window)
	count := 0
	for _, t := range rec.attempts {
		if t.After(cutoff) {
			count++
		}
	}

	return count, false, 0
}

// startCleanup 启动后台清理任务
func (l *Limiter) startCleanup() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stopCleanup:
			return
		}
	}
}

// cleanup 清理过期的记录
func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	for ip, rec := range l.records {
		// 如果封禁已过期且没有有效尝试记录，删除该 IP
		if now.After(rec.bannedUntil) {
			hasValidAttempts := false
			for _, t := range rec.attempts {
				if t.After(cutoff) {
					hasValidAttempts = true
					break
				}
			}
			if !hasValidAttempts {
				delete(l.records, ip)
			}
		}
	}
}

// Stop 停止速率限制器（清理后台任务）
func (l *Limiter) Stop() {
	close(l.stopCleanup)
}

// GetTotalRecords 获取当前记录总数（用于监控）
func (l *Limiter) GetTotalRecords() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.records)
}
