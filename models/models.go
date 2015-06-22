package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Model interface{}

func GetModels() []interface{} {
	return []interface{}{}
}

func GetStores() []interface{} {
	return []interface{}{}
}
