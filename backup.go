package mback

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/yowaimono/mback/logger"
)

type Mback struct {
	cfg       *Config
	reader    Reader
	writer    Writer
	recoverer Recoverer
}

func (m *Mback) DumpAll(tables ...interface{}) error {
	var tableNames []string

	if len(tables) == 0 {
		logger.Info("dump all tables...")
		allTables, err := m.reader.GetAllTables()
		if err != nil {
			logger.Info("get all table error: %s", err)
			return err
		}
		tableNames = allTables
	} else {

		for _, table := range tables {
			if tableName, ok := table.(string); ok {

				tableNames = append(tableNames, tableName)
			} else if tableNameInterface, ok := table.(TableName); ok {

				tableNames = append(tableNames, tableNameInterface.TableName())
			} else {
				logger.Error("handle table %v error，skip it", table)
			}
		}
	}

	for _, tableName := range tableNames {
		structDump, err := m.reader.DumpTableStruct(tableName)
		if err != nil {
			return err
		}

		dataDump, err := m.reader.DumpTableData(tableName)
		if err != nil {
			return err
		}

		err = m.writer.Write(tableName, structDump+"\n"+dataDump)
		if err != nil {
			return err
		}
	}

	return m.writer.Close()
}

func (m *Mback) DumpAllStruct(tables ...interface{}) error {
	var tableNames []string

	if len(tables) == 0 {
		logger.Info("dump all tables...")
		allTables, err := m.reader.GetAllTables()
		if err != nil {
			logger.Info("get all table error: %s", err)
			return err
		}
		tableNames = allTables
	} else {

		for _, table := range tables {
			if tableName, ok := table.(string); ok {

				tableNames = append(tableNames, tableName)
			} else if tableNameInterface, ok := table.(TableName); ok {

				tableNames = append(tableNames, tableNameInterface.TableName())
			} else {
				logger.Error("handle table %v error，skip it", table)
			}
		}
	}

	for _, tableName := range tableNames {

		structDump, err := m.reader.DumpTableStruct(tableName)
		if err != nil {
			return err
		}
		logger.Debug("开始处理表：%s\n\n%s\n\n", tableName, structDump)

		err = m.writer.Write(tableName, structDump)
		if err != nil {
			return err
		}
	}

	return m.writer.Close()
}

func NewMback(cfg *Config) (*Mback, error) {
	if cfg.Source == MYSQL {
		reader, err := NewMySQLReader(cfg)

		logger.Info("config.is %v", *cfg)
		if err != nil {
			logger.Error("ailed to create MySQL reader: %s", err.Error())
			return nil, fmt.Errorf("failed to create MySQL reader: %s", err.Error())
		}

		writer, err := NewSQLFileWriter(cfg.OutputFile) // Default to SQL file writer
		if err != nil {
			logger.Error("failed to create SQL writer: %s", err.Error())
			return nil, fmt.Errorf("failed to create SQL writer: %s", err.Error())
		}

		recoverer, err := NewMySQLRecoverer(cfg)
		if err != nil {
			panic(err)
		}

		return &Mback{
			cfg:       cfg,
			reader:    reader,
			writer:    writer,
			recoverer: recoverer,
		}, nil
	} else if cfg.Source == SQLITE {
		// 创建 SQLiteReader
		reader, err := NewSQLiteReader(cfg.Database)
		if err != nil {
			fmt.Println("Failed to create SQLite reader:", err)
			return nil, err
		}

		// 创建 SqlWriter
		writer, err := NewSQLFileWriter(cfg.OutputFile)
		if err != nil {
			fmt.Println("Failed to create SQL writer:", err)
			return nil, err
		}
		return &Mback{
			cfg:    cfg,
			reader: reader,
			writer: writer,
		}, nil
	} else {
		logger.Fatal("reader is not implement")
		return nil, fmt.Errorf("not implement")
	}
}
