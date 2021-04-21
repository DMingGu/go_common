package lib

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

func InitDBPool(path string) (err error) {
	//普通的db方式
	DbConfMap := &MysqlMapConf{}
	err = ParseConfig(path, DbConfMap)
	if err != nil {
		return
	}
	if len(DbConfMap.List) == 0 {
		Log.TagInfo(NewTrace(),DLTagMySqlEmpty, map[string]interface{}{})
	}
	GORMMapPool = map[string]*gorm.DB{}
	newLogger := initGormLog()
	for confName, DbConf := range DbConfMap.List {
		//gorm连接方式
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN:  DbConf.DataSourceName, // DSN data source name
			DefaultStringSize: 256, // string 类型字段的默认长度
			DisableDatetimePrecision: true, // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
			DontSupportRenameIndex: true, // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
			DontSupportRenameColumn: true, // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
			SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
		}), &gorm.Config{
			Logger: newLogger,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: DbConf.TablePrefix,   // 表名前缀，`User`表为`t_users`
				SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
			},
		})
		if err != nil {
			return err
		}
		sqlDB, err := db.DB()
		if err!=nil {
			return err
		}
		err = sqlDB.Ping()
		if err != nil {
			return err
		}
		// SetMaxIdleConns 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(DbConf.MaxIdleConn)
		// SetMaxOpenConns 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(DbConf.MaxOpenConn)
		// SetConnMaxLifetime 设置了连接可复用的最大时间。
		sqlDB.SetConnMaxLifetime(time.Duration(DbConf.MaxConnLifeTime) * time.Second)

		GORMMapPool[confName] = db
	}

	return
}





func GetGormPool(name string) (*gorm.DB, error) {
	if db, ok := GORMMapPool[name]; ok {
		return db, nil
	}
	return nil, errors.New("get db error")
}
//关闭数据库链接
func CloseDB() error {
	for _, dbpool := range GORMMapPool {
		sqlDb,err:=dbpool.DB()
		if err!=nil {
			sqlDb.Close()
		}
	}
	return nil
}

func initGormLog() (newLogger logger.Interface) {

	if GetConfEnv()=="dev" {
		newLogger = LogNew(
			logger.Config{
				SlowThreshold: time.Second,   // 慢 SQL 阈值
				LogLevel:      logger.Info, // Log level
				Colorful:      false,         // 禁用彩色打印
			},
		)
	}else{
		newLogger = LogNew(
			logger.Config{
				SlowThreshold: time.Second,   // 慢 SQL 阈值
				LogLevel:      logger.Error, // Log level
				Colorful:      false,         // 禁用彩色打印
			},
		)
	}
	return
}
//mysql日志打印类
// Logger default logger
type MysqlGormLogger struct {
	Config logger.Config
}
func LogNew(config logger.Config) logger.Interface{
	return &MysqlGormLogger{
		config,
	}
}
func (m *MysqlGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *m
	newlogger.Config.LogLevel = level
	return &newlogger
}
// Info print info
func (m MysqlGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if m.Config.LogLevel >= logger.Info {
		info:=fmt.Sprintf(msg,data...)
		Log.TagInfo(NewTrace(),DLTagMySqlEmpty, map[string]interface{}{"info":info})
	}
}
// Warn print warn messages
func (m MysqlGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if m.Config.LogLevel >= logger.Warn {
		info:=fmt.Sprintf(msg,data...)
		Log.TagWarn(NewTrace(),DLTagMySqlEmpty, map[string]interface{}{"warn":info})
	}
}

// Error print error messages
func (m MysqlGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if m.Config.LogLevel >= logger.Error {
		info:=fmt.Sprintf(msg,data...)
		Log.TagError(NewTrace(),DLTagMySqlEmpty, map[string]interface{}{"error":info})
	}
}

func (m MysqlGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

	if m.Config.LogLevel >= logger.Silent {
		elapsed := time.Since(begin)
		switch {
		case err != nil && m.Config.LogLevel >= logger.Error:
			sql, rows := fc()
			Log.TagError(NewTrace(),DLTagMySqlFailed, map[string]interface{}{"err":err,"elapsed":elapsed,"rows":rows,"sql":sql})
		case elapsed > m.Config.SlowThreshold && m.Config.SlowThreshold != 0 && m.Config.LogLevel >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", m.Config.SlowThreshold)
			Log.TagWarn(NewTrace(),DLTagMySqlWarn, map[string]interface{}{"slowLog":slowLog,"elapsed":elapsed,"rows":rows,"sql":sql})
		case m.Config.LogLevel == logger.Info:
			sql, rows := fc()
			Log.TagInfo(NewTrace(),DLTagMySqlSuccess, map[string]interface{}{"elapsed":elapsed,"rows":rows,"sql":sql})
		}
	}
}

