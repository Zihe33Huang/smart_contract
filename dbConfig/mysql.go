package dbConfig

import (
	"awesomeProject2/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var NaixueDB *gorm.DB
var MiddleDB *gorm.DB

func init() {
	var err error
	// 1、 从数据库读数据
	// 1.1、 连接奈雪数据库
	naixue := config.Config().GetString("database.nayuki")
	NaixueDB, err = gorm.Open(mysql.Open(naixue), &gorm.Config{})
	if err != nil {
		log.Fatalf("models.Setup err: %v", err)
	}
	// 1.2、 连接中台数据库
	middle := config.Config().GetString("database.middle")
	MiddleDB, err = gorm.Open(mysql.Open(middle), &gorm.Config{})
	if err != nil {
		log.Fatalf("models.Setup err: %v", err)
	}
}
