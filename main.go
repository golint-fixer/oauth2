package main

import (
	"runtime"

	"gopkg.in/redis.v3"

	"github.com/RangelReale/osin"
	"github.com/codegangsta/cli"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/cmd"
	"github.com/iogo-framework/logs"
	"github.com/iogo-framework/router"
	"github.com/quorumsco/oauth2/components"
	"github.com/quorumsco/oauth2/controllers"
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
		cli.StringFlag{Name: "listen, l", Value: "0.0.0.0:8080", Usage: "server listening host:port"},
		cli.StringFlag{Name: "redis, r", Value: "localhost:6379", Usage: "redis host:port"},
		cli.StringFlag{Name: "postgres, s", Value: "localhost:5432", Usage: "postgresql host:port"},
		cli.BoolFlag{Name: "debug, d", Usage: "print debug information"},
		cli.HelpFlag,
	}...)
	cmd.RunAndExitOnError()
}

func serve(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		logs.Level(logs.DebugLevel)
	}

	var app = application.New()

	client := redis.NewClient(&redis.Options{Addr: ctx.String("redis")})
	if _, err := client.Ping().Result(); err != nil {
		return err
	}
	logs.Debug("Connected to Redis at %s", ctx.String("redis"))
	app.Components["Redis"] = client

	cfg := osin.NewServerConfig()
	cfg.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE,
		osin.REFRESH_TOKEN, osin.PASSWORD}

	server := osin.NewServer(cfg, components.NewRedisStorage(client))
	app.Components["OAuth"] = server
	app.Components["Mux"] = router.New()

	if ctx.Bool("debug") {
		app.Use(router.Logger)
	}

	app.Use(app.Apply)
	app.Get("/authorize", controllers.Authorize)
	app.Post("/token", controllers.Token)
	app.Get("/info", controllers.Info)

	app.Get("/test", controllers.Test)

	app.Serve(ctx.String("listen"))

	return nil
}
