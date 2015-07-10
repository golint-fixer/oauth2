package main

import (
	"os"
	"runtime"

	"gopkg.in/redis.v3"

	"github.com/RangelReale/osin"
	"github.com/codegangsta/cli"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/cmd"
	"github.com/iogo-framework/logs"
	"github.com/iogo-framework/router"
	"github.com/iogo-framework/settings"
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
		cli.StringFlag{Name: "config, c", Usage: "configuration file", EnvVar: "CONFIG"},
		cli.BoolFlag{Name: "debug, d", Usage: "print debug information"},
		cli.HelpFlag,
	}...)
	cmd.RunAndExitOnError()
}

func serve(ctx *cli.Context) error {
	var app = application.New()

	config, err := settings.Parse(ctx.String("config"))
	if err != nil && ctx.String("config") != "" {
		logs.Error(err)
	}

	if ctx.Bool("debug") || config.Debug() {
		logs.Level(logs.DebugLevel)
	}

	redisSettings, err := config.Redis()
	client := redis.NewClient(&redis.Options{Addr: redisSettings.String()})
	if _, err := client.Ping().Result(); err != nil {
		return err
	}
	logs.Debug("Connected to Redis at %s", redisSettings.String())
	app.Components["Redis"] = client

	cfg := osin.NewServerConfig()
	cfg.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE,
		osin.REFRESH_TOKEN, osin.PASSWORD}

	oauthServer := osin.NewServer(cfg, components.NewRedisStorage(client))
	app.Components["OAuth"] = oauthServer
	app.Components["Mux"] = router.New()

	if ctx.Bool("debug") || config.Debug() {
		app.Use(router.Logger)
	}

	app.Use(app.Apply)

	app.Get("/authorize", controllers.Authorize)
	app.Post("/token", controllers.Token)
	app.Get("/info", controllers.Info)

	app.Get("/test", controllers.Test)

	server, err := config.Server()
	if err != nil {
		logs.Critical(err)
		os.Exit(1)
	}
	return app.Serve(server.String())
}
