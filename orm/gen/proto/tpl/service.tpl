//{{.tableNameComment}}
service {{.upperTableName}} {
  //{{.tableNameComment}}-创建一条数据
  rpc Create{{.upperTableName}}(Create{{.upperTableName}}Req) returns (Create{{.upperTableName}}Reply) {
    option (google.api.http) = {
      post: "/{{ .urlPrefix }}/{{.tableNameUnderScore}}/create"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      parameters: {
        headers: {
          name: "Authorization"
          description: "Bearer Token"
          type: STRING
          required: true
        }
      }
    };
  }
  //{{.tableNameComment}}-更新一条数据
  rpc Update{{.upperTableName}}(Update{{.upperTableName}}Req) returns (Update{{.upperTableName}}Reply) {
    option (google.api.http) = {
      post: "/{{ .urlPrefix }}/{{.tableNameUnderScore}}/update"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      parameters: {
        headers: {
          name: "Authorization"
          description: "Bearer Token"
          type: STRING
          required: true
        }
      }
    };
  }
  {{- if .status }}
  //{{.tableNameComment}}-更新状态
  rpc Update{{.upperTableName}}Status(Update{{.upperTableName}}StatusReq) returns (Update{{.upperTableName}}StatusReply) {
    option (google.api.http) = {
      post: "/{{ .urlPrefix }}/{{.tableNameUnderScore}}/update/status"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      parameters: {
        headers: {
          name: "Authorization"
          description: "Bearer Token"
          type: STRING
          required: true
        }
      }
    };
  }
	{{- end }}
  //{{.tableNameComment}}-删除多条数据
  rpc Delete{{.upperTableName}}(Delete{{.upperTableName}}Req) returns (Delete{{.upperTableName}}Reply) {
    option (google.api.http) = {
      post: "/{{ .urlPrefix }}/{{.tableNameUnderScore}}/delete"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      parameters: {
        headers: {
          name: "Authorization"
          description: "Bearer Token"
          type: STRING
          required: true
        }
      }
    };
  }
  //{{.tableNameComment}}-单条数据查询
  rpc Get{{.upperTableName}}Info(Get{{.upperTableName}}InfoReq) returns (Get{{.upperTableName}}InfoReply) {
    option (google.api.http) = {get: "/{{ .urlPrefix }}/{{.tableNameUnderScore}}/info"};
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      parameters: {
        headers: {
          name: "Authorization"
          description: "Bearer Token"
          type: STRING
          required: true
        }
      }
    };
  }
  //{{.tableNameComment}}-列表数据查询
  rpc Get{{.upperTableName}}List(Get{{.upperTableName}}ListReq) returns (Get{{.upperTableName}}ListReply) {
    option (google.api.http) = {get: "/{{ .urlPrefix }}/{{.tableNameUnderScore}}/list"};
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      parameters: {
        headers: {
          name: "Authorization"
          description: "Bearer Token"
          type: STRING
          required: true
        }
      }
    };
  }
}