// CreateBatchCache 批量创建数据, 并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) CreateBatchCache(ctx context.Context, data []*{{.dbName}}_model.{{.upperTableName}}, batchSize int) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	err := dao.WithContext(ctx).CreateInBatches(data,batchSize)
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx, data...)
	if err != nil {
		return err
	}
	return nil
}