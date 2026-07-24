//go:build integration

package plugin

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func integrationPostgresDSN(database string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=1 TimeZone=Asia/Shanghai",
		quotePostgresDSNValue(envOrDefault("PGHOST", "127.0.0.1")),
		quotePostgresDSNValue(envOrDefault("PGPORT", "5432")),
		quotePostgresDSNValue(envOrDefault("PGUSER", "postgres")),
		quotePostgresDSNValue(envOrDefault("PGPASSWORD", "123456")),
		quotePostgresDSNValue(database),
	)
}

func envOrDefault(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

func quotePostgresDSNValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return "'" + value + "'"
}

func openIntegrationShardDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, err := sql.Open("pgx", integrationPostgresDSN("fkratos_sys"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close PostgreSQL connection: %v", err)
		}
	})
	require.NoError(t, sqlDB.Ping())

	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Info),
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gormConfig)
	require.NoError(t, err)
	return db
}

// TestNewMonthShardingPlugin 验证按月分片插件配置。
func TestNewMonthShardingPlugin(t *testing.T) {
	db := openIntegrationShardDB(t)
	require.NoError(t, db.Exec(`
CREATE TABLE IF NOT EXISTS sys_admin_202301 (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
)`).Error)
	require.NoError(t, db.Use(NewMonthShardingPlugin("sys_admin", "created_at")))
	// 按月分片仅支持等值条件
	err := db.Exec("SELECT * FROM sys_admin WHERE created_at = ?", "2023-01-13 20:58:35").Error
	require.NoError(t, err)
}

// TestNewShardingPlugin 验证通用分片插件配置。
func TestNewShardingPlugin(t *testing.T) {
	db := openIntegrationShardDB(t)
	for i := 0; i < 64; i++ {
		require.NoError(t, db.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS sys_admin_%02d (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
)`, i)).Error)
	}
	require.NoError(t, db.Use(NewShardingPlugin("sys_admin", "created_at", 64)))
	err := db.Exec("SELECT * FROM sys_admin WHERE created_at = ?", "2023-01-13 20:58:01").Error
	require.NoError(t, err)
}
