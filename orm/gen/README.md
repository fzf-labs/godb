# orm/gen

## 简述
`orm/gen`是一个通过数据库表结构生成Go代码的工具。目前有两种工具，`GORM代码生成`和 `proto代码`生成。
- `GORM代码生成`:根据表结构自动生成`model`,`dao`,`repo`代码并内置查询缓存功能，简化数据库操作与缓存的开发工作。
- `proto代码生成`:根据表结构生成包含基本的 CURD方法的 pb 文件，提升编写 pb 文件的效率。

## 功能特点
- 按gorm/gen的方式自动生成代码
- 可自定义数据表命名规则和字段映射规则
- 根据表索引构建CURD方法
- 根据表中索引生成查询缓存方法
- 根据表结构生成proto文件

## 使用方法
```go
    db, err := orm.NewGormPostgresClient(&orm.GormPostgresClientConfig{
      DataSourceName:  "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai",
      MaxIdleConn:     0,
      MaxOpenConn:     0,
      ConnMaxLifeTime: 0,
      ShowLog:         false,
      Tracing:         false,
    })
    if err != nil {
        return
    }
    // 生成代码
    NewGenerationDB(
      db,
      "./example/postgres/",
      WithOutRepo(),
      WithDBNameOpts(DBNameOpts()),
      WithTables([]string{"admin_demo", "admin_log_demo", "admin_role_demo"}),
      WithDataMap(DataTypeMap()), // 设置数据类型映射
      WithDBOpts(ModelOptionRemoveDefault(), ModelOptionPgDefaultString(), ModelOptionRemoveGormTypeTag(), ModelOptionUnderline("UL")), // 设置自定义选项
      WithFieldNullable(),
    ).Do()
```

## 业界方案
### go-zero：
- 地址：[https://go-zero.dev/docs/tutorials/mysql/cache](https://go-zero.dev/docs/tutorials/mysql/cache)
- 实现方式：框架自行实现了ORM和Cache库，可通过数据库表结构生成数据库CURD代码和缓存代码。SQL是手动拼装的模式编写。缓存策略是先更新后删除缓存。

### go-frame：
- 地址：[https://goframe.org/pages/viewpage.action?pageId=1114346](https://goframe.org/pages/viewpage.action?pageId=1114346)
- 实现方式：框架自行实现了ORM和Cache库，仅支持链式操作，在事务操作下不可用。在写SQL的时候会使用一个缓存方法，把当前查询的数据缓存进去。

### gorm-cache：
- 地址：[https://github.com/go-gorm/caches](https://github.com/go-gorm/caches)
- 实现方式：官方gorm的插件，抽象了一个缓存接口，目前该插件不保证缓存和数据库之间的一致性。

### sponge：
- 地址：[https://github.com/zhufuyi/sponge/blob/main/assets/readme-cn.md](https://github.com/zhufuyi/sponge/blob/main/assets/readme-cn.md)
- 实现方式：基于gorm，使用主键ID做缓存。

## 常见问题

* 是否支持其他数据库类型？ 
  - gorm本身支持多种数据库，在orm-gen中理论上同样支持gorm所支持的数据库,但是目前只适配了PostgreSQL，MySQL。

* 什么样的索引生成查询缓存？ 
  - 所有的索引都会生成查询缓存,并且联合索引会按照最左匹配原则拆分索引，生成多个查询缓存方法.

* 为什么查询缓存中的Cache是一个接口类型？ 
  - 将缓存抽象为接口,是为了与具体的缓存实现解耦,您可以自行更换缓存实现,例:Redis,Memcached,本地缓存等.
