package plugin

import (
	"fmt"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/fzf-labs/godb/orm/utils/strutil"
	"gorm.io/sharding"
)

// NewShardingPlugin 按雪花算法
func NewShardingPlugin(table, shardingKey string, num uint) *sharding.Sharding {
	return sharding.Register(sharding.Config{
		ShardingKey:         shardingKey,
		NumberOfShards:      num,
		PrimaryKeyGenerator: sharding.PKSnowflake,
	}, table)
}

// NewMonthShardingPlugin 按月份分表
// 查询时必须传分表的主键,且只能取等判断
func NewMonthShardingPlugin(table, shardingKey string) *sharding.Sharding {
	return sharding.Register(sharding.Config{
		ShardingKey:         shardingKey,
		PrimaryKeyGenerator: sharding.PKCustom,
		ShardingAlgorithm: func(columnValue any) (suffix string, err error) {
			return monthShardingSuffix(shardingKey, columnValue)
		},
		PrimaryKeyGeneratorFn: func(tableIDx int64) int64 {
			return tableIDx
		},
	}, table)
}

func monthShardingSuffix(shardingKey string, columnValue any) (string, error) {
	t, err := monthShardingTime(shardingKey, columnValue)
	if err != nil {
		return "", err
	}
	return "_" + t.Format("200601"), nil
}

func monthShardingTime(shardingKey string, columnValue any) (time.Time, error) {
	if columnValue == nil {
		return time.Time{}, fmt.Errorf("sharding key %s cannot be nil", shardingKey)
	}

	switch value := columnValue.(type) {
	case time.Time:
		return value, nil
	case *time.Time:
		if value == nil {
			return time.Time{}, fmt.Errorf("sharding key %s cannot be nil", shardingKey)
		}
		return *value, nil
	default:
		parsed := carbon.Parse(strutil.ConvToString(columnValue))
		if parsed == nil || !parsed.IsValid() {
			return time.Time{}, fmt.Errorf("sharding key %s must be a valid time, got %T", shardingKey, columnValue)
		}
		return parsed.StdTime(), nil
	}
}
