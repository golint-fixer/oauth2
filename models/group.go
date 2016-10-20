// Definition of the structures and SQL interaction functions
package models

import (
	"time"

	// "github.com/asaskevich/govalidator"
)

// Group represents the componenents of a group
type Group struct {
	ID           uint       `json:"id"`
	Name         *string    `json:"name,omitempty"`
	Endofcampain *time.Time `json:"endofcampain"`
	Users        []User     `json:"contacts,omitempty"`
}
