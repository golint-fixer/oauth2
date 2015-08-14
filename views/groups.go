package views

import "github.com/quorumsco/user/models"

type Groups struct {
	Groups []models.Group `json:"groups"`
}

type Group struct {
	Group *models.Group `json:"group"`
}
