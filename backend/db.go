package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // import để đăng ký driver
)

var db *sql.DB

func connectDB() error {
	dsn := os.Getenv("DATABASE_URL")

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	// Giới hạn pool — tránh mở quá nhiều kết nối tới Postgres
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return db.Ping() // sql.Open chưa thật sự kết nối — Ping mới kiểm tra
}
