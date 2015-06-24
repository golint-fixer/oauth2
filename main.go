package main

import (
	"runtime"

	"gopkg.in/redis.v3"

	"github.com/Quorumsco/oauth2/controllers"
	"github.com/codegangsta/cli"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/cmd"
	"github.com/iogo-framework/logs"
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
		cli.StringFlag{Name: "listen, l", Value: "localhost:8080", Usage: "listening host:port"},
		cli.StringFlag{Name: "redis, r", Value: "localhost:6379", Usage: "redis host:port"},
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

	client := redis.NewClient(&redis.Options{Addr: ctx.String("redis")})
	app.Components["Redis"] = client

	if _, err := client.Ping().Result(); err != nil {
		return err
	}
	logs.Debug("Connected to Redis at %s", ctx.String("redis"))

	app.Mux = router.New()
	app.Use(router.Logger)
	app.Get("/authorize", controllers.Authorize)
	app.Post("/authorize", controllers.Authorize)
	app.Get("/token", controllers.Token)
	app.Post("/token", controllers.Token)

	app.Serve(ctx.String("listen"))

	return nil
}
