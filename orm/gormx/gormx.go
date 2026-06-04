package gormx

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/opentelemetry/tracing"
)

const (
	unhealthy = "unhealthy"
	health    = "health"
)

const (
	// MySQL 表示 MySQL 数据库驱动名称。
	MySQL = "mysql"
	// Postgres 表示 PostgreSQL 数据库驱动名称。
	Postgres = "postgres"
)

var sqlOpen = sql.Open

// ClientConfig 配置
type ClientConfig struct {
	Driver          string        `json:"driver"`          // 数据库类型 mysql/postgres
	DataSourceName  string        `json:"dataSourceName"`  // 数据源名称
	MaxIdleConn     int           `json:"maxIdleConn"`     // 最大空闲连接数 默认10
	MaxOpenConn     int           `json:"maxOpenConn"`     // 最大打开连接数 默认100
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime"` // 连接最大空闲时间 默认10分钟
	ConnMaxLifeTime time.Duration `json:"connMaxLifeTime"` // 连接最大生命周期 默认1小时
	ShowLog         bool          `json:"showLog"`         // 是否显示日志 默认false
	Tracing         bool          `json:"tracing"`         // 是否开启链路追踪 默认false
}

// NewDebugGormClient 创建调试模式的gorm客户端
func NewDebugGormClient(driver, dsn string) (*gorm.DB, error) {
	return newDirectGormClient(driver, dsn, logger.Info)
}

// NewSimpleGormClient 创建数据库连接
func NewSimpleGormClient(driver, dsn string) (*gorm.DB, error) {
	return newDirectGormClient(driver, dsn, logger.Silent)
}

// newDirectGormClient 按驱动和日志级别创建直连 gorm 客户端。
func newDirectGormClient(driver, dsn string, logLevel logger.LogLevel) (*gorm.DB, error) {
	switch driver {
	case MySQL:
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to open mysql connection")
		}
		return db, nil
	case Postgres:
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to open postgres connection")
		}
		return db, nil
	default:
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}
}

// NewGormClient 初始化gorm客户端
// mysql: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
// postgres: "host=localhost user=postgres password=123456 dbname=godb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
func NewGormClient(cfg *ClientConfig) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("client config cannot be nil")
	}
	switch cfg.Driver {
	case MySQL:
		return NewMySQLGormClient(cfg)
	case Postgres:
		return NewPostgresGormClient(cfg)
	default:
		return nil, fmt.Errorf("unknown database driver: %s", cfg.Driver)
	}
}

// NewMySQLGormClient 创建 MySQL gorm 客户端。
func NewMySQLGormClient(cfg *ClientConfig) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("client config cannot be nil")
	}
	sqlDB, err := sqlOpen("mysql", cfg.DataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open mysql connection")
	}
	// 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	// 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	// 设置连接可以重复使用的最长时间.
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifeTime)
	// 设置连接可以重复使用的最长时间.
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}
	if cfg.ShowLog {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB}), &gormConfig)
	if err != nil {
		_ = sqlDB.Close()
		return nil, errors.Wrap(err, "failed to open mysql connection")
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	if cfg.Tracing {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			_ = sqlDB.Close()
			return nil, errors.Wrap(err, "failed to enable tracing")
		}
	}
	return db, nil
}

// NewPostgresGormClient 创建 PostgreSQL gorm 客户端。
func NewPostgresGormClient(cfg *ClientConfig) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("client config cannot be nil")
	}
	sqlDB, err := sqlOpen("pgx", cfg.DataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open postgres connection")
	}
	// 用于设置最大打开的连接数。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	// 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	// 设置连接可以重复使用的最长时间.
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifeTime)
	// 设置连接可以重复使用的最长时间.
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}
	if cfg.ShowLog {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gormConfig)
	if err != nil {
		_ = sqlDB.Close()
		return nil, errors.Wrap(err, "failed to open postgres connection")
	}
	if cfg.Tracing {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			_ = sqlDB.Close()
			return nil, errors.Wrap(err, "failed to enable tracing")
		}
	}
	return db, nil
}

// GetHealthStatus 检查链接是否健康
func GetHealthStatus(gormDB *gorm.DB) string {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return unhealthy
	}
	// 验证与数据库的连接是否仍然存在
	err = sqlDB.Ping()
	if err != nil {
		return unhealthy
	}
	err = runHealthCheckQuery(gormDB, `select 1`)
	if err != nil {
		return unhealthy
	}
	return health
}

func runHealthCheckQuery(gormDB *gorm.DB, query string) error {
	var probe int
	return gormDB.Raw(query).Scan(&probe).Error
}

// GetState 获取目前数据库状态参数
func GetState(gormDB *gorm.DB) *sql.DBStats {
	db, err := gormDB.DB()
	if err != nil {
		return nil
	}
	state := db.Stats()
	return &state
}
