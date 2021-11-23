package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/neverteaser/golf"
)

func main() {
	var dsn = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	globalDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		AllowGlobalUpdate: false,
	})
	if err != nil {
		log.Fatal("open db failed")

	}
	if err := globalDB.AutoMigrate(&testModel{}); err != nil {
		log.Fatal("db  migration failed")

	}
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		golfQ := golf.NewGolf(globalDB)
		var tests testModel
		if err := golfQ.Build(&testModel{}, c.Request.URL.Query()).Find(&tests).Error; err != nil {
			log.Println("find failed", err)
		}
		fmt.Println(c.Request.URL.Query())
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	func() { _ = r.Run() }()
	// URL path /ping?eq_id=1&like_username=test
	// sql log should be  SELECT * FROM "test_model" WHERE username LIKE 'test' AND id = 1 LIMIT 10
}
