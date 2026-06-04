package gormx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewGormPostgresClient(t *testing.T) {
	config := ClientConfig{
		Driver:          "postgres",
		DataSourceName:  "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=user sslmode=disable TimeZone=Asia/Shanghai",
		MaxIdleConn:     0,
		MaxOpenConn:     0,
		ConnMaxLifeTime: 0,
		ShowLog:         false,
		Tracing:         false,
	}
	_, err := NewGormClient(&config)
	fmt.Println(err)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	assert.Equal(t, nil, err)
}

func TestRunHealthCheckQuery_ExecutesSQL(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = runHealthCheckQuery(db, "select * from definitely_missing_table")
	assert.Error(t, err)
}

func TestGetHealthStatus_Healthy(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, health, GetHealthStatus(db))
}
