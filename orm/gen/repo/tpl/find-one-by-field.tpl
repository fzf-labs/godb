// FindOneBy{{.upperField}} 根据{{.lowerField}}查询一条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindOneBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error) {
    dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
    result, err := dao.WithContext(ctx).Where({{.whereFields}}).First()
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
	return result, nil
}
{{- if .haveDeletedAt }}
// FindOneUnscopedBy{{.upperField}} 根据{{.lowerField}}查询一条数据（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindOneUnscopedBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error) {
    dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
    result, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).First()
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
	return result, nil
}
{{- end }}