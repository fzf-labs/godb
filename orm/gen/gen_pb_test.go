package gen

import (
	"testing"

	"github.com/fzf-labs/godb/orm/gormx"
)

func TestNewGenerationPb(t *testing.T) {
	db := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return
	}
	NewGenerationPB(
		db,
		"./example/postgres/pb",
		"api.gorm_gen.v1",
		"api/gorm_gen/v1;v1",
		WithPBOpts(ModelOptionRemoveDefault(), ModelOptionUnderline("ul_")),
	).Do()
}
