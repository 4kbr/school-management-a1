package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb(dbname string) (*sql.DB, error) {
	fmt.Println("try connecting to MariaDB")

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	fmt.Println("connected to MariaDB")
	return db, nil
}
