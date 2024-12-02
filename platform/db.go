package platform

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

var (
	DB *gorm.DB
)

// Config 包含数据库连接的配置信息
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func InitDB() {
	// 配置数据库连接信息
	config := Config{
		Host:     os.Getenv("SQL_HOST"),
		Port:     os.Getenv("SQL_PORT"),
		User:     os.Getenv("SQL_USER"),
		Password: os.Getenv("SQL_PASSWORD"),
		DBName:   os.Getenv("SQL_DBNAME"),
	}

	// 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	DB = db
	return
}
