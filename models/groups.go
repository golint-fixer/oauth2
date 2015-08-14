package models

import (
	"github.com/asaskevich/govalidator"
	"github.com/quorumsco/contacts/models"
)

type User struct {
	ID   int64   `json:"id" gorm"primary_key"`
	Name *string `json:"name,omitempty"`

	Contacts []models.Contact `json:"contacts,omitempty"`
}
