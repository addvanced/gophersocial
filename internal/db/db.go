package db

type DatabaseType string

const (
	PostgresDatabaseType DatabaseType = "postgres"
	MongoDBDatabaseType  DatabaseType = "mongodb"
)

type DBConfiger interface {
	ConnString() string
}
