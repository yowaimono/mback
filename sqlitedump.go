package mback

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" 
)


type SQLiteReader struct {
	db *sql.DB
}


func NewSQLiteReader(dbPath string) (*SQLiteReader, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return &SQLiteReader{db: db}, nil
}


func (r *SQLiteReader) GetAllTables() ([]string, error) {
	rows, err := r.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
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


func (r *SQLiteReader) DumpTableStruct(tableName string) (string, error) {
	rows, err := r.db.Query(fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", tableName))
	if err != nil {
		return "", fmt.Errorf("failed to query table structure: %w", err)
	}
	defer rows.Close()

	var createTableSQL string
	if rows.Next() {
		if err := rows.Scan(&createTableSQL); err != nil {
			return "", fmt.Errorf("failed to scan table structure: %w", err)
		}
	}

	return createTableSQL, nil
}


func (r *SQLiteReader) DumpTableData(tableName string) (string, error) {
	rows, err := r.db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return "", fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %w", err)
	}

	var data strings.Builder
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		rowData := make([]string, len(columns))
		for i, val := range values {
			rowData[i] = fmt.Sprintf("%v", val)
		}
		data.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (\"%s\");\n",
			tableName, strings.Join(columns, ", "), strings.Join(rowData, "\", \"")))
	}

	return data.String(), nil
}


func (r *SQLiteReader) Close() error {
	return r.db.Close()
}
