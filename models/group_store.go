package models

import "github.com/jinzhu/gorm"

type GroupDS interface {
	Save(*Group) error
	Delete(*Group) error
	First(*Group) error
	Find() ([]Group, error)
}

type GroupSQL struct {
	DB *gorm.DB
}

func GroupStore(db *gorm.DB) GroupDS {
	return &GroupSQL{DB: db}
}

// func (s *UserSQL) Save(u *User) error {
// 	if u.ID == 0 {
// 		s.DB.Create(u)

// 		return s.DB.Error
// 	}

// 	s.DB.Save(u)

// 	return s.DB.Error
// }

// func (s *UserSQL) Delete(u *User) error {
// 	s.DB.Delete(u)

// 	return s.DB.Error
// }

// func (s *UserSQL) First(u *User) error {
// 	if u.Mail != nil && u.Password != nil {
// 		s.DB.Where("mail = ?", u.Mail).Find(u)
// 	} else {
// 		s.DB.Find(u)
// 	}

// 	return s.DB.Error
// }

// func (s *UserSQL) Find() ([]User, error) {
// 	var users []User
// 	s.DB.Find(&users)
// 	if s.DB.Error != nil {
// 		return users, nil
// 	}
// 	return users, s.DB.Error
// }
