package proto

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils"
	"github.com/fzf-labs/godb/orm/utils/file"
	"github.com/fzf-labs/godb/orm/utils/template"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GenerationPB 生成
func GenerationPB(db *gorm.DB, outPutPath, packageStr, goPackageStr, table string, columnNameToName map[string]string, columnNameToDataType map[string]string) error {
	var f string
	p := &Proto{
		gorm:                 db,
		outPutPath:           outPutPath,
		packageStr:           packageStr,
		goPackageStr:         goPackageStr,
		tableName:            table,
		tableNameComment:     "",
		tableNameUnderScore:  strcase.ToSnake(table),
		lowerTableName:       "",
		upperTableName:       "",
		columnNameToName:     columnNameToName,
		columnNameToDataType: columnNameToDataType,
	}
	p.tableNameComment = p.getTableComment(table)
	p.lowerTableName = p.lowerName(table)
	p.upperTableName = p.upperName(table)
	f += p.genSyntax()
	f += p.genPackage()
	f += p.genImport()
	f += p.genOption()
	f += p.genService()
	f += p.genMessage()
	outputFile := p.outPutPath + "/" + table + ".proto"
	return p.output(outputFile, f)
}

type Proto struct {
	gorm                 *gorm.DB          // 数据库
	outPutPath           string            // 生成文件路径
	packageStr           string            // proto中的package名称
	goPackageStr         string            // proto中的goPackage名称
	tableName            string            // 表名称
	tableNameComment     string            // 表注释
	tableNameUnderScore  string            // 表下划线名称
	lowerTableName       string            // 表名称首字母小写
	upperTableName       string            // 表名称首字母大写
	columnNameToName     map[string]string // 字段名称对应的Go名称
	columnNameToDataType map[string]string // 字段名称对应的Go类型
}

func (p *Proto) output(filePath, content string) error {
	if file.Exists(filePath) {
		return errors.New(fmt.Sprintf("%s exist", filePath))
	}
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0775); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = dstFile.WriteString(content)
	if err != nil {
		return err
	}
	return err
}

func (p *Proto) getTableComment(table string) string {
	tableComments, err := gormx.GetTableComments(p.gorm)
	if err != nil {
		return ""
	}
	for k, v := range tableComments {
		if k == table {
			return v
		}
	}
	return ""
}

func (p *Proto) genSyntax() string {
	str, _ := template.NewTemplate().Parse(Syntax).Execute(map[string]any{})
	return fmt.Sprintln(str.String())
}

func (p *Proto) genPackage() string {
	str, _ := template.NewTemplate().Parse(Package).Execute(map[string]any{
		"packageStr": p.packageStr,
	})
	return fmt.Sprintln(str.String())
}

func (p *Proto) genImport() string {
	str, _ := template.NewTemplate().Parse(Import).Execute(map[string]any{})
	return fmt.Sprintln(str.String())
}

func (p *Proto) genOption() string {
	str, _ := template.NewTemplate().Parse(Option).Execute(map[string]any{
		"goPackageStr": p.goPackageStr,
	})
	return fmt.Sprintln(str.String())
}

func (p *Proto) genService() string {
	str, _ := template.NewTemplate().Parse(Service).Execute(map[string]any{
		"upperTableName":      p.upperTableName,
		"tableNameComment":    p.tableNameComment,
		"tableNameUnderScore": p.tableNameUnderScore,
	})
	return fmt.Sprintln(str.String())
}

func (p *Proto) genMessage() string {
	var info string
	var createReq string
	var createReply string
	var updateReq string
	var deleteReq string
	var getReq string
	columnTypes, err := p.gorm.Migrator().ColumnTypes(p.tableName)
	if err != nil {
		return ""
	}
	// 获取索引
	indexes, err := gormx.GetIndexes(p.gorm, p.tableName)
	if err != nil {
		return ""
	}
	var primaryKeyColumn string
	for _, index := range indexes {
		if index.Primary {
			primaryKeyColumn = index.ColumnName
			break
		}
	}
	columnTypeInfo := make(map[string]gorm.ColumnType)
	num := 0
	createNum := 0
	for _, v := range columnTypes {
		num++
		columnTypeInfo[v.Name()] = v
		pbType := dataTypeToPbType(p.columnNameToDataType[v.Name()])
		pbName := lowerFieldName(p.columnNameToName[v.Name()])
		comment, _ := v.Comment()
		nullable, _ := v.Nullable()
		length, _ := v.Length()
		validate := pbTypeToValidate(pbType, nullable, length)
		if utils.StrSliFind([]string{"deletedAt", "deleted_at", "deletedTime", "deleted_time"}, v.Name()) {
			continue
		}
		info += fmt.Sprintf("	%s %s = %d; // %s\n", pbType, pbName, num, comment)
		if utils.StrSliFind([]string{"createdAt", "created_at", "createdTime", "created_time", "updatedAt", "updated_at", "updatedTime", "updated_time"}, v.Name()) {
			continue
		}
		if v.Name() != primaryKeyColumn {
			createNum++
			createReq += fmt.Sprintf("	%s %s = %d %s; // %s\n", pbType, pbName, createNum, validate, comment)
		}
		updateReq += fmt.Sprintf("	%s %s = %d %s; // %s\n", pbType, pbName, num, validate, comment)
	}
	if primaryKeyColumn != "" {
		primaryKeyComment, _ := columnTypeInfo[primaryKeyColumn].Comment()
		pbType := dataTypeToPbType(p.columnNameToDataType[primaryKeyColumn])
		nullable, _ := columnTypeInfo[primaryKeyColumn].Nullable()
		length, _ := columnTypeInfo[primaryKeyColumn].Length()
		validate := pbTypeToValidate(pbType, nullable, length)
		pbName := lowerFieldName(p.columnNameToName[primaryKeyColumn])
		createReply = fmt.Sprintf("	%s %s = %d; // %s", pbType, pbName, 1, primaryKeyComment)
		getReq = fmt.Sprintf("	%s %s = %d %s; // %s\n", pbType, pbName, 1, validate, primaryKeyComment)
		deleteReq = fmt.Sprintf("repeated %s %s = %d %s; // %s\n", pbType, plural(pbName), 1, validate, primaryKeyComment+"集合")
	}
	info = strings.TrimSpace(strings.TrimRight(info, "\n"))
	createReq = strings.TrimSpace(strings.TrimRight(createReq, "\n"))
	updateReq = strings.TrimSpace(strings.TrimRight(updateReq, "\n"))
	deleteReq = strings.TrimSpace(strings.TrimRight(deleteReq, "\n"))
	getReq = strings.TrimSpace(strings.TrimRight(getReq, "\n"))
	str, _ := template.NewTemplate().Parse(Message).Execute(map[string]any{
		"tableNameComment": p.tableNameComment,
		"upperTableName":   p.upperTableName,
		"info":             info,
		"createReq":        createReq,
		"createReply":      createReply,
		"updateReq":        updateReq,
		"deleteReq":        deleteReq,
		"getReq":           getReq,
	})
	return fmt.Sprintln(str.String())
}

// upperName 大写
func (p *Proto) upperName(s string) string {
	return p.gorm.NamingStrategy.SchemaName(s)
}

// plural 复数形式
func plural(s string) string {
	str := inflection.Plural(s)
	if str == s {
		str += "plural"
	}
	return str
}

// lowerName 小写
func (p *Proto) lowerName(s string) string {
	str := p.upperName(s)
	if str == "" {
		return str
	}
	words := []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "ttl", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	// 如果第一个单词命中  则不处理
	for _, v := range words {
		if strings.HasPrefix(str, v) {
			return str
		}
	}
	rs := []rune(str)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		str = string(unicode.ToLower(f)) + string(rs[1:])
	}
	return str
}

// lowerFieldName 字段名称小写
func lowerFieldName(str string) string {
	words := []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "ttl", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	// 如果第一个单词命中  则不处理
	for _, v := range words {
		if strings.HasPrefix(str, v) {
			return str
		}
	}
	// 替换ID为Id
	str = strings.ReplaceAll(str, "ID", "Id")
	rs := []rune(str)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		str = string(unicode.ToLower(f)) + string(rs[1:])
	}
	if token.Lookup(str).IsKeyword() {
		str = "_" + str
	}
	return str
}

// dataTypeToPbType 根据数据库类型转换为proto类型
// go 语言类型转换为proto类型
func dataTypeToPbType(dataType string) string {
	var fieldType string
	switch dataType {
	case "int", "int8", "int16", "int32", "int64":
		fieldType = "int32" // 64位存在溢出问题
	case "uint", "uint8", "uint16", "uint32", "uint64":
		fieldType = "uint32" // 64位存在溢出问题
	case "float32":
		fieldType = "float"
	case "float64":
		fieldType = "double"
	case "bool":
		fieldType = "bool"
	case "string":
		fieldType = "string"
	case "time.Time":
		fieldType = "google.protobuf.Timestamp"
	case "[]byte":
		fieldType = "bytes"
	default:
		fieldType = "string"
	}
	return fieldType
}

// pbTypeToValidate 根据pb类型转换为validate类型
func pbTypeToValidate(pbType string, isNull bool, length int64) string {
	switch pbType {
	case "string":
		if isNull {
			if length <= 0 {
				return "[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).string={min_len: 1}]"
			}
			return fmt.Sprintf("[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).string={min_len: 1, max_len: %d}]", length)
		}
		if length <= 0 {
			return "[(buf.validate.field).string={min_len: 1}]"
		}
		return fmt.Sprintf("[(buf.validate.field).string={min_len: 1, max_len: %d}]", length)
	case "int32":
		if isNull {
			if length <= 0 {
				return "[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).int32={gt: 0}]"
			}
			return fmt.Sprintf("[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).int32={gt: 0, lte: %d}]", length)
		}
		if length <= 0 {
			return "[(buf.validate.field).int32={gt: 0}]"
		}
		return fmt.Sprintf("[(buf.validate.field).int32={gt: 0, lte: %d}]", length)
	case "int64":
		if isNull {
			if length <= 0 {
				return "[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).int64={gt: 0}]"
			}
			return fmt.Sprintf("[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).int64={gt: 0, lte: %d}]", length)
		}
		if length <= 0 {
			return "[(buf.validate.field).int64={gt: 0}]"
		}
		return fmt.Sprintf("[(buf.validate.field).int64={gt: 0, lte: %d}]", length)
	case "float":
		if isNull {
			if length <= 0 {
				return "[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).float={gt: 0}]"
			}
			return fmt.Sprintf("[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).float={gt: 0, lte: %d}]", length)
		}
		if length <= 0 {
			return "[(buf.validate.field).float={gt: 0}]"
		}
		return fmt.Sprintf("[(buf.validate.field).float={gt: 0, lte: %d}]", length)
	case "double":
		if isNull {
			if length <= 0 {
				return "[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).double={gt: 0}]"
			}
			return fmt.Sprintf("[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED,(buf.validate.field).double={gt: 0, lte: %d}]", length)
		}
		if length <= 0 {
			return "[(buf.validate.field).double={gt: 0}]"
		}
		return fmt.Sprintf("[(buf.validate.field).double={gt: 0, lte: %d}]", length)
	case "google.protobuf.Timestamp":
		if isNull {
			return "[(buf.validate.field).ignore=IGNORE_IF_UNPOPULATED]"
		}
		return "[(buf.validate.field).required=true]"
	default:
		return ""
	}
}
