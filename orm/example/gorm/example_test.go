package gorm

import (
	"context"
	"testing"

	"github.com/fzf-labs/godb/internal/testenv"
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
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newDB 创建示例测试用 PostgreSQL 连接。
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gormx.NewDebugGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	return db
}

// newRedis 创建示例测试用 Redis 客户端。
func newRedis(t *testing.T) *redis.Client {
	t.Helper()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     testenv.RedisAddr(),
		Password: testenv.RedisPassword(),
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		_ = redisClient.Close()
		testenv.SkipIfUnavailable(t, "redis unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = redisClient.Close()
	})
	return redisClient
}

// Test_DeepCopy 验证模型深拷贝逻辑。
func Test_DeepCopy(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewAdminRoleDemoRepo(cfg)
	data := repo.NewData()
	data.ID = "182a65a0-ee20-4fe0-a0e8-ba30edcf402b"
	data.Name = "admin"
	data.Admins = []*gorm_gen_model2.AdminDemo{
		{
			ID:       "182a65a0-ee20-4fe0-a0e8-ba30edcf402b",
			Username: "admin",
			Nickname: "admin",
			Gender:   0,
		},
	}
	copyData := repo.DeepCopy(data)
	// 修改值Admins的值
	data.Name = "admin2"
	data.Admins[0].Username = "admin2"
	data.Admins[0].Nickname = "admin2"
	assert.Equal(t, "admin", copyData.Name)
	require.Len(t, copyData.Admins, 1)
	assert.Equal(t, "admin", copyData.Admins[0].Username)
	assert.Equal(t, "admin", copyData.Admins[0].Nickname)
}

// Test_FindOneCacheByID 根据ID查询单条数据
func Test_FindOneCacheByID(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	result, err := repo.FindOneByID(ctx, "182a65a0-ee20-4fe0-a0e8-ba30edcf402b")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "182a65a0-ee20-4fe0-a0e8-ba30edcf402b", result.ID)
	assert.Equal(t, "user-1", result.UID)
	assert.Equal(t, "a", result.Username)
	assert.Equal(t, "user-a", result.Nickname)
}

// Test_FindMultiCacheByUsernames 验证按用户名批量查询缓存。
func Test_FindMultiCacheByUsernames(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	result, err := repo.FindMultiCacheByUsernames(ctx, []string{"a", "b", "c", "d", "e", "f", "g"})
	require.NoError(t, err)
	usernames := make([]string, 0, len(result))
	for _, item := range result {
		usernames = append(usernames, item.Username)
	}
	assert.ElementsMatch(t, []string{"a", "b"}, usernames)
}

// Test_UpdateOneCache 验证单条记录缓存更新。
func Test_UpdateOneCache(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	data, err := repo.FindOneByID(ctx, "182a65a0-ee20-4fe0-a0e8-ba30edcf402b")
	require.NoError(t, err)
	oldData := repo.DeepCopy(data)
	data.Remark = "123"
	err = repo.UpdateOneCache(ctx, data, oldData)
	require.NoError(t, err)
	updated, err := repo.FindOneByID(ctx, data.ID)
	require.NoError(t, err)
	assert.Equal(t, "123", updated.Remark)
}

// Test_UpsertOneWithZeroCache 验证带零值字段的 upsert 缓存更新。
func Test_UpsertOneWithZeroCache(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	data := repo.NewData()
	data.ID = "182a65a0-ee20-4fe0-a0e8-ba30edcf402b"
	data.Remark = "123"
	err := repo.UpsertOneCache(ctx, data)
	require.NoError(t, err)
	updated, err := repo.FindOneByID(ctx, data.ID)
	require.NoError(t, err)
	assert.Equal(t, "123", updated.Remark)
}

// Test_UpsertOneCacheByFieldsTx 验证事务内按字段 upsert 缓存。
func Test_UpsertOneCacheByFieldsTx(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	data := repo.NewData()
	data.ID = "182a65a0-ee20-4fe0-a0e8-ba30edcf402b"
	data.Remark = "123"
	err := gorm_gen_dao.Use(db).Transaction(func(tx *gorm_gen_dao.Query) error {
		err := repo.UpsertOneCacheByFieldsTx(ctx, tx, data, []string{"id"})
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)
	updated, err := repo.FindOneByID(ctx, data.ID)
	require.NoError(t, err)
	assert.Equal(t, "123", updated.Remark)
}

// Test_UpdateBatchByIDS 验证按 ID 批量更新。
func Test_UpdateBatchByIDS(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	err := repo.UpdateBatchByIDS(ctx, []string{"182a65a0-ee20-4fe0-a0e8-ba30edcf402b", "2cc31ef9-7d6b-438b-874c-01d84a332b57"}, map[string]interface{}{
		"remark": "test",
	})
	require.NoError(t, err)
	for _, id := range []string{"182a65a0-ee20-4fe0-a0e8-ba30edcf402b", "2cc31ef9-7d6b-438b-874c-01d84a332b57"} {
		updated, err := repo.FindOneByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "test", updated.Remark)
	}
}

// Test_FindMultiCacheByTenantIDS 验证按租户 ID 批量查询缓存。
func Test_FindMultiCacheByTenantIDS(t *testing.T) {
	db := newDB(t)
	redisClient := newRedis(t)
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	repo := gorm_gen_repo2.NewUserDemoRepo(cfg)
	result, err := repo.FindMultiCacheByTenantIDS(ctx, []int64{1, 2})
	require.NoError(t, err)
	tenantIDs := make([]int64, 0, len(result))
	for _, item := range result {
		tenantIDs = append(tenantIDs, item.TenantID)
	}
	assert.ElementsMatch(t, []int64{1, 2}, tenantIDs)
}

// Test_FindMultiByCustom 自定义查询
func Test_FindMultiByCondition(t *testing.T) {
	db := newDB(t)
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
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, int32(1), p.Page)
	assert.Equal(t, int32(10), p.PageSize)
	assert.LessOrEqual(t, len(result), 10)
}

// Test_Tx 使用事务
func Test_Tx(t *testing.T) {
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	client, _ := redismock.NewClientMock()
	dbCache := goredisdbcache.NewGoRedisDBCache(client)
	ctx := context.Background()
	cfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	adminDemoRepo := gorm_gen_repo2.NewAdminDemoRepo(cfg)
	adminLogDemoRepo := gorm_gen_repo2.NewAdminLogDemoRepo(cfg)
	err = gorm_gen_dao.Use(db).Transaction(func(tx *gorm_gen_dao.Query) error {
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
	require.NoError(t, err)
	created, err := adminDemoRepo.FindOneByID(ctx, "c8ddd930-339a-408b-8acb-fac22f5b43aa")
	require.NoError(t, err)
	assert.Equal(t, "admin", created.Username)
	assert.Equal(t, int16(1), created.Status)

	var logCount int64
	err = db.WithContext(ctx).Table("admin_log_demo").Where("admin_id = ?", "c8ddd930-339a-408b-8acb-fac22f5b43aa").Count(&logCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), logCount)
}
