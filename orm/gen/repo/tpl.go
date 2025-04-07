package repo

import (
	_ "embed"
)

//go:embed tpl/pkg.tpl
var Pkg string

//go:embed tpl/import.tpl
var Import string

//go:embed tpl/var.tpl
var Var string

//go:embed tpl/var-cache.tpl
var VarCache string

//go:embed tpl/var-cache-keys.tpl
var VarCacheKeys string

//go:embed tpl/var-cache-global.tpl
var VarCacheGlobal string

//go:embed tpl/var-cache-del-key.tpl
var VarCacheDelKey string

//go:embed tpl/types.tpl
var Types string

//go:embed tpl/new.tpl
var New string

//go:embed tpl/interface-deep-copy.tpl
var InterfaceDeepCopy string

//go:embed tpl/deep-copy.tpl
var DeepCopy string

// 创建

//go:embed tpl/interface-create-one.tpl
var InterfaceCreateOne string

//go:embed tpl/create-one.tpl
var CreateOne string

//go:embed tpl/interface-create-one-cache.tpl
var InterfaceCreateOneCache string

//go:embed tpl/create-one-cache.tpl
var CreateOneCache string

//go:embed tpl/interface-create-one-by-tx.tpl
var InterfaceCreateOneByTx string

//go:embed tpl/create-one-by-tx.tpl
var CreateOneByTx string

//go:embed tpl/interface-create-one-cache-by-tx.tpl
var InterfaceCreateOneCacheByTx string

//go:embed tpl/create-one-cache-by-tx.tpl
var CreateOneCacheByTx string

//go:embed tpl/interface-create-batch.tpl
var InterfaceCreateBatch string

//go:embed tpl/create-batch.tpl
var CreateBatch string

//go:embed tpl/interface-create-batch-cache.tpl
var InterfaceCreateBatchCache string

//go:embed tpl/create-batch-cache.tpl
var CreateBatchCache string

//go:embed tpl/interface-create-batch-by-tx.tpl
var InterfaceCreateBatchByTx string

//go:embed tpl/create-batch-by-tx.tpl
var CreateBatchByTx string

//go:embed tpl/interface-create-batch-cache-by-tx.tpl
var InterfaceCreateBatchCacheByTx string

//go:embed tpl/create-batch-cache-by-tx.tpl
var CreateBatchCacheByTx string

//go:embed tpl/interface-upsert-one.tpl
var InterfaceUpsertOne string

//go:embed tpl/upsert-one.tpl
var UpsertOne string

//go:embed tpl/interface-upsert-one-cache.tpl
var InterfaceUpsertOneCache string

//go:embed tpl/upsert-one-cache.tpl
var UpsertOneCache string

//go:embed tpl/interface-upsert-one-by-tx.tpl
var InterfaceUpsertOneByTx string

//go:embed tpl/upsert-one-by-tx.tpl
var UpsertOneByTx string

//go:embed tpl/interface-upsert-one-cache-by-tx.tpl
var InterfaceUpsertOneCacheByTx string

//go:embed tpl/upsert-one-cache-by-tx.tpl
var UpsertOneCacheByTx string

//go:embed tpl/interface-upsert-one-by-fields.tpl
var InterfaceUpsertOneByFields string

//go:embed tpl/upsert-one-by-fields.tpl
var UpsertOneByFields string

//go:embed tpl/interface-upsert-one-cache-by-fields.tpl
var InterfaceUpsertOneCacheByFields string

//go:embed tpl/upsert-one-cache-by-fields.tpl
var UpsertOneCacheByFields string

//go:embed tpl/interface-upsert-one-by-fields-tx.tpl
var InterfaceUpsertOneByFieldsTx string

//go:embed tpl/upsert-one-by-fields-tx.tpl
var UpsertOneByFieldsTx string

//go:embed tpl/interface-upsert-one-cache-by-fields-tx.tpl
var InterfaceUpsertOneCacheByFieldsTx string

//go:embed tpl/upsert-one-cache-by-fields-tx.tpl
var UpsertOneCacheByFieldsTx string

// 更新

//go:embed tpl/interface-update-one.tpl
var InterfaceUpdateOne string

//go:embed tpl/update-one.tpl
var UpdateOne string

//go:embed tpl/interface-update-one-cache.tpl
var InterfaceUpdateOneCache string

//go:embed tpl/update-one-cache.tpl
var UpdateOneCache string

//go:embed tpl/interface-update-one-by-tx.tpl
var InterfaceUpdateOneByTx string

//go:embed tpl/update-one-by-tx.tpl
var UpdateOneByTx string

//go:embed tpl/interface-update-one-cache-by-tx.tpl
var InterfaceUpdateOneCacheByTx string

//go:embed tpl/update-one-cache-by-tx.tpl
var UpdateOneCacheByTx string

//go:embed tpl/interface-update-one-with-zero.tpl
var InterfaceUpdateOneWithZero string

//go:embed tpl/update-one-with-zero.tpl
var UpdateOneWithZero string

//go:embed tpl/interface-update-one-cache-with-zero.tpl
var InterfaceUpdateOneCacheWithZero string

//go:embed tpl/update-one-cache-with-zero.tpl
var UpdateOneCacheWithZero string

//go:embed tpl/interface-update-one-with-zero-by-tx.tpl
var InterfaceUpdateOneWithZeroByTx string

//go:embed tpl/update-one-with-zero-by-tx.tpl
var UpdateOneWithZeroByTx string

//go:embed tpl/interface-update-one-cache-with-zero-by-tx.tpl
var InterfaceUpdateOneCacheWithZeroByTx string

//go:embed tpl/update-one-cache-with-zero-by-tx.tpl
var UpdateOneCacheWithZeroByTx string

//go:embed tpl/interface-update-batch-by-primary-keys.tpl
var InterfaceUpdateBatchByPrimaryKeys string

//go:embed tpl/update-batch-by-primary-keys.tpl
var UpdateBatchByPrimaryKeys string

//go:embed tpl/interface-update-batch-by-primary-keys-tx.tpl
var InterfaceUpdateBatchByPrimaryKeysTx string

//go:embed tpl/update-batch-by-primary-keys-tx.tpl
var UpdateBatchByPrimaryKeysTx string

// 删除

//go:embed tpl/interface-delete-one-by-field.tpl
var InterfaceDeleteOneByField string

//go:embed tpl/delete-one-by-field.tpl
var DeleteOneByField string

//go:embed tpl/interface-delete-one-cache-by-field.tpl
var InterfaceDeleteOneCacheByField string

//go:embed tpl/delete-one-cache-by-field.tpl
var DeleteOneCacheByField string

//go:embed tpl/interface-delete-one-by-fields.tpl
var InterfaceDeleteOneByFields string

//go:embed tpl/delete-one-by-fields.tpl
var DeleteOneByFields string

//go:embed tpl/interface-delete-one-cache-by-fields.tpl
var InterfaceDeleteOneCacheByFields string

//go:embed tpl/delete-one-cache-by-fields.tpl
var DeleteOneCacheByFields string

//go:embed tpl/interface-delete-one-by-field-tx.tpl
var InterfaceDeleteOneByFieldTx string

//go:embed tpl/delete-one-by-field-tx.tpl
var DeleteOneByFieldTx string

//go:embed tpl/interface-delete-one-cache-by-field-tx.tpl
var InterfaceDeleteOneCacheByFieldTx string

//go:embed tpl/delete-one-cache-by-field-tx.tpl
var DeleteOneCacheByFieldTx string

//go:embed tpl/interface-delete-one-by-fields-tx.tpl
var InterfaceDeleteOneByFieldsTx string

//go:embed tpl/delete-one-by-fields-tx.tpl
var DeleteOneByFieldsTx string

//go:embed tpl/interface-delete-one-cache-by-fields-tx.tpl
var InterfaceDeleteOneCacheByFieldsTx string

//go:embed tpl/delete-one-cache-by-fields-tx.tpl
var DeleteOneCacheByFieldsTx string

//go:embed tpl/interface-delete-multi-by-field.tpl
var InterfaceDeleteMultiByField string

//go:embed tpl/delete-multi-by-field.tpl
var DeleteMultiByField string

//go:embed tpl/interface-delete-multi-cache-by-field.tpl
var InterfaceDeleteMultiCacheByField string

//go:embed tpl/delete-multi-cache-by-field.tpl
var DeleteMultiCacheByField string

//go:embed tpl/interface-delete-multi-by-field-tx.tpl
var InterfaceDeleteMultiByFieldTx string

//go:embed tpl/delete-multi-by-field-tx.tpl
var DeleteMultiByFieldTx string

//go:embed tpl/interface-delete-multi-cache-by-field-tx.tpl
var InterfaceDeleteMultiCacheByFieldTx string

//go:embed tpl/delete-multi-cache-by-field-tx.tpl
var DeleteMultiCacheByFieldTx string

//go:embed tpl/interface-delete-multi-by-field-plural.tpl
var InterfaceDeleteMultiByFieldPlural string

//go:embed tpl/delete-multi-by-field-plural.tpl
var DeleteMultiByFieldPlural string

//go:embed tpl/interface-delete-multi-cache-by-field-plural.tpl
var InterfaceDeleteMultiCacheByFieldPlural string

//go:embed tpl/delete-multi-cache-by-field-plural.tpl
var DeleteMultiCacheByFieldPlural string

//go:embed tpl/interface-delete-multi-by-field-plural-tx.tpl
var InterfaceDeleteMultiByFieldPluralTx string

//go:embed tpl/delete-multi-by-field-plural-tx.tpl
var DeleteMultiByFieldPluralTx string

//go:embed tpl/interface-delete-multi-cache-by-field-plural-tx.tpl
var InterfaceDeleteMultiCacheByFieldPluralTx string

//go:embed tpl/delete-multi-cache-by-field-plural-tx.tpl
var DeleteMultiCacheByFieldPluralTx string

//go:embed tpl/interface-delete-multi-by-fields.tpl
var InterfaceDeleteMultiByFields string

//go:embed tpl/delete-multi-by-fields.tpl
var DeleteMultiByFields string

//go:embed tpl/interface-delete-multi-cache-by-fields.tpl
var InterfaceDeleteMultiCacheByFields string

//go:embed tpl/delete-multi-cache-by-fields.tpl
var DeleteMultiCacheByFields string

//go:embed tpl/interface-delete-multi-by-fields-tx.tpl
var InterfaceDeleteMultiByFieldsTx string

//go:embed tpl/delete-multi-by-fields-tx.tpl
var DeleteMultiByFieldsTx string

//go:embed tpl/interface-delete-multi-cache-by-fields-tx.tpl
var InterfaceDeleteMultiCacheByFieldsTx string

//go:embed tpl/delete-multi-cache-by-fields-tx.tpl
var DeleteMultiCacheByFieldsTx string

//go:embed tpl/interface-delete-index-cache.tpl
var InterfaceDeleteIndexCache string

//go:embed tpl/delete-index-cache.tpl
var DeleteIndexCache string

// 查询

//go:embed tpl/interface-find-one-by-field.tpl
var InterfaceFindOneByField string

//go:embed tpl/find-one-by-field.tpl
var FindOneByField string

//go:embed tpl/interface-find-one-cache-by-field.tpl
var InterfaceFindOneCacheByField string

//go:embed tpl/find-one-cache-by-field.tpl
var FindOneCacheByField string

//go:embed tpl/interface-find-multi-by-field-plural.tpl
var InterfaceFindMultiByFieldPlural string

//go:embed tpl/find-multi-by-field-plural.tpl
var FindMultiByFieldPlural string

//go:embed tpl/interface-find-multi-cache-by-field-plural-unique-true.tpl
var InterfaceFindMultiCacheByFieldPluralUniqueTrue string

//go:embed tpl/find-multi-cache-by-field-plural-unique-true.tpl
var FindMultiCacheByFieldPluralUniqueTrue string

//go:embed tpl/interface-find-multi-cache-by-field-plural-unique-false.tpl
var InterfaceFindMultiCacheByFieldPluralUniqueFalse string

//go:embed tpl/find-multi-cache-by-field-plural-unique-false.tpl
var FindMultiCacheByFieldPluralUniqueFalse string

//go:embed tpl/interface-find-one-by-fields.tpl
var InterfaceFindOneByFields string

//go:embed tpl/find-one-by-fields.tpl
var FindOneByFields string

//go:embed tpl/interface-find-one-cache-by-fields.tpl
var InterfaceFindOneCacheByFields string

//go:embed tpl/find-one-cache-by-fields.tpl
var FindOneCacheByFields string

//go:embed tpl/interface-find-multi-by-field.tpl
var InterfaceFindMultiByField string

//go:embed tpl/find-multi-by-field.tpl
var FindMultiByField string

//go:embed tpl/interface-find-multi-cache-by-field.tpl
var InterfaceFindMultiCacheByField string

//go:embed tpl/find-multi-cache-by-field.tpl
var FindMultiCacheByField string

//go:embed tpl/interface-find-multi-by-fields.tpl
var InterfaceFindMultiByFields string

//go:embed tpl/find-multi-by-fields.tpl
var FindMultiByFields string

//go:embed tpl/interface-find-multi-cache-by-fields.tpl
var InterfaceFindMultiCacheByFields string

//go:embed tpl/find-multi-cache-by-fields.tpl
var FindMultiCacheByFields string

//go:embed tpl/interface-find-multi-by-condition.tpl
var InterfaceFindMultiByCondition string

//go:embed tpl/find-multi-by-condition.tpl
var FindMultiByCondition string

//go:embed tpl/interface-find-multi-by-cache-condition.tpl
var InterfaceFindMultiByCacheCondition string

//go:embed tpl/find-multi-by-cache-condition.tpl
var FindMultiByCacheCondition string
