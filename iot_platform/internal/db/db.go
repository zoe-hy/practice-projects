package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// OpenDB 打开数据库并设置连接池参数
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn+"&charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 2)
	return db, nil
}

// ExecPrepared 执行预编译的 Insert/Update 操作，带重试
func ExecPrepared(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		stmt, err := db.Prepare(query)
		if err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(time.Millisecond * time.Duration(100*(i+1)))
			}
			continue
		}
		defer stmt.Close()

		result, err := stmt.Exec(args...)
		if err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(time.Millisecond * time.Duration(100*(i+1)))
			}
			continue
		}
		return result, nil
	}
	return nil, fmt.Errorf("execution failed after %d retries: %w", maxRetries, lastErr)
}

// RunMigrations 从 sql 文件中按分号执行语句（简单实现，仅用于示例）
func RunMigrations(db *sql.DB, sqlFilePath string) error {
	b, err := ioutil.ReadFile(sqlFilePath)
	if err != nil {
		return err
	}
	content := string(b)
	parts := strings.Split(content, ";")
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}
