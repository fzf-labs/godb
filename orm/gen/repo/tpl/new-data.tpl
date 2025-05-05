// NewData 实例化
func ({{.firstTableChar}} *{{.upperTableName}}Repo) NewData() *{{.dbName}}_model.{{.upperTableName}} {
    return &{{.dbName}}_model.{{.upperTableName}}{}
}