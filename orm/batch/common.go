package batch

import "math"

// getSQLQuantity 计算需要生成的 SQL 语句数量
func getSQLQuantity(length, batchSize int) int {
	return int(math.Ceil(float64(length) / float64(batchSize)))
}
