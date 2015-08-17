package models

import "github.com/jinzhu/gorm"

type GroupDS interface {
	Save(*Group) error
	Delete(*Group) error
	First(*Group, uint) error
	Find() ([]Group, error)
}

type GroupSQL struct {
	DB *gorm.DB
}

func GroupStore(db *gorm.DB) GroupDS {
	return &GroupSQL{DB: db}
}

func (s *GroupSQL) Save(g *Group) error {
	if g.ID == 0 {
		s.DB.Create(g)

		return s.DB.Error
	}

	s.DB.Save(g)

	return s.DB.Error
}

func (s *GroupSQL) Delete(g *Group) error {
	s.DB.Delete(g)

	return s.DB.Error
}

func (s *GroupSQL) First(g *Group, ID uint) error {
	s.DB.Where("ID = ?", ID).Find(g)

	return s.DB.Error
}

func (s *GroupSQL) Find() ([]Group, error) {
	var groups []Group
	s.DB.Find(&groups)
	if s.DB.Error != nil {
		return groups, nil
	}
	return groups, s.DB.Error
}
