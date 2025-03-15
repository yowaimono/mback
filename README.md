# 🚀 Mback - MySQL Backup and Restore Tool

**Mback** is a lightweight and efficient Go package designed to simplify the process of backing up and restoring MySQL databases. Whether you're managing a small project or a large-scale application, Mback provides a flexible and easy-to-use API to dump database tables (both structure and data) into SQL files and restore them back into a database.

---

## ✨ Features

- **📦 Backup Database Tables**: Dump the structure and data of MySQL tables into SQL files.
- **🔄 Restore Database Tables**: Restore tables from SQL files back into a MySQL database.
- **⚙️ Flexible Configuration**: Supports multiple database sources (MySQL, SQLite, etc.) and customizable output files.
- **📝 Integrated Logging**: Built-in logging for better debugging and monitoring.
- **🚀 Lightweight and Fast**: Designed for performance and simplicity.

---

## 🛠 Installation

To use Mback in your Go project, install it using `go get`:

```bash
go get github.com/yowaimono/mback
```

---

## 🚀 Usage

### 📤 Backup Database Tables

To backup **all tables** in a MySQL database:

```go
package main

import (
	"github.com/yowaimono/mback"
)

func main() {
	cfg := &mback.Config{
		Source:    mback.MYSQL,
		UserName:  "root",
		PassWord:  "password",
		Host:      "localhost",
		Port:      3306,
		Database:  "mydatabase",
		OutputFile: "backup.sql",
	}

	mbackInstance, err := mback.NewMback(cfg)
	if err != nil {
		panic(err)
	}

	err = mbackInstance.DumpAll()
	if err != nil {
		panic(err)
	}
}
```

To backup **specific tables**:

```go
err = mbackInstance.DumpAll("table1", "table2")
if err != nil {
    panic(err)
}
```

---

### 📥 Restore Database Tables

To restore tables from a SQL file:

```go
package main

import (
	"github.com/yowaimono/mback"
)

func main() {
	cfg := &mback.Config{
		Source:    mback.MYSQL,
		UserName:  "root",
		PassWord:  "password",
		Host:      "localhost",
		Port:      3306,
		Database:  "mydatabase",
	}

	recoverer, err := mback.NewMySQLRecoverer(cfg)
	if err != nil {
		panic(err)
	}
	defer recoverer.Close()

	// Assuming schema and data are read from a SQL file
	schema := "DROP TABLE IF EXISTS `table1`; CREATE TABLE `table1` (...);"
	data := "INSERT INTO `table1` VALUES (...);"

	err = recoverer.Recover("table1", schema, data)
	if err != nil {
		panic(err)
	}
}
```

---

## ⚙️ Configuration

The `Config` struct is used to configure the database connection and backup/restore settings:

```go
type Config struct {
	Source    int    // Database source (e.g., MYSQL, SQLITE)
	UserName  string // Database username
	PassWord  string // Database password
	Host      string // Database host
	Port      int    // Database port
	Database  string // Database name
	OutputFile string // Output file for backup
}
```

---

## 📝 Logging

Mback uses a built-in logger for logging messages. You can control the logging level and output as needed.

```go
logger.Info("Starting backup process...")
logger.Debug("Processing table: %s", tableName)
logger.Error("Failed to backup table: %s", err)
```

---

## 🤝 Contributing

Contributions are welcome! If you'd like to contribute, please:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request.

---

## 📜 License

Mback is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for more details.

---

## 🙏 Acknowledgments

- Thanks to the **Go community** for providing excellent libraries and tools.
- Special thanks to the contributors who helped improve this package.

---

For more information, visit the [GitHub repository](https://github.com/yowaimono/mback). Happy coding! 🎉
