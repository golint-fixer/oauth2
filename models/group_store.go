// Definition of the structures and SQL interaction functions
package models

import "github.com/jinzhu/gorm"

// GroupDS implements the GroupSQL methods
type GroupDS interface {
	Save(*Group) error
	Delete(*Group) error
	First(*Group, uint) error
	Find() ([]Group, error)
}

// GroupSQL contains a Gorm client and the group and gorm related methods
type GroupSQL struct {
	DB *gorm.DB
}

// Groupstore returns a GroupDS implementing CRUD methods for the groups and containing a gorm client
func GroupStore(db *gorm.DB) GroupDS {
	return &GroupSQL{DB: db}
}

// Save inserts a new group into the database
func (s *GroupSQL) Save(g *Group) error {
	if g.ID == 0 {
		s.DB.Create(g)

		return s.DB.Error
	}

	s.DB.Save(g)

	return s.DB.Error
}

// Delete removes a group from the database
func (s *GroupSQL) Delete(g *Group) error {
	s.DB.Delete(g)

	return s.DB.Error
}

// First returns a group from the database using it's ID
func (s *GroupSQL) First(g *Group, ID uint) error {
	s.DB.Where("ID = ?", ID).Find(g)

	return s.DB.Error
}

// First returns every group from the database
func (s *GroupSQL) Find() ([]Group, error) {
	var groups []Group
	s.DB.Find(&groups)
	if s.DB.Error != nil {
		return groups, nil
	}
	return groups, s.DB.Error
}
