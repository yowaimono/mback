package mback

// Writer 接口定义了将数据写入目标格式的操作
type Writer interface {
	Write(tableName, content string) error
	Close() error
}
