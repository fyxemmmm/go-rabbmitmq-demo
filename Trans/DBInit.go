package Trans

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

var db *sqlx.DB
func DBInit(dbName string) error {
	var err error
	db, err = sqlx.Connect("mysql",
		"root:root@tcp(localhost:3306)/"+dbName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Duration(8*3600) * time.Second)
	return nil
}
func GetDB() *sqlx.DB {
	return db
}

