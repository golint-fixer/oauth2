// Definition of the structures and SQL interaction functions
package models

// Group represents the componenents of a group
type Group struct {
	ID   uint    `json:"id"`
	Name *string `json:"name,omitempty"`

	Users []User `json:"contacts,omitempty"`
}
