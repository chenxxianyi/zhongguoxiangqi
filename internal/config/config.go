package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
	EngineMoveTime  time.Duration
	MaxUploadBytes  int64
	AllowedOrigin   string
	DataMode        string

	// 数据库配置（DataMode = "mysql" 时使用）
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	DBCharset      string
	DBLoc          string
	DBMaxOpenConns int
	DBMaxIdleConns int
	DBMaxLifetime  time.Duration
}

func Load() (Config, error) {
	// 自动加载 .env 文件（如果存在）
	_ = loadDotEnv(".env")       // 项目根目录
	_ = loadDotEnv("../.env")    // 从 cmd/api/ 或 cmd/worker/ 运行时

	shutdownTimeout, err := durationEnv("XIANGQI_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	engineMoveTime, err := durationEnv("XIANGQI_ENGINE_MOVE_TIME", 600*time.Millisecond)
	if err != nil {
		return Config{}, err
	}
	maxUploadBytes, err := int64Env("XIANGQI_MAX_UPLOAD_BYTES", 2<<20)
	if err != nil {
		return Config{}, err
	}
	dbMaxLifetime, err := durationEnv("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:        env("XIANGQI_HTTP_ADDR", ":8080"),
		ShutdownTimeout: shutdownTimeout,
		EngineMoveTime:  engineMoveTime,
		MaxUploadBytes:  maxUploadBytes,
		AllowedOrigin:   env("XIANGQI_ALLOWED_ORIGIN", "http://localhost:5666"),
		DataMode:        env("XIANGQI_DATA_MODE", "memory"),

		DBHost:         firstEnv("DATABASE_HOST", "XIANGQI_DB_HOST", "127.0.0.1"),
		DBPort:         firstEnv("DATABASE_PORT", "XIANGQI_DB_PORT", "3306"),
		DBName:         firstEnv("DATABASE_NAME", "XIANGQI_DB_NAME", "xiangqi"),
		DBUser:         firstEnv("DATABASE_USER", "XIANGQI_DB_USER", "xiangqi"),
		DBPassword:     firstEnv("DATABASE_PASSWORD", "XIANGQI_DB_PASSWORD", "change-me"),
		DBCharset:      firstEnv("DATABASE_CHARSET", "XIANGQI_DB_CHARSET", "utf8mb4"),
		DBLoc:          firstEnv("DATABASE_LOC", "XIANGQI_DB_LOC", "UTC"),
		DBMaxOpenConns: intEnv("DATABASE_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns: intEnv("DATABASE_MAX_IDLE_CONNS", 10),
		DBMaxLifetime:  dbMaxLifetime,
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("XIANGQI_HTTP_ADDR must not be empty")
	}
	if cfg.ShutdownTimeout <= 0 || cfg.EngineMoveTime <= 0 {
		return Config{}, fmt.Errorf("timeouts must be positive")
	}
	if cfg.MaxUploadBytes < 1024 || cfg.MaxUploadBytes > 64<<20 {
		return Config{}, fmt.Errorf("XIANGQI_MAX_UPLOAD_BYTES must be between 1 KiB and 64 MiB")
	}
	switch cfg.DataMode {
	case "memory":
	case "mysql":
	default:
		return Config{}, fmt.Errorf("unsupported XIANGQI_DATA_MODE %q; supported: memory, mysql", cfg.DataMode)
	}
	return cfg, nil
}

// DSN 构建 MySQL 连接串。
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=%s&loc=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBCharset,
		c.DBLoc,
	)
}

// DSNWithoutDB 构建一个不指定数据库的连接串（用于创建数据库）。
func (c Config) DSNWithoutDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&charset=%s&loc=%s&multiStatements=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBCharset,
		c.DBLoc,
	)
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// firstEnv 返回第一个存在的环境变量值，都不存在时返回 fallback。
func firstEnv(keys ...string) string {
	for _, key := range keys[:len(keys)-1] {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
	}
	return keys[len(keys)-1]
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}

func int64Env(key string, fallback int64) (int64, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}

func intEnv(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

// loadDotEnv 读取 .env 格式文件，将未设置的环境变量注入进程。
// 已由外部设置的变量不会被覆盖。
func loadDotEnv(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 支持两种格式: KEY=value 和 KEY=value # comment
		if idx := strings.IndexByte(line, '#'); idx > 0 && (idx == 0 || line[idx-1] != '\\') {
			line = strings.TrimSpace(line[:idx])
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// 移除首尾引号
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		// 不覆盖已设置的变量
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return nil
}
