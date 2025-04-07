package gormx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGormPostgresClient(t *testing.T) {
	config := ClientConfig{
		DB:              "",
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
		return
	}
	assert.Equal(t, nil, err)
}
