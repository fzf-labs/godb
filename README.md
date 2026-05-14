# godb

godb 是一个 Go 语言数据库工具集合，提供了对 MySQL、PostgreSQL 等数据库的便捷操作，以及代码生成等功能。

## ✨ 功能特性

- 🛠 基于 GORM 的数据库操作封装
- 🔌 支持 MySQL 和 PostgreSQL 数据库
- 📦 查询缓存支持
- 📁 代码生成器(支持生成 dao、model、repo 文件)
- 📁 数据表结构导出工具
- 📁 数据表生成proto文件工具
## 📦 安装

```bash
go get github.com/fzf-labs/godb
```

## 🚀 快速开始

### MySQL 示例

```go
package main

import (
    "github.com/fzf-labs/godb/orm/gormx"
)

func main() {
    db, err := gormx.NewGormClient(&gormx.ClientConfig{
        Driver:          gormx.MySQL,
        DataSourceName:  "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        MaxIdleConn:    10,
        MaxOpenConn:    100,
        ShowLog:        true,
    })
    if err != nil {
        panic(err)
    }
    _ = db
    // 使用 db 进行数据库操作...
}
```

### PostgreSQL 示例

```go
package main

import (
    "github.com/fzf-labs/godb/orm/gormx"
)

func main() {
    db, err := gormx.NewGormClient(&gormx.ClientConfig{
        Driver:          gormx.Postgres,
        DataSourceName:  "host=localhost port=5432 user=postgres password=123456 dbname=test sslmode=disable",
        MaxIdleConn:    10,
        MaxOpenConn:    100,
        ShowLog:        true,
    })
    if err != nil {
        panic(err)
    }
    _ = db
    // 使用 db 进行数据库操作...
}
```

## 📚 工具使用文档
### 工具安装
```
go install github.com/fzf-labs/godb/cmd/godb@latest
```
### 代码生成

```shell
godb ormgen --db $(DB_TYPE) --dsn $(DB_DSN) --tables $(DB_TABLES)
```

### 数据库导出

```shell
godb sqldump --db $(DB_TYPE) --dsn $(DB_DSN) --tables $(DB_TABLES) -f true
```

### 数据库表生成proto文件

```shell
godb sqltopb --db $(DB_TYPE) --dsn $(DB_DSN) -p '' -g '' -o ''

```

## 🤝 贡献

欢迎提交 issue 和 Pull Request。

## 📄 开源协议

本项目采用 [MIT 许可证](LICENSE)。
