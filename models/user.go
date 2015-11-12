// Definition of the structures and SQL interaction functions
package models

import "github.com/asaskevich/govalidator"

// User represent a user in the database
type User struct {
	ID        int64   `json:"id"`
	Mail      *string `sql:"not null;unique"`
	Password  *string `sql:"not null"`
	Firstname *string `sql:"not null" json:"firstname"`
	Surname   *string `sql:"not null" json:"surname"`
	GroupID   uint    `json:"group_id"`
}

// Validate is used to check if the user infos are correct before inserting it into the database
func (u *User) Validate() map[string]string {
	var errs = make(map[string]string)

	switch {
	case u.Mail == nil:
		errs["mail"] = "is required"
	case u.Mail != nil && !govalidator.IsEmail(*u.Mail):
		errs["mail"] = "is not valid"
	case u.Password == nil:
		errs["password"] = "is required"
	}

	return errs
}
