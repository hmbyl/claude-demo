package data

import (
	"context"
	"demo/internal/conf"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo, NewUserLoginLogRepo, NewCaptchaRepo)

// Data holds database and cache clients.
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
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

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		ReadTimeout:  c.Redis.ReadTimeout,
		WriteTimeout: c.Redis.WriteTimeout,
	})

	// Test Redis connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, nil, err
	}

	logHelper.Info("redis connection initialized")

	d := &Data{
		db:  db,
		rdb: rdb,
	}

	cleanup := func() {
		logHelper.Info("closing the data resources")
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		if rdb != nil {
			_ = rdb.Close()
		}
	}
	return d, cleanup, nil
}
