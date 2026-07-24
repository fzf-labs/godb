//go:build integration

package gormx

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func cleanupIntegrationGormDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, sqlDB.Close())
	})
}

func TestNewGormPostgresClient(t *testing.T) {
	config := ClientConfig{
		Driver:          "postgres",
		DataSourceName:  integrationPostgresDSN("user"),
		MaxIdleConn:     0,
		MaxOpenConn:     0,
		ConnMaxLifeTime: 0,
		ShowLog:         false,
		Tracing:         false,
	}
	db, err := NewGormClient(&config)
	require.NoError(t, err)
	cleanupIntegrationGormDB(t, db)
}

func TestNewDirectGormClientPostgresSuccess(t *testing.T) {
	db, err := newDirectGormClient(Postgres, integrationPostgresDSN("gorm_gen"), logger.Silent)
	require.NoError(t, err)
	cleanupIntegrationGormDB(t, db)
	assert.NotNil(t, db)
}
