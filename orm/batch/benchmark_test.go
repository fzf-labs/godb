package batch

import (
	"fmt"
	"testing"
	"time"
)

type benchmarkBatchRow struct {
	ID        int       `gorm:"column:id"`
	UserName  string    `gorm:"column:user_name"`
	Enabled   bool      `gorm:"column:enabled"`
	Score     float64   `gorm:"column:score"`
	Payload   []byte    `gorm:"column:payload"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

var benchBatchSQL []string

func benchmarkBatchRows() []*benchmarkBatchRow {
	rows := make([]*benchmarkBatchRow, 200)
	for i := range rows {
		rows[i] = &benchmarkBatchRow{
			ID:        i + 1,
			UserName:  fmt.Sprintf("user-%d's", i+1),
			Enabled:   i%2 == 0,
			Score:     float64(i) * 1.25,
			Payload:   []byte(fmt.Sprintf("payload-%d", i+1)),
			UpdatedAt: time.Date(2026, 6, 4, 12, 34, 56, 0, time.UTC),
		}
	}
	return rows
}

func BenchmarkMysqlBatchUpdateToSQLArray(b *testing.B) {
	rows := benchmarkBatchRows()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sqls, err := MysqlBatchUpdateToSQLArray("app.user_demo", rows)
		if err != nil {
			b.Fatal(err)
		}
		benchBatchSQL = sqls
	}
}

func BenchmarkPostgresBatchUpdateToSQLArray(b *testing.B) {
	rows := benchmarkBatchRows()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sqls, err := PostgresBatchUpdateToSQLArray("app.user_demo", rows)
		if err != nil {
			b.Fatal(err)
		}
		benchBatchSQL = sqls
	}
}
