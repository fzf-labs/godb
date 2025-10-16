package gorm

import (
	"context"
	"fmt"
	"testing"

	"github.com/fzf-labs/godb/orm/condition"
	"github.com/fzf-labs/godb/orm/dbcache/goredisdbcache"
	"github.com/fzf-labs/godb/orm/encoding"
	"github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_dao"
	gorm_gen_model2 "github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_model"
	gorm_gen_repo2 "github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_repo"
	"github.com/fzf-labs/godb/orm/gen/config"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func newDB() *gorm.DB {
	db := gormx.NewDebugGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return nil
	}
	return db
}

func newRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6379",
		Password: "123456",
	})
	return redisClient
}

// Test_FindOneCacheByID 根据ID查询单条数据
func Test_FindOneCacheByID(t *testing.T) {
	db := newDB()
	redisClient := newRedis()
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	result, err := repo.FindOneByID(ctx, "182a65a0-ee20-4fe0-a0e8-ba30edcf402b")
	if err != nil {
		return
	}
	fmt.Println(result)
	assert.Equal(t, nil, err)
}

func Test_FindMultiCacheByUsernames(t *testing.T) {
	db := newDB()
	redisClient := newRedis()
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	result, err := repo.FindMultiCacheByUsernames(ctx, []string{"a", "b", "c", "d", "e", "f", "g"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
	assert.Equal(t, nil, err)
}

func Test_UpdateOneCache(t *testing.T) {
	db := newDB()
	redisClient := newRedis()
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	data, err := repo.FindOneByID(ctx, "182a65a0-ee20-4fe0-a0e8-ba30edcf402b")
	if err != nil {
		return
	}
	oldData := repo.DeepCopy(data)
	data.Remark = "123"
	err = repo.UpdateOneCache(ctx, data, oldData)
	if err != nil {
		fmt.Println(err)
		return
	}
	assert.Equal(t, nil, err)
}

func Test_UpsertOneWithZeroCache(t *testing.T) {
	db := newDB()
	redisClient := newRedis()
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	data := repo.NewData()
	data.ID = "182a65a0-ee20-4fe0-a0e8-ba30edcf402b"
	data.Remark = "123"
	err := repo.UpsertOneCache(ctx, data)
	if err != nil {
		fmt.Println(err)
		return
	}
	assert.Equal(t, nil, err)
}

func Test_UpdateBatchByIDS(t *testing.T) {
	db := newDB()
	redisClient := newRedis()
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	err := repo.UpdateBatchByIDS(ctx, []string{"182a65a0-ee20-4fe0-a0e8-ba30edcf402b", "2cc31ef9-7d6b-438b-874c-01d84a332b57"}, map[string]interface{}{
		"remark": "test",
	})
	if err != nil {
		return
	}
	assert.Equal(t, nil, err)
}

func Test_FindMultiCacheByTenantIDS(t *testing.T) {
	db := newDB()
	redisClient := newRedis()
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	result, err := repo.FindMultiCacheByTenantIDS(ctx, []int64{1, 2})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
	assert.Equal(t, nil, err)
}

// Test_FindMultiByCustom 自定义查询
func Test_FindMultiByCondition(t *testing.T) {
	db := newDB()
	client, _ := redismock.NewClientMock()
	dbCache := goredisdbcache.NewGoRedisDBCache(client)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewAdminDemoRepo(cfg)
	result, p, err := repo.FindMultiByCondition(ctx, &condition.Req{
		Page:     1,
		PageSize: 10,
		Order: []*condition.OrderParam{
			{
				Field: "created_at",
				Order: condition.DESC,
			},
		},
		Query: []*condition.QueryParam{
			{
				Field: "username",
				Value: "admin",
				Exp:   condition.EQ,
				Logic: condition.AND,
			},
			{
				Field: "username",
				Value: []interface{}{"admin", "admin2"},
				Exp:   condition.IN,
				Logic: "",
			},
			{
				Field: "username",
				Value: "123",
				Exp:   condition.LIKE,
				Logic: "",
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", result)
	fmt.Printf("%+v\n", p)
	assert.Equal(t, nil, err)
}

// Test_Tx 使用事务
func Test_Tx(t *testing.T) {
	db := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return
	}
	client, _ := redismock.NewClientMock()
	dbCache := goredisdbcache.NewGoRedisDBCache(client)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	adminDemoRepo := gorm_gen_repo2.NewAdminDemoRepo(cfg)
	adminLogDemoRepo := gorm_gen_repo2.NewAdminLogDemoRepo(cfg)
	err := gorm_gen_dao.Use(db).Transaction(func(tx *gorm_gen_dao.Query) error {
		err2 := adminDemoRepo.UpsertOneByTx(ctx, tx, &gorm_gen_model2.AdminDemo{
			ID:       "c8ddd930-339a-408b-8acb-fac22f5b43aa",
			Username: "admin",
			Nickname: "admin",
			Gender:   0,
			RoleIds:  nil,
			Salt:     "123",
			Status:   1,
		})
		if err2 != nil {
			return err2
		}
		err2 = adminLogDemoRepo.CreateOneByTx(ctx, tx, &gorm_gen_model2.AdminLogDemo{
			AdminID:   "c8ddd930-339a-408b-8acb-fac22f5b43aa",
			IP:        "0.0.0.0",
			URI:       "www.baidu.com",
			Useragent: "apifox",
			Header:    nil,
			Req:       nil,
			Resp:      nil,
		})
		if err2 != nil {
			return err2
		}
		return nil
	})
	assert.Equal(t, nil, err)
}
