package db

import (
	"fmt"
	"log"
	"os"
	"time"

	myLooger "edgelink-api/internal/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbConfig struct {
	Username    string
	Password    string
	Host        string
	DBName      string
	Port        int
	ShowSQL     bool
	MaxIdleConn int
	MaxOpenConn int
}

func NewMySQLWithGORM(dbCfg dbConfig) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCfg.Username,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.DBName,
	)

	gromCfg := &gorm.Config{}
	if dbCfg.ShowSQL {
		gormLog := logger.New(
			log.New(os.Stdout, "\r\n[GORM] ", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             200 * time.Millisecond, // 慢 SQL 阈值
				LogLevel:                  logger.Info,            // Log level
				IgnoreRecordNotFoundError: true,                   // 忽略 ErrRecordNotFound
				Colorful:                  true,                   // 彩色打印
			},
		)
		gromCfg.Logger = gormLog
	}
	db, err := gorm.Open(mysql.Open(dsn), gromCfg)
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConn)
	myLooger.Info("connect to mysql success", "addr", fmt.Sprintf("%s:%d", dbCfg.Host, dbCfg.Port))
	return db
}

func InitDB() *gorm.DB {
	return NewMySQLWithGORM(dbConfig{
		Username:    viper.GetString("databases.edgelink.username"),
		Password:    viper.GetString("databases.edgelink.password"),
		Host:        viper.GetString("databases.edgelink.host"),
		Port:        viper.GetInt("databases.edgelink.port"),
		DBName:      viper.GetString("databases.edgelink.dbname"),
		ShowSQL:     true,
		MaxIdleConn: 2,
		MaxOpenConn: 5,
	})
}
