package lib

import (
	"flychat/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
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

// NewDB 初始化并返回一个 GORM DB 对象
func NewDB(config Config) (*gorm.DB, error) {
	// 构建 DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	// 可以在这里进行数据库迁移等操作
	if err := db.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	return db, nil
}

func InstallDB() {
	// 配置数据库连接信息
	config := Config{
		Host:     "*",
		Port:     "3306",
		User:     "*",
		Password: "*",
		DBName:   "flychat",
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

	// 可以在这里进行数据库迁移等操作
	if err := db.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	DB = db
	return
}
