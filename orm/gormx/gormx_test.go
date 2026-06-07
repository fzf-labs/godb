package gormx

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/fzf-labs/godb/internal/testenv"
)

func TestNewGormPostgresClient(t *testing.T) {
	config := ClientConfig{
		Driver:          "postgres",
		DataSourceName:  testenv.PostgresDSN("user"),
		MaxIdleConn:     0,
		MaxOpenConn:     0,
		ConnMaxLifeTime: 0,
		ShowLog:         false,
		Tracing:         false,
	}
	_, err := NewGormClient(&config)
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	assert.NoError(t, err)
}

func TestNewGormClientRejectsNilAndUnknownDriver(t *testing.T) {
	db, err := NewGormClient(nil)
	assert.Nil(t, db)
	assert.Error(t, err)

	db, err = NewGormClient(&ClientConfig{Driver: "sqlite"})
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestNewDirectGormClientRejectsUnknownDriver(t *testing.T) {
	db, err := newDirectGormClient("sqlite", ":memory:", 0)
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestNewDirectGormClientKnownDriversRejectBadDSN(t *testing.T) {
	db, err := newDirectGormClient(MySQL, "%", logger.Silent)
	assert.Nil(t, db)
	assert.Error(t, err)

	db, err = newDirectGormClient(Postgres, "bad dsn", logger.Silent)
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestNewDirectGormClientPostgresSuccess(t *testing.T) {
	db, err := newDirectGormClient(Postgres, testenv.PostgresDSN("gorm_gen"), logger.Silent)
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	assert.NotNil(t, db)
}

func TestDirectClientWrappersRejectBadDSN(t *testing.T) {
	db, err := NewDebugGormClient(MySQL, "%")
	assert.Nil(t, db)
	assert.Error(t, err)

	db, err = NewSimpleGormClient(Postgres, "bad dsn")
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestDriverSpecificClientsRejectNilConfig(t *testing.T) {
	db, err := NewMySQLGormClient(nil)
	assert.Nil(t, db)
	assert.Error(t, err)

	db, err = NewPostgresGormClient(nil)
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestNewGormClientMySQLWithMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	restore := replaceSQLOpen(t, func(driverName, dsn string) (*sql.DB, error) {
		assert.Equal(t, "mysql", driverName)
		assert.Equal(t, "dsn", dsn)
		return sqlDB, nil
	})
	defer restore()

	mock.ExpectQuery("SELECT VERSION\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.0"))

	db, err := NewGormClient(&ClientConfig{
		Driver:         MySQL,
		DataSourceName: "dsn",
		MaxIdleConn:    1,
		MaxOpenConn:    2,
		ShowLog:        true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, MySQL, db.Dialector.Name())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewMySQLGormClientReturnsGormOpenError(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	restore := replaceSQLOpen(t, func(driverName, dsn string) (*sql.DB, error) {
		assert.Equal(t, "mysql", driverName)
		assert.Equal(t, "dsn", dsn)
		return sqlDB, nil
	})
	defer restore()

	db, err := NewMySQLGormClient(&ClientConfig{DataSourceName: "dsn"})
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestNewGormClientPostgresWithMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	restore := replaceSQLOpen(t, func(driverName, dsn string) (*sql.DB, error) {
		assert.Equal(t, "pgx", driverName)
		assert.Equal(t, "dsn", dsn)
		return sqlDB, nil
	})
	defer restore()

	db, err := NewGormClient(&ClientConfig{
		Driver:         Postgres,
		DataSourceName: "dsn",
		MaxIdleConn:    1,
		MaxOpenConn:    2,
		ShowLog:        true,
		Tracing:        true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, Postgres, db.Dialector.Name())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnableTracingClosesSQLDBWhenTracingFails(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	restoreTracing := replaceTracingPlugin(t, func(*gorm.DB) error {
		return context.Canceled
	})
	defer restoreTracing()

	closed := 0
	restoreClose := replaceCloseSQLDB(t, func(*sql.DB) error {
		closed++
		return nil
	})
	defer restoreClose()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = enableTracing(db, sqlDB)
	assert.ErrorIs(t, err, context.Canceled)
	assert.ErrorContains(t, err, "failed to enable tracing")
	assert.Equal(t, 1, closed)
}

func TestDriverSpecificClientsReturnSQLOpenErrors(t *testing.T) {
	restore := replaceSQLOpen(t, func(string, string) (*sql.DB, error) {
		return nil, context.Canceled
	})
	defer restore()

	db, err := NewMySQLGormClient(&ClientConfig{DataSourceName: "dsn"})
	assert.Nil(t, db)
	assert.ErrorIs(t, err, context.Canceled)

	db, err = NewPostgresGormClient(&ClientConfig{DataSourceName: "dsn"})
	assert.Nil(t, db)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRunHealthCheckQuery_ExecutesSQL(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = runHealthCheckQuery(db, "select * from definitely_missing_table")
	assert.Error(t, err)
}

func TestGetHealthStatus_Healthy(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, health, GetHealthStatus(db))
}

func TestGetHealthStatus_UnhealthyWhenDBUnavailable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, sqlDB.Close())
	assert.Equal(t, unhealthy, GetHealthStatus(db))
}

func TestGetHealthStatusAndStateInvalidDB(t *testing.T) {
	db := &gorm.DB{Config: &gorm.Config{}}
	assert.Equal(t, unhealthy, GetHealthStatus(db))
	assert.Nil(t, GetState(db))
}

func TestGetHealthStatusAndStateNilDB(t *testing.T) {
	assert.Equal(t, unhealthy, GetHealthStatus(nil))
	assert.Nil(t, GetState(nil))
}

func TestGetState(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if state := GetState(db); state == nil {
		t.Fatal("expected state")
	}
}

type namedDialector struct {
	gorm.Dialector
	name string
}

// Name 返回测试包装后的方言名称。
func (d namedDialector) Name() string {
	return d.name
}

func openNamedSQLite(t *testing.T, name string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(namedDialector{Dialector: sqlite.Open(":memory:"), name: name}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func replaceSQLOpen(t *testing.T, fn func(string, string) (*sql.DB, error)) func() {
	t.Helper()
	old := sqlOpen
	sqlOpen = fn
	return func() { sqlOpen = old }
}

func replaceTracingPlugin(t *testing.T, fn func(*gorm.DB) error) func() {
	t.Helper()
	old := useTracingPlugin
	useTracingPlugin = fn
	return func() { useTracingPlugin = old }
}

func replaceCloseSQLDB(t *testing.T, fn func(*sql.DB) error) func() {
	t.Helper()
	old := closeSQLDB
	closeSQLDB = fn
	return func() { closeSQLDB = old }
}

func openMockMySQL(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func openMockPostgres(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func expectMySQLCurrentDatabase(mock sqlmock.Sqlmock, dbName string) {
	query := regexp.QuoteMeta("SELECT SCHEMA_NAME from Information_schema.SCHEMATA where SCHEMA_NAME LIKE ? ORDER BY SCHEMA_NAME=? DESC,SCHEMA_NAME limit 1")
	mock.ExpectQuery(query).
		WithArgs("%", "").
		WillReturnRows(sqlmock.NewRows([]string{"SCHEMA_NAME"}).AddRow(dbName))
}
