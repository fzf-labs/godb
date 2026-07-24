//go:build integration

package proto

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/orm/gormx"
)

func newIntegrationPostgresDB(t *testing.T, dbname string) *gorm.DB {
	t.Helper()
	envOrDefault := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=1 TimeZone=Asia/Shanghai",
		envOrDefault("PGHOST", "127.0.0.1"),
		envOrDefault("PGPORT", "5432"),
		envOrDefault("PGUSER", "postgres"),
		envOrDefault("PGPASSWORD", "123456"),
		dbname,
	)
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, dsn)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})
	return db
}

// TestGenerationPB 验证从数据库表生成 proto。
func TestGenerationPB(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")
	type args struct {
		db                   *gorm.DB
		outPutPath           string
		packageStr           string
		goPackageStr         string
		table                string
		columnNameToName     map[string]string
		columnNameToDataType map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				db:           db,
				outPutPath:   "../../example/pb",
				packageStr:   "api.gorm_gen.v1",
				goPackageStr: "api/gorm_gen/v1;v1",
				table:        "admin_log_demo",
				columnNameToName: map[string]string{
					"id":         "ID",
					"admin_id":   "adminID",
					"ip":         "IP",
					"uri":        "URI",
					"useragent":  "Useragent",
					"header":     "Header",
					"req":        "Req",
					"resp":       "Resp",
					"created_at": "CreatedAt",
					"status":     "Status",
				},
				columnNameToDataType: map[string]string{
					"id":        "int64",
					"admin_id":  "int64",
					"ip":        "string",
					"uri":       "string",
					"useragent": "string",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenerationPB(tt.args.db, tt.args.outPutPath, tt.args.packageStr, tt.args.goPackageStr, tt.args.table, tt.args.columnNameToName, tt.args.columnNameToDataType); (err != nil) != tt.wantErr {
				t.Errorf("GenerationPB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
