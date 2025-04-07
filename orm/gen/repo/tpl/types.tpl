type (
	I{{.upperTableName}}Repo interface{
		{{.methods}}
	}
	{{.upperTableName}}Repo struct {
		db       *gorm.DB
		cache    dbcache.IDBCache
		encoding encoding.API
	}
)
