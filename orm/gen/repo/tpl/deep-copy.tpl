// DeepCopy 深拷贝
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeepCopy(data *{{.dbName}}_model.{{.upperTableName}}) *{{.dbName}}_model.{{.upperTableName}} {
    newData := new({{.dbName}}_model.{{.upperTableName}})
    *newData = *data
    return newData
}