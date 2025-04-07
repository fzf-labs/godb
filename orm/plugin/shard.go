package plugin

import (
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/fzf-labs/godb/orm/utils"
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
			var t time.Time
			// columnValue要是time类型
			switch value := columnValue.(type) {
			case time.Time:
				t = value
			case *time.Time:
				t = *value
			default:
				// 时间转换
				t = carbon.Parse(utils.ConvToString(columnValue)).StdTime()
			}
			return "_" + t.Format("200601"), nil
		},
		PrimaryKeyGeneratorFn: func(tableIDx int64) int64 {
			return tableIDx
		},
	}, table)
}
