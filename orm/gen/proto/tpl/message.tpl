//{{.tableNameComment}}信息
message {{.upperTableName}}Info {
  {{.info}}
}

//请求-{{.tableNameComment}}-创建一条数据
message Create{{.upperTableName}}Req {
  {{.createReq}}
}

//响应-{{.tableNameComment}}-创建一条数据
message Create{{.upperTableName}}Reply {
  {{.createReply}}
}

//请求-{{.tableNameComment}}-更新一条数据
message Update{{.upperTableName}}Req {
  {{.updateReq}}
}

//响应-{{.tableNameComment}}-更新一条数据
message Update{{.upperTableName}}Reply {}

{{- if .status }}

//请求-{{.tableNameComment}}-更新状态
message Update{{.upperTableName}}StatusReq {
  {{.updateStatusReq}}
}

//响应-{{.tableNameComment}}-更新状态
message Update{{.upperTableName}}StatusReply {}
{{- end }}

//请求-{{.tableNameComment}}-删除多条数据
message Delete{{.upperTableName}}Req {
  {{.deleteReq}}
}

//响应-{{.tableNameComment}}-删除多条数据
message Delete{{.upperTableName}}Reply {}

//请求-{{.tableNameComment}}-单条数据查询
message Get{{.upperTableName}}InfoReq {
  {{.getReq}}
}

//响应-{{.tableNameComment}}-单条数据查询
message Get{{.upperTableName}}InfoReply {
  {{.upperTableName}}Info info = 1;
}

//请求-{{.tableNameComment}}-列表数据查询
message Get{{.upperTableName}}ListReq {
  int32 page = 1; //页码
  int32 pageSize = 2; //页数
}

//响应-{{.tableNameComment}}-列表数据查询
message Get{{.upperTableName}}ListReply {
  int32 total = 1; //总数
  repeated {{.upperTableName}}Info list = 2; // 列表数据
}
