// Bundle of functions managing the CRUD
package controllers

import (
	"net/http"
	"text/template"

	"github.com/jinzhu/gorm"
	"github.com/quorumsco/application"
	"github.com/quorumsco/router"
)

// Allows you to get the gorm client defined in the main
func getDB(r *http.Request) *gorm.DB {
	return router.Context(r).Env["Application"].(*application.Application).Components["DB"].(*gorm.DB)
}

// Allows you to get the templates
func getTemplates(r *http.Request) map[string]*template.Template {
	return router.Context(r).Env["Application"].(*application.Application).Components["Templates"].(map[string]*template.Template)
}
