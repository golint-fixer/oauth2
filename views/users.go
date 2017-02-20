// Views for JSON responses
package views

import "github.com/quorumsco/oauth2/models"

// Users represents the json response for users
type Users struct {
	Users []models.User `json:"users"`
	Count int `json:"count"`
}

// User represents the json response for user
type User struct {
	User *models.User `json:"user"`
}
