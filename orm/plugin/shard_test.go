package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
