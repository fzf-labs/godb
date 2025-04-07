// UpdateOneWithZeroByTx 更新一条数据(事务),包含零值，
// data 中主键字段必须有值,并且会更新所有字段,包括零值
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Select(dao.ALL.WithTable("")).Updates(newData)
	if err != nil {
		return err
	}
	return err
}