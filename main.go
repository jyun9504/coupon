package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// init 變數
var db *gorm.DB

func main() {
	// START POINT
	fmt.Println("Start!")

	var err error
	// init mysql 連線
	dsn := "root:password@tcp(127.0.0.1:3306)/coupon_db"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	// Check DB connection
	fmt.Println("DB connection Success!")
}