package models

import "github.com/quorumsco/contacts/models"

type Group struct {
	ID   int64   `json:"id" gorm"primary_key"`
	Name *string `json:"name,omitempty"`

	Contacts []models.Contact `json:"contacts,omitempty"`
}
