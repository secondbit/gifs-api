package main

import (
	"database/sql"

	"secondbit.org/gifs/api"

	_ "github.com/go-sql-driver/mysql"
)

func initDB(dsn, name string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	return (*api.SQLStore)(db).Init(name)
}
