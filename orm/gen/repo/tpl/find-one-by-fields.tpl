// FindOneBy{{.upperFields}} 根据{{.upperFields}}查询一条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindOneBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) (*{{.dbName}}_model.{{.upperTableName}}, error) {
    dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
    result, err := dao.WithContext(ctx).Where({{.whereFields}}).First()
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
	return result, nil
}
{{- if .haveDeletedAt }}
// FindOneUnscopedBy{{.upperFields}} 根据{{.upperFields}}查询一条数据（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindOneUnscopedBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) (*{{.dbName}}_model.{{.upperTableName}}, error) {
    dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
    result, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).First()
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
	return result, nil
}
{{- end }}