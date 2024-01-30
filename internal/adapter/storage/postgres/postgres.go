package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func ConnectDB(databaseHost, databasePort, databaseUser, databasePassword, databaseName, databaseSchemaName string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable search_path=%s",
		databaseHost,
		databasePort,
		databaseUser,
		databasePassword,
		databaseName,
		databaseSchemaName,
	)

	dbInstance, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = dbInstance.Ping()
	if err != nil {
		return nil, err
	}

	return dbInstance, nil
}
