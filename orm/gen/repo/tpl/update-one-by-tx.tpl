// UpdateOneByTx 更新一条数据(事务)
// data 中主键字段必须有值，零值不会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Updates(newData)
	if err != nil {
		return err
	}
	return err
}