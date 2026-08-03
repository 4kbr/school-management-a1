package sqlconnect

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb(dbname string) (*sql.DB, error) {
	fmt.Println("try connecting to MariaDB")
	// todo: nanti bagian credential diambil dari .env
	connectionString := "root:root@tcp(127.0.0.1:3306)/" + dbname
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	fmt.Println("connected to MariaDB")
	return db, nil
}
