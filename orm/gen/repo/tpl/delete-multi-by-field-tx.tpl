// DeleteMultiBy{{.upperField}}Tx 根据{{.lowerField}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.Eq({{.lowerField}})).Delete()
	if err != nil {
		return err
	}
	return nil
}
