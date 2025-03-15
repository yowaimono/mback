package mback

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/yowaimono/mback/logger"
)

type MySQLRecoverer struct {
	db *sql.DB
}

func NewMySQLRecoverer(cfg *Config) (*MySQLRecoverer, error) {

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

	return &MySQLRecoverer{db: db}, nil
}

func (r *MySQLRecoverer) Recover(tableName string, schema string, data string) error {

	_, err := r.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}


	_, err = r.db.Exec(data)
	if err != nil {
		return fmt.Errorf("failed to execute data: %w", err)
	}

	return nil
}


func (r *MySQLRecoverer) Close() error {
	return r.db.Close()
}

func (m *Mback) Recover(filePath string) error {
	return m.recoverer.Restore(filePath)
}
func (r *MySQLRecoverer) Restore(filePath string) error {
	logger.Info("Recovering file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open SQL file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sqlStatements []string
	var currentStatement strings.Builder
	var lineNumber int

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "/*") {
			logger.Debug("Skipping line %d: %s", lineNumber, line)
			continue
		}

		currentStatement.WriteString(line + " ")

		
		if strings.HasSuffix(line, ";") {
			sqlStatements = append(sqlStatements, currentStatement.String())
			currentStatement.Reset()
		}
	}


	if currentStatement.Len() > 0 {
		sqlStatements = append(sqlStatements, currentStatement.String())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	for _, statement := range sqlStatements {
		logger.Debug("Executing statement: %s", statement)
		_, err := r.db.Exec(statement)
		if err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w, statement: %s", err, statement)
		}
	}

	logger.Info("Restore completed successfully!")
	return nil
}
