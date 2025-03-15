package mback

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// SQLFileWriter 实现了 Writer 接口，将数据写入 SQL 文件
type SQLFileWriter struct {
	file *os.File
}

func NewSQLFileWriter(filename string) (*SQLFileWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL file: %w", err)
	}
	return &SQLFileWriter{file: file}, nil
}

func (w *SQLFileWriter) Write(tableName, content string) error {
	_, err := w.file.WriteString(content + "\n")
	return err
}

func (w *SQLFileWriter) Close() error {
	return w.file.Close()
}

// MySQLReader 实现了 Reader 接口，从 MySQL 数据库读取数据
type MySQLReader struct {
	db *sql.DB
}

func NewMySQLReader(cfg *Config) (*MySQLReader, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.UserName,
		cfg.PassWord,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLReader{db: db}, nil
}

func (r *MySQLReader) GetAllTables() ([]string, error) {
	rows, err := r.db.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to show tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (r *MySQLReader) DumpTableStruct(tableName string) (string, error) {
	rows, err := r.db.Query(fmt.Sprintf("SHOW CREATE TABLE %s", tableName))
	if err != nil {
		return "", fmt.Errorf("failed to query table structure: %w", err)
	}
	defer rows.Close()

	var tableNameResult, createStatement string
	if rows.Next() {
		if err := rows.Scan(&tableNameResult, &createStatement); err != nil {
			return "", fmt.Errorf("failed to scan table structure: %w", err)
		}
	}

	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n%s;\n", tableName, createStatement), nil
}

func (r *MySQLReader) DumpTableData(tableName string) (string, error) {
	rows, err := r.db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return "", fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %w", err)
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var data strings.Builder
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		data.WriteString(fmt.Sprintf("INSERT INTO `%s` VALUES (", tableName))
		for i, col := range values {
			if col == nil {
				data.WriteString("NULL")
			} else {
				data.WriteString(fmt.Sprintf("'%s'", strings.ReplaceAll(string(col), "'", "\\'")))
			}
			if i < len(values)-1 {
				data.WriteString(",")
			}
		}
		data.WriteString(");\n")
	}

	return data.String(), nil
}

