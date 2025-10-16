	Cache{{.upperTableName}}By{{.cacheFields}}Prefix = "DBCache:{{.dbName}}:{{.upperTableName}}By{{.cacheFields}}"
	{{- if .haveDeletedAt }}
	Cache{{.upperTableName}}UnscopedBy{{.cacheFields}}Prefix = "DBCache:{{.dbName}}:{{.upperTableName}}UnscopedBy{{.cacheFields}}"
	{{- end }}
