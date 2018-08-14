// Definition of the structures and SQL interaction functions
package models

import "github.com/jinzhu/gorm"

// TeamDS implements the TeamSQL methods
type TeamDS interface {
	Save(*Team, TeamArgs) error
	Delete(*Team, TeamArgs) error
	Find(TeamArgs) ([]Team, error)
}

// Teamstore returns a TeamDS implementing CRUD methods for the tags and containing a gorm client
func TeamStore(db *gorm.DB) TeamDS {
	return &TeamSQL{DB: db}
}
