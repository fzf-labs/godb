    KeyMap[{{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.cacheFields}}Prefix, {{.delCacheFieldsParam}})] = struct{}{}
    {{- if .haveDeletedAt }}
    KeyMap[{{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}UnscopedBy{{.cacheFields}}Prefix, {{.delCacheFieldsParam}})] = struct{}{}
    {{- end }}