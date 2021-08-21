package AppInit

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB
func DBInit() error {
	var err error
	db, err = sqlx.Connect("mysql",
		"root:root@tcp(192.168.0.165:3306)/go?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return nil
}
func  GetDB() *sqlx.DB {
	return db
}

