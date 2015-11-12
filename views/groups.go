// Views for JSON responses
package views

import "github.com/quorumsco/oauth2/models"

// Groups represents the json response for groups
type Groups struct {
	Groups []models.Group `json:"groups"`
}

// Group represents the json response for groups
type Group struct {
	Group *models.Group `json:"group"`
}
