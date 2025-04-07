## 索引生成查询缓存规则

### 创建：

- 普通:
  - 单个创建
    - createOne ，createOneCache
    - CreateOneByTx，CreateOneCacheByTx
  - 批量创建
    - CreateBatch，CreateBatchCache
    - CreateBatchByTx，CreateBatchCacheByTx
- upsert:
  - UpsertOne，UpsertOneCache
  - UpsertOneByTx，UpsertOneCacheByTx
  - UpsertOneByFields，UpsertOneCacheByFields
  - UpsertOneByFieldsTx，UpsertOneCacheByFieldsTx

### 更新：
- 忽略 0 值：
  - UpdateOne，UpdateOneCache
  - UpdateOneByTx，UpdateOneCacheByTx
- 不忽略 0 值
  - UpdateOneWithZero，UpdateOneCacheWithZero
  - UpdateOneWithZeroByTx，UpdateOneCacheWithZeroByTx

### 删除：
- 唯一性索引：
  - 字段数为 1：
    - DeleteOneBy{{Field}}，DeleteOneCacheBy{{Field}}
    - DeleteOneBy{{Field}}Tx，DeleteOneCacheBy{{Field}}Tx
    - DeleteMultiBy{{FieldPlural}}，DeleteMultiCacheBy{{FieldPlural}}
    - DeleteMultiBy{{FieldPlural}}Tx，DeleteMultiCacheBy{{FieldPlural}}Tx
  - 字段数不为 1：
    - DeleteOneBy{{Fields}}，DeleteOneCacheBy{{Fields}}
    - DeleteOneBy{{Fields}}Tx，DeleteOneCacheBy{{Fields}}Tx
- 非唯一性索引：
  - 字段数为 1：
    - DeleteMultiBy{{Field}}，DeleteMultiCacheBy{{Field}}
    - DeleteMultiBy{{Field}}Tx，DeleteMultiCacheBy{{Field}}Tx
    - DeleteMultiBy{{FieldPlural}}，DeleteMultiCacheBy{{FieldPlural}}
    - DeleteMultiBy{{FieldPlural}}Tx，DeleteMultiCacheBy{{FieldPlural}}Tx
  - 字段数不为 1：
    - DeleteMultiBy{{Fields}}，DeleteMultiCacheBy{{Fields}}
    - DeleteMultiBy{{Fields}}Tx，DeleteMultiCacheBy{{Fields}}Tx

### 查询：
- 唯一性索引：
  - 字段数为 1：
    - FindOneBy{{Field}}，FindOneCacheBy{{Field}}
    - FindMultiBy{{FieldPlural}}，FindMultiCacheBy{{FieldPlural}}
  - 字段数不为 1：
    - FindOneBy{{Fields}}，FindOneCacheBy{{Fields}}
- 非唯一性索引：
  - 字段数为 1：
    - FindMultiBy{{Field}}，FindMultiCacheBy{{Field}}
    - FindMultiBy{{FieldPlural}}，FindMultiCacheBy{{FieldPlural}}
  - 字段数不为 1：
    - FindMultiBy{{Fields}}，FindMultiCacheBy{{Fields}}