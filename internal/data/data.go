package data

import (
	"demo/internal/conf"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo, NewUserLoginLogRepo)

// Data holds database and cache clients.
type Data struct {
	db *gorm.DB
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)

	// Initialize GORM DB
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logHelper.Info("database connection initialized")

	d := &Data{
		db: db,
	}

	cleanup := func() {
		logHelper.Info("closing the data resources")
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return d, cleanup, nil
}
