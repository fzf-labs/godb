	Cache{{.upperTableName}}ByConditionPrefix = "DBCache:{{.dbName}}:{{.upperTableName}}ByCondition"
	{{- if .haveDeletedAt }}
	Cache{{.upperTableName}}UnscopedByConditionPrefix = "DBCache:{{.dbName}}:{{.upperTableName}}ByCondition"
	{{- end }}
