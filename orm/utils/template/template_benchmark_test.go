package template

import "testing"

var benchTemplateString string

func BenchmarkDefaultTemplateExecute(b *testing.B) {
	tpl := `package {{.Package}}

type {{.Name}} struct {
{{- range .Fields }}
	{{ .Name }} {{ .Type }} ` + "`json:\"{{ .JSON }}\"`" + `
{{- end }}
}
`
	data := map[string]any{
		"Package": "model",
		"Name":    "UserDemo",
		"Fields": []map[string]string{
			{"Name": "ID", "Type": "int64", "JSON": "id"},
			{"Name": "UserName", "Type": "string", "JSON": "userName"},
			{"Name": "CreatedAt", "Type": "time.Time", "JSON": "createdAt"},
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf, err := NewTemplate().Parse(tpl).Execute(data)
		if err != nil {
			b.Fatal(err)
		}
		benchTemplateString = buf.String()
	}
}
