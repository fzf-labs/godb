package repo

import (
	"testing"

	"github.com/fzf-labs/godb/orm/gormx"
	"gorm.io/gorm"
)

func newDB() *gorm.DB {
	db := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return nil
	}
	return db
}

func TestGenerationTable(t *testing.T) {
	db := newDB()
	type args struct {
		db                    *gorm.DB
		dbname                string
		daoPath               string
		modelPath             string
		repoPath              string
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
			name: "test1",
			args: args{
				db:        db,
				dbname:    "gorm_gen",
				daoPath:   "../example/postgres/gorm_gen_dao",
				modelPath: "../example/postgres/gorm_gen_model",
				repoPath:  "../example/postgres/gorm_gen_repo",
				table:     "data_type_demo",
				columnNameToDataType: map[string]string{
					"data_type_int8":      "int64",
					"data_type_time":      "time.Time",
					"_id":                 "string",
					"array_int8":          "pq.Int64Array",
					"id":                  "string",
					"data_type_json":      "datatypes.JSON",
					"data_type_jsonb":     "datatypes.JSON",
					"data_type_date":      "time.Time",
					"data_type_float8":    "float64",
					"data_type_text":      "string",
					"data_type_time_null": "sql.NullTime",
					"array_varchar":       "pq.StringArray",
					"array_int2":          "pq.Int32Array",
					"array_int4":          "pq.Int32Array",
					"data_type_int2":      "int16",
					"data_type_varchar":   "string",
					"created_at":          "time.Time",
					"updated_at":          "time.Time",
					"deleted_at":          "gorm.DeletedAt",
					"data_type_byte":      "[]uint8",
					"data_type_float4":    "float32",
					"cacheKey":            "string",
					"data_type_bool":      "bool",
				},
				columnNameToName: map[string]string{
					"data_type_varchar":   "DataTypeVarchar",
					"created_at":          "CreatedAt",
					"id":                  "ID",
					"data_type_json":      "DataTypeJSON",
					"deleted_at":          "DeletedAt",
					"data_type_time":      "DataTypeTime",
					"data_type_byte":      "DataTypeByte",
					"data_type_date":      "DataTypeDate",
					"array_int2":          "ArrayInt2",
					"data_type_bool":      "DataTypeBool",
					"data_type_int2":      "DataTypeInt2",
					"data_type_text":      "DataTypeText",
					"data_type_time_null": "DataTypeTimeNull",
					"data_type_float8":    "DataTypeFloat8",
					"cacheKey":            "CacheKey",
					"array_varchar":       "ArrayVarchar",
					"array_int4":          "ArrayInt4",
					"array_int8":          "ArrayInt8",
					"data_type_int8":      "DataTypeInt8",
					"updated_at":          "UpdatedAt",
					"data_type_jsonb":     "DataTypeJsonb",
					"data_type_float4":    "DataTypeFloat4",
					"_id":                 "ULid",
				},
				columnNameToFieldType: map[string]string{
					"data_type_int2":      "Int16",
					"data_type_int8":      "Int64",
					"data_type_time_null": "Field",
					"_id":                 "String",
					"array_varchar":       "Field",
					"data_type_json":      "Field",
					"created_at":          "Time",
					"deleted_at":          "Field",
					"data_type_jsonb":     "Field",
					"data_type_byte":      "Field",
					"data_type_varchar":   "String",
					"data_type_text":      "String",
					"data_type_date":      "Time",
					"array_int4":          "Field",
					"array_int8":          "Field",
					"data_type_float8":    "Float64",
					"cacheKey":            "String",
					"array_int2":          "Field",
					"id":                  "String",
					"data_type_bool":      "Bool",
					"updated_at":          "Time",
					"data_type_time":      "Time",
					"data_type_float4":    "Float32",
				},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				db:        db,
				dbname:    "gorm_gen",
				daoPath:   "../example/postgres/gorm_gen_dao",
				modelPath: "../example/postgres/gorm_gen_model",
				repoPath:  "../example/postgres/gorm_gen_repo",
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
			name: "test2",
			args: args{
				db:             db,
				dbname:         "gorm_gen",
				daoPath:        "../example/postgres/gorm_gen_dao",
				modelPath:      "../example/postgres/gorm_gen_model",
				repoPath:       "../example/postgres/gorm_gen_repo",
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
			if err := GenerationTable(tt.args.db, tt.args.dbname, tt.args.daoPath, tt.args.modelPath, tt.args.repoPath, tt.args.table, tt.args.partitionTable, tt.args.columnNameToDataType, tt.args.columnNameToName, tt.args.columnNameToFieldType); (err != nil) != tt.wantErr {
				t.Errorf("GenerationTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
