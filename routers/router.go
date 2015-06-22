package routers

import (
	"net/http"
	"path/filepath"

	"../components/application"
	"../controllers"
	"github.com/silverwyrda/iogo"
)

func URLs(app *application.Application) {
	app.Use(iogo.Logger)
	//app.Use(app.ApplyTemplates)
	//app.Use(app.ApplyDB)

	app.Get("/authorize", controllers.Authorize)
	app.Post("/authorize", controllers.Authorize)
	app.Get("/token", controllers.Token)
	app.Post("/token", controllers.Token)

	wd, _ := filepath.Abs("public")
	app.Get("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir(wd))).ServeHTTP)
}
