// Views for JSON responses
package views

import "github.com/quorumsco/oauth2/models"

// Teams represents the json response for teams
type Teams struct {
	Teams []models.Team `json:"teams"`
}

// Team represents the json response for teams
type Team struct {
	Team *models.Team `json:"team"`
}
