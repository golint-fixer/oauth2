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
		s.DB.Create(u)

		return s.DB.Error
	}

	s.DB.Save(u)

	return s.DB.Error
}

// Delete removes a user from the database
func (s *UserSQL) Delete(u *User) error {
	s.DB.Delete(u)

	return s.DB.Error
}

// First return a user from the database using his ID
func (s *UserSQL) First(u *User) error {
	if u.Mail != nil && u.Password != nil {
		s.DB.Where("mail = ?", u.Mail).Find(u)
	} else {
		s.DB.Find(u)
	}

	return s.DB.Error
}

// Find returns every user with a given groupID from the database
func (s *UserSQL) Find() ([]User, error) {
	var users []User
	s.DB.Find(&users)
	if s.DB.Error != nil {
		return users, nil
	}
	return users, s.DB.Error
}
