package models

type Group struct {
	ID   uint    `json:"id" gorm"primary_key"`
	Name *string `json:"name,omitempty"`

	Users []User `json:"contacts,omitempty"`
}
