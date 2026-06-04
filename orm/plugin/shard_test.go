package plugin

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
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

// TestNewMonthShardingPlugin 验证按月分片插件配置。
func TestNewMonthShardingPlugin(t *testing.T) {
	sqlDB, err := sql.Open("pgx", "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=fkratos_sys sslmode=disable TimeZone=Asia/Shanghai")
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer sqlDB.Close()
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}
	gormConfig.Logger = logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gormConfig)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	err = db.Use(NewMonthShardingPlugin("sys_admin", "created_at"))
	if err != nil {
		fmt.Printf("gormopentracing new failed!  err: %+v", err)
	}
	// this record will insert to orders_03
	err = db.Exec("SELECT * FROM sys_admin WHERE created_at in ('2023-01-13 20:58:35')  ").Error
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, nil, err)
}

// TestNewShardingPlugin 验证通用分片插件配置。
func TestNewShardingPlugin(t *testing.T) {
	sqlDB, err := sql.Open("pgx", "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=fkratos_sys sslmode=disable TimeZone=Asia/Shanghai")
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer sqlDB.Close()
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}
	gormConfig.Logger = logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gormConfig)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	err = db.Use(NewShardingPlugin("sys_admin", "created_at", 64))
	if err != nil {
		fmt.Printf("gormopentracing new failed!  err: %+v", err)
	}
	// this record will insert to orders_03
	err = db.Exec("SELECT * FROM sys_admin WHERE created_at  ='2023-01-13 20:58:01'  ").Error
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, nil, err)
}
