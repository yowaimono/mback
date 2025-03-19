package mback

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/yowaimono/mback/logger"
)

// SQLFileWriter 实现了 Writer 接口，将数据写入 SQL 文件
type SQLFileWriter struct {
	file *os.File
}

func NewSQLFileWriter(filename string) (*SQLFileWriter, error) {
	logger.Debug("Entering NewSQLFileWriter with filename: %s", filename)
	file, err := os.Create(filename)
	if err != nil {
		logger.Error("Failed to create SQL file: %v", err)
		return nil, fmt.Errorf("failed to create SQL file: %w", err)
	}
	logger.Info("Successfully created SQL file: %s", filename)
	logger.Debug("Exiting NewSQLFileWriter with success")
	return &SQLFileWriter{file: file}, nil
}

func (w *SQLFileWriter) Write(tableName, content string) error {
	logger.Debug("Entering SQLFileWriter.Write with tableName: %s", tableName)
	_, err := w.file.WriteString(content + "\n")
	if err != nil {
		logger.Error("Failed to write to SQL file: %v", err)
		return err
	}
	logger.Debug("Successfully wrote to SQL file for table: %s", tableName)
	logger.Debug("Exiting SQLFileWriter.Write")
	return err
}

func (w *SQLFileWriter) Close() error {
	logger.Debug("Entering SQLFileWriter.Close")
	err := w.file.Close()
	if err != nil {
		logger.Error("Failed to close SQL file: %v", err)
		return err
	}
	logger.Info("Successfully closed SQL file")
	logger.Debug("Exiting SQLFileWriter.Close")
	return err
}

// MySQLReader 实现了 Reader 接口，从 MySQL 数据库读取数据
type MySQLReader struct {
	db *sql.DB
}

func NewMySQLReader(cfg *Config) (*MySQLReader, error) {
	logger.Debug("Entering NewMySQLReader")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.UserName,
		cfg.PassWord,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	logger.Debug("DSN: %s", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Error("Failed to open database connection: %v", err)
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	logger.Info("Successfully opened database connection")

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("Successfully pinged database")

	logger.Debug("Exiting NewMySQLReader with success")
	return &MySQLReader{db: db}, nil
}

func (r *MySQLReader) GetAllTables() ([]string, error) {
	logger.Debug("Entering MySQLReader.GetAllTables")
	rows, err := r.db.Query("SHOW TABLES")
	if err != nil {
		logger.Error("Failed to show tables: %v", err)
		return nil, fmt.Errorf("failed to show tables: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Error("Failed to close rows: %v", err)
		}
	}()
	logger.Debug("Successfully executed SHOW TABLES query")

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			logger.Error("Failed to scan table name: %v", err)
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
		logger.Debug("Found table: %s", table)
	}
	logger.Info("Found tables: %v", tables)
	logger.Debug("Exiting MySQLReader.GetAllTables with success")
	return tables, nil
}

func (r *MySQLReader) DumpTableStruct(tableName string) (string, error) {
	logger.Debug("Entering MySQLReader.DumpTableStruct with tableName: %s", tableName)
	query := fmt.Sprintf("SHOW CREATE TABLE %s", tableName)
	logger.Debug("Executing query: %s", query)
	rows, err := r.db.Query(query)
	if err != nil {
		logger.Error("Failed to query table structure: %v", err)
		return "", fmt.Errorf("failed to query table structure: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Error("Failed to close rows: %v", err)
		}
	}()
	logger.Debug("Successfully executed SHOW CREATE TABLE query")

	var tableNameResult, createStatement string
	if rows.Next() {
		if err := rows.Scan(&tableNameResult, &createStatement); err != nil {
			logger.Error("Failed to scan table structure: %v", err)
			return "", fmt.Errorf("failed to scan table structure: %w", err)
		}
	}

	result := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n%s;\n", tableName, createStatement)
	logger.Debug("Table structure: %s", result)
	logger.Info("Successfully dumped table structure for table: %s", tableName)
	logger.Debug("Exiting MySQLReader.DumpTableStruct with success")
	return result, nil
}

func (r *MySQLReader) DumpTableData(tableName string) (string, error) {
	logger.Debug("Entering MySQLReader.DumpTableData with tableName: %s", tableName)
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	logger.Debug("Executing query: %s", query)
	rows, err := r.db.Query(query)
	if err != nil {
		logger.Error("Failed to query table data: %v", err)
		return "", fmt.Errorf("failed to query table data: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Error("Failed to close rows: %v", err)
		}
	}()
	logger.Debug("Successfully executed SELECT * query")

	columns, err := rows.Columns()
	if err != nil {
		logger.Error("Failed to get columns: %v", err)
		return "", fmt.Errorf("failed to get columns: %w", err)
	}
	logger.Debug("Columns: %v", columns)

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var data strings.Builder
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			logger.Error("Failed to scan row: %v", err)
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		data.WriteString(fmt.Sprintf("INSERT INTO `%s` VALUES (", tableName))
		for i, col := range values {
			if col == nil {
				data.WriteString("NULL")
			} else {
				escapedValue := strings.ReplaceAll(string(col), "'", "\\'")
				data.WriteString(fmt.Sprintf("'%s'", escapedValue))
			}
			if i < len(values)-1 {
				data.WriteString(",")
			}
		}
		data.WriteString(");\n")
	}

	logger.Info("Successfully dumped table data for table: %s", tableName)
	logger.Debug("Exiting MySQLReader.DumpTableData with success")
	return data.String(), nil
}
