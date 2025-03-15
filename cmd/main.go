package main

import (
	"fmt"

	"github.com/yowaimono/mback"
	"github.com/yowaimono/mback/logger"
)

func main() {
	logger.Info("开始备份...")

	cfg := &mback.Config{
		Host:       "localhost",
		Port:       3306,
		UserName:   "root",
		PassWord:   "123456",
		Database:   "tests",
		OutputFile: "./backup-struct.sql",
		Source:     mback.MYSQL,
	}

	mback, err := mback.NewMback(cfg)
	if err != nil {
		return
	}

	if err = mback.DumpAll(&User{}, &Like{}); err != nil {
		fmt.Println("Failed to dump data:", err)
		return
	}

	fmt.Println("Backup completed successfully!")
}

type User struct {
}

func (u User) TableName() string {
	return "users"
}

type Like struct {
}

func (l Like) TableName() string {
	return "likes"
}
