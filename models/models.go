// Definition of the structures and SQL interaction functions
package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func Models() []interface{} {
	return []interface{}{
		&User{}, &Group{}, &Team{},
	}
}
