// UpsertOneCacheByFields 根据fields字段Upsert一条数据, 并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpsertOneCacheByFields(ctx context.Context, data *{{.dbName}}_model.{{.upperTableName}},fields []string) error {
	if len(fields) == 0 {
        return errors.New("UpsertOneByFields fields is empty")
    }
	fieldNameToValue := make(map[string]interface{})
	typ := reflect.TypeOf(data).Elem()
	val := reflect.ValueOf(data).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			gormTags := strings.Split(gormTag, ";")
			for _, v := range gormTags {
				if strings.Contains(v, "column") {
					columnName := strings.TrimPrefix(v, "column:")
					fieldValue := val.Field(i).Interface()
					fieldNameToValue[columnName] = fieldValue
					break
				}
			}
		}
	}
	whereExpressions := make([]clause.Expression, 0)
	columns := make([]clause.Column, 0)
	for _, v := range fields {
		whereExpressions = append(whereExpressions,clause.And(clause.Eq{Column: v, Value: fieldNameToValue[v]}))
		columns = append(columns, clause.Column{Name: v})
	}
	oldData := &{{.dbName}}_model.{{.upperTableName}}{}
	err := {{.firstTableChar}}.db.Model(&{{.dbName}}_model.{{.upperTableName}}{}).Clauses(whereExpressions...).Unscoped().First(oldData).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	err = dao.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   columns,
		UpdateAll: true,
	}).Create(data)
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx,oldData,data)
	if err != nil {
		return err
	}
	return nil
}