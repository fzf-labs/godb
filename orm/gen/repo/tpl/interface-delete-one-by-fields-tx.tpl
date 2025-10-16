// DeleteOneBy{{.upperFields}}Tx 根据{{.upperFields}}删除一条数据(事务)
DeleteOneBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error
{{- if .haveDeletedAt }}
// DeleteOneUnscopedBy{{.upperFields}}Tx 根据{{.upperFields}}删除一条数据(事务)
DeleteOneUnscopedBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error
{{- end }}