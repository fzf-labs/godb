# Batch Update 工具包

这是一个用于 MySQL 批量更新操作的高效工具包。该工具包主要提供了将结构体切片转换为批量更新 SQL 语句的功能，支持大数据量分批处理。

## 功能特性

- 支持结构体切片到 SQL 的自动转换
- 自动处理字段类型转换和 SQL 注入防护
- 支持大数据量分批处理（默认每批 200 条）
- 支持多种数据类型的 ID 字段（整数、字符串等）
- 自动处理字段值的转义和格式化

## SQL 示例

MySQL 批量更新语句示例：

```sql
UPDATE users SET 
    age = CASE id 
        WHEN 1 THEN 25
        WHEN 2 THEN 30
        WHEN 3 THEN 35
    END,
    status = CASE id
        WHEN 1 THEN 1
        WHEN 2 THEN 0
        WHEN 3 THEN 1
    END
WHERE id IN (1,2,3);
```

PostgreSQL 批量更新语句示例：
```sql
UPDATE "users" SET 
    "age" = CASE "id" 
        WHEN 1 THEN 25
        WHEN 2 THEN 30
        WHEN 3 THEN 35
    END,
    "status" = CASE "id"
        WHEN 1 THEN true  -- PostgreSQL使用true/false而不是1/0表示布尔值
        WHEN 2 THEN false
        WHEN 3 THEN true
    END
WHERE "id" IN (1,2,3);
```