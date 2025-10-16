// FindMultiBy{{.upperField}} 根据{{.lowerField}}查询多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where({{.whereFields}}).Find()
	if err != nil {
		return nil, err
	}
	return result, nil
}
{{- if .haveDeletedAt }}
// FindMultiUnscopedBy{{.upperField}} 根据{{.lowerField}}查询多条数据（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiUnscopedBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	result, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Find()
	if err != nil {
		return nil, err
	}
	return result, nil
}
{{- end }}