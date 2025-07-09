package mysqlops

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// Insert, Update, Delete 쿼리 실행 함수 (결과 집합이 없는 쿼리용)
func ExecQuery(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {

	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Query execution failed: %s, args: %v, error: %v", query, args, err)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return result, nil
}

// 구조체의 db 태그 또는 필드명을 기반으로 컬럼에 매핑
/**
 * strucnt example:
 * type Container struct {
 * 	ID    string `db:"id"`
 * 	Image string `db:"image"`
 * 	Tag   string `db:"tag"`}
 */
func SelectQueryRowsToStructs[T any](db *sql.DB, query string, args ...interface{}) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Query failed: %s, args: %v, error: %v", query, args, err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Printf("Failed to get columns: %s, args: %v, error: %v", query, args, err)
		return nil, err
	}

	var results []T
	for rows.Next() {
		var obj T
		objVal := reflect.ValueOf(&obj).Elem()
		objType := objVal.Type()

		fieldMap := make(map[string]reflect.Value)

		for i := 0; i < objType.NumField(); i++ {
			field := objType.Field(i)
			colName := field.Tag.Get("db")
			if colName == "" {
				colName = field.Name
			}
			fieldMap[strings.ToLower(colName)] = objVal.Field(i)
		}

		fields := make([]interface{}, len(columns))
		for i, col := range columns {
			if f, ok := fieldMap[strings.ToLower(col)]; ok && f.CanSet() {
				fields[i] = f.Addr().Interface()
			} else {
				var dummy interface{}
				fields[i] = &dummy
			}
		}

		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
		results = append(results, obj)
	}

	return results, nil
}
