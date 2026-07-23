package repo

import (
	"fmt"

	"github.com/fzf-labs/godb/orm/utils/template"
)

// generatePkg
func (r *Repo) generatePkg() (string, error) {
	tplParams := map[string]any{
		"dbName": r.dbName,
	}
	tpl, err := template.NewTemplate().Parse(Pkg).Execute(tplParams)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// generateImport
func (r *Repo) generateImport() (string, error) {
	tplParams := map[string]any{
		"daoPkgPath":   r.daoPkgPath,
		"modelPkgPath": r.modelPkgPath,
	}
	tpl, err := template.NewTemplate().Parse(Import).Execute(tplParams)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// generateVar
func (r *Repo) generateVar() (string, error) {
	var varStr string
	var cacheKeys string
	varCacheGlobalTpl, err := template.NewTemplate().Parse(VarCacheGlobal).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"haveDeletedAt":  r.haveDeletedAt,
	})
	if err != nil {
		return "", err
	}
	cacheKeys += varCacheGlobalTpl.String()
	// 生成缓存key
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		varCacheTpl, err := template.NewTemplate().Parse(VarCache).Execute(map[string]any{
			"dbName":         r.dbName,
			"upperTableName": r.upperTableName,
			"cacheFields":    r.cacheFields(v.Columns),
			"haveDeletedAt":  r.haveDeletedAt,
		})
		if err != nil {
			return "", err
		}
		cacheKeys += varCacheTpl.String()
	}
	// 生成变量
	varTpl, err := template.NewTemplate().Parse(Var).Execute(map[string]any{
		"upperTableName": r.upperTableName,
	})
	if err != nil {
		return "", err
	}
	varStr += fmt.Sprintln(varTpl.String())
	if len(cacheKeys) > 0 {
		varCacheKeysTpl, err := template.NewTemplate().Parse(VarCacheKeys).Execute(map[string]any{
			"cacheKeys": cacheKeys,
		})
		if err != nil {
			return "", err
		}
		varStr += fmt.Sprintln(varCacheKeysTpl.String())
	}
	return varStr, nil
}

// generateTypes
func (r *Repo) generateTypes() (string, error) {
	var methods string
	commonMethods, err := r.generateCommonMethods()
	if err != nil {
		return "", err
	}
	createMethods, err := r.generateCreateMethods()
	if err != nil {
		return "", err
	}
	updateMethods, err := r.generateUpdateMethods()
	if err != nil {
		return "", err
	}
	readMethods, err := r.generateReadMethods()
	if err != nil {
		return "", err
	}
	delMethods, err := r.generateDelMethods()
	if err != nil {
		return "", err
	}
	methods += commonMethods
	methods += createMethods
	methods += updateMethods
	methods += readMethods
	methods += delMethods
	typesTpl, err := template.NewTemplate().Parse(Types).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
		"methods":        methods,
	})
	if err != nil {
		return "", err
	}
	return typesTpl.String(), nil
}

// generateNew
func (r *Repo) generateNew() (string, error) {
	newTpl, err := template.NewTemplate().Parse(New).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	})
	if err != nil {
		return "", err
	}
	return newTpl.String(), nil
}

// generateCommonFunc
func (r *Repo) generateCommonFunc() (string, error) {
	var commonFunc string
	tplParams := map[string]any{
		"firstTableChar": r.firstTableChar,
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	}
	newData, err := template.NewTemplate().Parse(NewData).Execute(tplParams)
	if err != nil {
		return "", err
	}
	commonFunc += fmt.Sprintln(newData.String())
	deepCopy, err := template.NewTemplate().Parse(DeepCopy).Execute(tplParams)
	if err != nil {
		return "", err
	}
	commonFunc += fmt.Sprintln(deepCopy.String())
	return commonFunc, nil
}
