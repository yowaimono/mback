package mback

type Recoverer interface {
	Recover(tableName string, schema string, data string) error
	Restore(filePath string) error
	Close() error
}
