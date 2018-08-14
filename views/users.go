// Views for JSON responses
package views

import (
	usermodels "github.com/quorumsco/oauth2/models"
	teammodels "github.com/quorumsco/teams/models"
)

// Users represents the json response for users
type Users struct {
	Users []usermodels.User `json:"users"`
	Teams []teammodels.Team `json:"teams"`
	Count int               `json:"count"`
}

// User represents the json response for user
type User struct {
	User *usermodels.User `json:"user"`
}
