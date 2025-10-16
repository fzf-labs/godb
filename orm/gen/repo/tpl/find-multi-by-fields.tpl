// FindMultiBy{{.upperFields}} 根据{{.upperFields}}查询多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where({{.whereFields}}).Find()
    if err != nil {
        return nil, err
    }
	return result, nil
}
{{- if .haveDeletedAt }}
// FindMultiUnscopedBy{{.upperFields}} 根据{{.upperFields}}查询多条数据（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiUnscopedBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	result, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Find()
    if err != nil {
        return nil, err
    }
	return result, nil
}
{{- end }}
