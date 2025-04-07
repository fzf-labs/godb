# 数据库备份

```sh
godb sqldump -d postgres -s "host=localhost user=postgres password=123456 dbname=godb port=5432 sslmode=disable TimeZone=Asia/Shanghai" -o "./orm/example/sql"
```

# 数据库恢复

```sh
godb sqltopb -d postgres -s "host=localhost user=postgres password=123456 dbname=godb port=5432 sslmode=disable TimeZone=Asia/Shanghai" -o "./orm/example/pb"
```

# 数据库代码生成器

```sh
godb ormgen -d postgres -s "host=localhost user=postgres password=123456 dbname=godb port=5432 sslmode=disable TimeZone=Asia/Shanghai" -o "./orm/example/gorm"
```
