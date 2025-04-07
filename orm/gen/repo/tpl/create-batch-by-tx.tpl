// CreateBatchByTx 批量创建数据(事务)
func ({{.firstTableChar}} *{{.upperTableName}}Repo) CreateBatchByTx(ctx context.Context,tx *{{.dbName}}_dao.Query, data []*{{.dbName}}_model.{{.upperTableName}}, batchSize int) error {
	dao := tx.{{.upperTableName}}
	err := dao.WithContext(ctx).CreateInBatches(data,batchSize)
	if err != nil {
		return err
	}
	return nil
}