package mback


type Reader interface {
	GetAllTables() ([]string, error)
	DumpTableStruct(tableName string) (string, error)
	DumpTableData(tableName string) (string, error)
}
