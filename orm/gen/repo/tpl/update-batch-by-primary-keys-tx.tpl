// UpdateBatchBy{{.upperSinglePrimaryKeyPlural}}Tx 根据主键{{.upperSinglePrimaryKeyPlural}}批量更新(事务)
// 零值会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchBy{{.upperSinglePrimaryKeyPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerSinglePrimaryKeyPlural}} []{{.dataTypeSinglePrimaryKey}}, data map[string]interface{}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where(dao.{{.upperSinglePrimaryKey}}.In({{.lowerSinglePrimaryKeyPlural}}...)).Updates(data)
	if err != nil {
		return err
	}
	return nil
}