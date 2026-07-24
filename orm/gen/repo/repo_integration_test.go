//go:build integration

package repo

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

// TestGenerationTable 验证表级 repo 代码生成。
func TestGenerationTable(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")
	type args struct {
		db                    *gorm.DB
		dbname                string
		daoPath               string
		modelPath             string
		table                 string
		partitionTable        []string
		columnNameToDataType  map[string]string
		columnNameToName      map[string]string
		columnNameToFieldType map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "user_demo",
			args: args{
				db:        db,
				dbname:    "gorm_gen",
				daoPath:   "../../example/gorm/postgres/gorm_gen_dao",
				modelPath: "../../example/gorm/postgres/gorm_gen_model",
				table:     "user_demo",
				columnNameToDataType: map[string]string{
					"password":   "string",
					"avatar":     "string",
					"login_ip":   "string",
					"login_date": "time.Time",
					"created_at": "time.Time",
					"id":         "string",
					"uid":        "string",
					"post_ids":   "string",
					"email":      "string",
					"mobile":     "string",
					"status":     "int16",
					"tenant_id":  "int64",
					"username":   "string",
					"nickname":   "string",
					"dept_id":    "int64",
					"deleted_at": "gorm.DeletedAt",
					"updated_at": "time.Time",
					"remark":     "string",
					"sex":        "int16",
				},
				columnNameToName: map[string]string{
					"login_ip":   "LoginIP",
					"id":         "ID",
					"remark":     "Remark",
					"password":   "Password",
					"nickname":   "Nickname",
					"mobile":     "Mobile",
					"sex":        "Sex",
					"status":     "Status",
					"created_at": "CreatedAt",
					"uid":        "UID",
					"username":   "Username",
					"tenant_id":  "TenantID",
					"updated_at": "UpdatedAt",
					"deleted_at": "DeletedAt",
					"dept_id":    "DeptID",
					"post_ids":   "PostIds",
					"login_date": "LoginDate",
					"email":      "Email",
					"avatar":     "Avatar",
				},
				columnNameToFieldType: map[string]string{
					"post_ids":   "String",
					"email":      "String",
					"status":     "Int16",
					"login_ip":   "String",
					"tenant_id":  "Int64",
					"created_at": "Time",
					"username":   "String",
					"nickname":   "String",
					"dept_id":    "Int64",
					"mobile":     "String",
					"updated_at": "Time",
					"id":         "String",
					"password":   "String",
					"login_date": "Time",
					"deleted_at": "Field",
					"remark":     "String",
					"avatar":     "String",
					"uid":        "String",
					"sex":        "Int16",
				},
			},
			wantErr: false,
		},
		{
			name: "partition_table",
			args: args{
				db:             db,
				dbname:         "gorm_gen",
				daoPath:        "../../example/gorm/postgres/gorm_gen_dao",
				modelPath:      "../../example/gorm/postgres/gorm_gen_model",
				table:          "partition_table",
				partitionTable: []string{},
				columnNameToDataType: map[string]string{
					"id":         "string",
					"user_id":    "string",
					"created_at": "time.Time",
				},
				columnNameToName: map[string]string{
					"id":         "ID",
					"user_id":    "UserID",
					"created_at": "CreatedAt",
				},
				columnNameToFieldType: map[string]string{
					"id":         "String",
					"user_id":    "String",
					"created_at": "Time",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.table == "partition_table" {
				if err := tt.args.db.Exec(`
CREATE TABLE IF NOT EXISTS partition_table (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL
)`).Error; err != nil {
					t.Fatalf("create partition_table: %v", err)
				}
			}
			repoPath := t.TempDir()
			if err := GenerationTable(tt.args.db, tt.args.dbname, tt.args.daoPath, tt.args.modelPath, repoPath, tt.args.table, tt.args.partitionTable, tt.args.columnNameToDataType, tt.args.columnNameToName, tt.args.columnNameToFieldType); (err != nil) != tt.wantErr {
				t.Errorf("GenerationTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
