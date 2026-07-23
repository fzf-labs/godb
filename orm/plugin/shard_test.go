package plugin

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/fzf-labs/godb/internal/testenv"
)

func TestMonthShardingSuffix(t *testing.T) {
	fixedTime := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	suffix, err := monthShardingSuffix("created_at", fixedTime)
	assert.NoError(t, err)
	assert.Equal(t, "_202401", suffix)
}

func TestMonthShardingSuffixRejectsInvalidString(t *testing.T) {
	_, err := monthShardingSuffix("created_at", "not-a-time")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a valid time")
}

func TestMonthShardingSuffixCoversPointerNilAndString(t *testing.T) {
	fixedTime := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	suffix, err := monthShardingSuffix("created_at", &fixedTime)
	assert.NoError(t, err)
	assert.Equal(t, "_202402", suffix)

	suffix, err = monthShardingSuffix("created_at", "2024-03-04 05:06:07")
	assert.NoError(t, err)
	assert.Equal(t, "_202403", suffix)

	var nilTime *time.Time
	_, err = monthShardingSuffix("created_at", nilTime)
	assert.Error(t, err)

	_, err = monthShardingSuffix("created_at", nil)
	assert.Error(t, err)
}

func TestShardingPluginConstructors(t *testing.T) {
	assert.NotNil(t, NewShardingPlugin("orders", "user_id", 8))
	assert.NotNil(t, NewMonthShardingPlugin("orders", "created_at"))
}

func TestMonthShardingPluginRoutesSQLiteQueries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE orders_202401 (id integer, created_at datetime, name text)`).Error)
	require.NoError(t, db.Use(NewMonthShardingPlugin("orders", "created_at")))

	err = db.Exec(`INSERT INTO orders (created_at, name) VALUES (?, ?)`, "2024-01-02 03:04:05", "new-year").Error
	require.NoError(t, err)

	var count int64
	err = db.Table("orders_202401").Where("id = ? AND name = ?", 202401, "new-year").Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	err = db.Exec(`SELECT * FROM orders WHERE created_at = ?`, "2024-01-02 03:04:05").Error
	assert.NoError(t, err)
}

// TestNewMonthShardingPlugin 验证按月分片插件配置。
func TestNewMonthShardingPlugin(t *testing.T) {
	sqlDB, err := sql.Open("pgx", testenv.PostgresDSN("fkratos_sys"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}
	gormConfig.Logger = logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gormConfig)
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	require.NoError(t, db.Exec(`
CREATE TABLE IF NOT EXISTS sys_admin_202301 (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
)`).Error)
	require.NoError(t, db.Use(NewMonthShardingPlugin("sys_admin", "created_at")))
	// 按月分片仅支持等值条件
	err = db.Exec("SELECT * FROM sys_admin WHERE created_at = ?", "2023-01-13 20:58:35").Error
	require.NoError(t, err)
}

// TestNewShardingPlugin 验证通用分片插件配置。
func TestNewShardingPlugin(t *testing.T) {
	sqlDB, err := sql.Open("pgx", testenv.PostgresDSN("fkratos_sys"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}
	gormConfig.Logger = logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gormConfig)
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	for i := 0; i < 64; i++ {
		require.NoError(t, db.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS sys_admin_%02d (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
)`, i)).Error)
	}
	require.NoError(t, db.Use(NewShardingPlugin("sys_admin", "created_at", 64)))
	err = db.Exec("SELECT * FROM sys_admin WHERE created_at = ?", "2023-01-13 20:58:01").Error
	require.NoError(t, err)
}
