package mysqlops

import (
	"database/sql"
	"fmt"

	"docker-server-mgr/config"

	clog "docker-server-mgr/utils/log" //custom log

	_ "github.com/go-sql-driver/mysql"
)

func MysqlConnection(mysqlConfig *config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&allowNativePasswords=true",
		mysqlConfig.User,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database,
	)

	// DB 연결
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB 연결 실패", "err", err)
	}

	// 연결 확인
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("DB Ping 실패", "err", err)
	}

	clog.Debug("✅ MySQL 연결 성공")
	return db, nil
}
