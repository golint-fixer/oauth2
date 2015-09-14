package models

type Group struct {
	ID   uint    `json:"id"`
	Name *string `json:"name,omitempty"`

	Users []User `json:"contacts,omitempty"`
}
