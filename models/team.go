// Definition of the structures and SQL interaction functions
package models

import (
	"time"
	// "github.com/asaskevich/govalidator"
)

// Group represents the components of a group
type Team struct {
	ID            uint       `json:"id";gorm:"primary_key"`
	Name          *string    `json:"name,omitempty"`
	GroupID       uint       `json:"groupid"`
	Created       *time.Time `json:"created,omitempty"`
	User_referent *string    `json:"user_referent,omitempty"`
	Users         []User     `gorm:"many2many:team_users"`
}

type TeamArgs struct {
	GroupID uint
	UserID  int64
	TeamID  uint
	Team    *Team
}

type TeamReply struct {
	Team  *Team
	Teams []Team
}
