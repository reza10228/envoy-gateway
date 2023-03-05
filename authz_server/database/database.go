package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"main/model"
	"os"
)

type DbInstance struct {
	Db *gorm.DB
}

var DB DbInstance

func ConnectDb()  {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True","srcaccess","gfKIgxwnM2YixszA","192.168.132.18","3306","hooshang")
	//dsn := "root:root@tcp(localhost:3306)/userMediana?parseTime=True"

	db , err := gorm.Open(mysql.Open(dsn),&gorm.Config{

	})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}
	log.Println("Data base connected")
	db.Logger = logger.Default.LogMode(logger.Info)
	DB = DbInstance {
		Db: db,
	}
}
func (manager *DbInstance) CheckKey(key string) (*model.ApiKey,error) {
	var userKey model.ApiKey
	result := manager.Db.Model(model.ApiKey{}).Preload("User").Where("api_keys.key = ? and revoked = 0",key).First(&userKey)
	if result.Error != nil{
		return nil ,result.Error
	}
	return &userKey,nil
}