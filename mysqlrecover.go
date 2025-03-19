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
	logger.Debug("Entering NewMySQLRecoverer")

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

	logger.Debug("Exiting NewMySQLRecoverer with success")
	return &MySQLRecoverer{db: db}, nil
}

func (r *MySQLRecoverer) Recover(tableName string, schema string, data string) error {
	logger.Debug("Entering MySQLRecoverer.Recover with tableName: %s", tableName)
	logger.Debug("Schema: %s", schema)
	logger.Debug("Data: %s", data)

	_, err := r.db.Exec(schema)
	if err != nil {
		logger.Error("Failed to execute schema: %v", err)
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	logger.Info("Successfully executed schema")

	_, err = r.db.Exec(data)
	if err != nil {
		logger.Error("Failed to execute data: %v", err)
		return fmt.Errorf("failed to execute data: %w", err)
	}
	logger.Info("Successfully executed data")

	logger.Debug("Exiting MySQLRecoverer.Recover with success")
	return nil
}

func (r *MySQLRecoverer) Close() error {
	logger.Debug("Entering MySQLRecoverer.Close")

	err := r.db.Close()
	if err != nil {
		logger.Error("Failed to close database connection: %v", err)
		return err
	}
	logger.Info("Successfully closed database connection")

	logger.Debug("Exiting MySQLRecoverer.Close")
	return err
}

func (m *Mback) Recover(filePath string) error {
	logger.Debug("Entering Mback.Recover with filePath: %s", filePath)

	err := m.recoverer.Restore(filePath)
	if err != nil {
		logger.Error("Failed to restore from file: %v", err)
		return err
	}
	logger.Info("Successfully restored from file")

	logger.Debug("Exiting Mback.Recover with success")
	return nil
}

func (r *MySQLRecoverer) Restore(filePath string) error {
	logger.Info("Entering MySQLRecoverer.Restore with filePath: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("Failed to open SQL file: %v", err)
		return fmt.Errorf("failed to open SQL file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Error("Failed to close SQL file: %v", err)
		}
	}()
	logger.Info("Successfully opened SQL file: %s", filePath)

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
			logger.Debug("Added SQL statement: %s", sqlStatements[len(sqlStatements)-1])
		}
	}

	if currentStatement.Len() > 0 {
		sqlStatements = append(sqlStatements, currentStatement.String())
		logger.Debug("Added remaining SQL statement: %s", sqlStatements[len(sqlStatements)-1])
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Failed to read SQL file: %v", err)
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	for _, statement := range sqlStatements {
		logger.Debug("Executing statement: %s", statement)
		_, err := r.db.Exec(statement)
		if err != nil {
			logger.Error("Failed to execute SQL statement: %v, statement: %s", err, statement)
			return fmt.Errorf("failed to execute SQL statement: %w, statement: %s", err, statement)
		}
		logger.Debug("Successfully executed SQL statement")
	}

	logger.Info("Restore completed successfully!")
	logger.Debug("Exiting MySQLRecoverer.Restore with success")
	return nil
}
