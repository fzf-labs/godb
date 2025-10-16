// DeleteMultiBy{{.upperFields}}Tx 根据{{.lowerField}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where({{.whereFields}}).Delete()
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedBy{{.upperFields}}Tx 根据{{.lowerField}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiUnscopedBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Delete()
	if err != nil {
		return err
	}
	return nil
}
{{- end }}