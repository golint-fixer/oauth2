// Definition of the structures and SQL interaction functions
package models

import "github.com/jinzhu/gorm"

// GroupDS implements the GroupSQL methods
type UserDS interface {
	Save(*User) error
	Delete(*User) error
	First(*User) error
	Find() ([]User, error)
}

// UserSQL contains a Gorm client and the user and gorm related methods
type UserSQL struct {
	DB *gorm.DB
}

// Userstore returns a UserDS implementing CRUD methods for the users and containing a gorm client
func UserStore(db *gorm.DB) UserDS {
	return &UserSQL{DB: db}
}

// Save inserts a new user into the database
func (s *UserSQL) Save(u *User) error {
	if u.ID == 0 {
		err := s.DB.Create(u).Error

		return err
	}

	err := s.DB.Save(u).Error

	return err
}

// Delete removes a user from the database
func (s *UserSQL) Delete(u *User) error {
	err := s.DB.Delete(u).Error

	return err
}

// First return a user from the database using his ID
func (s *UserSQL) First(u *User) error {
	var err error

	if u.Mail != nil && u.Password != nil {
		err = s.DB.Where("mail = ?", u.Mail).Find(u).Error
	} else {
		err = s.DB.Find(u).Error
	}

	return err
}

// Find returns every user with a given groupID from the database
func (s *UserSQL) Find() ([]User, error) {
	var users []User
	err := s.DB.Find(&users).Error
	if err != nil {
		return users, nil
	}
	return users, err
}
