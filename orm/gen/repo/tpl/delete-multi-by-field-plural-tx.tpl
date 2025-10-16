// DeleteMultiBy{{.upperFieldPlural}}Tx 根据{{.lowerFieldPlural}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Delete()
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedBy{{.upperFieldPlural}}Tx 根据{{.lowerFieldPlural}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiUnscopedBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Delete()
	if err != nil {
		return err
	}
	return nil
}
{{- end }}
