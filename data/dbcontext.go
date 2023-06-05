package data

import (
	"fmt"

	"github.com/glebarez/sqlite"
	_"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"

)

var DB *gorm.DB

// func openMySql(server, database, username, password string, port int) *gorm.DB {
// 	var url string
// 	url = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
// 		username, password, server, port, database)

// 	// db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
// 	if err != nil {
// 		panic("failed to connect database")
// 	}
// 	return db
// }

func openPostgres() *gorm.DB {
	var err error
	dsn := "host=postgresql.traininglog.svc.cluster.local user=postgres-user dbname=postgres-db sslmode=disable password=POSTGRES_PASSWORD"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("db err: ", err)
	}
	return db
}

func InitDatabase(file, server, database, username, password string, port int) {
	if len(file) == 0 {
		// DB = openMySql(server, database, username, password, port)
		DB = openPostgres()
	} else {
		DB, _ = gorm.Open(sqlite.Open(file), &gorm.Config{})
	}

	DB.AutoMigrate(&User{})
	// DB.AutoMigrate(&Employee{})
	// var antal int64
	// DB.Model(&Employee{}).Count(&antal) // Seed
	// if antal == 0 {
	// 	DB.Create(&Employee{Age: 50, Namn: "Stefan", City: "Test"})
	// 	DB.Create(&Employee{Age: 14, Namn: "Oliver", City: "Test"})
	// 	DB.Create(&Employee{Age: 20, Namn: "Josefine", City: "Test"})
	// }
}
