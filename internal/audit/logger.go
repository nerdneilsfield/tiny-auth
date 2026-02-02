package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// Event 表示一条审计日志事件（JSON Lines）
type Event struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestID    string    `json:"request_id,omitempty"`
	ClientIP     string    `json:"client_ip,omitempty"`
	DirectIP     string    `json:"direct_ip,omitempty"`
	TrustedProxy bool      `json:"trusted_proxy"`
	Host         string    `json:"host,omitempty"`
	URI          string    `json:"uri,omitempty"`
	Method       string    `json:"method,omitempty"`
	AuthMethod   string    `json:"auth_method,omitempty"`
	AuthName     string    `json:"auth_name,omitempty"`
	User         string    `json:"user,omitempty"`
	Roles        []string  `json:"roles,omitempty"`
	Policy       string    `json:"policy,omitempty"`
	Result       string    `json:"result"`
	Reason       string    `json:"reason,omitempty"`
	Status       int       `json:"status"`
	LatencyMs    int64     `json:"latency_ms"`
}

// Logger 负责输出审计日志（独立结构化事件流）
type Logger struct {
	enabled bool
	mu      sync.Mutex
	encoder *json.Encoder
	closer  io.Closer
}

// NewLogger 创建审计日志记录器
func NewLogger(cfg config.AuditConfig) (*Logger, error) {
	if !cfg.Enabled {
		return &Logger{enabled: false}, nil
	}

	output := strings.TrimSpace(cfg.Output)
	if output == "" {
		return nil, fmt.Errorf("audit output cannot be empty when enabled")
	}

	var (
		writer io.Writer
		closer io.Closer
	)

	switch output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
		writer = file
		closer = file
	}

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	return &Logger{
		enabled: true,
		encoder: encoder,
		closer:  closer,
	}, nil
}

// Log 输出一条审计日志
func (l *Logger) Log(event *Event) error {
	if l == nil || !l.enabled || event == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return l.encoder.Encode(event)
}

// Close 关闭底层资源（如果有）
func (l *Logger) Close() error {
	if l == nil || l.closer == nil {
		return nil
	}
	return l.closer.Close()
}
