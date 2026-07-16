package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xiangqi-lab/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect 创建 MySQL 连接池。
func Connect(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBMaxLifetime)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// Migrate 执行数据库迁移。
func Migrate(db *sql.DB, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查表是否已存在
	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'matches'",
	).Scan(&count); err != nil {
		return fmt.Errorf("check migration state: %w", err)
	}

	if count > 0 {
		logger.Info("database already migrated, skipping")
		return nil
	}

	sqlBytes, err := migrationsFS.ReadFile("migrations/0001_initial.sql")
	if err != nil {
		return fmt.Errorf("read embedded migration: %w", err)
	}

	statements := splitStatements(string(sqlBytes))
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		logger.Info("running migration statement", "idx", i+1, "total", len(statements))
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration statement %d: %w", i+1, err)
		}
	}

	logger.Info("database migration completed", "statements", len(statements))
	return nil
}

// AutoCreateDatabase 如果 DATABASE_AUTO_CREATE 环境变量为 "true"，则先创建数据库再建立连接。
func AutoCreateDatabase(cfg config.Config) (*sql.DB, error) {
	autoCreate := strings.ToLower(os.Getenv("DATABASE_AUTO_CREATE"))
	if autoCreate != "true" && autoCreate != "1" {
		return Connect(cfg)
	}

	adminDSN := cfg.DSNWithoutDB()
	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return nil, fmt.Errorf("open admin mysql: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s COLLATE %s_general_ci",
		cfg.DBName, cfg.DBCharset, cfg.DBCharset)
	if _, err := admin.ExecContext(ctx, createSQL); err != nil {
		_ = admin.Close()
		return nil, fmt.Errorf("create database %s: %w", cfg.DBName, err)
	}
	_ = admin.Close()

	return Connect(cfg)
}

// splitStatements 按空白行或 CREATE TABLE 边界分割 SQL 文本。
func splitStatements(sql string) []string {
	var result []string
	var current strings.Builder

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		// 到达语句边界（空白行或新的 CREATE）
		if (trimmed == "" || strings.HasPrefix(trimmed, "CREATE TABLE")) && current.Len() > 0 {
			result = append(result, current.String())
			current.Reset()
		}
		if trimmed != "" {
			current.WriteString(line)
			current.WriteByte('\n')
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
