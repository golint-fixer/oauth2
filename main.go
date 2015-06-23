package main

import (
	"runtime"

	"github.com/Quorumsco/oauth2/controllers"
	"github.com/codegangsta/cli"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/cmd"
	"github.com/iogo-framework/router"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	cmd := cmd.New()
	cmd.Name = "oauth2"
	cmd.Usage = "Quorums OAuth2 server"
	cmd.Version = "0.0.1"
	cmd.Before = serve
	cmd.Flags = append(cmd.Flags, []cli.Flag{
		cli.StringFlag{"cpu, cpuprofile", "", "cpu profiling", ""},
		cli.IntFlag{"port, p", 8080, "server listening port", ""},
		cli.HelpFlag,
	}...)
	cmd.RunAndExitOnError()
}

func serve(ctx *cli.Context) error {
	var app *application.Application
	var err error

	if app, err = application.New(); err != nil {
		return err
	}

	app.Mux = router.New()
	app.Use(router.Logger)
	app.Get("/authorize", controllers.Authorize)
	app.Post("/authorize", controllers.Authorize)
	app.Get("/token", controllers.Token)
	app.Post("/token", controllers.Token)

	app.Serve(ctx.Int("port"))

	return nil
}
