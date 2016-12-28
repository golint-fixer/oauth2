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
	Parti				 *string    `json:"parti,omitempty"`
	Echelle 		 *string    `json:"echelle,omitempty"`
	Zone 				 *string    `json:"zone,omitempty"`
	Users        []User     `json:"contacts,omitempty"`
}
