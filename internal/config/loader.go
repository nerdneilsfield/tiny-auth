package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	apperrors "github.com/nerdneilsfield/tiny-auth/internal/errors"
)

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 如果未指定路径，使用环境变量或默认值
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
		if path == "" {
			path = "config.toml"
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, apperrors.ConfigNotFound(path)
	}

	// 检查文件权限
	if err := CheckFilePermissions(path); err != nil {
		// 只警告，不阻止启动
		fmt.Fprintf(os.Stderr, "⚠ Warning: %v\n", err)
		fmt.Fprintf(os.Stderr, "⚠ Recommendation: chmod 0600 %s\n", path)
	}

	// 解析 TOML
	cfg := &Config{}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, apperrors.ConfigInvalid(err)
	}

	// 应用默认值
	ApplyDefaults(cfg)

	// 解析环境变量
	if err := ResolveEnvVars(cfg); err != nil {
		return nil, fmt.Errorf("failed to resolve environment variables: %w", err)
	}

	// 验证配置
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// CheckFilePermissions 检查配置文件权限
// 如果权限过于宽松（可被组或其他用户读取），返回警告
func CheckFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot check file permissions: %w", err)
	}

	mode := info.Mode().Perm()

	// 检查组和其他用户的读权限
	if mode&0077 != 0 {
		return apperrors.ConfigPermissionError(path, fmt.Sprintf("%o", mode))
	}

	return nil
}
