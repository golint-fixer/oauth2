package urls

import (
	"net/http"
	"path/filepath"

	"github.com/iogo-framework/applications"
	"github.com/iogo-framework/router"
)

func URLs(app *applications.Application) {
	app.Use(router.Logger)
	//app.Use(app.ApplyTemplates)
	//app.Use(app.ApplyDB)

	app.Get("/authorize", controllers.Authorize)
	app.Post("/authorize", controllers.Authorize)
	app.Get("/token", controllers.Token)
	app.Post("/token", controllers.Token)

	wd, _ := filepath.Abs("public")
	app.Get("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir(wd))).ServeHTTP)
}
