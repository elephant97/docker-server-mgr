package mysqlops

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func MysqlConnection() (*sql.DB, error) {
	dsn := "sherpa:sherpa@tcp(128.10.30.70:3306)/sherpa?parseTime=true&allowNativePasswords=true"

	// DB 연결
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB 연결 실패: %v", err)
	}

	// 연결 확인
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("DB Ping 실패: %v", err)
	}

	log.Println("✅ MySQL 연결 성공")
	return db, nil
}
