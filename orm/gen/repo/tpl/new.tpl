func New{{.upperTableName}}Repo(cfg *config.Repo) *{{.upperTableName}}Repo {
	return &{{.upperTableName}}Repo{
		db:       cfg.DB,
		cache:    cfg.Cache,
		encoding: cfg.Encoding,
	}
}