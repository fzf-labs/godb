package repo

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

var benchRepoString string

func benchmarkRepoContext() *Repo {
	return &Repo{
		gorm:           &gorm.DB{},
		daoPath:        "github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_dao",
		modelPath:      "github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_model",
		repoPath:       "github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_repo",
		table:          "user_demo",
		dbName:         "godb",
		firstTableChar: "u",
		lowerTableName: "userDemo",
		upperTableName: "UserDemo",
		index:          benchmarkRepoIndexes(),
		haveDeletedAt:  true,
		columnNameToDataType: map[string]string{
			"id":         "int64",
			"user_name":  "string",
			"enabled":    "bool",
			"org_id":     "int64",
			"role_id":    "int64",
			"created_at": "time.Time",
			"deleted_at": "gorm.DeletedAt",
		},
		columnNameToName: map[string]string{
			"id":         "ID",
			"user_name":  "UserName",
			"enabled":    "Enabled",
			"org_id":     "OrgID",
			"role_id":    "RoleID",
			"created_at": "CreatedAt",
			"deleted_at": "DeletedAt",
		},
		columnNameToFieldType: map[string]string{
			"id":         "int64",
			"user_name":  "string",
			"enabled":    "bool",
			"org_id":     "int64",
			"role_id":    "int64",
			"created_at": "time.Time",
			"deleted_at": "gorm.DeletedAt",
		},
	}
}

func benchmarkRepoIndexes() []DBIndex {
	return []DBIndex{
		{
			Name:       "PRIMARY",
			ColumnKey:  "id",
			PrimaryKey: true,
			Unique:     true,
			Columns:    []string{"id"},
		},
		{
			Name:       "uk_user_name",
			ColumnKey:  "user_name",
			PrimaryKey: false,
			Unique:     true,
			Columns:    []string{"user_name"},
		},
		{
			Name:       "uk_org_role",
			ColumnKey:  "org_id:role_id",
			PrimaryKey: false,
			Unique:     true,
			Columns:    []string{"org_id", "role_id"},
		},
		{
			Name:       "idx_enabled_created_at",
			ColumnKey:  "enabled:created_at",
			PrimaryKey: false,
			Unique:     false,
			Columns:    []string{"enabled", "created_at"},
		},
	}
}

func BenchmarkRepoTemplateGeneration(b *testing.B) {
	r := benchmarkRepoContext()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		parts := []func() (string, error){
			r.generatePkg,
			r.generateImport,
			r.generateVar,
			r.generateTypes,
			r.generateNew,
			r.generateCommonMethods,
			r.generateCreateMethods,
			r.generateUpdateMethods,
			r.generateReadMethods,
			r.generateDelMethods,
			r.generateCommonFunc,
			r.generateCreateFunc,
			r.generateUpdateFunc,
			r.generateReadFunc,
			r.generateDelFunc,
		}
		for _, fn := range parts {
			got, err := fn()
			if err != nil {
				b.Fatal(err)
			}
			builder.WriteString(got)
			builder.WriteByte('\n')
		}
		benchRepoString = builder.String()
	}
}
